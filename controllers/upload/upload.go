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
	file, info, err := this.GetFile("fileToUpload")
	if err != nil {
		beego.Critical(err)
	}
	file_size := file.(Sizer).Size()
	beego.Info(file, info.Filename, file_size, info.Header)
	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "%.1f KB", float64(file_size)/1024.0)
	result := UploadResult{MessageBody{info.Filename, buf.String()}, ""}
	b, err := json.Marshal(&result)
	if err != nil {
		beego.Critical(err)
	}
	this.Ctx.WriteString(string(b))
}
