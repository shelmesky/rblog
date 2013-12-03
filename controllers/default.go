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
	"errors"
)

type MainController struct {
	beego.Controller
}


var (
	urllist cache.Cache
	Site_config models.SiteConfig
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


func GetCategoryName(id int) (string, error) {
	var category models.Category
	o := orm.NewOrm()
	err := o.QueryTable(new(models.Category)).Filter("Id", id).One(&category)
		if err == orm.ErrMultiRows {
	    // 多条的时候报错
	    return "", errors.New("Returned Multi Rows Not One")
	}
	if err == orm.ErrNoRows {
	    // 没有找到记录
	    return "", errors.New("Not row found")
	}
	return category.Name, nil
}


func GetCategoryId(name string) (int, error) {
	var category models.Category
	o := orm.NewOrm()
	err := o.QueryTable(new(models.Category)).Filter("Name", name).One(&category)
		if err == orm.ErrMultiRows {
	    // 多条的时候报错
	    return 0, errors.New("Returned Multi Rows Not One")
	}
	if err == orm.ErrNoRows {
	    // 没有找到记录
	    return 0, errors.New("Not row found")
	}
	return category.Id, nil
}


//HOME
func (this *MainController) Get() {
	o := orm.NewOrm()
	var p []*models.Post
	o.QueryTable(new(models.Post)).Limit(Site_config.NumPerPage).All(&p)
	
	this.TplNames = "index.html"
	this.Data["Posts"] = p
	this.Data["BlogName"] = Site_config.BlogName
	this.Data["BlogUrl"] = Site_config.BlogUrl
	this.Data["AdminEmail"] = Site_config.AdminEmail
	this.Data["CopyRight"] = Site_config.CopyRight
	
	this.Data["OldPage"] = 1
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
			
			
			this.Data["BlogName"] = Site_config.BlogName
			this.Data["BlogUrl"] = Site_config.BlogUrl
			this.Data["AdminEmail"] = Site_config.AdminEmail
			this.Data["CopyRight"] = Site_config.CopyRight
			
			if body != nil {
				beego.Debug("Hit cache for Post.")
				this.Data["Body"] = body.Body
				this.Data["Title"] = body.Title
				this.Data["CreatedTime"] = body.Time
				category_name, err := GetCategoryName(body.CategoryId)
				if err != nil {
					beego.Error(err)
				}
				this.Data["CategoryName"] = category_name
			} else {
				beego.Debug("Cache missed for Post.")
				category_name, _ := GetCategoryName(p.CategoryId)
				this.Data["CategoryName"] = category_name
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

type CategoryController struct {
	beego.Controller
}

func (this *CategoryController) Get() {
	category_name := this.Ctx.Input.Params(":name")
	category_id, err := GetCategoryId(category_name)
	if err != nil {
		beego.Error(err)
	}
	
	o := orm.NewOrm()
	var posts []*models.Post
	num, err := o.QueryTable(new(models.Post)).Filter("CategoryId", category_id).Limit(10).All(&posts)
	if err != nil {
		beego.Error(err)
	}
	
	this.Data["Posts"] = posts
			
	this.Data["BlogName"] = Site_config.BlogName
	this.Data["BlogUrl"] = Site_config.BlogUrl
	this.Data["AdminEmail"] = Site_config.AdminEmail
	this.Data["CopyRight"] = Site_config.CopyRight
	this.Data["CategoryCounts"] = num
	this.Data["CategoryName"] = category_name
	
	this.TplNames = "category.html"
	this.Render()
	
}


type CategoryPageController struct {
	beego.Controller
}


func (this *CategoryPageController) Get() {
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
	fmt.Println(Site_config.NumPerPage, page_id)
	qs := o.QueryTable(new(models.Post))
	_, err = qs.Limit(Site_config.NumPerPage, page_id*Site_config.NumPerPage).All(&posts)
	
	if err != nil {
		beego.Error(err)
	}
	
	this.Data["Posts"] = posts
	this.Data["BlogName"] = Site_config.BlogName
	this.Data["BlogUrl"] = Site_config.BlogUrl
	this.Data["AdminEmail"] = Site_config.AdminEmail
	this.Data["CopyRight"] = Site_config.CopyRight
	
	/*
	算出总的文章数
	再根据当前页和每页数量，计算出还剩几条记录
	如果剩余记录数的大于每页数量，就显示Older按钮
	否则不显示
	*/
	count, _ := qs.Count()
	remain_page := int(count) - (page_id * Site_config.NumPerPage)
	if remain_page > Site_config.NumPerPage {
		this.Data["OldPage"] = page_id + 1
	} else if remain_page <= Site_config.NumPerPage {
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


type AdminController struct {
	beego.Controller
}


//管理后台
func (this *AdminController) Get() {
	this.Ctx.WriteString("admin page")
}


