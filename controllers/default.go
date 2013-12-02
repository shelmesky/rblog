package controllers

import (
	"fmt"
	"strconv"
	"rblog/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
)

type MainController struct {
	beego.Controller
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
			var site_config models.SiteConfig
			o.QueryTable(new(models.SiteConfig)).One(&site_config)
			
			this.Data["Posts"] = p
			this.Data["BlogName"] = site_config.BlogName
			this.Data["BlogUrl"] = site_config.BlogUrl
			this.Data["AdminEmail"] = site_config.AdminEmail
			this.Data["CopyRight"] = site_config.CopyRight
			
			this.Data["Title"] = p.Title
			this.Data["CreatedTime"] = p.Time
			this.Data["Body"] = p.Body
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


