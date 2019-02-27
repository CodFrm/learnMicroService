package post

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/CodFrm/learnMicroService/core"
	"github.com/CodFrm/learnMicroService/ddd/commands"

	micro "github.com/CodFrm/learnMicroService/proto"

	"github.com/CodFrm/learnMicroService/common"
)

var authService micro.AuthClient
var db *common.Db

func post(w http.ResponseWriter, req *http.Request) {
	ret := ""
	switch strings.ToLower(req.Method) {
	case "post":
		{
			//远程调用权限验证微服务,判断是否拥有权限
			//这里应该写在其他地方,这里相当于是Controller层
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			userMsg, err := authService.Isvalid(ctx, &micro.TokenMsg{
				Token: req.PostFormValue("token"),
				Api:   "post",
			}) //rpc调用isvalid方法
			if err != nil {
				ret = "rpc调用错误"
			} else if !userMsg.Access {
				ret = "没有权限"
			} else {
				err = commands.CommandBus(&commands.PostCommand{Uid: int(userMsg.Uid), Title: req.PostFormValue("title")})
				if err != nil {
					ret = err.Error()
				} else {
					ret = userMsg.Name + " post 请求成功"
				}
				//ret = userMsg.Name + " post 请求成功"
				//db.Exec("insert into posts(uid,title) values(?,?)", userMsg.Uid, req.PostFormValue("title"))
			}
			break
		}
	case "get":
		{
			//查询直接从数据库查询
			rows, err := db.Query("select id,name,title from posts")
			if err != nil {
				ret += "帖子列表错误 error:" + err.Error()
			} else {
				for rows.Next() {
					var id, name, title string
					err := rows.Scan(&id, &name, &title)
					if err != nil {
						ret += err.Error()
						break
					}
					ret += fmt.Sprintf("帖子id:%v 发帖用户:%v 帖子标题:%v\n", id, name, title)
				}
			}
			break
		}
	}
	w.Write([]byte(ret))
}

func init_database() {
	sql := `
	CREATE TABLE IF NOT EXISTS posts (
		id int(11) NOT NULL AUTO_INCREMENT,
		uid int(11) NOT NULL,
		name varchar(64) CHARACTER SET utf8 DEFAULT NULL,
		title varchar(255) CHARACTER SET utf8 NOT NULL,
		createtime int(11) DEFAULT NULL,
		PRIMARY KEY (id)
	  ) ENGINE=InnoDB DEFAULT CHARSET=latin1;
	`
	db.Exec(sql)
}

func Start() {
	//http 接口列表
	apis := make([]core.HttpApi, 1)
	apis[0] = core.HttpApi{Pattern: "/post", Handler: post}
	//注册服务
	services := make(map[string]common.Service)
	services["post_http"] = common.Service{
		Name:    "post_micro",
		Tags:    []string{"rest"},
		Address: common.LocalIP(),
		Port:    8004,
	}
	services["auth_rpc"] = common.Service{
		Name: "auth_micro",
		Tags: []string{"rpc"},
		//Address: common.LocalIP(),
		Port: 5000,
	}
	//数据库配置
	dbs := make(map[string]core.DbConfig)
	dbs["post_db"] = core.DbConfig{
		"post_db", 3306, "post", "micro_db_pwd", "post",
	}
	//微服务配置
	err := core.StartService(core.AppConfig{
		Http:    core.HttpConfig{Port: 8004, Api: apis},
		Service: services,
		Db:      dbs,
		Mq:      core.MqConfig{[]string{"kafka_mq:9092"}},
	}, func() {
		var err error
		db, err = core.GetDb("post_db")
		if err != nil {
			println("db connect error")
		}
		rpc, err := core.GetRPCService("auth_rpc")
		if err != nil {
			println("rpc connect error")
		}
		authService = micro.NewAuthClient(rpc)
		init_database()
	})
	if err != nil {
		println(err.Error())
	}
}
