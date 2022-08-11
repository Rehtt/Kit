package main

import (
	"fmt"
	"github.com/Rehtt/Kit/net/client"
	"github.com/Rehtt/Kit/net/server"
	"log"
	"time"
)

func main() {
	b, err := client.New("tcp", "0.0.0.0:8125", "rehtt.com:7220", true)
	if err != nil {
		log.Panicln(err)
	}
	b.OnResponse = func(ctx *client.Context) {
		var buf = make([]byte, 64)
		n, _ := ctx.Read(buf)
		fmt.Println(string(buf[:n]))
		for i := 0; i < 3; i++ {
			fmt.Println("try")
			c, err := client.New("tcp", "0.0.0.0:8125", string(buf[:n]), true)
			if err != nil {
				log.Panicln(err)
				continue
			}
			c.Timeout = 1 * time.Second
			c.Dial()

		}
		ctx.Write([]byte("done"))
		ctx.Send()
		s()
	}
	fmt.Println(b.Dial())
	b.Write([]byte("b"))
	b.Send()
	b.Wait()
	fmt.Println(123)
}

func s() {
	s := server.New("tcp", "0.0.0.0:8125")
	s.Handle = func(ctx *server.Context) {
		fmt.Println(ctx.Body().ToString(true))
	}
	s.Run()
}
