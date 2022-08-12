/**
 * @Author: dsreshiram@gmail.com
 * @Date: 2022/7/16 下午 08:50
 */

package goweb

import (
	"context"
	"net/http"
	"sync"
)

type Context struct {
	Request *http.Request
	Writer  http.ResponseWriter

	param map[string]string

	lock   sync.RWMutex
	values map[interface{}]interface{}

	survive bool

	context.Context
}

type HandlerFunc func(ctx *Context)

func (c *Context) GetParam(key string) string {
	return c.param[key]
}

func (c *Context) Get(key interface{}) interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.values[key]
}
func (c *Context) Set(key interface{}, value interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.values[key] = value
}

func (c *Context) Stop() {
	c.survive = false
}

func (c *Context) runFunc(handlerFunc HandlerFunc) {
	if !c.survive {
		return
	}
	handlerFunc(c)
}

func (c *Context) Value(key interface{}) interface{} {
	return c.Get(key)
}
