package uploadcontrollers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/shelmesky/rblog/common/utils"
	"github.com/shelmesky/rblog/models"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

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
	/* 客户端的ajaxuploadfile需要的类型为text/html */
	this.Ctx.Output.Header("Content-Type", "text/html;charset=UTF-8")

	count, _ := this.GetInt64("count")
	file, info, err := this.GetFile("fileToUpload")
	if err != nil {
		utils.Error(err)
		this.Ctx.WriteString(`{"Error": "No selectd file"}`)
		return
	}
	defer file.Close()

	buffer, err := ioutil.ReadAll(file)
	if err != nil {
		utils.Error(err)
		this.Ctx.WriteString(`{"Error": "ReadFull error"}`)
		return
	}

	file_size := int64(len(buffer))

	// if filesize more than 128MB
	if file_size > 134217728 {
		this.Ctx.WriteString(`{"Error": "Filesize more than 128MB"}`)
		return
	}

	filename := info.Filename
	filesize := file_size
	hashname := utils.MakeRandomID()
	//beego.Debug(filename, filesize, hashname, fullpath)

	// 保存文件
	if !utils.Exist("upload") {
		os.Mkdir("upload", 0775)
	}

	// 检查文件是否有后缀名
	file_ext := filepath.Ext(filename)
	if file_ext == "" {
		utils.Error(err)
		this.Ctx.WriteString(`{"Error": "Filename incorrect, need suffix"}`)
		return
	}

	hashname = hashname + file_ext
	tofile := path.Join("upload", hashname)
	f, err := os.OpenFile(tofile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		utils.Error(err)
		this.Ctx.WriteString(`{"Error": "Create file error"}`)
		return
	}
	defer f.Close()

	file_reader := bytes.NewReader(buffer)
	io.Copy(f, file_reader)

	// 增加DB记录
	o := orm.NewOrm()
	upload_file := new(models.UploadFile)
	upload_file.Filesize = filesize
	upload_file.Hashname = hashname
	upload_file.Filename = filename
	_, err = o.Insert(upload_file)
	if err != nil {
		utils.Error(err)
		this.Ctx.WriteString(`{"Error": "Insert database error"}`)
		return
	}

	// 返回文件信息到client
	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "%.1f KB", float64(filesize)/1024.0)
	result := UploadResult{MessageBody{filename, buf.String(), count}, ""}
	b, err := json.Marshal(&result)
	if err != nil {
		utils.Error(err)
	}

	this.Ctx.WriteString(string(b))

}

type DownloadController struct {
	beego.Controller
}

// file download handler
func (this *DownloadController) Get() {
	// 根据hash从DB中查询文件上传记录
	hashname := this.Ctx.Input.Param(":filehash")
	var upload_file models.UploadFile
	o := orm.NewOrm()
	err := o.QueryTable(new(models.UploadFile)).Filter("Hashname", hashname).One(&upload_file)
	if err == orm.ErrNoRows {
		utils.Error(err)
		this.Ctx.WriteString(`{"Error": "Can not find file"}`)
		return
	}

	// 获取文件完整路径
	current_path, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		utils.Error(err)
		this.Ctx.WriteString(`{"Error": "Get current dir error"}`)
		return
	}

	// 判断文件是否存在
	fullpath := path.Join(current_path, "upload", upload_file.Hashname)
	if !utils.Exist(fullpath) {
		utils.Error(fullpath, "not exist")
		this.Ctx.WriteString(`{"Error": "File not exist"}`)
		return
	}

	// 开始下载

	this.Ctx.Output.Header("Content-Description", "File Transfer")
	this.Ctx.Output.Header("Content-Disposition", "attachment; filename="+upload_file.Filename)
	this.Ctx.Output.Header("Content-Transfer-Encoding", "binary")
	this.Ctx.Output.Header("Expires", "0")
	this.Ctx.Output.Header("Cache-Control", "must-revalidate")
	this.Ctx.Output.Header("Pragma", "public")

	ctype := mime.TypeByExtension(filepath.Ext(upload_file.Filename))
	this.Ctx.Output.Header("Content-Type", ctype)

	http.ServeFile(this.Ctx.Output.Context.ResponseWriter,
		this.Ctx.Output.Context.Request,
		fullpath)
}
