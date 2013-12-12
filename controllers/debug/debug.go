package debugcontrollers

import (
	"encoding/base64"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"net/http"
	"net/http/pprof"
	"strings"
)

// HTTP Server Performance profile
type ProfController struct {
	beego.Controller
}

func (this *ProfController) CheckAuth(w http.ResponseWriter, r *context.Context) bool {
	authorization_array := r.Request.Header["Authorization"]
	if len(authorization_array) > 0 {
		authorization := strings.TrimSpace(authorization_array[0])
		var splited []string
		splited = strings.Split(authorization, " ")
		data, err := base64.StdEncoding.DecodeString(splited[1])
		if err != nil {
			this.SetBasicAuth(w)
		}
		auth_info := strings.Split(string(data), ":")
		if auth_info[0] == "admin" && auth_info[1] == "password" {
			return true
		}
		this.SetBasicAuth(w)
	} else {
		this.SetBasicAuth(w)
	}
	return false
}

func (this *ProfController) SetBasicAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Basic realm=\"Performace Profile\"")
	http.Error(w, http.StatusText(401), 401)
}

func (this *ProfController) Get() {
	if !this.CheckAuth(this.Ctx.ResponseWriter, this.Ctx) {
		beego.Error("Pprof check user failed.")
		return
	}

	switch this.Ctx.Input.Params(":pp") {
	default:
		pprof.Index(this.Ctx.ResponseWriter, this.Ctx.Request)
	case "":
		pprof.Index(this.Ctx.ResponseWriter, this.Ctx.Request)
	case "cmdline":
		pprof.Cmdline(this.Ctx.ResponseWriter, this.Ctx.Request)
	case "profile":
		pprof.Profile(this.Ctx.ResponseWriter, this.Ctx.Request)
	case "symbol":
		pprof.Symbol(this.Ctx.ResponseWriter, this.Ctx.Request)
	}
	this.Ctx.ResponseWriter.WriteHeader(200)
}
