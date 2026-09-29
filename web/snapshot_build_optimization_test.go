package web

import (
	"fmt"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestDeepSnapshotMiddlewareStack(t *testing.T) {
	var calls []string
	record := func(s string) HandlerFunc { return func(*Context) { calls = append(calls, s) } }
	g := New(WithRoutes(func(root *RouterGroup) {
		root.Middlewares(record("root"))
		for _, branch := range []string{"left", "right"} {
			r := root.Grep("/" + branch)
			for i := 0; i < 10; i++ {
				r = r.Grep("/level")
				r.Middlewares(record(fmt.Sprint(branch, i)))
			}
			r.GET("/leaf", record("get"))
			r.Middlewares(record("late"))
			r.POST("/leaf", record("post"))
		}
	}))
	old := g.host.snapshot.Load()
	for _, branch := range []string{"left", "right"} {
		path := "/" + branch
		want := []string{"root"}
		for i := 0; i < 10; i++ {
			path += "/level"
			want = append(want, fmt.Sprint(branch, i))
		}
		path += "/leaf"
		for _, method := range []string{"GET", "POST"} {
			calls = nil
			expected := append([]string(nil), want...)
			if method == "POST" {
				expected = append(expected, "late", "post")
			} else {
				expected = append(expected, "get")
			}
			g.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(method, path, nil))
			if !reflect.DeepEqual(calls, expected) {
				t.Fatalf("%s %s: %v != %v", method, path, calls, expected)
			}
			if old.flatRoutes[path[1:]] == nil {
				t.Fatal("static path workspace corrupted")
			}
		}
	}
}
