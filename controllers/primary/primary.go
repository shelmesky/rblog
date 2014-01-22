package controllers

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"html/template"
	"rblog/common/utils"
	"rblog/models"
	"strconv"
)

type ArticleComment struct {
	PostId string `form:"PostId"`
	User   string `form:"User"`
	Email  string `form:"Email"`
	Body   string `form:"Body"`
}

type MainController struct {
	beego.Controller
}

//HOME
func (this *MainController) Get() {
	o := orm.NewOrm()
	var p []*models.Post
	qs := o.QueryTable(new(models.Post))
	_, err := qs.Limit(utils.Site_config.NumPerPage).OrderBy("-CreatedTime").All(&p)
	if err != nil {
		utils.Error(err)
	}

	this.TplNames = "index.html"

	this.Data["Catagories"] = utils.Category_map.Items()
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

/*
	如果是检查文章密码的表单，检测XSRF字段
	如果是提交评论的表单，则直接返回true
*/
func (this *ArticleController) CheckXsrfCookie() bool {
	FormType := this.GetString("FormType")
	if FormType == "Encrypt" {
		return this.Controller.CheckXsrfCookie()
	} else if FormType == "Comment" {
		return true
	}
	return true
}

/*
	查询文章
	根据Id或者Shortname
*/
func (this *ArticleController) Get() {
	page_id, err := this.GetInt("comment")

	id_str := this.Ctx.Input.Param(":id")
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
		utils.Error(err)
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

			this.Data["Catagories"] = utils.Category_map.Items()
			this.Data["ArchiveCount"] = utils.ArCount
			this.Data["BlogName"] = utils.Site_config.BlogName
			this.Data["BlogUrl"] = utils.Site_config.BlogUrl
			this.Data["AdminEmail"] = utils.Site_config.AdminEmail
			this.Data["CopyRight"] = utils.Site_config.CopyRight

			if body != nil {
				// utils.Debug("Hit cache for Post.")
				this.Data["Id"] = body.Id
				this.Data["Summary"] = body.Summary
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
				// utils.Debug("Cache missed for Post.")
				category_name := utils.GetCategoryName(p.CategoryId)
				this.Data["CategoryName"] = category_name
				this.Data["Id"] = p.Id
				this.Data["Summary"] = p.Summary
				this.Data["Body"] = p.Body
				this.Data["User"] = p.User
				this.Data["Title"] = p.Title
				this.Data["Password"] = p.Password
				this.Data["CreatedTime"] = p.CreatedTime
				this.Data["UpdateTime"] = p.UpdateTime
				this.Data["xsrfdata"] = template.HTML(this.XsrfFormHtml())
				utils.Urllist.Put(url_hash, &p, 3600)
			}

			// 检查是否URL中传递的页数超过总的页数
			comment_count, err := o.QueryTable(new(models.Comment)).Filter("PostId", this.Data["Id"].(int)).Count()
			if (int(page_id) * utils.Site_config.CommentNumPerPage) > int(comment_count) {
				beego.Error(err)
				this.Abort("404")
			}

			// 获取每篇问站的评论
			var comments []*models.Comment
			comment_per_page := utils.Site_config.CommentNumPerPage
			qs := o.QueryTable(new(models.Comment)).Filter("PostId", this.Data["Id"].(int))
			_, err = qs.Limit(comment_per_page, int(page_id)*comment_per_page).All(&comments)
			if err != nil {
				utils.Error(err)
			}

			// 最大显示几页
			max_per_page := 5

			// 计算总的页数，如果不能被max_per_page整除，则加1页
			var max_pages int
			max_pages = (int(comment_count) / comment_per_page)
			if (int(comment_count) % comment_per_page) > 0 {
				max_pages += 1
			}

			var comment_count_elements []int
			// 默认从第0页开始
			var start int = 0
			var end int = 0

			// 如果总页数大于5，则默认到第5页结束
			// 否则到最大页数结束
			if max_pages >= max_per_page {
				end = max_per_page
			} else {
				end = max_pages
			}

			/*
				如果当前页数，大于等于最大页数
				就根据当前页数，算出当前页落在哪个区间
				例如当前第7页，处于5~10这个区间
			*/
			if int(page_id) >= max_per_page {
				current_five_page := int(page_id) / max_per_page
				start = current_five_page * max_per_page
				end = start + 5

				/*
					根据 总页数 % 最大显示页数 = 剩余的页数
					如果有剩余的页数，则end等于剩余的页数
					意味着页数不能被max_per_page整除

					如果end大于最大的页数
					说明已经达到末尾
					应该根据remain_page_nums重新计算end
				*/
				if end > max_pages {
					remain_page_nums := max_pages % max_per_page
					if remain_page_nums > 0 {
						end = start + remain_page_nums
					}
				}
			}

			for i := start; i < end; i++ {
				comment_count_elements = append(comment_count_elements, i)
			}

			page_id := int(page_id)
			this.Data["CommentCountNums"] = comment_count
			this.Data["Comments"] = comments
			this.Data["CommentsCount"] = comment_count_elements
			this.Data["CurrentCommentPage"] = page_id
			this.Data["PrevCommentPage"] = page_id - 1
			this.Data["NextCommentPage"] = page_id + 1
			this.Data["MaxCommentPage"] = max_pages - 1
			this.Data["MinCommentPage"] = 0

			this.TplNames = "post.html"
			this.Render()
		} else {
			utils.Error(err)
			this.Abort("500")
		}
	}
}

