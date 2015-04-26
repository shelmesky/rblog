package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/shelmesky/rblog/common/utils"
	"github.com/shelmesky/rblog/models"
	"strconv"
)

type ArchiveController struct {
	beego.Controller
}

func (this *ArchiveController) Get() {
	archive := this.Ctx.Input.Param(":name")

	o := orm.NewOrm()
	var posts []*models.Post
	qs := o.QueryTable(new(models.Post)).OrderBy("-CreatedTime").Filter("Archive", archive)
	_, err := qs.Limit(utils.Site_config.NumPerPage).All(&posts)
	if err != nil {
		this.Abort("404")
		utils.Error(err)
	}

	count, _ := qs.Count()
	if int(count) <= utils.Site_config.NumPerPage {
		this.Data["OldPage"] = -1
	} else {
		this.Data["OldPage"] = 1
	}
	this.Data["NewPage"] = -1

	this.Data["Catagories"] = utils.CatCount
	this.Data["Posts"] = posts
	this.Data["ArchiveCount"] = utils.ArCount
	this.Data["BlogName"] = utils.Site_config.BlogName
	this.Data["BlogUrl"] = utils.Site_config.BlogUrl
	this.Data["AdminEmail"] = utils.Site_config.AdminEmail
	this.Data["CopyRight"] = utils.Site_config.CopyRight
	this.Data["ArchiveCounts"] = count
	this.Data["ArchiveName"] = archive

	this.TplNames = "archive.html"
	this.Render()

}

type ArchivePageController struct {
	beego.Controller
}

func (this *ArchivePageController) Get() {
	archive := this.Ctx.Input.Param(":name")

	page_id_str := this.Ctx.Input.Param(":page_id")
	page_id, err := strconv.Atoi(page_id_str)
	if err != nil {
		page_id = 0
	}

	o := orm.NewOrm()
	var posts []*models.Post
	qs := o.QueryTable(new(models.Post)).OrderBy("-CreatedTime").Filter("Archive", archive)
	_, err = qs.Limit(utils.Site_config.NumPerPage, page_id*utils.Site_config.NumPerPage).All(&posts)

	if err != nil {
		utils.Error(err)
	}

	count, _ := qs.Count()

	this.Data["Catagories"] = utils.Category_map.Items()
	this.Data["ArchiveCount"] = utils.ArCount
	this.Data["ArchiveName"] = archive
	this.Data["ArchiveCounts"] = count
	this.Data["Posts"] = posts
	this.Data["BlogName"] = utils.Site_config.BlogName
	this.Data["BlogUrl"] = utils.Site_config.BlogUrl
	this.Data["AdminEmail"] = utils.Site_config.AdminEmail
	this.Data["CopyRight"] = utils.Site_config.CopyRight

	/*
		算出总的文章数
		再根据当前页和每页数量，计算出还剩几条记录
		如果剩余记录数的大于每页数量，就显示Older按钮
		否则不显示
	*/
	remain_page := int(count) - (page_id * utils.Site_config.NumPerPage)
	if remain_page > utils.Site_config.NumPerPage {
		this.Data["OldPage"] = page_id + 1
	} else if remain_page <= utils.Site_config.NumPerPage {
		this.Data["OldPage"] = -1
	}

	/*
		当page_id==1，NewPage==0，显示第一页
		当page_id==0，NewPage==-1，不显示Newer按钮
		以上是在index.html中判断
	*/
	this.Data["NewPage"] = page_id - 1
	this.TplNames = "archive.html"
	this.Render()
}
