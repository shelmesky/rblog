package models

import (
	"github.com/astaxie/beego/orm"
	"time"
)

//站点配置
type SiteConfig struct {
	Id int `orm:"auto"`
	BlogName string `orm:"size(256)"`
	NumPerPage int
	BlogUrl string `orm:"size(256)"`
	AdminEmail string `orm:"size(256)"`
	CopyRight string `orm:"size(256)"`
	AdminUser string `orm:"size(256)"`
	AdminPassword string `orm:"size(256)"`
}

//文章
type Post struct {
	Id int `orm:"auto"`
	CategoryId int
	User string
	Shortname string
	Title string `orm:"size(256)"`
	Body string `orm:"type(text)"`
	CreatedTime time.Time `orm:"auto_now_add;type(datetime)"`
	UpdateTime time.Time `orm:"auto_now;type(datetime)"`
	Ip string
	Archive time.Time `orm:"auto_now_add;type(date)"`
}

//评论
type Comment struct {
	Id int `orm:"auto"`
	PostId int
	Body string `orm:"type(text)"`
	User string
	CreatedTime time.Time `orm:"auto_now_add;type(datetime)"`
	Ip string
}

//分类
type Category struct {
	Id int `orm:"auto"`
	Name string `orm:"size(256)"`
}

func init() {
	orm.RegisterModel(new(Post))
	orm.RegisterModel(new(SiteConfig))
	orm.RegisterModel(new(Comment))
	orm.RegisterModel(new(Category))
}
