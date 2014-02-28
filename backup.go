package main

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/smtp"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"rblog/common/utils"
	"time"
)

type SMTPAuth struct {
	Host     string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
	FromHost string `json:"from_host"`
}

type GlobalConfig struct {
	SMTPAuth       SMTPAuth `json:"smtp_auth"`
	MailFrom       string   `json:"mail_from"`
	MailSendto     string   `json:"mail_sendto"`
	MySQLHost      string   `json:"mysql_host"`
	MySQLPort      string   `json:"mysql_port"`
	MySQLUsername  string   `json:"mysql_username"`
	MySQLPassword  string   `json:"mysql_password"`
	MySQLDatabases string   `json:"mysql_databases"`
	BackupDir      string   `json:"backup_dir"`
	UploadDir      string   `json:"upload_dir"`
	LogFile        string   `json:"log_file"`
}

const backup_config_file = "./backup.config.json"

var (
	config  GlobalConfig
	logfile *os.File
)

func init() {
	// parse  JSON config file
	data, err := ioutil.ReadFile(backup_config_file)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatal(err)
	}

	// init log
	logfile, err = os.OpenFile(config.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("error opening file: %v", err)
	}
	log.SetOutput(logfile)
	log.Println("Backup Start")
}

/*
	备份MySQL数据库到文件
	@username MySQL用户名
	@password MySQL密码
	@username MySQL数据库
	备份多个数据库可以用空格隔开，如："db1 db2 db3"
	返回备份的内容
*/
func DumpMysql(username, password, database string) ([]byte, error) {
	var ret []byte
	var mysqldump string

	// 在$PATH中寻找mysqldump命令
	// 没有找到则默认使用/usr/bin/mysqldump
	mysqldump, err := exec.LookPath("mysqldump")
	if err != nil {
		log.Println("Can not find mysqldump in $PATH, we will use /usr/bin/dump.")
		mysqldump = "/usr/bin/mysqldump"
	}

	username = fmt.Sprintf("-u%s", username)
	password = fmt.Sprintf("-p%s", password)

	// 执行备份命令
	cmd := exec.Command(mysqldump, username, password, "-Y", "--default-character-set=utf8", database)

	var std_out bytes.Buffer
	var std_err bytes.Buffer
	cmd.Stdout = &std_out
	cmd.Stderr = &std_err

	fmt.Printf("Dump MySQL DB: %s\n", database)

	// 成功返回标准输出中的内容
	// 失败返回标准错误中的内容
	err = cmd.Start()
	if err != nil {
		return std_err.Bytes(), err
	}
	err = cmd.Wait()
	if err != nil {
		return std_err.Bytes(), err
	}
	ret = std_out.Bytes()

	return ret, err
}

/*
	写文件函数
*/
func WriteBackupFile(data []byte, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		log.Println("Can file %s failed: ", filename)
		return err
	}
	n, err := file.Write(data)
	if err != nil {
		log.Println("Write file %s failed: ", filename)
		return err
	}
	if n != len(data) {
		log.Println("Write file %s failed: ", filename)
		return errors.New(fmt.Sprintf("Only write %d byte, less than %d.", n, len(data)))
	}

	return nil
}

/*
	将备份目录打包为tar文件
*/
func CompressDir(src_dirname, dst_filename string, ifoverride bool) error {
	err := Tar(src_dirname, dst_filename, ifoverride)
	if err != nil {
		return err
	}
	return nil
}

/*
	将打包后的文件发送到制定EMAIL地址
*/
func SendFileEmail(auth smtp.Auth, smtp_host, from, subject string,
	to []string, message string, files []string) error {

	err := utils.SendEmailWithAttachments(
		auth, smtp_host, from, subject, to, message, files,
	)

	if err != nil {
		return err
	}
	return nil
}

func main() {
	defer logfile.Close()

	const LayoutFile = "20060102150405"
	const LayoutMail = "2006-01-02 15:04:05"
	now_time := time.Now().Format(LayoutFile)
	now_time_mail := time.Now().Format(LayoutMail)

	username := config.MySQLUsername
	password := config.MySQLPassword
	database := config.MySQLDatabases

	backup_dir := config.BackupDir

	// 备份MySQL数据库
	data, err := DumpMysql(username, password, database)
	if err != nil {
		log.Println(err)
		return
	}

	date_backup_dir := filepath.Join(backup_dir, now_time)

	if !DirExists(date_backup_dir) {
		err := os.MkdirAll(date_backup_dir, 0775)
		if err != nil {
			log.Println(err, date_backup_dir)
			return
		}
	}

	// 写入备份文件
	bak_file := database + "." + now_time + ".sql"
	full_path := filepath.Join(date_backup_dir, bak_file)
	if err = WriteBackupFile(data, full_path); err != nil {
		log.Println(err)
		return
	}

	// 打包数据库备份目录
	compress_file := filepath.Join(backup_dir, now_time+".db.tar")
	err = CompressDir(date_backup_dir, compress_file, true)
	if err != nil {
		log.Println(err)
		return
	}

	// 打包上传文件目录
	upload_dir := config.UploadDir
	upload_dir_dst := filepath.Join(backup_dir, now_time+".upload.tar")
	err = CompressDir(upload_dir, upload_dir_dst, true)
	if err != nil {
		log.Println(err)
		return
	}

	// 发送邮件
	auth := smtp.PlainAuth(
		"",
		config.SMTPAuth.Username,
		config.SMTPAuth.Password,
		config.SMTPAuth.FromHost,
	)
	smtp_host := config.SMTPAuth.Host

	from := config.MailFrom
	to := []string{config.MailSendto}
	subject := fmt.Sprintf("Data Backup %s", now_time_mail)
	message := subject
	attachments := []string{compress_file, upload_dir_dst}

	err = SendFileEmail(auth, smtp_host, from, subject, to, message, attachments)
	if err != nil {
		log.Println(err)
	}

	log.Println("Backup End")
}

