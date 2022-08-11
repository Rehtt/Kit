package main

import (
	"fmt"
	"github.com/Rehtt/Kit/net/client"
	"log"
)

func main() {
	a, err := client.New("tcp", "0.0.0.0:8340", "rehtt.com:7220", true)
	if err != nil {
		log.Panicln(err)
	}
	a.OnResponse = func(ctx *client.Context) {
		i := ctx.Body().ToString(true)
		t, err := client.New("tcp", "0.0.0.0:8340", i, true)
		if err != nil {
			log.Fatalln(err)
			return
		}
		fmt.Println(t.Dial())
		t.Write([]byte("test"))
		t.Send()
	}
	fmt.Println(a.Dial())
	a.Write([]byte("a"))
	a.Send()
	a.Wait()
}
