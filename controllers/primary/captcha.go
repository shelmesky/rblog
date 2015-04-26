package controllers

import (
	"github.com/astaxie/beego"
	"github.com/shelmesky/rblog/common/api"
	"strconv"
)

type CaptchaController struct {
	beego.Controller
}

// Captcha
func (this *CaptchaController) Get() {
	d := make([]byte, 4)
	s := api.NewLen(4)
	ss := ""
	d = []byte(s)
	for v := range d {
		d[v] %= 10
		ss += strconv.FormatInt(int64(d[v]), 32)
	}
	this.Ctx.ResponseWriter.Header().Set("Content-Type", "image/png")
	api.NewImage(d, 100, 40).WriteTo(this.Ctx.ResponseWriter)
	this.SetSession("captcha", ss)
}
