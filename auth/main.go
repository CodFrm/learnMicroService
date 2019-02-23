package auth

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	common "github.com/CodFrm/learnMicroService/common"
	"google.golang.org/grpc"

	micro "github.com/CodFrm/learnMicroService/proto"
)

const (
	port = ":5000"
)

type server struct{}

var userData = make([]map[string]string, 0, 1)
var tokenList = make(map[string]int32)
var authList = make(map[string]string)
var db = common.Db{}

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
		user varchar(255) NOT NULL,
		pwd varchar(255) NOT NULL,
		token varchar(255),
		PRIMARY KEY (uid)
	  ) ENGINE=InnoDB DEFAULT CHARSET=latin1;
	`
	db.Exec(sql)
}

func Start() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Printf("net error:%v\n", err)
	}
	//注册rpc
	rpcService := common.Service{
		Name:    "auth_micro",
		Tags:    []string{"rpc"},
		Address: common.LocalIP(),
		Port:    5000,
	}
	defer rpcService.Deregister()
	err = rpcService.Register()
	if err != nil {
		log.Printf("service Register error:%v\n", err)
	}

	//注册对外的restful服务
	httpService := common.Service{
		Name:    "auth_micro",
		Tags:    []string{"rest"},
		Address: common.LocalIP(),
		Port:    8004,
	}
	defer httpService.Deregister()
	err = httpService.Register()
	if err != nil {
		log.Printf("service Register error:%v\n", err)
	}

	//连接数据库
	err = db.Connect("auth_db", 3306, "auth", "micro_db_pwd", "auth")
	if err != nil {
		log.Printf("database connect error:%v\n", err)
	}
	init_database()

	s := grpc.NewServer()
	micro.RegisterAuthServer(s, &server{})
	go s.Serve(lis)
	http.HandleFunc("/login", login)
	http.HandleFunc("/register", register)
	http.ListenAndServe(":8004", nil)
}
