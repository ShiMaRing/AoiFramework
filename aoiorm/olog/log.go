package olog

import (
	"io"
	"log"
	"os"
	"sync"
)

var (
	errorLog = log.New(os.Stdout, "\033[31m[error]\033[0m ", log.LstdFlags|log.Lshortfile)
	infoLog  = log.New(os.Stdout, "\033[34m[info]\033[0m ", log.LstdFlags|log.Lshortfile)
	loggers  = []*log.Logger{errorLog, infoLog}
	mu       sync.Mutex
)

var (
	Error  = errorLog.Println
	Errorf = errorLog.Printf
	Info   = infoLog.Println
	Infof  = infoLog.Printf
)

//设置日志等级，会将部分日志的输出重定向
const (
	InfoLevel  = iota //通知
	ErrorLevel        //错误
	Disabled          //禁止
)

func SetLevel(level int) {
	//加锁防止竞态
	for _, logger := range loggers {
		logger.SetOutput(os.Stdout)
	}
	if InfoLevel < level {
		infoLog.SetOutput(io.Discard)
	}
	if ErrorLevel < level {
		errorLog.SetOutput(io.Discard)
	}
}
