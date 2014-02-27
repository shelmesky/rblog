package admincontrollers

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"rblog/common/utils"
	"rblog/models"
	"regexp"
	"strings"
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
	page_id, _ := this.GetInt("page")
	var limit int64 = 10

	action := this.GetString("action")
	id, _ := this.GetInt("id")

	o := orm.NewOrm()

	var article models.Post
	if action != "" && id >= 0 {
		if action == "update" {
			err := o.QueryTable(new(models.Post)).Filter("Id", id).One(&article)
			if err != nil {
				utils.Error(err)
			}
		}
		if action == "delete" {
			// 删除DB记录
			_, err := o.QueryTable(new(models.Post)).Filter("Id", id).Delete()
			if err != nil {
				utils.Error(err)
				this.Abort("404")
				return
			}
			this.Redirect("/admin/article", 301)
		}
	}

	post_count, err := o.QueryTable(new(models.Post)).Count()
	if (page_id * limit) > post_count {
		utils.Error(err)
		this.Abort("404")
	}

	var posts []*models.Post
	o.QueryTable(new(models.Post)).Limit(limit, page_id*limit).OrderBy("-CreatedTime").All(&posts)

	/* 开始处理分页 */

	// 最大显示几页
	max_per_page := 5

	// 计算总的页数，如果不能被max_per_page整除，则加1页
	var max_pages int
	max_pages = (int(post_count) / int(limit))
	if (post_count % limit) > 0 {
		max_pages += 1
	}

	var post_count_elements []int
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
		post_count_elements = append(post_count_elements, i)
	}

	this.Data["PostCountNums"] = post_count
	this.Data["Posts"] = posts
	this.Data["PostsCount"] = post_count_elements
	this.Data["CurrentPostPage"] = page_id
	this.Data["PrevPostPage"] = page_id - 1
	this.Data["NextPostPage"] = page_id + 1
	this.Data["MaxPostPage"] = max_pages - 1
	this.Data["MinPostPage"] = 0

	if page_id >= 0 {
		this.Data["PageId"] = page_id
	}

	/* 结束处理分页 */

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
						}
					} else {
						this.Data["MessageOK"] = "Post new article success."
					}

					// 重新载入Archives和Categories
					utils.LoadArchivesAndCategory(o)

					// 验证成功则删除session
					// 解决由于失败也删除session
					// 导致验证失败后，再次提交时直接刷新页面，无任何响应的BUG
					this.DelSession(session_key)
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
	page_id, _ := this.GetInt("page")
	var limit int64 = 5

	action := this.GetString("action")
	id, _ := this.GetInt("id")

	o := orm.NewOrm()

	if action != "" && id >= 0 {
		if action == "delete" {
			// 删除DB记录
			_, err := o.QueryTable(new(models.Comment)).Filter("Id", id).Delete()
			if err != nil {
				utils.Error(err)
				this.Abort("404")
				return
			}
			this.Redirect("/admin/comment", 301)
		}
	}

	comment_count, err := o.QueryTable(new(models.Comment)).Count()
	if (page_id * limit) > comment_count {
		utils.Error(err)
		this.Abort("404")
	}

	var comments []*models.Comment
	o.QueryTable(new(models.Comment)).Limit(limit, page_id*limit).OrderBy("-CreatedTime").All(&comments)

	/* 开始处理分页 */

	// 最大显示几页
	max_per_page := 5

	// 计算总的页数，如果不能被max_per_page整除，则加1页
	var max_pages int
	max_pages = (int(comment_count) / int(limit))
	if (comment_count % limit) > 0 {
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

	this.Data["CommentCountNums"] = comment_count
	this.Data["Comments"] = comments
	this.Data["CommentsCount"] = comment_count_elements
	this.Data["CurrentCommentPage"] = page_id
	this.Data["PrevCommentPage"] = page_id - 1
	this.Data["NextCommentPage"] = page_id + 1
	this.Data["MaxCommentPage"] = max_pages - 1
	this.Data["MinCommentPage"] = 0

	/* 结束处理分页 */

	this.Data["MaxComments"] = comment_count
	this.Data["BlogUrl"] = utils.Site_config.BlogUrl

	this.Data["xsrfdata"] = template.HTML(this.XsrfFormHtml())

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
	page_id, _ := this.GetInt("page")
	var limit int64 = 5

	action := this.GetString("action")
	id, _ := this.GetInt("id")

	o := orm.NewOrm()

	var upload_file models.UploadFile
	if action != "" && id >= 0 {
		if action == "delete" {
			err := o.QueryTable(new(models.UploadFile)).Filter("Id", id).One(&upload_file)
			if err != nil {
				utils.Error(err)
				this.Abort("404")
				return
			}

			// 获取文件完整路径
			current_path, err := filepath.Abs(filepath.Dir(os.Args[0]))
			if err != nil {
				utils.Error(err)
				this.Abort("404")
				return
			}

			var file_exist bool = true
			// 判断文件是否存在
			fullpath := path.Join(current_path, "upload", upload_file.Hashname)
			if !utils.Exist(fullpath) {
				file_exist = false
			}

			// 删除DB记录
			_, err = o.QueryTable(new(models.UploadFile)).Filter("Id", id).Delete()
			if err != nil {
				utils.Error(err)
				this.Abort("404")
				return
			}

			// 从文件系统上删除文件
			if file_exist {
				utils.Debug("Delete file:", fullpath)
				err = os.Remove(fullpath)
				if err != nil {
					// 删除文件失败后，只记录日志
					utils.Error(err)
				}
			}

			this.Redirect("/admin/files", 301)
		}
	}

	upload_count, err := o.QueryTable(new(models.UploadFile)).Count()
	if (page_id * limit) > upload_count {
		utils.Error(err)
		this.Abort("404")
	}

	var uploads []*models.UploadFile
	o.QueryTable(new(models.UploadFile)).Limit(limit, page_id*limit).OrderBy("-UploadTime").All(&uploads)

	/* 开始处理分页 */

	// 最大显示几页
	max_per_page := 5

	// 计算总的页数，如果不能被max_per_page整除，则加1页
	var max_pages int
	max_pages = (int(upload_count) / int(limit))
	if (upload_count % limit) > 0 {
		max_pages += 1
	}

	var upload_count_elements []int
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
		upload_count_elements = append(upload_count_elements, i)
	}

	this.Data["UploadCountNums"] = upload_count
	this.Data["UploadFiles"] = uploads
	this.Data["UploadsCount"] = upload_count_elements
	this.Data["CurrentUploadPage"] = page_id
	this.Data["PrevUploadPage"] = page_id - 1
	this.Data["NextUploadPage"] = page_id + 1
	this.Data["MaxUploadPage"] = max_pages - 1
	this.Data["MinUploadPage"] = 0

	/* 结束处理分页 */

	this.Data["MaxFiles"] = upload_count
	this.Data["BlogUrl"] = utils.Site_config.BlogUrl

	this.Data["xsrfdata"] = template.HTML(this.XsrfFormHtml())
	this.TplNames = "admin/file.html"
	this.Render()
}
