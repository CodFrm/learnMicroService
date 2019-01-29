package main

import (
	"context"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc"

	micro "../proto"
)

var posts = make([]string, 0, 1)
var rpcConn *grpc.ClientConn
var authService micro.AuthClient

func genPosts() {
	//生成帖子
	posts = append(posts, "我是帖子1")
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
			})//rpc调用isvalid方法
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
	rpcConn, err = grpc.Dial("localhost:5000", grpc.WithInsecure())
	if err != nil {
		println("rpc Service error:%v", err)
	}
	authService = micro.NewAuthClient(rpcConn)

	http.HandleFunc("/post", post)
	http.ListenAndServe(":8004", nil)
}
