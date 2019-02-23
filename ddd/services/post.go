package services

import (
	"github.com/CodFrm/learnMicroService/ddd/repository"
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

	self.Repository = &repository.PostRepository{}
	self.Repository.Save()
	return nil
}
