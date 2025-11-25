/**
 * @Author: dsreshiram@gmail.com
 * @Date: 2022/7/16 下午 08:50
 */

package web

import (
	"context"
	"maps"
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

func (c *Context) AllUrlPathParam() map[string]string {
	return maps.Clone(c.param)
}

func (c *Context) GetUrlPathParam(key string) string {
	if c.param == nil {
		return ""
	}
	return c.param[key]
}

func (c *Context) GetContextValue(key any) any {
	return c.Value(key)
}

func (c *Context) SetContextValue(key any, value any) {
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
