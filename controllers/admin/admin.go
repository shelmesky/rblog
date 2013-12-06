package admincontrollers

import (
	//"fmt"
	//"strconv"
	"rblog/models"
	"rblog/controllers/primary"
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
	"strings"
	"html/template"
	"regexp"
	//"reflect"
	"rblog/common/utils"
)

var (
	Category_map *beego.BeeMap
)


func GetCategoryName(content interface{}) (string) {
	//fmt.Println(reflect.TypeOf(content))
	var category_name string
	if value, ok := content.(int); ok {
		if Category_map.Check(value) {
			category := Category_map.Get(value)
			category_name, _ := category.(string)
			return category_name
		} else {
			o := orm.NewOrm()
			var category models.Category
			err := o.QueryTable(new(models.Category)).Filter("Id", value).One(&category)
			if err != nil {
				beego.Error(err)
				return string(value)
			}
			return category.Name
		}
	}
	
	return category_name
}


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
	var posts []*models.Post
	o := orm.NewOrm()
	o.QueryTable(new(models.Post)).All(&posts)
	this.Data["Posts"] = posts
	
	this.Data["BlogUrl"] = controllers.Site_config.BlogUrl
	
	var categories []*models.Category
	o.QueryTable(new(models.Category)).All(&categories)
	this.Data["Categories"] = categories
	
	this.Data["xsrfdata"] = template.HTML(this.XsrfFormHtml())
	
	// 防止重复提交，设置session
	session_key := "admin_article_get"
	this.SetSession(session_key, utils.MakeRandomID())
	
	this.TplNames = "admin/article.html"
	this.Render()
}

func (this *AdminArticleController) Post() {
	var MessageError string
	
	// 检查是否为重复提交
	session_key := "admin_article_get"
	session := this.GetSession(session_key)
	if session == nil {
		this.Ctx.Redirect(301, "/admin/article")
		return
	}
	
	article := Article{}
	if err := this.ParseForm(&article); err != nil {
		beego.Error(err)
	}
	o := orm.NewOrm()
	var post models.Post
	post.CategoryId = article.Category
	post.User = strings.Trim(article.User, " ")
	post.Shortname = strings.Trim(article.Shortname, " ")
	post.Title = strings.Trim(article.Title, " ")
	post.Body = article.Body
	
	only_words_match, _ := regexp.Match(`[^\d]+$`, []byte(post.Shortname))
	if !only_words_match {
		MessageError = "短名称不能为纯数字!"
	} else {
		post.Ip = this.Ctx.Input.IP()
		o.Insert(&post)
		
		this.Data["MessageOK"] = "Post new article success."
		
		// 验证成功则删除session
		// 解决由于失败也删除session
		// 导致验证失败后，再次提交时直接刷新页面，无任何响应的BUG
		this.DelSession(session_key)
	}
	
	// send articles to template
	var posts []*models.Post
	o.QueryTable(new(models.Post)).All(&posts)
	this.Data["Posts"] = posts
	
	this.Data["BlogUrl"] = controllers.Site_config.BlogUrl
	
	var categories []*models.Category
	o.QueryTable(new(models.Category)).All(&categories)
	this.Data["Categories"] = categories
	
	this.Data["xsrfdata"] = template.HTML(this.XsrfFormHtml())
	
	this.Data["MessageError"] = MessageError
	this.TplNames = "admin/article.html"
	this.Render()
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
