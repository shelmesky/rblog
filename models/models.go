package models

import (
	"github.com/astaxie/beego/orm"
)

type SiteConfig struct {
	Id int `orm:"auto"`
	BlogName string `orm:"size(256)"`
	NumPerPage int
	BlogUrl string `orm:"size(256)"`
	AdminEmail string `orm:"size(256)"`
	CopyRight string `orm:"size(256)"`
}

//字段名只允许首字母大写？
//文章
type Post struct {
	Id int `orm:"auto"`
	CategoryId int
	Shortname string
	Title string `orm:"size(256)"`
	Body string
	Time string
	Ip string
}

//目录
type Category struct {
	Id int `orm:"auto"`
	Name string `orm:"size(256)"`
}

func init() {
	orm.RegisterModel(new(Post))
	orm.RegisterModel(new(SiteConfig))
	orm.RegisterModel(new(Category))
}
