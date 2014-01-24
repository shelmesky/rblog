package main

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	beego_utils "github.com/astaxie/beego/utils"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"rblog/common/utils"
	"rblog/controllers/admin"
	"rblog/controllers/debug"
	"rblog/controllers/feed"
	"rblog/controllers/primary"
	"rblog/controllers/search"
	"rblog/controllers/upload"
	"rblog/models"
	"runtime"
)

type MySQL_Config struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
}

type Main_Config struct {
	MySQL       MySQL_Config
	Static_Path string
	Log_Path    string
}

func init() {
	var config_file string = "config.json"
	var file_size int64
	var config Main_Config

	// Check config file
	stat, err := os.Stat(config_file)
	if os.IsNotExist(err) {
		fmt.Println("Config file does not exists, Please Check it out.")
		fmt.Println("Server exit now.")
		os.Exit(1)
	}

	file, err := os.Open(config_file)
	defer file.Close()

	if err == nil {
		file_size = stat.Size()
	}

	buf := make([]byte, file_size)
	for {
		count, _ := file.Read(buf)
		if count == 0 {
			break
		}
	}

	file.Close()

	err = json.Unmarshal(buf, &config)
	if err == nil {
		// Set logger
		log_config_str := `{"filename": "%s", "daily": true, "maxdays": 15, "maxsize": 67108864, "rotate": true}`
		log_config := fmt.Sprintf(log_config_str, config.Log_Path)
		beego.BeeLogger.SetLogger("file", log_config)
		beego.SetLevel(beego.LevelTrace)

		orm.RegisterDriver("mysql", orm.DR_MySQL)
		// Init DB Connection
		mc := config.MySQL
		conn_uri := mc.Username + ":" + mc.Password + "@tcp(" + mc.Host + ":" + mc.Port + ")" + "/" + mc.Database + "?charset=utf8"
		orm.RegisterDataBase("default", "mysql", conn_uri, 30)
	} else {
		fmt.Println(err)
		beego.Debug(err)
	}

	// Set Config fit ORM
	orm.RunCommand()
	orm.Debug = false

	// Set static library path
	beego.SetStaticPath("./static", config.Static_Path)

	//init global site config
	o := orm.NewOrm()
	o.QueryTable(new(models.SiteConfig)).One(&utils.Site_config)

	// int corotine safe map
	utils.Category_map = beego_utils.NewBeeMap()

	// insert catagories to map
	var categories []*models.Category
	o.QueryTable(new(models.Category)).All(&categories)

	for _, category := range categories {
		utils.Category_map.Set(category.Id, category.Name)
	}

	// cache the archives count
	utils.ArCount, err = utils.GetArchives()
	if err != nil {
		beego.Error(err)
	}

	// cache the categories count
	utils.CatCount, err = utils.GetCategories()
	if err != nil {
		utils.Error(err)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// global filter
	beego.AddFilter("/admin", "AfterStatic", utils.AuthFilter)
	beego.AddFilter("/admin/:all", "AfterStatic", utils.AuthFilter)
	beego.AddFilter("/post/upload", "AfterStatic", utils.AuthFilter)

	beego.AddFuncMap("markdown", utils.RenderMarkdown)
	beego.AddFuncMap("categoryname", utils.GetCategoryName)
	beego.AddFuncMap("filesize", utils.FileSize)
	beego.EnableAdmin = true
	beego.AdminHttpAddr = "0.0.0.0"
	beego.AdminHttpPort = 8088

	beego.Router("/", &controllers.MainController{})

	beego.Router("/post/:id([^/]+).html", &controllers.ArticleController{})
	beego.Router("/page/:page_id([^/]+)", &controllers.PageController{})

	beego.Router("/category/:name([^/]+)", &controllers.CategoryController{})
	beego.Router("/category/:name([^/]+)/page/:page_id([^/]+)", &controllers.CategoryPageController{})

	beego.Router("/archive/:name([^/]+)", &controllers.ArchiveController{})
	beego.Router("/archive/:name([^/]+)/page/:page_id([^/]+)", &controllers.ArchivePageController{})

	// search handler
	beego.Router("/post/search", &searchcontrollers.SearchController{})
	beego.Router("/post/search/:keyword([^/]+)/page/:page_id([^/]+)", &searchcontrollers.SearchPageController{})

	// admin console
	beego.Router("/admin", &admincontrollers.AdminController{})
	beego.Router("/admin/login", &admincontrollers.AdminLoginController{})
	beego.Router("/admin/logout", &admincontrollers.AdminLogoutController{})
	beego.Router("/admin/article", &admincontrollers.AdminArticleController{})
	beego.Router("/admin/category", &admincontrollers.AdminCategoryController{})
	beego.Router("/admin/comment", &admincontrollers.AdminCommentController{})
	beego.Router("/admin/site", &admincontrollers.AdminSiteController{})
	beego.Router("/admin/files", &admincontrollers.AdminFileController{})

	//add http pprof url handler
	beego.Router("/debug/pprof", &debugcontrollers.ProfController{})
	beego.Router("/debug/pprof/:pp([^/]+)", &debugcontrollers.ProfController{})

	// add feed handler
	beego.Router("/feed", &feedcontrollers.RssController{})

	// file upload handler
	beego.Router("/post/upload", &uploadcontrollers.UploadController{})

	// file download handler
	beego.Router("/post/download/:filehash([^/]+)", &uploadcontrollers.DownloadController{})

	/*
		go utils.SendEmailWithAttachments(
			"ox55aa@126.com",
			"来自126的测试邮件",
			[]string{"33326771@qq.com"},
			"附件列表",
			[]string{"/home/roy/coding/Golang_SourceCode/rblog/src/rblog/测试1.log",
				"/home/roy/coding/Golang_SourceCode/rblog/src/rblog/测试2.log"},
		)
	*/

	beego.Run()
}