func (this *ArticleController) Post() {
	page_id, _ := this.GetInt("comment")
	FormType := this.GetString("FormType")
	if FormType == "Encrypt" {
		Password := this.GetString("ArticlePassword")
		Id := this.GetString("ArticleId")
		if Password == "" || Id == "" {
			this.Abort("404")
		}
		IdInt, err := strconv.Atoi(Id)
		if err != nil {
			utils.Error(err)
			this.Abort("500")
		}

		url := this.Ctx.Input.Uri()

		if Password != "" {
			var p models.Post
			o := orm.NewOrm()
			err = o.QueryTable(new(models.Post)).Filter("Id", IdInt).One(&p)
			if err != nil {
				utils.Error(err)
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

				this.Data["Catagories"] = utils.Category_map.Items()
				this.Data["ArchiveCount"] = utils.ArCount
				this.Data["BlogName"] = utils.Site_config.BlogName
				this.Data["BlogUrl"] = utils.Site_config.BlogUrl
				this.Data["AdminEmail"] = utils.Site_config.AdminEmail
				this.Data["CopyRight"] = utils.Site_config.CopyRight

				if body != nil {
					beego.Debug("Hit cache for Post.")
					this.Data["Id"] = body.Id
					this.Data["Summary"] = body.Summary
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
					this.Data["Summary"] = p.Summary
					this.Data["Body"] = p.Body
					this.Data["User"] = p.User
					this.Data["Title"] = p.Title
					this.Data["CreatedTime"] = p.CreatedTime
					this.Data["UpdateTime"] = p.UpdateTime
					utils.Urllist.Put(url_hash, &p, 3600)
				}

				// check if the `page` number in url
				comment_count, err := o.QueryTable(new(models.Comment)).Filter("PostId", this.Data["Id"].(int)).Count()
				if (int(page_id) * utils.Site_config.CommentNumPerPage) > int(comment_count) {
					utils.Error(err)
					this.Abort("404")
				}

				// Get comment for article
				var comments []*models.Comment
				comment_per_page := utils.Site_config.CommentNumPerPage
				qs := o.QueryTable(new(models.Comment)).Filter("PostId", this.Data["Id"].(int))
				_, err = qs.Limit(comment_per_page, int(page_id)*comment_per_page).All(&comments)
				if err != nil {
					utils.Error(err)
				}

				/*
					计算总的页数，如取模有余数则加1
				*/
				// 最大显示几页
				max_per_page := 5

				max_pages := (int(comment_count) / comment_per_page)
				if (int(comment_count) % comment_per_page) > 0 {
					max_pages += 1
				}

				var comment_count_elements []int
				// 默认从第0页开始
				var start int = 0
				var end int = 0

				// 如果总页数大于5，则默认到第5页结束
				// 否则到最大页数结束
				if max_pages >= max_per_page {
					end = max_per_page
				} else {
					end = max_pages
				}

				/*
					如果当前页数，大于等于最大页数
					就根据当前页数，算出当前页落在哪个区间
					例如当前第7页，处于5~10这个区间
				*/
				if int(page_id) >= max_per_page {
					current_five_page := int(page_id) / max_per_page
					start = current_five_page * max_per_page
					end = start + 5

					/*
						根据 总页数 % 最大显示页数 = 剩余的页数
						如果有剩余的页数，则end等于剩余的页数
						意味着页数不能被max_per_page整除
					*/
					remain_page_nums := max_pages % max_per_page
					if remain_page_nums > 0 {
						end = start + remain_page_nums
					}
				}

				for i := start; i < end; i++ {
					comment_count_elements = append(comment_count_elements, i)
				}

				page_id := int(page_id)
				this.Data["CommentCountNums"] = comment_count
				this.Data["Comments"] = comments
				this.Data["CommentsCount"] = comment_count_elements
				this.Data["CurrentCommentPage"] = page_id
				this.Data["PrevCommentPage"] = page_id - 1
				this.Data["NextCommentPage"] = page_id + 1
				this.Data["MaxCommentPage"] = max_pages - 1
				this.Data["MinCommentPage"] = 0

				this.TplNames = "post.html"
				this.Render()
				return
			}
			this.Ctx.Redirect(301, url)
		} else {
			this.Ctx.Redirect(301, url)
		}

	} else if FormType == "Comment" {
		var comment_form = ArticleComment{}
		if err := this.ParseForm(&comment_form); err != nil {
			utils.Error(err)
		}
		var comment models.Comment
		PostId, err := strconv.Atoi(comment_form.PostId)
		if err != nil {
			utils.Error(err)
		}
		o := orm.NewOrm()
		UserIp := this.Ctx.Input.IP()
		comment.PostId = PostId
		comment.User = comment_form.User
		comment.Email = comment_form.Email
		comment.Body = comment_form.Body
		comment.Ip = UserIp
		_, err = o.Insert(&comment)
		if err != nil {
			utils.Error(err)
		}
		url := this.Ctx.Input.Uri()
		this.Ctx.Redirect(301, url)
	}
}

