#!/bin/bash

case $1 in
	windows_386)
		CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -o server_windows_x86.exe main.go
		;;

	windows_amd64)
		CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o server_windows_x64.exe main.go
		;;

	linux_386)
		CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -o server_linux_x86 main.go
		;;

	linux_amd64)
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server_linux_x64 main.go
		;;

	freebsd_386)
		CGO_ENABLED=0 GOOS=freebsd GOARCH=386 go build -o server_freebsd_x86 main.go
		;;

	freebsd_amd64)
		CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -o server_freebsd_x64 main.go
		;;

	darwin_386)
		CGO_ENABLED=0 GOOS=darwin GOARCH=386 go build -o server_darwin_x86 main.go
		;;

	darwin_amd64)
		CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o server_darwin_x64 main.go
		;;

	*)
		echo "$0 [windows_386 windows_amd64 linux_386 linux_amd64 freebsd_386 freebsd_amd64 darwin_386 darwin_amd64]"
		;;
esac
