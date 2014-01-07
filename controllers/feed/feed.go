package feedcontrollers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	. "github.com/gorilla/feeds"
	"rblog/common/utils"
	"rblog/models"
	"time"
)

const ns = "http://www.w3.org/2005/Atom"

type RssController struct {
	beego.Controller
}

func (this *RssController) Get() {
	o := orm.NewOrm()
	var posts []*models.Post
	_, err := o.QueryTable(new(models.Post)).Limit(20).OrderBy("-id").All(&posts)
	if err != nil {
		beego.Error(err)
	}

	feed := &AtomFeed{
		Xmlns:    ns,
		Title:    utils.Site_config.BlogName,
		Link:     &AtomLink{Href: utils.Site_config.BlogUrl},
		Subtitle: "编程/生活/思考",
		Author:   &AtomAuthor{AtomPerson: AtomPerson{Name: "Roy Lieu", Email: "roy@rootk.com"}},
		Updated:  utils.Now(),
	}

	for _, post := range posts {
		body := post.Body
		body = utils.RenderMarkdown(body)
		item := &AtomEntry{
			Title:     post.Title,
			Content:   &AtomContent{Type: "text/html", Content: body},
			Link:      &AtomLink{Rel: "alternate", Href: utils.Site_config.BlogUrl + "/post/" + post.Shortname + ".html"},
			Updated:   post.CreatedTime.Format(time.RFC3339),
			Published: post.CreatedTime.Format(time.RFC3339),
		}
		feed.Entries = append(feed.Entries, item)
	}

	xmlstr, err := ToXML(feed)
	if err != nil {
		beego.Error(err)
	}

	this.Ctx.WriteString(xmlstr)
}
