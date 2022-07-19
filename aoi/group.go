package aoi

import (
	"net/http"
	"path"
)

// RouterGroup 提供分组功能
type RouterGroup struct {
	prefix      string       //分组前缀
	middlewares []HandleFunc //提供中间件功能
	parent      *RouterGroup //支持嵌套分组
	engine      *Engine      //使用engine结构体的各个方法
}

//addRoute 向Engine Map中添加新的规则
func (group *RouterGroup) addRoute(method, pre string, handler HandleFunc) {
	pattern := group.prefix + pre
	group.engine.router.addRoute(method, pattern, handler)
}

// Get 添加Get方法路径，还需要判断路径合法性，暂时没有实现
func (group *RouterGroup) Get(pattern string, handler HandleFunc) {
	group.addRoute("GET", pattern, handler)
}

// Post 添加 Post方法路径，还需要判断路径合法性，暂时没有实现
func (group *RouterGroup) Post(pattern string, handler HandleFunc) {
	group.addRoute("POST", pattern, handler)
}

// Group 传入前缀返回一个分组,当前分组前缀由创建它的分组前缀与当前传入参数拼接取得
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine //获取当前的engine对象
	g := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	engine.groups = append(engine.groups, g)
	return g
}

// Use 指定的group使用中间件
func (group *RouterGroup) Use(middlewares ...HandleFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

//创建静态文件处理器，参数为相对路径以及文件系统
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandleFunc {
	absolutePath := path.Join(group.prefix, relativePath) //拼接相对路径以及group的路径获得绝对路径
	server := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		// Check if file exists and/or if we have permission to access it
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		server.ServeHTTP(c.Writer, c.Request)
	}
}

func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	join := path.Join(relativePath, "/*filePath")
	group.Get(join, handler)
}
