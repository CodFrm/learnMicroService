package repository

import (
	"github.com/CodFrm/learnMicroService/core"
	"github.com/CodFrm/learnMicroService/ddd/domain"
)

type PostRepository struct {
}

func (self *PostRepository) Save(post domain.PostAggregate) error {
	db, _ := core.GetDb("post_db")
	//感觉换成mongodb存储会更方便
	_, err := db.Exec("insert into posts(`uid`,`name`,`title`,`createtime`) values(?,?,?,?)",
		post.User.Uid, post.User.Name, post.Title, post.CreateTime,
	)
	return err
}
