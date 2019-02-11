package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc"

	micro "github.com/CodFrm/learnMicroService/proto"

	common "github.com/CodFrm/learnMicroService/common"
)

var rpcConn *grpc.ClientConn
var authService micro.AuthClient
var db = common.Db{}

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

func main() {
	//初始化rpc客户端
	var err error
	// rpc客户端的配置,这里是要auth_micro的
	rpcService := common.Service{
		Name: "auth_micro",
		Tags: []string{"rpc"},
		Port: 5000,
	}
	rpcConn, err = rpcService.GetRPCService() //直接返回rpc
	if err != nil {
		log.Printf("rpc Service error:%v\n", err)
	}
	authService = micro.NewAuthClient(rpcConn)

	//注册对外的restful服务
	httpService := common.Service{
		Name:    "post_micro",
		Tags:    []string{"rest"},
		Address: common.LocalIP(),
		Port:    8004,
	}
	defer httpService.Deregister()
	err = httpService.Register()
	if err != nil {
		log.Printf("service Register error:%v\n", err)
	}

	err = db.Connect("post_db", 3306, "post", "micro_db_pwd", "post")
	if err != nil {
		log.Printf("database connect error:%v\n", err)
	}
	init_database()

	http.HandleFunc("/post", post)
	http.ListenAndServe(":8004", nil)
}