/*
	将文件或目录打包成 .tar 文件
	src 是要打包的文件或目录的路径
	dstTar 是要生成的 .tar 文件的路径
	failIfExist 标记如果 dstTar 文件存在，是否放弃打包，如果否，则会覆盖已存在的文件
*/
func Tar(src string, dstTar string, failIfExist bool) (err error) {
	// 清理路径字符串
	src = path.Clean(src)

	// 判断要打包的文件或目录是否存在
	if !Exists(src) {
		return errors.New("要打包的文件或目录不存在：" + src)
	}

	// 判断目标文件是否存在
	if FileExists(dstTar) {
		if failIfExist { // 不覆盖已存在的文件
			return errors.New("目标文件已经存在：" + dstTar)
		} else { // 覆盖已存在的文件
			if er := os.Remove(dstTar); er != nil {
				return er
			}
		}
	}

	// 创建空的目标文件
	fw, er := os.Create(dstTar)
	if er != nil {
		return er
	}
	defer fw.Close()

	// 创建 tar.Writer，执行打包操作
	tw := tar.NewWriter(fw)
	defer func() {
		// 这里要判断 tw 是否关闭成功，如果关闭失败，则 .tar 文件可能不完整
		if er := tw.Close(); er != nil {
			err = er
		}
	}()

	// 获取文件或目录信息
	fi, er := os.Stat(src)
	if er != nil {
		return er
	}

	// 获取要打包的文件或目录的所在位置和名称
	srcBase, srcRelative := path.Split(path.Clean(src))

	// 开始打包
	if fi.IsDir() {
		tarDir(srcBase, srcRelative, tw, fi)
	} else {
		tarFile(srcBase, srcRelative, tw, fi)
	}

	return nil
}

// 因为要执行遍历操作，所以要单独创建一个函数
func tarDir(srcBase, srcRelative string, tw *tar.Writer, fi os.FileInfo) (err error) {
	// 获取完整路径
	srcFull := srcBase + srcRelative

	// 在结尾添加 "/"
	last := len(srcRelative) - 1
	if srcRelative[last] != os.PathSeparator {
		srcRelative += string(os.PathSeparator)
	}

	// 获取 srcFull 下的文件或子目录列表
	fis, er := ioutil.ReadDir(srcFull)
	if er != nil {
		return er
	}

	// 开始遍历
	for _, fi := range fis {
		if fi.IsDir() {
			tarDir(srcBase, srcRelative+fi.Name(), tw, fi)
		} else {
			tarFile(srcBase, srcRelative+fi.Name(), tw, fi)
		}
	}

	// 写入目录信息
	if len(srcRelative) > 0 {
		hdr, er := tar.FileInfoHeader(fi, "")
		if er != nil {
			return er
		}
		hdr.Name = srcRelative

		if er = tw.WriteHeader(hdr); er != nil {
			return er
		}
	}

	return nil
}

// 因为要在 defer 中关闭文件，所以要单独创建一个函数
func tarFile(srcBase, srcRelative string, tw *tar.Writer, fi os.FileInfo) (err error) {
	// 获取完整路径
	srcFull := srcBase + srcRelative

	// 写入文件信息
	hdr, er := tar.FileInfoHeader(fi, "")
	if er != nil {
		return er
	}
	hdr.Name = srcRelative

	if er = tw.WriteHeader(hdr); er != nil {
		return er
	}

	// 打开要打包的文件，准备读取
	fr, er := os.Open(srcFull)
	if er != nil {
		return er
	}
	defer fr.Close()

	// 将文件数据写入 tw 中
	if _, er = io.Copy(tw, fr); er != nil {
		return er
	}

	return nil
}

// 判断档案是否存在
func Exists(name string) bool {
	_, err := os.Stat(name)
	return err == nil || os.IsExist(err)
}

// 判断文件是否存在
func FileExists(filename string) bool {
	fi, err := os.Stat(filename)
	return (err == nil || os.IsExist(err)) && !fi.IsDir()
}

// 判断目录是否存在
func DirExists(dirname string) bool {
	fi, err := os.Stat(dirname)
	return (err == nil || os.IsExist(err)) && fi.IsDir()
}
