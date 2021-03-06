package models

import (
	"github.com/astaxie/beego/orm"
	"time"
)

//站点配置
type SiteConfig struct {
	Id                int    `orm:"auto"`
	BlogName          string `orm:"size(256)"`
	NumPerPage        int
	CommentNumPerPage int
	BlogUrl           string `orm:"size(256)"`
	AdminEmail        string `orm:"size(256)"`
	CopyRight         string `orm:"size(256)"`
	AdminUser         string `orm:"size(256)"`
	AdminPassword     string `orm:"size(256)"`
}

//文章
type Post struct {
	Id          int `orm:"auto"`
	CategoryId  int
	User        string
	Shortname   string
	Title       string    `orm:"size(256)"`
	Summary     string    `orm:"type(text);null"`
	Body        string    `orm:"type(text)"`
	Password    string    `orm:"size(256);null"`
	CreatedTime time.Time `orm:"auto_now_add;type(datetime)"`
	UpdateTime  time.Time `orm:"auto_now;type(datetime)"`
	Ip          string    `orm:"size(256);null"`
	Archive     string    `orm:"size(256)"`
}

//评论
type Comment struct {
	Id          int `orm:"auto"`
	PostId      int
	Body        string `orm:"type(text)"`
	Email       string `orm:"type(text)"`
	User        string
	CreatedTime time.Time `orm:"auto_now_add;type(datetime)"`
	Ip          string
}

//分类
type Category struct {
	Id   int    `orm:"auto"`
	Name string `orm:"size(256)"`
}

//上传的文件
type UploadFile struct {
	Id         int       `orm:"auto"`
	UploadTime time.Time `orm:"auto_now_add;type(datetime)"`
	Filesize   int64
	Filename   string `orm:"size(512)"`
	Hashname   string `orm:"size(512)"`
}

//关于
type About struct {
	Id      int    `orm:"auto"`
	Content string `orm:"type(text)"`
}

//项目
type Projects struct {
	Id      int    `orm:"auto"`
	Content string `orm:"type(text)"`
}

func init() {
	orm.RegisterModel(new(Post))
	orm.RegisterModel(new(SiteConfig))
	orm.RegisterModel(new(Comment))
	orm.RegisterModel(new(Category))
	orm.RegisterModel(new(UploadFile))
	orm.RegisterModel(new(About))
	orm.RegisterModel(new(Projects))
}
