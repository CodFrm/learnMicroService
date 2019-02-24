package services

import (
	"context"
	"encoding/json"
	"github.com/CodFrm/learnMicroService/core"
	"github.com/CodFrm/learnMicroService/ddd/domain"
	"github.com/CodFrm/learnMicroService/ddd/repository"
	"github.com/CodFrm/learnMicroService/proto"
	"time"
)

/**
 * post 领域服务 处理业务逻辑
 * 例如判断帖子是否发过,用户是否有发帖权限
 */

type PostService struct {
	Repository *repository.PostRepository
}

func (self *PostService) Post(uid int, title string) error {
	//这里处理一些业务逻辑,比如判断用户是否有权限发帖
	//不过我们的接口只判断了登录,而且之前入口已经实现了
	//那么这里我就只获取用户信息了(虽然这两个感觉起来是一样的)
	client, _ := core.GetRPCService("auth_rpc")
	auth := lms.NewAuthClient(client)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	userMsg, err := auth.GetUser(ctx, &lms.UserMsgRequest{
		Uid: int32(uid),
	}) //rpc调用isvalid方法
	if err != nil {
		return err
	}
	model := domain.PostAggregate{
		User:       domain.UserEntity{int(userMsg.Uid), userMsg.Name},
		Title:      title,
		CreateTime: int(time.Now().Unix()),
	}

	//将模型存入仓库
	self.Repository = &repository.PostRepository{}
	err = self.Repository.Save(model)
	if err != nil {
		return err
	}
	//成功 发送事件
	data, _ := json.Marshal(model)
	err = core.SendMessage("post_msg", string(data[:]))
	if err != nil {
		//必要的话可以回滚本地事务
		return err
	}
	return nil
}
