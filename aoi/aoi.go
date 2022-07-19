package aoi

import (
	"html/template"
	"net/http"
	"strings"
)

// HandleFunc 该类型实现了handleFunc
type HandleFunc func(ctx *Context)

//Engine 其中保存各个路径对应的函数映射
type Engine struct {
	router       *router
	*RouterGroup                //本身就作为一个routeGroup
	groups       []*RouterGroup //存储所有的分组

	htmlTemplates *template.Template // 添加html模板支持
	funcMap       template.FuncMap   // 模板的渲染支持函数
}

func (e *Engine) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	//需要根据参数判断需要执行的中间件，engine中存了所有的group
	var middlewares []HandleFunc
	for _, group := range e.groups {
		if strings.HasPrefix(request.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	//需要开始配对
	c := newContext(writer, request)
	c.handlers = middlewares
	c.engine = e
	e.router.handle(c)
}

//New 返回空的Engine对象
func New() *Engine {
	e := &Engine{router: newRouter()}
	e.RouterGroup = &RouterGroup{
		engine: e,
	}
	e.groups = make([]*RouterGroup, 0)
	e.groups = append(e.groups, e.RouterGroup)
	return e
}

func (e *Engine) Run(address string) error {
	return http.ListenAndServe(address, e)
}

func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}
