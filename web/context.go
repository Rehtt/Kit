/**
 * @Author: dsreshiram@gmail.com
 * @Date: 2022/7/16 下午 08:50
 */

package goweb

import (
	"context"
	"net/http"
)

type Context struct {
	Request *http.Request
	Writer  http.ResponseWriter

	param map[string]string

	context.Context
	cancel context.CancelFunc
}

type HandlerFunc func(ctx *Context)

func (c *Context) GetParam(key string) string {
	if c.param == nil {
		return ""
	}
	return c.param[key]
}

func (c *Context) GetValue(key interface{}) interface{} {
	return c.Value(key)
}
func (c *Context) SetValue(key interface{}, value interface{}) {
	c.Context = context.WithValue(c.Context, key, value)
}

func (c *Context) Stop() {
	c.cancel()
}

func (c *Context) runFunc(handlerFunc HandlerFunc) {
	select {
	case <-c.Done():
		return
	default:
		handlerFunc(c)
	}
}
