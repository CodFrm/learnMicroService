package auth

import (
	"context"
	"database/sql"
	"github.com/CodFrm/learnMicroService/core"
	"github.com/CodFrm/learnMicroService/ddd/events"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/CodFrm/learnMicroService/common"
	"google.golang.org/grpc"

	micro "github.com/CodFrm/learnMicroService/proto"
)

const (
	port = ":5000"
)

type server struct{}

var db *common.Db

func (s *server) Isvalid(ctx context.Context, in *micro.TokenMsg) (*micro.UserMsg, error) {
	//验证token是否有权限访问
	log.Printf("token:%v,api:%v\n", in.Token, in.Api)
	var uid, name string
	err := db.QueryRow("select uid,user from user where token=?", in.Token).Scan(&uid, &name)
	if err != sql.ErrNoRows {
		//Token存在
		id, _ := strconv.Atoi(uid)
		return &micro.UserMsg{
			Uid:    int32(id),
			Access: true, //数据库版本后取消权限验证了
			Name:   name,
			Group:  "user",
		}, nil
	}
	//验证失败
	return &micro.UserMsg{Access: false}, nil
}

func (s *server) GetUser(ctx context.Context, in *micro.UserMsgRequest) (*micro.UserMsgResponse, error) {
	user := &micro.UserMsgResponse{}
	var uid, name string
	err := db.QueryRow("select uid,user from user where uid=?", in.Uid).Scan(&uid, &name)
	if err != sql.ErrNoRows {
		id, _ := strconv.Atoi(uid)
		return &micro.UserMsgResponse{
			Uid:  int32(id),
			Name: name,
		}, nil
	}
	return user, err
}

func login(w http.ResponseWriter, req *http.Request) {
	var ret string
	if strings.ToLower(req.Method) != "post" {
		ret = "error method"
	} else {
		user, pwd := req.PostFormValue("user"), req.PostFormValue("pwd")
		var val string
		err := db.QueryRow("select * from user where user=? and pwd=?", user, pwd).Scan(&val)
		if err == sql.ErrNoRows {
			ret = "null account"
		} else {
			token := common.RandStringRunes(16)
			ret = "login success token:" + token
			db.Exec("update user set token=? where user=? and pwd=?", token, user, pwd)
		}
	}
	w.Write([]byte(ret))
}

func register(w http.ResponseWriter, req *http.Request) {
	var ret string
	if strings.ToLower(req.Method) != "post" {
		ret = "error method"
	} else {
		user, pwd := req.PostFormValue("user"), req.PostFormValue("pwd")
		var val string
		err := db.QueryRow("select * from user where user=?", user).Scan(&val)
		if err != sql.ErrNoRows {
			ret = "user exist"
		} else {
			ret = "register success"
			db.Exec("insert into user(user,pwd) values(?,?)", user, pwd)
		}
	}
	w.Write([]byte(ret))
}

func init_database() {
	sql := `
	CREATE TABLE IF NOT EXISTS user (
		uid int(11) NOT NULL AUTO_INCREMENT,
		user varchar(255) CHARACTER SET utf8 NOT NULL,
		pwd varchar(255) CHARACTER SET utf8 NOT NULL,
        credit int(11) DEFAULT '0',
		token varchar(255) CHARACTER SET utf8 DEFAULT NULL,
		PRIMARY KEY (uid)
	  ) ENGINE=InnoDB DEFAULT CHARSET=latin1;
	`
	db.Exec(sql)
	sql = `
	CREATE TABLE IF NOT EXISTS credit (
		id int(11) NOT NULL AUTO_INCREMENT,
		uid int(11) NOT NULL COMMENT '关联用户id',`
	sql += "`change` int(11) NOT NULL COMMENT '积分增长',"
	sql += `createtime int(11) NOT NULL COMMENT '产生时间',
  		PRIMARY KEY (id)
	  ) ENGINE=InnoDB DEFAULT CHARSET=latin1;
	`
	db.Exec(sql)
}

func Start() {
	//grpc 服务端
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Printf("net error:%v", err)
	}
	s := grpc.NewServer()
	micro.RegisterAuthServer(s, &server{})
	go s.Serve(lis)
	//http 接口
	apis := make([]core.HttpApi, 2)
	apis[0] = core.HttpApi{Pattern: "/login", Handler: login}
	apis[1] = core.HttpApi{Pattern: "/register", Handler: register}
	//注册服务
	services := make(map[string]common.Service)
	services["auth_http"] = common.Service{
		Name:    "auth_micro",
		Tags:    []string{"rest"},
		Address: common.LocalIP(),
		Port:    8004,
	}
	services["auth_rpc"] = common.Service{
		Name:    "auth_micro",
		Tags:    []string{"rpc"},
		Address: common.LocalIP(),
		Port:    5000,
	}
	//数据库配置
	dbs := make(map[string]core.DbConfig)
	dbs["auth_db"] = core.DbConfig{
		"127.0.0.1", 3307, "auth", "micro_db_pwd", "auth",
	}
	//微服务配置
	err = core.StartService(core.AppConfig{
		Http:    core.HttpConfig{Port: 8023, Api: apis},
		Service: services,
		Db:      dbs,
		Mq:      core.MqConfig{[]string{"127.0.0.1:9092"}},
	}, func() {
		var err error
		db, err = core.GetDb("auth_db")
		if err != nil {
			println("db connect error")
		}
		init_database()
		//event bus这里监听事件
		err = events.RegisterEvent(&events.PostEvent{
			"auth",
			[]string{"post_msg"},
		})
		if err != nil {
			println(err.Error())
		}
	})
	if err != nil {
		println(err)
	}
}
