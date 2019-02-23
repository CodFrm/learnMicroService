package commands

import (
	"errors"

	"github.com/CodFrm/learnMicroService/ddd/services"
)

/**
 * post命令 对数据进行校验 分配到业务处理
 */

type PostCommand struct {
	Uid   int
	Title string
}

func (p *PostCommand) ResolveHandler() error {
	if p.Uid <= 0 {
		return errors.New("错误的用户id")
	}
	if p.Title == "" {
		return errors.New("标题不能为空")
	}
	if len(p.Title) > 64 {
		return errors.New("标题过长")
	}
	postService := services.PostService{}
	return postService.Post(p.Uid, p.Title)
}
