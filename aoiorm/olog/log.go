package olog

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
)

var (
	errorLog = log.New(os.Stdout, "\033[31m[error]\033[0m ", log.LstdFlags|log.Lshortfile)
	infoLog  = log.New(os.Stdout, "\033[34m[info]\033[0m ", log.LstdFlags|log.Lshortfile)
	loggers  = []*log.Logger{errorLog, infoLog}
	mu       sync.Mutex
)

var (
	Error  = deepShow
	Errorf = errorLog.Printf
	Info   = infoLog.Println
	Infof  = infoLog.Printf
)

func deepShow(value any) {

	maxCallerDepth := 10
	minCallerDepth := 1
	callers := []string{}
	pcs := make([]uintptr, maxCallerDepth)
	depth := runtime.Callers(minCallerDepth, pcs)
	frames := runtime.CallersFrames(pcs[:depth])
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		s := fmt.Sprintf("%s: %d %s ", frame.File, frame.Line, frame.Function)
		callers = append(callers, s)
		if !more {
			break
		}
	}
	frame := strings.Join(callers, "\n")
	errorLog.Println(value, "\n", frame)
}

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
