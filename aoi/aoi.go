package aoi

import (
	"net/http"
)

// HandleFunc 该类型实现了handleFunc
type HandleFunc func(ctx *Context)

//Engine 其中保存各个路径对应的函数映射
type Engine struct {
	router *router
}

func (e *Engine) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	//需要开始配对
	c := newContext(writer, request)
	e.router.handle(c)
}

//New 返回空的Engine对象
func New() *Engine {
	return &Engine{router: new(router)}
}

//addRoute 向Engine Map中添加新的规则
func (e *Engine) addRoute(method, pattern string, handler HandleFunc) {
	e.router.addRoute(method, pattern, handler)
}

// Get 添加Get方法路径，还需要判断路径合法性，暂时没有实现
func (e *Engine) Get(pattern string, handler HandleFunc) {
	e.addRoute("GET", pattern, handler)
}

// Post 添加 Post方法路径，还需要判断路径合法性，暂时没有实现
func (e *Engine) Post(pattern string, handler HandleFunc) {
	e.addRoute("POST", pattern, handler)
}

func (e *Engine) Run(address string) error {
	return http.ListenAndServe(address, e)
}
