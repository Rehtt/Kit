package web

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkRequestOptimization(b *testing.B) {
	for _, cancelable := range []bool{false, true} {
		for _, count := range []int{0, 5, 20} {
			for _, kind := range []string{"static", "param", "wildcard", "404", "405"} {
				b.Run(fmt.Sprintf("%s/mw%d/cancelable%t", kind, count, cancelable), func(b *testing.B) {
					g := New(WithRoutes(func(r *RouterGroup) {
						for i := 0; i < count; i++ {
							r.Middlewares(benchHandler)
						}
						r.GET("/static", benchHandler)
						r.GET("/param/#id", benchHandler)
						r.GET("/wild/#...", benchHandler)
					}))
					path, method := "/static", GET
					switch kind {
					case "param":
						path = "/param/1"
					case "wildcard":
						path = "/wild/a/b"
					case "404":
						path = "/missing"
					case "405":
						method = "POST"
					}
					req := httptest.NewRequest(method, path, nil)
					if cancelable {
						ctx, cancel := context.WithCancel(context.Background())
						defer cancel()
						req = req.WithContext(ctx)
					}
					rw := &reusableResponseWriter{header: make(http.Header)}
					g.ServeHTTP(rw, req)
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						rw.reset()
						g.ServeHTTP(rw, req)
					}
				})
			}
		}
	}
}

var globalValueSink any

func BenchmarkGlobalOptimization(b *testing.B) {
	for _, history := range []int{0, 1000, 100000} {
		for _, operation := range []string{"hit", "miss", "update"} {
			b.Run(fmt.Sprintf("history%d/%s", history, operation), func(b *testing.B) {
				g := New()
				g.SetValue("stable", 1)
				for i := 0; i < history; i++ {
					g.SetValue("config", i)
				}
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					switch operation {
					case "hit":
						globalValueSink = g.Value("stable")
					case "miss":
						globalValueSink = g.Value("missing")
					case "update":
						g.SetValue("config", 1)
					}
				}
			})
		}
	}
}

func BenchmarkBuildOptimization(b *testing.B) {
	for _, count := range []int{100, 1000} {
		for _, depth := range []int{1, 10} {
			for _, methods := range []int{1, 3} {
				for _, mw := range []int{0, 5, 20} {
					b.Run(fmt.Sprintf("routes%d/depth%d/methods%d/mw%d", count, depth, methods, mw), func(b *testing.B) {
						g := New(WithRoutes(func(r *RouterGroup) {
							for d := 0; d < depth; d++ {
								r = r.Grep("/group")
								for m := 0; m < mw; m++ {
									r.Middlewares(benchHandler)
								}
							}
							for i := 0; i < count; i++ {
								p := fmt.Sprintf("/route%d", i)
								r.GET(p, benchHandler)
								if methods == 3 {
									r.POST(p, benchHandler)
									r.PUT(p, benchHandler)
								}
							}
						}))
						b.ReportAllocs()
						b.ResetTimer()
						for i := 0; i < b.N; i++ {
							g.host.publish(&g.RouterGroup)
						}
					})
				}
			}
		}
	}
}
