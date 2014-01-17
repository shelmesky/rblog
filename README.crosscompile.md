## 跨平台编译脚本

### 1. 首先确保GOPATH/GOROOT设置正确

### 2. 执行source ./crosscompile.bash

### 3. 执行后多出以下命令：

* go-crosscompile-build-all
* go-freebsd-386
* go-linux-386
* go-windows-386
* go-build-all
* go-darwin-386
* go-freebsd-amd64
* go-linux-amd64
* go-windows-amd64
* go-crosscompile-build
* go-darwin-amd64
* go-freebsd-arm
* go-linux-arm

### 4. 首先执行go-crosscompile-build-all，为所有平台编译runtime和标准库

### 5. 进入GOPATH，执行go-build-all xxx.go
