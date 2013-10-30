package models

import (
	"github.com/astaxie/beego/orm"
)

//字段名只允许首字母大写？
type Post struct {
	Id int
	Shortname string
	Title string
	Body string
	Time string
	Ip string
}

func init() {
	orm.RegisterModel(new(Post))
}
