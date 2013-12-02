package models

import (
	"github.com/astaxie/beego/orm"
)

type SiteConfig struct {
	Id int
	BlogName string
	BlogUrl string
	AdminEmail string
	CopyRight string
}

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
	orm.RegisterModel(new(SiteConfig))
}
