package mhttp

import (
	"encoding/json"
	"fmt"
	"net/http"
)

/*  ====== HttpEngine ======  */
type HandleFun func(*Context)
type HttpEngine struct{ route *route }

// 实例化Engine类
func New() *HttpEngine {
	return &HttpEngine{route: newRoute()}
}

// 注册请求方法,路由和处理函数
func (e *HttpEngine) registerRoute(method string, pattern string, handler HandleFun) {
	e.route.registerRoute(method, pattern, handler)
}

// GET请求处理函数
func (e *HttpEngine) Get(pattern string, handler HandleFun) {
	e.registerRoute("GET", pattern, handler)
}

// POST请求处理函数
func (e *HttpEngine) Post(pattern string, handler HandleFun) {
	e.registerRoute("POST", pattern, handler)
}

// 启动服务
func (e *HttpEngine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, e)
}

// 实现ServeHTTP接口
func (e *HttpEngine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := newContext(w, req)
	e.route.handle(c)
}

/*  ======== Context=======  */
type Context struct {
	Writer     http.ResponseWriter
	Request    *http.Request
	Path       string
	Method     string
	StatusCode int
}

// json返回类型
type H map[string]interface{}

// 创建上下文实例
func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer:  w,
		Request: r,
		Path:    r.URL.Path,
		Method:  r.Method,
	}
}

func (ctx *Context) PostForm(key string) string {
	return ctx.Request.FormValue(key)
}

func (ctx *Context) Query(key string) string {
	return ctx.Request.URL.Query().Get(key)
}

func (ctx *Context) Status(code int) {
	ctx.StatusCode = code
	ctx.Writer.WriteHeader(code)
}

func (ctx *Context) SetHeader(key, value string) {
	ctx.Writer.Header().Set(key, value)
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
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

/*  ========  Route  =======  */
type route struct{ handles map[string]HandleFun }

func newRoute() *route {
	return &route{handles: make(map[string]HandleFun)}
}

func (r *route) registerRoute(method string, pattern string, handler HandleFun) {
	key := method + "-" + pattern
	r.handles[key] = handler
}

func (r *route) handle(ctx *Context) {
	key := ctx.Method + "-" + ctx.Path
	if handler, ok := r.handles[key]; ok {
		handler(ctx)
	} else {
		ctx.String(http.StatusNotFound, "404 NOT FOUND: %s\n", ctx.Path)
	}
}