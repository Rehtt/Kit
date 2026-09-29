package web

import (
	"net/http/httptest"
	"testing"
)

func TestCachedAllowHeader(t *testing.T) {
	for _, path := range []string{"/static", "/param/#id", "/wild/#..."} {
		t.Run(path, func(t *testing.T) {
			g := New()
			g.POST(path, benchHandler)
			g.GET(path, benchHandler)
			requestPath := path
			if path == "/param/#id" {
				requestPath = "/param/1"
			}
			if path == "/wild/#..." {
				requestPath = "/wild/a/b"
			}
			old := g.host.snapshot.Load()
			_, _, leaf, _ := old.match(requestPath, "DELETE")
			if leaf.allowHeader != "GET, POST" {
				t.Fatal(leaf.allowHeader)
			}
			check := func(method, want string, status int) {
				t.Helper()
				rw := httptest.NewRecorder()
				g.ServeHTTP(rw, httptest.NewRequest(method, requestPath, nil))
				if rw.Code != status || rw.Header().Get("Allow") != want {
					t.Fatalf("%s: status=%d Allow=%q", method, rw.Code, rw.Header().Get("Allow"))
				}
			}
			check("DELETE", "GET, POST", 405)
			check("HEAD", "", 200)
			g.PUT(path, benchHandler)
			check("DELETE", "GET, POST, PUT", 405)
			if leaf.allowHeader != "GET, POST" {
				t.Fatal("old snapshot changed")
			}
			g = New()
			g.Any(path, benchHandler)
			check("DELETE", "", 200)
			_, _, current, _ := g.host.snapshot.Load().match(requestPath, "GET")
			if current.allowHeader != "" {
				t.Fatalf("ANY included: %q", current.allowHeader)
			}
		})
	}
}
