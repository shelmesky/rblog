package utils

import (
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/cache"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/utils"
	"github.com/russross/blackfriday"
	"html/template"
	"io"
	"math/rand"
	"mime"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"rblog/common/api/smtp"
	"rblog/models"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	Category_map *utils.BeeMap
	Urllist      cache.Cache
	Site_config  models.SiteConfig
	ArCount      []ArchiveCount
)

type ArchiveCount struct {
	Archive string
	Count   int
}

func init() {
	// init cache
	c, err := cache.NewCache("memory", `{"interval": 60}`)
	if err != nil {
		fmt.Println(err)
		beego.Debug(err)
	}
	Urllist = c
}

func MakeRandomID() string {
	nano := time.Now().UnixNano()
	rand.Seed(nano)
	rndNum := rand.Int63()

	md5_nano := MD5(strconv.FormatInt(nano, 10))
	md5_rand := MD5(strconv.FormatInt(rndNum, 10))
	RandomID := MD5(md5_nano + md5_rand)
	return RandomID
}

func MD5(text string) string {
	hashMD5 := md5.New()
	io.WriteString(hashMD5, text)
	return fmt.Sprintf("%x", hashMD5.Sum(nil))
}

func GetFuncName(function interface{}) string {
	func_pointer := reflect.ValueOf(function).Pointer()
	return runtime.FuncForPC(func_pointer).Name()
}

type NewTime struct {
	time.Time
}

func (t NewTime) YearMonthString() string {
	const layout = "2006-01"
	return t.Format(layout)
}

func (t NewTime) NowString() string {
	const layout = "2006-01-02 15:04:05"
	return t.Format(layout)
}

func YearMonth() string {
	ta := time.Now()
	t := NewTime{ta}
	return t.YearMonthString()
}

func Now() string {
	ta := time.Now()
	t := NewTime{ta}
	return t.NowString()
}

func GetCategoryName(content interface{}) string {
	//fmt.Println(reflect.TypeOf(content))
	var category_name string
	if value, ok := content.(int); ok {
		if Category_map.Check(value) {
			category := Category_map.Get(value)
			category_name, _ := category.(string)
			return category_name
		} else {
			o := orm.NewOrm()
			var category models.Category
			err := o.QueryTable(new(models.Category)).Filter("Id", value).One(&category)
			if err != nil {
				Error(err)
				return string(value)
			}
			return category.Name
		}
	}

	return category_name
}

func RenderMarkdown(content interface{}) string {
	var output []byte
	if value, ok := content.(template.HTML); ok {
		output = blackfriday.MarkdownCommon([]byte(value))
	} else if value, ok := content.(string); ok {
		output = blackfriday.MarkdownCommon([]byte(value))
	}
	return string(output)

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

func GetArchives() ([]ArchiveCount, error) {
	var sql = `select distinct archive as ar,count(archive) as count
	from post group by archive`
	o := orm.NewOrm()
	var archives []ArchiveCount
	_, err := o.Raw(sql).QueryRows(&archives)
	if err != nil {
		return archives, err
	}
	return archives, nil
}

var AuthFilter = func(ctx *context.Context) {
	user := ctx.Input.Session("admin_user")
	if user == nil {
		if !CheckAuth(ctx.ResponseWriter, ctx) {
			beego.Error(ctx.Input.Url() + " Admin check user failed.")
			return
		}
	} else {
		return
	}
}

func CheckAuth(w http.ResponseWriter, r *context.Context) bool {
	authorization_array := r.Request.Header["Authorization"]
	if len(authorization_array) > 0 {
		authorization := strings.TrimSpace(authorization_array[0])
		var splited []string
		splited = strings.Split(authorization, " ")
		data, err := base64.StdEncoding.DecodeString(splited[1])
		if err != nil {
			Error(r.Input.Url() + " Decode Base64 Auth failed.")
			SetBasicAuth(w)
		}
		auth_info := strings.Split(string(data), ":")
		if auth_info[0] == "admin" && auth_info[1] == "password" {
			r.Output.Session("admin_user", auth_info[0])
			return true
		}
		SetBasicAuth(w)
	} else {
		SetBasicAuth(w)
	}
	return false
}

func SetBasicAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Basic realm=\"Admin console\"")
	http.Error(w, http.StatusText(401), 401)
}

func SendEmailWithAttachments(from, subject string, to []string, message string, attach_file []string) (err error) {
	auth := smtp.PlainAuth(
		"",
		"ox55aa@126.com",
		"63897100",
		"smtp.126.com",
	)
	smtp_host := "smtp.126.com:25"

	if from == "" || subject == "" || message == "" {
		err := errors.New("from or subject or body is empty.")
		Error(err)
		return err
	}
	if len(to) == 0 {
		err := errors.New("to address is empty.")
		Error(err)
		return err
	}

	e := email.NewEmail()
	e.From = from
	e.To = to
	e.Subject = subject
	e.Text = message
	e.HTML = "<h3>" + message + "</h3>"
	for _, file := range attach_file {
		_, err := AttachFile(e, file)
		if err != nil {
			Error(err)
			return err
		}
	}
	err = e.Send(smtp_host, auth)
	if err != nil {
		Error(err)
		return err
	}
	return nil
}

func AttachFile(e *email.Email, filename string) (a *email.Attachment, err error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		beego.Critical(err)
		return nil, err
	}
	f, err := os.Open(filename)
	if err != nil {
		Error(err)
		return nil, err
	}
	_, file := filepath.Split(filename)
	ct := mime.TypeByExtension(filepath.Ext(filename))
	return e.Attach(f, file, ct)
}

func Exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

func FileSize(content interface{}) string {
	value := content.(int64)
	if value < 1000000 {
		return fmt.Sprintf("%.2f KB", float64(value)/1000)
	} else {
		return fmt.Sprintf("%.2f MB", float64(value)/1000/1000)
	}
}

func Debug(args ...interface{}) {
	pc, _, line, _ := runtime.Caller(1)
	function := runtime.FuncForPC(pc)
	funcname := function.Name()
	beego.Debug(funcname, ":", line, args)
}

func Error(args ...interface{}) {
	pc, _, line, _ := runtime.Caller(1)
	function := runtime.FuncForPC(pc)
	funcname := function.Name()
	beego.Error(funcname, ":", line, args)
}
