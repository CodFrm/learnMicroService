package main

import (
	"context"
	"fmt"
	"log"
	"net"
)
import "google.golang.org/grpc"
import micro "../proto"

const (
	port = ":5000"
)

type server struct{}

var userData = make([]map[string]string, 0, 1)
var tokenList = make(map[string]int32)
var authList = make(map[string]string)

func genUser() {
	//生成用户,绑定token
	user := make(map[string]string, 1)
	user["name"] = "admin1"
	user["pwd"] = "qwe123"
	user["group"] = "admin"
	userData = append(userData, user)
	user = make(map[string]string, 1)
	user["name"] = "user1"
	user["pwd"] = "userpwd"
	user["group"] = "user"
	userData = append(userData, user)
	//value表示uid吧,从0开始,hhh不管了
	tokenList["token1"] = 0
	tokenList["token2"] = 1
	//权限列表
	authList["post"] = "admin" //意思发帖需要admin权限
}

func (s *server) Isvalid(ctx context.Context, in *micro.TokenMsg) (*micro.UserMsg, error) {
	//验证token是否有权限访问
	fmt.Printf("token:%v,api:%v\n", in.Token, in.Api)
	if index, ok := tokenList[in.Token]; ok {
		//登录了
		return &micro.UserMsg{
			Uid:    index,
			Access: (authList[in.Api] == userData[index]["group"]), //但是不一定有权限
			Name:   userData[index]["name"],
			Group:  userData[index]["group"],
		}, nil
	}
	//验证失败
	return &micro.UserMsg{Access: false}, nil
}

func main() {
	genUser()
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Printf("net error:%v", err)
	}
	s := grpc.NewServer()
	micro.RegisterAuthServer(s, &server{})
	s.Serve(lis)
}
