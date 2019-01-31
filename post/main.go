package main

import (
	"context"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc"

	micro "../proto"

	consul "github.com/CodFrm/learnMicroService/common"
)

var posts = make([]string, 0, 1)
var rpcConn *grpc.ClientConn
var authService micro.AuthClient

func genPosts() {
	//生成帖子
	posts = append(posts, "我是帖子124")
}

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
			println(userMsg.Name)
			if err != nil {
				ret = "rpc调用错误"
			} else if !userMsg.Access {
				ret = "没有权限"
			} else {
				ret = userMsg.Name + " post 请求成功"
				posts = append(posts, req.PostFormValue("title"))
			}
			break
		}
	case "get":
		{
			for i := range posts {
				ret += posts[i] + "\n"
			}
			break
		}
	}
	w.Write([]byte(ret))
}

func main() {
	genPosts()
	//初始化rpc客户端
	var err error
	//rpc客户端的配置,这里是要auth_micro的
	rpcService := consul.Service{
		Name: "auth_micro",
		Tags: []string{"rpc"},
	}
	rpcConn, err = rpcService.GetRPCService() //直接返回rpc
	if err != nil {
		println("rpc Service error:%v", err)
	}
	authService = micro.NewAuthClient(rpcConn)

	//注册对外的restful服务
	httpService := consul.Service{
		Name:    "post_micro",
		Tags:    []string{"rest"},
		Address: consul.LocalIP(),
		Port:    8004,
	}
	defer httpService.Deregister()
	err = httpService.Register()
	if err != nil {
		println("service Register error:%v", err)
	}
	http.HandleFunc("/post", post)
	http.ListenAndServe(":8004", nil)
}
