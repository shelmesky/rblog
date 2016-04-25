package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/shelmesky/rblog/common/utils"
	"github.com/shelmesky/rblog/models"
)

type AboutController struct {
	beego.Controller
}

func (this *AboutController) Get() {
	this.TplName = "about.html"

	this.Data["Catagories"] = utils.CatCount
	this.Data["ArchiveCount"] = utils.ArCount
	this.Data["BlogName"] = utils.Site_config.BlogName
	this.Data["BlogUrl"] = utils.Site_config.BlogUrl
	this.Data["AdminEmail"] = utils.Site_config.AdminEmail
	this.Data["CopyRight"] = utils.Site_config.CopyRight

	o := orm.NewOrm()
	var about models.About
	err := o.QueryTable(new(models.About)).One(&about)
	if err != nil {
		utils.Error(err)
		this.Abort("404")
	}

	this.Data["About"] = about

	this.Render()
}
