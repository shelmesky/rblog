#!/bin/bash

case $1 in
	windows_386)
		CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -o server_windows_x86.exe
		;;

	windows_amd64)
		CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o server_windows_x64.exe
		;;

	linux_386)
		CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -o server_linux_x86
		;;

	linux_amd64)
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server_linux_x64
		;;

	freebsd_386)
		CGO_ENABLED=0 GOOS=freebsd GOARCH=386 go build -o server_freebsd_x86
		;;

	freebsd_amd64)
		CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -o server_freebsd_x64
		;;

	darwin_386)
		CGO_ENABLED=0 GOOS=darwin GOARCH=386 go build -o server_darwin_x86
		;;

	darwin_amd64)
		CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o server_darwin_x64
		;;

	*)
		echo "$0 [windows_386 windows_amd64 linux_386 linux_amd64 freebsd_386 freebsd_amd64 darwin_386 darwin_amd64]"
		;;
esac
