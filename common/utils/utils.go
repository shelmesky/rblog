package utils

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/cache"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/utils"
	"github.com/russross/blackfriday"
	"html/template"
	"io"
	"math/rand"
	"rblog/models"
	"reflect"
	"runtime"
	"strconv"
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
				beego.Error(err)
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
