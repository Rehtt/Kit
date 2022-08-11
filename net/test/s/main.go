package main

import (
	"github.com/Rehtt/Kit/net/server"
)

func main() {
	var b, a *server.Context
	e := server.New("tcp", "0.0.0.0:7220")
	e.Handle = func(ctx *server.Context) {
		switch ctx.Body().ToString() {
		case "b":
			b = ctx
		case "a":
			if b == nil {
				ctx.Write([]byte("nil"))
				ctx.Send()
				ctx.ConnClose()
				return
			}
			b.Write([]byte(ctx.RemoteAddr().String()))
			b.Send()
			a = ctx
			return
		case "done":
			a.Write([]byte(b.RemoteAddr().String()))
			a.Send()
			a.ConnClose()
			b.ConnClose()
		}
	}
	panic(e.Run())
}
