一个纯原生简单的web api框架

支持中间件，路由编写更友好

```shell
go get github.com/Rehtt/Kit/web
```

使用jsoniter:

```shell
go build -tags=jsoniter
```

```go
package main

import (
	"fmt"
	"github.com/Rehtt/Kit/web"
	"net/http"
)

func main() {
	web := web.New()
	web.SetValue("test","123")
	web.HeadMiddleware(func(ctx *web.Context) {
		fmt.Println("中间件")
	})
	web.NoRoute(func(ctx *web.Context) {
		ctx.Writer.Write([]byte("找不到啊大佬"))
	})

	web.Any("/123/#asd/234", func(ctx *web.Context) {
		fmt.Println(ctx.GetUrlPathParam("asd"), "获取动态路由参数")
	})
    // curl 127.0.0.1:9090/123/zxcv/234
    // print: zxcv 获取动态路由参数

    web.Any("/1234/#...",func(ctx *web.Context){
        fmt.Println(ctx.GetUrlPathParam("#"), "获取参数")
    })
    // curl 127.0.0.1:9090/1234/qwe/asd/sdf
    // print: qwe/asd/sdf 获取参数

	api := web.Grep("/api")
	api.GET("/test", func(ctx *web.Context) {
		fmt.Println(ctx.GetContextValue("test"))
	})

    // /#... 最后匹配
    web.GET("/#...",func(ctx *web.Context){
        fmt.Println(ctx.GetUrlPathParam("#"))
    })
    // curl 127.0.0.1:9090/asd/asd
    // print: asd/asd

	http.ListenAndServe(":9090", web)
}

```

并发测试:

```go
g := web.New()
g.GET("/ping", func(ctx *web.Context) {
	ctx.Writer.Write([]byte("pong"))
})
http.ListenAndServe(":8070", g)
```

```shell
$ wrk -d 100s -c 1024 -t 8 http://127.0.0.1:8070/ping
Running 2m test @ http://127.0.0.1:8070/ping
  8 threads and 1024 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     4.30ms    5.17ms  92.06ms   86.25%
    Req/Sec    42.37k     7.90k  130.44k    69.17%
  33674619 requests in 1.67m, 3.76GB read
  Socket errors: connect 0, read 0, write 0, timeout 38
Requests/sec: 336435.08
Transfer/sec:     38.50MB

```

gin:

```go
g := gin.New()
g.GET("/ping", func(context *gin.Context) {
	context.Writer.Write([]byte("pong"))
})
http.ListenAndServe(":8060", g)
```

```shell
wrk -d 100s -c 1024 -t 8 http://127.0.0.1:8060/ping
Running 2m test @ http://127.0.0.1:8060/ping
  8 threads and 1024 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     4.43ms    5.99ms 224.24ms   87.84%
    Req/Sec    43.33k     9.81k  112.97k    71.84%
  34451839 requests in 1.67m, 3.85GB read
  Socket errors: connect 0, read 0, write 0, timeout 100
Requests/sec: 344178.03
Transfer/sec:     39.39MB
```
