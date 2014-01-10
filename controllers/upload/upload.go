package uploadcontrollers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
)

// 获取文件大小
type Sizer interface {
	Size() int64
}

type MessageBody struct {
	Filename string
	Filesize string
	Count    int64
}

type UploadResult struct {
	Message MessageBody
	Error   string
}

type UploadController struct {
	beego.Controller
}

// file upload handler
func (this *UploadController) Post() {
	count, _ := this.GetInt("count")
	file, info, err := this.GetFile("fileToUpload")
	if err != nil {
		beego.Critical(err)
		return
	}
	file_size := file.(Sizer).Size()
	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "%.1f KB", float64(file_size)/1024.0)
	result := UploadResult{MessageBody{info.Filename, buf.String(), count}, ""}
	b, err := json.Marshal(&result)
	if err != nil {
		beego.Critical(err)
	}
	/* 客户端的ajaxuploadfile需要的类型为text/html */
	this.Ctx.Output.Header("Content-Type", "text/html;charset=UTF-8")
	this.Ctx.WriteString(string(b))

}
