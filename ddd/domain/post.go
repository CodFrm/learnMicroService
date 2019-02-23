package domain

/**
 * 领域模型 聚合根 实体 值对象
 */

//如果允许用户名修改的话,我们将用户名设置为一个实体
//聚合根
type PostAggregate struct {
	Id         int
	User       UserEntity
	Title      string //值对象
	CreateTime int    //值对象
}

//实体 在这个发帖的微服务下是实体,在用户微服务下是聚合根
type UserEntity struct {
	Uid  int
	Name string
}

//聚合根 在回帖的场景下,回帖将作为主体,他们都是独立的聚合根
type ReplyAggregate struct {
	Id         int
	User       UserEntity
	PostId     int
	Content    string
	CreateTime int
}

//行为 获取用户权限
func (self *UserEntity) GetAuth() {

}
