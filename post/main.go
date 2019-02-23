package post

import (
	"context"
	"fmt"
	"github.com/CodFrm/learnMicroService/core"
	"net/http"
	"strings"
	"time"

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
				ret = userMsg.Name + " post 请求成功"
				db.Exec("insert into posts(uid,title) values(?,?)", userMsg.Uid, req.PostFormValue("title"))
			}
			break
		}
	case "get":
		{
			rows, err := db.Query("select a.id,b.user,a.title from posts as a join user as b on a.uid=b.uid")
			if err != nil {
				ret = "帖子列表错误 error:" + err.Error()
			} else {
				for rows.Next() {
					var id, name, title string
					rows.Scan(&id, &name, &title)
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
		title varchar(255) NOT NULL,
		PRIMARY KEY (id)
	  ) ENGINE=InnoDB DEFAULT CHARSET=latin1;
	`
	db.Exec(sql)
}

func Start() {
	apis := make([]core.HttpApi, 1)
	services := make(map[string]common.Service)
	services["post_http"] = common.Service{
		Name: "post_micro",
		Tags: []string{"rest"},
		//Address: common.LocalIP(),
		Port: 8004,
	}
	services["auth_rpc"] = common.Service{
		Name: "auth_micro",
		Tags: []string{"rpc"},
		Port: 5000,
	}
	dbs := make(map[string]core.DbConfig)
	dbs["post_db"] = core.DbConfig{
		"127.0.0.1", 3308, "post", "micro_db_pwd", "post",
	}
	apis[0] = core.HttpApi{Pattern: "/post", Handler: post}
	err := core.StartService(core.AppConfig{
		Http:    core.HttpConfig{Port: 8001, Api: apis},
		Service: services,
		Db:      dbs,
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
		println(err)
	}
}
