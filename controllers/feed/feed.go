package feedcontrollers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/gorilla/feeds"
	"rblog/common/utils"
	"rblog/models"
	"time"
)

type RssController struct {
	beego.Controller
}

func (this *RssController) Get() {
	o := orm.NewOrm()
	var posts []*models.Post
	_, err := o.QueryTable(new(models.Post)).Limit(20).All(&posts)
	if err != nil {
		beego.Error(err)
	}

	now := time.Now()
	feed := &feeds.Feed{
		Title:       utils.Site_config.BlogName,
		Link:        &feeds.Link{Href: utils.Site_config.BlogUrl},
		Description: "编程/生活/思考",
		Author:      &feeds.Author{"Roy Lieu", "roy@rootk.com"},
		Created:     now,
	}

	for _, post := range posts {
		item := &feeds.Item{
			Title:   post.Title,
			Link:    &feeds.Link{Href: utils.Site_config.BlogUrl + "/post/" + post.Shortname + ".html"},
			Author:  &feeds.Author{"Roy Lieu", "roy@rootk.com"},
			Created: post.CreatedTime,
		}
		feed.Add(item)
	}

	rss, err := feed.ToAtom()
	if err != nil {
		beego.Error(err)
	}

	this.Ctx.WriteString(rss)
}

type AtomController struct {
	beego.Controller
}

func (this *AtomController) Get() {

}