type CategoryController struct {
	beego.Controller
}

func (this *CategoryController) Get() {
	category_name := this.Ctx.Input.Param(":name")
	category_id, err := utils.GetCategoryId(category_name)
	if err != nil {
		utils.Error(err)
	}

	o := orm.NewOrm()
	var posts []*models.Post
	qs := o.QueryTable(new(models.Post)).OrderBy("-CreatedTime").Filter("CategoryId", category_id)
	_, err = qs.Limit(utils.Site_config.NumPerPage).All(&posts)
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

	this.Data["Catagories"] = utils.Category_map.Items()
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
	category_name := this.Ctx.Input.Param(":name")
	category_id, err := utils.GetCategoryId(category_name)
	if err != nil {
		this.Abort("404")
		utils.Error(err)
	}

	page_id_str := this.Ctx.Input.Param(":page_id")
	page_id, err := strconv.Atoi(page_id_str)
	if err != nil {
		page_id = 0
	}

	o := orm.NewOrm()
	var posts []*models.Post
	qs := o.QueryTable(new(models.Post)).OrderBy("-CreatedTime").Filter("CategoryId", category_id)
	_, err = qs.Limit(utils.Site_config.NumPerPage, page_id*utils.Site_config.NumPerPage).All(&posts)

	if err != nil {
		utils.Error(err)
	}

	count, _ := qs.Count()

	this.Data["Catagories"] = utils.Category_map.Items()
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
	page_id_str := this.Ctx.Input.Param(":page_id")
	page_id, err := strconv.Atoi(page_id_str)
	if err != nil {
		page_id = 0
	}
	o := orm.NewOrm()
	var posts []*models.Post
	qs := o.QueryTable(new(models.Post)).OrderBy("-CreatedTime")
	_, err = qs.Limit(utils.Site_config.NumPerPage, page_id*utils.Site_config.NumPerPage).All(&posts)

	if err != nil {
		utils.Error(err)
	}

	this.Data["Catagories"] = utils.Category_map.Items()
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
