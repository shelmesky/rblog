package utils

import (
	"time"
	"crypto/md5"
	"math/rand"
	"strconv"
	"io"
	"fmt"
	"reflect"
	"runtime"
)


func MakeRandomID()(string) {
	nano := time.Now().UnixNano()
	rand.Seed(nano)
	rndNum := rand.Int63()
	
	md5_nano := MD5(strconv.FormatInt(nano, 10))
	md5_rand := MD5(strconv.FormatInt(rndNum, 10))
	RandomID := MD5(md5_nano + md5_rand)
	return RandomID
}

func MD5(text string) string {
	hashMD5 := md5.New()
	io.WriteString(hashMD5, text)
	return fmt.Sprintf("%x", hashMD5.Sum(nil))
}


func GetFuncName(function interface{}) string {
	func_pointer := reflect.ValueOf(function).Pointer()
	return runtime.FuncForPC(func_pointer).Name()
}


type NewTime struct {
    time.Time
}

func (t NewTime)YearMonthString() string {
    const layout = "2006-01"
    return t.Format(layout)
}

func (t NewTime)NowString() string {
    const layout = "2006-01-02 15:04:05"
    return t.Format(layout)
}

func YearMonth() string {
    ta := time.Now()
    t := NewTime{ta}
    return t.YearMonthString()
}

func Now() string {
    ta := time.Now()
    t := NewTime{ta}
    return t.NowString()
}


