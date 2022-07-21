package aoiweb

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
)

func Recovery() HandleFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				log.Printf("%s\n\n", trace(message))
				c.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()

		c.Next()
	}
}

const DEEP = 32

func trace(message string) any {
	var pcs [DEEP]uintptr
	num := runtime.Callers(3, pcs[:])
	var builder strings.Builder
	builder.WriteString(message + "\n Traceback:")
	for _, pc := range pcs[:num] {
		forPC := runtime.FuncForPC(pc)
		file, line := forPC.FileLine(pc)
		builder.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return builder.String()
}
