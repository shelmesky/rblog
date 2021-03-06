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
	"github.com/shelmesky/rblog/common/api/smtp"
	"github.com/shelmesky/rblog/models"
	"html/template"
	"io"
	"math/rand"
	"mime"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
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
	CatCount     []CategoryCount
	Layout       = "2006-01-02 15:04:05"
)

// 每个Archive的文章数
type ArchiveCount struct {
	Archive string
	Count   int
}

// 每个分类的文章数
type CategoryCount struct {
	Id    int
	Name  string
	Count int
}

// 前一篇和后一篇文章
type PrevNextPage struct {
	Title     string
	Shortname string
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

func Page_not_found(rw http.ResponseWriter, r *http.Request) {
	t, _ := template.New("404.html").ParseFiles("views/404.html")
	t.Execute(rw, nil)
}

func LoadArchivesAndCategory(o orm.Ormer) {
	// insert catagories to map
	var categories []*models.Category
	o.QueryTable(new(models.Category)).All(&categories)

	for _, category := range categories {
		Category_map.Set(category.Id, category.Name)
	}

	// cache the archives count
	ar_count, err := GetArchives()
	if err != nil {
		Error(err)
	}
	ArCount = ar_count

	// cache the categories count
	CatCount, err = GetCategories()
	if err != nil {
		Error(err)
	}
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

type SimplePostInfo struct {
	Title     string
	Shortname string
}

func GetPostInfo(content interface{}) string {
	var ret string
	var Id int
	if Id, ok := content.(int); ok {
		var post_info SimplePostInfo
		sql_raw := `SELECT title, shortname
				FROM post
				WHERE id=%d`
		sql := fmt.Sprintf(sql_raw, Id)
		o := orm.NewOrm()
		err := o.Raw(sql).QueryRow(&post_info)
		if err != nil {
			Error(err)
			return ret
		}
		ret_raw := `<a href="/post/%s.html">%s</a>`
		ret = fmt.Sprintf(ret_raw, post_info.Shortname, post_info.Title)
		return ret
	}
	return string(Id)
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
	var sql = `select distinct archive as archive,count(archive) as count
	from post group by archive`
	o := orm.NewOrm()
	var archives []ArchiveCount
	_, err := o.Raw(sql).QueryRows(&archives)
	if err != nil {
		return archives, err
	}
	return archives, nil
}

func GetCategories() ([]CategoryCount, error) {
	var sql = `SELECT DISTINCT a.category_id AS id, b.name,
		COUNT( a.category_id ) AS count
		FROM post a, category b
		WHERE a.category_id = b.id
		GROUP BY b.id`
	o := orm.NewOrm()
	var categories []CategoryCount
	_, err := o.Raw(sql).QueryRows(&categories)
	if err != nil {
		return categories, err
	}
	return categories, err
}

func GetPrevArticle(created_time string) (prev PrevNextPage, err error) {
	var sql_raw = `
		SELECT title, shortname 
		FROM post
		WHERE created_time < "%s"
		ORDER BY created_time DESC 
		LIMIT 1`
	sql := fmt.Sprintf(sql_raw, created_time)
	o := orm.NewOrm()
	var prev_page PrevNextPage
	err = o.Raw(sql).QueryRow(&prev_page)
	if err != nil {
		return prev_page, err
	}
	return prev_page, err
}

func GetNextArticle(created_time string) (next PrevNextPage, err error) {
	var sql_raw = `
		SELECT title, shortname 
		FROM post
		WHERE created_time > "%s"
		LIMIT 1`
	sql := fmt.Sprintf(sql_raw, created_time)
	o := orm.NewOrm()
	var next_page PrevNextPage
	err = o.Raw(sql).QueryRow(&next_page)
	if err != nil {
		return next_page, err
	}
	return next_page, err
}

var AuthFilter = func(ctx *context.Context) {
	user := ctx.Input.Session("admin_user")
	if user == nil {
		if !CheckAuth(ctx.ResponseWriter, ctx) {
			beego.Error(ctx.Input.URL() + " Admin check user failed.")
			return
		}
	} else {
		return
	}
}

func VerifyAuth(username, password string) bool {
	if username == Site_config.AdminUser && password == Site_config.AdminPassword {
		return true
	}
	return false
}

func CheckAuth(w http.ResponseWriter, r *context.Context) bool {
	authorization_array := r.Request.Header["Authorization"]
	if len(authorization_array) > 0 {
		authorization := strings.TrimSpace(authorization_array[0])
		var splited []string
		splited = strings.Split(authorization, " ")
		data, err := base64.StdEncoding.DecodeString(splited[1])
		if err != nil {
			Error(r.Input.URL() + " Decode Base64 Auth failed.")
			SetBasicAuth(w)
		}
		auth_info := strings.Split(string(data), ":")
		if VerifyAuth(auth_info[0], auth_info[1]) {
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

/*
发送带附件的邮件
utils.SendEmailWithAttachments(
		"ox55aa@126.com",
		"来自126的测试邮件",
		[]string{"33326771@qq.com"},
		"附件列表",
		[]string{"/home/roy/coding/Golang_SourceCode/rblog/src/rblog/测试1.log",
			"/home/roy/coding/Golang_SourceCode/rblog/src/rblog/测试2.log"},
	)
*/
func SendEmailWithAttachments(auth smtp.Auth, smtp_host, from, subject string,
	to []string, message string, attach_file []string) (err error) {

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
