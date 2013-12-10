package controllers

import (
	"strconv"
	"rblog/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"crypto/md5"
	"encoding/hex"
	"html/template"
	"rblog/common/utils"
)


type MainController struct {
	beego.Controller
}


//HOME
func (this *MainController) Get() {
	o := orm.NewOrm()
	var p []*models.Post
	qs := o.QueryTable(new(models.Post))
	_, err := qs.Limit(utils.Site_config.NumPerPage).OrderBy("-id").All(&p)
	if err != nil {
		beego.Error(err)
	}
	
	this.TplNames = "index.html"
	this.Data["ArchiveCount"] = utils.ArCount
	this.Data["Posts"] = p
	this.Data["BlogName"] = utils.Site_config.BlogName
	this.Data["BlogUrl"] = utils.Site_config.BlogUrl
	this.Data["AdminEmail"] = utils.Site_config.AdminEmail
	this.Data["CopyRight"] = utils.Site_config.CopyRight
	
	count, _ := qs.Count()
	if int(count) <= utils.Site_config.NumPerPage {
		this.Data["OldPage"] = -1
	} else {
		this.Data["OldPage"] = 1
	}
	this.Data["NewPage"] = -1
	
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
		this.Abort("404")
	} else if err == orm.ErrMissPK {
		beego.Debug(err)
		this.Abort("500")
	} else {
		if err == nil {
			// query cache for article body
			url := this.Ctx.Input.Uri()
			hash := md5.New()
			hash.Write([]byte(url))
			var url_hash string
			url_hash = hex.EncodeToString(hash.Sum(nil))
			var body *models.Post
			if ok := utils.Urllist.IsExist(url_hash); ok {
				value := utils.Urllist.Get(url_hash)
				if value != nil {
					body = value.(*models.Post)
				}
			}
			
			this.Data["ArchiveCount"] = utils.ArCount
			this.Data["BlogName"] = utils.Site_config.BlogName
			this.Data["BlogUrl"] = utils.Site_config.BlogUrl
			this.Data["AdminEmail"] = utils.Site_config.AdminEmail
			this.Data["CopyRight"] = utils.Site_config.CopyRight
			
			if body != nil {
				beego.Debug("Hit cache for Post.")
				this.Data["Id"] = body.Id
				this.Data["Body"] = body.Body
				this.Data["User"] = body.User
				this.Data["Title"] = body.Title
				this.Data["Password"] = body.Password
				this.Data["CreatedTime"] = body.CreatedTime
				this.Data["UpdateTime"] = body.UpdateTime
				this.Data["xsrfdata"] = template.HTML(this.XsrfFormHtml())
				category_name := utils.GetCategoryName(body.CategoryId)
				this.Data["CategoryName"] = category_name
			} else {
				beego.Debug("Cache missed for Post.")
				category_name := utils.GetCategoryName(p.CategoryId)
				this.Data["CategoryName"] = category_name
				this.Data["Id"] = p.Id
				this.Data["Body"] = p.Body
				this.Data["User"] = p.User
				this.Data["Title"] = p.Title
				this.Data["Password"] = p.Password
				this.Data["CreatedTime"] = p.CreatedTime
				this.Data["UpdateTime"] = p.UpdateTime
				this.Data["xsrfdata"] = template.HTML(this.XsrfFormHtml())
				utils.Urllist.Put(url_hash, &p, 3600)
			}
			this.TplNames = "post.html"
			this.Render()
		} else {
			beego.Debug(err)
			this.Abort("500")
		}
	}
}


