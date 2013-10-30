package main

import (
	"fmt"
	"os"
	"encoding/json"
	"rblog/controllers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

type MySQL_Config struct {
	Host string
	Username string
	Password string
	Database string
}

type Main_Config struct {
	MySQL MySQL_Config
	Static_Path string
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
		orm.RegisterDriver("mysql", orm.DR_MySQL)
		// Init DB Connection
		mc := config.MySQL
		conn_uri := mc.Username + ":" + mc.Password + "@/" + mc.Database + "?charset=utf8"
		orm.RegisterDataBase("default", "mysql", conn_uri, 30)
	} else {
		fmt.Println(err)
	}

	// Set static library path
	beego.SetStaticPath("/static", config.Static_Path)
}


func main() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/admin", &controllers.AdminController{})
	beego.Router("/post/:id([^/]+)", &controllers.ArticleController{})
	beego.Run()
}

