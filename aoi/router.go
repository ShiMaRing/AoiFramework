package aoi

import (
	"log"
	"net/http"
)

type router struct {
	handlers map[string]HandleFunc
}

func newRouter(handlers map[string]HandleFunc) *router {
	return &router{handlers: handlers}
}
func (r *router) addRoute(method string, pattern string, handler HandleFunc) {
	log.Printf("add %s with  method %s", pattern, method)
	key := method + "-" + pattern
	r.handlers[key] = handler
}

func (r *router) handle(c *Context) {
	key := c.Method + "-" + c.Path
	if handler, ok := r.handlers[key]; ok {
		handler(c)
	} else {
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
	}
}
