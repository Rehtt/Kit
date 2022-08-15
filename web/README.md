一个简单的web api框架

支持中间件，路由编写更友好

```go
package main

import (
	"fmt"
	goweb "github.com/Rehtt/Kit/web"
	"net/http"
)

func main() {
	web := goweb.New()
	web.SetValue("test","123")
	web.Middleware(func(ctx *goweb.Context) {
		fmt.Println("中间件")
	})
	web.NoRoute(func(ctx *goweb.Context) {
		ctx.Writer.Write([]byte("找不到啊大佬"))
	})

	web.Any("/123/#asd/234", func(ctx *goweb.Context) {
		fmt.Println(ctx.GetParam("asd"), "获取动态路由参数")
	})
	api := web.Grep("/api")
	api.GET("/test", func(ctx *goweb.Context) {
		fmt.Println(ctx.GetValue("test"))
	})

	http.ListenAndServe(":9090", web)
}

```
