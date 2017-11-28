package feedcontrollers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/gorilla/feeds"
	"github.com/shelmesky/rblog/common/utils"
	"github.com/shelmesky/rblog/models"
	"time"
)

const ns = "http://www.w3.org/2005/Atom"

type RssController struct {
	beego.Controller
}

func (this *RssController) Get() {
	o := orm.NewOrm()
	var posts []*models.Post
	_, err := o.QueryTable(new(models.Post)).Limit(20).OrderBy("-CreatedTime").All(&posts)
	if err != nil {
		utils.Error(err)
	}

	feed := &feeds.AtomFeed{
		Xmlns:    ns,
		Title:    utils.Site_config.BlogName,
		Link:     &feeds.AtomLink{Href: utils.Site_config.BlogUrl},
		Subtitle: "编程/生活/思考",
		Author:   &feeds.AtomAuthor{AtomPerson: feeds.AtomPerson{Name: "Roy Lieu", Email: "roy@rootk.com"}},
		Updated:  utils.Now(),
	}

	for _, post := range posts {
		body := post.Body
		body = utils.RenderMarkdown(body)
		links := make([]feeds.AtomLink, 0)
		link := feeds.AtomLink{Rel: "alternate", Href: utils.Site_config.BlogUrl + "/post/" + post.Shortname + ".html"}
		links = append(links, link)
		item := &feeds.AtomEntry{
			Title:     post.Title,
			Content:   &feeds.AtomContent{Type: "text/html", Content: body},
			Links:     links,
			Updated:   post.CreatedTime.Format(time.RFC3339),
			Published: post.CreatedTime.Format(time.RFC3339),
		}
		feed.Entries = append(feed.Entries, item)
	}

	xmlstr, err := feeds.ToXML(feed)
	if err != nil {
		utils.Error(err)
	}

	this.Ctx.WriteString(xmlstr)
}
