package controllers

import (
	"fmt"
	"strconv"
	"rblog/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/cache"
	"crypto/md5"
	"encoding/hex"
)

type MainController struct {
	beego.Controller
}


var (
	urllist cache.Cache
)


func init() {
	// init cache
	c, err := cache.NewCache("memory", `{"interval": 60}`)
	if err != nil {
		fmt.Println(err)
		beego.Debug(err)
	}
	urllist = c
}

//HOME
func (this *MainController) Get() { 
	o := orm.NewOrm()
	var p []*models.Post
	o.QueryTable(new(models.Post)).All(&p)
	
	var site_config models.SiteConfig
	o.QueryTable(new(models.SiteConfig)).One(&site_config)
	
	this.TplNames = "index.html"
	this.Data["Posts"] = p
	this.Data["BlogName"] = site_config.BlogName
	this.Data["BlogUrl"] = site_config.BlogUrl
	this.Data["AdminEmail"] = site_config.AdminEmail
	this.Data["CopyRight"] = site_config.CopyRight
	this.Render()
}

type ArticleController struct {
	beego.Controller
}

//查询文章，根据Id或者Shortname
func (this *ArticleController) Get() {
	id_str := this.Ctx.Input.Params(":id")
	id, err := strconv.ParseInt(id_str, 10, 32)

	o := orm.NewOrm()

	var use_Id bool = false
	var p models.Post
	if err == nil {
		p = models.Post{Id: int(id)}
		use_Id = true
	} else {
		p = models.Post{Shortname: id_str}
	}

	if use_Id {
		err = o.Read(&p)
	} else {
		err = o.Read(&p, "Shortname")
	}

	if err == orm.ErrNoRows {
		beego.Debug(err)
		this.Ctx.WriteString("查询不到")
	} else if err == orm.ErrMissPK {
		beego.Debug(err)
		this.Ctx.WriteString("找不到主键")
	} else {
		if err == nil {
			// query cache for article body
			url := this.Ctx.Input.Uri()
			hash := md5.New()
			hash.Write([]byte(url))
			var url_hash string
			url_hash = hex.EncodeToString(hash.Sum(nil))
			var body *models.Post
			if ok := urllist.IsExist(url_hash); ok {
				value := urllist.Get(url_hash)
				if value != nil {
					body = value.(*models.Post)
				}
			}
			
			var site_config models.SiteConfig
			o.QueryTable(new(models.SiteConfig)).One(&site_config)
			
			this.Data["Posts"] = p
			this.Data["BlogName"] = site_config.BlogName
			this.Data["BlogUrl"] = site_config.BlogUrl
			this.Data["AdminEmail"] = site_config.AdminEmail
			this.Data["CopyRight"] = site_config.CopyRight
			
			if body != nil {
				beego.Debug("Hit cache for Post.")
				this.Data["Body"] = body.Body
				this.Data["Title"] = body.Title
				this.Data["CreatedTime"] = body.Time
			} else {
				beego.Debug("Cache missed for Post.")
				this.Data["Body"] = p.Body
				this.Data["Title"] = p.Title
				this.Data["CreatedTime"] = p.Time
				urllist.Put(url_hash, &p, 3600)
			}
			this.TplNames = "post.html"
			this.Render()
		} else {
			beego.Debug(err)
			fmt.Println(err)
		}
	}
}

type AdminController struct {
	beego.Controller
}

//管理后台
func (this *AdminController) Get() {
	this.Ctx.WriteString("admin page")
}


