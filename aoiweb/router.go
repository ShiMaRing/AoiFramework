package aoiweb

import (
	"net/http"
	"strings"
)

type router struct {
	roots    map[string]*node      //存储各个请求方式的的树根节点
	handlers map[string]HandleFunc //存储每种请求方式的处理函数
}

//newRouter 创建新路由
func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandleFunc),
	}
}

//parsePattern 解析路径参数，排除非法参数，*必须要在最后的位置并且只能出现一次
//函数会进行修复
func parsePattern(pattern string) []string {
	parts := strings.Split(pattern, "/")
	result := make([]string, 0)
	for i := 0; i < len(parts); i++ {
		part := parts[i]
		if part != "" { //避免出现空字符串
			result = append(result, part)
			if strings.Contains(part, "*") {
				break
			}
		}
	}
	return result
}

//添加路由规则，支持 ：以及* 通配符
func (r *router) addRoute(method string, pattern string, handler HandleFunc) {
	parts := parsePattern(pattern)
	key := method + "-" + pattern
	//检查该方法是否又节点存在
	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = &node{}
	}
	r.roots[method].insert(pattern, parts, 0)
	r.handlers[key] = handler
}

//真正处理请求的方法
func (r *router) handle(c *Context) {
	route, m := r.getRoute(c.Method, c.Path)
	//说明有参数能够进行处理
	if route != nil {
		c.Params = m
		key := c.Method + "-" + route.pattern //获取对应路由的key
		handleFunc := r.handlers[key]
		c.handlers = append(c.handlers, handleFunc)
	} else {
		c.String(http.StatusNotFound, "404 NOT FOUND %s \n", c.Path)
	}
	c.Next()
}

func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	//输入参数，请求方法，请求路径，输出参数为对应的节点以及对应的通配符匹配值
	parts := parsePattern(path)
	root, ok := r.roots[method]

	params := make(map[string]string) //用来返回路径参数映射

	//说明该方法没有被添加入节点
	if !ok {
		return nil, nil
	}
	n := root.search(parts, 0) //传入parts切片搜索对应的节点
	if n != nil {
		pathParts := parsePattern(n.pattern) //路径的各个参数，需要进行映射

		for index, part := range pathParts {
			if part[0] == ':' { //双方的长度应当为相等，除非出现*号标识
				params[part[1:]] = parts[index]
			}
			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(parts[index:], "/")
				break
			}
		}
	}
	return n, params
}
