package admincontrollers

import (
	//"fmt"
	//"strconv"
	"rblog/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	//"github.com/astaxie/beego/cache"
	//"github.com/astaxie/beego/context"
	//"github.com/russross/blackfriday"
	//"crypto/md5"
	//"encoding/hex"
	//"encoding/base64"
	//"errors"
	//"net/http/pprof"
	//"net/http"
	//"strings"
	"html/template"
)


type Article struct {
	Title string `form:"Title"`
	User string `form:"User"`
	Category int `form:"Category"`
	Shortname string `form:"Shortname"`
	Body string `form:"Body"`
}


// Admin home
type AdminController struct {
	beego.Controller
}

func (this *AdminController) Get() {
	this.TplNames = "admin/home.html"
	this.Render()
}


// Login
type AdminLoginController struct {
	beego.Controller
}

func (this *AdminLoginController) Get() {
	this.TplNames = "admin/login.html"
	this.Render()
}


// Logout
type AdminLogoutController struct {
	beego.Controller
}

func (this *AdminLogoutController) Get() {
	this.TplNames = "admin/logout.html"
	this.Render()
}


// Article
type AdminArticleController struct {
	beego.Controller
}

func (this *AdminArticleController) Get() {
	this.Data["xsrfdata"] = template.HTML(this.XsrfFormHtml())
	this.TplNames = "admin/article.html"
	this.Render()
}

func (this *AdminArticleController) Post() {
	article := Article{}
	if err := this.ParseForm(&article); err != nil {
		beego.Error(err)
	}
	o := orm.NewOrm()
	var post models.Post
	post.CategoryId = article.Category
	post.User = article.User
	post.Shortname = article.Shortname
	post.Title = article.Title
	post.Body = article.Body
	
	post.Ip = this.Ctx.Input.IP()
	o.Insert(&post)
	
	this.Ctx.Redirect(301, "/")
}


// Category
type AdminCategoryController struct {
	beego.Controller
}

func (this *AdminCategoryController) Get() {
	this.TplNames = "admin/category.html"
	this.Render()
}


// Comment
type AdminCommentController struct {
	beego.Controller
}

func (this *AdminCommentController) Get() {
	this.TplNames = "admin/comment.html"
	this.Render()
}


// SiteConfig
type AdminSiteController struct {
	beego.Controller
}

func (this *AdminSiteController) Get() {
	this.TplNames = "admin/site.html"
	this.Render()
}
