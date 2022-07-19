package aoi

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// H 提供输出方法
type H map[string]interface{}

//Context 上下文，封装各类常用方法
type Context struct {
	//响应与请求
	Writer  http.ResponseWriter
	Request *http.Request

	//请求信息
	Path   string
	Method string

	//存放请求参数信息
	Params map[string]string

	//响应码
	StatusCode int

	// 中间件相关参数
	handlers []HandleFunc
	index    int
}

//newContext 创建并返回对应的上下文
func newContext(writer http.ResponseWriter, request *http.Request) *Context {
	return &Context{
		Writer:  writer,
		Request: request,
		Path:    request.URL.Path,
		Method:  request.Method,
		index:   -1,
	}
}

// GetFormValue 从表单各处获取键值对
func (c *Context) GetFormValue(key string) string {
	return c.Request.FormValue(key)
}

//Query 从query路径中获取参数
func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

//Status 设置http响应码
func (c *Context) Status(code int) {
	c.StatusCode = code
	//没有显式声明则会默认发送200
	c.Writer.WriteHeader(code)
}

//SetHeader 设置http响应头信息
func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

//String 返回format格式的响应信息
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) JSON(code int, obj interface{}) {
	c.Status(code)
	c.SetHeader("Content-Type", "application/json")
	encoder := json.NewEncoder(c.Writer)
	err := encoder.Encode(obj)
	if err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}
func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}

func (c *Context) Param(key string) string {
	s := c.Params[key]
	return s
}

// Next 不断遍历交给下一个处理函数
func (c *Context) Next() {
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}