func (this *ArticleController) Post() {
	Password := this.GetString("ArticlePassword")
	Id := this.Input().Get("ArticleId")
	IdInt, err := strconv.Atoi(Id)
	if err != nil {
		beego.Error(err)
		this.Abort("500")
	}
	
	url := this.Ctx.Input.Uri()
	
	if Password != "" {
		var p models.Post
		o := orm.NewOrm()
		err = o.QueryTable(new(models.Post)).Filter("Id", IdInt).One(&p)
		if err != nil {
			beego.Error(err)
			this.Abort("500")
		}
		
		if Password == p.Password {
			// query cache for article body
			hash := md5.New()
			hash.Write([]byte(url))
			var url_hash string
			url_hash = hex.EncodeToString(hash.Sum(nil))
			var body *models.Post
			if ok := utils.Urllist.IsExist(url_hash); ok {
				value := utils.Urllist.Get(url_hash)
				if value != nil {
					body = value.(*models.Post)
				}
			}
			
			this.Data["ArchiveCount"] = utils.ArCount
			this.Data["BlogName"] = utils.Site_config.BlogName
			this.Data["BlogUrl"] = utils.Site_config.BlogUrl
			this.Data["AdminEmail"] = utils.Site_config.AdminEmail
			this.Data["CopyRight"] = utils.Site_config.CopyRight
			
			if body != nil {
				beego.Debug("Hit cache for Post.")
				this.Data["Id"] = body.Id
				this.Data["Body"] = body.Body
				this.Data["User"] = body.User
				this.Data["Title"] = body.Title
				this.Data["CreatedTime"] = body.CreatedTime
				this.Data["UpdateTime"] = body.UpdateTime
				category_name := utils.GetCategoryName(body.CategoryId)
				if err != nil {
					beego.Error(err)
				}
				this.Data["CategoryName"] = category_name
			} else {
				beego.Debug("Cache missed for Post.")
				category_name := utils.GetCategoryName(p.CategoryId)
				this.Data["CategoryName"] = category_name
				this.Data["Id"] = p.Id
				this.Data["Body"] = p.Body
				this.Data["User"] = p.User
				this.Data["Title"] = p.Title
				this.Data["CreatedTime"] = p.CreatedTime
				this.Data["UpdateTime"] = p.UpdateTime
				utils.Urllist.Put(url_hash, &p, 3600)
			}
			this.TplNames = "post.html"
			this.Render()
			return
		}
		this.Ctx.Redirect(301, url)
	} else {
		this.Ctx.Redirect(301, url)
	}
}


type CategoryController struct {
	beego.Controller
}

func (this *CategoryController) Get() {
	category_name := this.Ctx.Input.Params(":name")
	category_id, err := utils.GetCategoryId(category_name)
	if err != nil {
		beego.Error(err)
	}
	
	o := orm.NewOrm()
	var posts []*models.Post
	qs := o.QueryTable(new(models.Post)).OrderBy("-id").Filter("CategoryId", category_id)
	_, err = qs.Limit(utils.Site_config.NumPerPage).All(&posts)
	if err != nil {
		this.Abort("404")
		beego.Error(err)
	}
	
	count, _ := qs.Count()
	if int(count) <= utils.Site_config.NumPerPage {
		this.Data["OldPage"] = -1
	} else {
		this.Data["OldPage"] = 1
	}
	this.Data["NewPage"] = -1
	
	this.Data["Posts"] = posts
	this.Data["ArchiveCount"] = utils.ArCount
	this.Data["BlogName"] = utils.Site_config.BlogName
	this.Data["BlogUrl"] = utils.Site_config.BlogUrl
	this.Data["AdminEmail"] = utils.Site_config.AdminEmail
	this.Data["CopyRight"] = utils.Site_config.CopyRight
	this.Data["CategoryCounts"] = count
	this.Data["CategoryName"] = category_name
	
	this.TplNames = "category.html"
	this.Render()
	
}


type CategoryPageController struct {
	beego.Controller
}


func (this *CategoryPageController) Get() {
	category_name := this.Ctx.Input.Params(":name")
	category_id, err := utils.GetCategoryId(category_name)
	if err != nil {
		this.Abort("404")
		beego.Error(err)
	}
	
	page_id_str := this.Ctx.Input.Params(":page_id")
	page_id, err := strconv.Atoi(page_id_str)
	if err != nil {
		page_id = 0
	}
	
	o := orm.NewOrm()
	var posts []*models.Post
	qs := o.QueryTable(new(models.Post)).OrderBy("-id").Filter("CategoryId", category_id)
	_, err = qs.Limit(utils.Site_config.NumPerPage, page_id*utils.Site_config.NumPerPage).All(&posts)
	
	if err != nil {
		beego.Error(err)
	}
	
	count, _ := qs.Count()
	this.Data["ArchiveCount"] = utils.ArCount
	this.Data["CategoryName"] = category_name
	this.Data["CategoryCounts"] = count
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
	this.TplNames = "category.html"
	this.Render()
}


type PageController struct {
	beego.Controller
}


func (this *PageController) Get() {
	page_id_str := this.Ctx.Input.Params(":page_id")
	page_id, err := strconv.Atoi(page_id_str)
	if err != nil {
		page_id = 0
	}
	o := orm.NewOrm()
	var posts []*models.Post
	qs := o.QueryTable(new(models.Post)).OrderBy("-id")
	_, err = qs.Limit(utils.Site_config.NumPerPage, page_id*utils.Site_config.NumPerPage).All(&posts)
	
	if err != nil {
		beego.Error(err)
	}
	this.Data["ArchiveCount"] = utils.ArCount
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
	count, _ := qs.Count()
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
	this.TplNames = "index.html"
	this.Render()
}


