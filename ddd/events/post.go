package events

import (
	"encoding/json"
	"github.com/CodFrm/learnMicroService/core"
	"github.com/CodFrm/learnMicroService/ddd/domain"
)

type PostEvent struct {
	GroupId   string
	EventName []string
}

func (self *PostEvent) GetGroupId() string {
	return self.GroupId
}

func (self *PostEvent) GetEventNames() []string {
	return self.EventName
}

func (self *PostEvent) Handler(name string, value []byte) error {
	model := &domain.PostAggregate{}
	if err := json.Unmarshal(value, model); err != nil {
		return err
	}
	db, _ := core.GetDb("auth_db")
	_, err := db.Exec("insert into credit(uid,`change`,createtime) values(?,?,?)", model.User.Uid, 2, model.CreateTime)
	if err != nil {
		println(err)
	}
	_, err = db.Exec("update user set credit=credit+2 where uid=?", model.User.Uid)
	if err != nil {
		println(err)
	}
	//积分记录+2 记录日志
	return nil
}
