package gee

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
)

//print stack trace for debug
func trace(message string) string {
	var pcs [32]uintptr
	// Callers 用来返回调用栈的程序计数器 
	// 第0个Caller是Callers本身，第1个是上一层trace，第2个是再上一层的defer func()
	n := runtime.Callers(3, pcs[:]) // skip first 3 caller

	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}

func Recovery() HandleFunc {
	return func(ctx *Context) {
		defer func () {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				log.Printf("%s\n\n", trace(message))
				ctx.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()

		ctx.Next()
	}
}