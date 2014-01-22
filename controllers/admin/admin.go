package admincontrollers

import (
	//"fmt"
	//"strconv"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"rblog/models"
	//"github.com/astaxie/beego/cache"
	//"github.com/astaxie/beego/context"
	//"github.com/russross/blackfriday"
	//"crypto/md5"
	//"encoding/hex"
	//"encoding/base64"
	//"errors"
	//"net/http/pprof"
	//"net/http"
	"html/template"
	"regexp"
	"strings"
	//"reflect"
	"crypto/md5"
	"encoding/hex"
	"rblog/common/utils"
)

type Article struct {
	Id        int    `form:"Id"`
	Title     string `form:"Title"`
	Password  string `form:"Password"`
	User      string `form:"User"`
	Category  int    `form:"Category"`
	Shortname string `form:"Shortname"`
	Summary   string `form:"Summary"`
	Body      string `form:"Body"`
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
	action := this.GetString("action")
	id, _ := this.GetInt("id")

	var article models.Post
	if action != "" && id >= 0 {
		if action == "update" {
			o := orm.NewOrm()
			err := o.QueryTable(new(models.Post)).Filter("Id", id).One(&article)
			if err != nil {
				utils.Error(err)
			}
		}
	}
	var posts []*models.Post
	o := orm.NewOrm()
	o.QueryTable(new(models.Post)).All(&posts)
	this.Data["Posts"] = posts
	this.Data["Article"] = article

	this.Data["BlogUrl"] = utils.Site_config.BlogUrl

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
		utils.Error(err)
	}
	o := orm.NewOrm()
	var post models.Post
	post.CategoryId = article.Category
	post.User = strings.Trim(article.User, " ")
	shortname := strings.Trim(article.Shortname, " ")
	post.Shortname = strings.Replace(shortname, " ", "-", -1)
	post.Title = strings.Trim(article.Title, " ")
	post.Summary = article.Summary
	post.Body = article.Body
	post.Password = strings.Trim(article.Password, " ")
	post.Archive = utils.YearMonth()

	only_digests_match, _ := regexp.Match(`^[\d]+$`, []byte(post.Shortname))
	if only_digests_match && article.Id == 0 {
		MessageError = "短名称不能为纯数字!"
	} else {
		//如果是更新文章
		if article.Id >= 0 {

			// 查询Id是否存在
			var exist bool = false
			var old_post models.Post
			err := o.QueryTable(new(models.Post)).Filter("Id", article.Id).One(&old_post)
			if err != orm.ErrNoRows {
				exist = true
			}

			if exist {
				_, err := o.QueryTable(new(models.Post)).Filter("Id", article.Id).Update(orm.Params{
					"CategoryId": post.CategoryId,
					"User":       post.User,
					"Shortname":  post.Shortname,
					"Title":      post.Title,
					"Summary":    post.Summary,
					"Body":       post.Body,
					"Password":   post.Password,
					"Archive":    post.Archive,
					"Ip":         this.Ctx.Input.IP(),
				})
				if err != nil {
					MessageError = "更新文章错误!"
				} else {
					this.Data["MessageOK"] = "Update article success."

					// update cache
					hash := md5.New()
					hash.Write([]byte("/post/" + post.Shortname + ".html"))
					var url_hash string
					url_hash = hex.EncodeToString(hash.Sum(nil))
					if ok := utils.Urllist.IsExist(url_hash); ok {
						value := utils.Urllist.Get(url_hash)
						if value != nil {
							// 更新CreatedTime UpdateTime
							utils.Urllist.Delete(url_hash)
							post.CreatedTime = old_post.CreatedTime
							post.UpdateTime = old_post.UpdateTime
							utils.Urllist.Put(url_hash, &post, 3600)
						}
					}
				}

			} else {
				// 检查短名称是否重复
				post_count, _ := o.QueryTable(new(models.Post)).Filter("Shortname", post.Shortname).Count()
				if post_count > 0 {
					MessageError = "短名称重复!"
				} else {

					post.Ip = this.Ctx.Input.IP()
					id, _ := o.Insert(&post)
					// 如果未指定Shortname，则使用id作为shortname
					if post.Shortname == "" {
						_, err := o.QueryTable(new(models.Post)).Filter("Id", id).Update(orm.Params{
							"Shortname": id,
						})
						if err != nil {
							MessageError = "更新Shortname错误!"
						} else {
							this.Data["MessageOK"] = "Post new article success."

							// 验证成功则删除session
							// 解决由于失败也删除session
							// 导致验证失败后，再次提交时直接刷新页面，无任何响应的BUG
							this.DelSession(session_key)
						}
					}
				}
			}
		}

	}

	// send articles to template
	var posts []*models.Post
	o.QueryTable(new(models.Post)).All(&posts)
	this.Data["Posts"] = posts

	this.Data["BlogUrl"] = utils.Site_config.BlogUrl

	var categories []*models.Category
	o.QueryTable(new(models.Category)).All(&categories)
	this.Data["Categories"] = categories

	this.Data["xsrfdata"] = template.HTML(this.XsrfFormHtml())

	// 再次设置session
	// 解决POST文章之后，立刻再POST一篇会失败的问题
	// 但导致了F5刷新重复提交的问题
	// 不过因为上面有检测Shortname是否重复，所以最终避免了这个问题
	session_key = "admin_article_get"
	this.SetSession(session_key, utils.MakeRandomID())

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

// FileUplaod
type AdminFileController struct {
	beego.Controller
}

func (this *AdminFileController) Get() {
	o := orm.NewOrm()
	var upload_files []models.UploadFile
	num, err := o.QueryTable(new(models.UploadFile)).All(&upload_files)
	if err != nil {
		utils.Error(err)
	}

	this.Data["MaxFiles"] = num
	this.Data["UploadFiles"] = upload_files
	this.Data["BlogUrl"] = utils.Site_config.BlogUrl

	this.Data["xsrfdata"] = template.HTML(this.XsrfFormHtml())
	this.TplNames = "admin/file.html"
	this.Render()
}
