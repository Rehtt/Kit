/**
 * @Author: dsreshiram@gmail.com
 * @Date: 2022/7/16 下午 08:50
 */

package goweb

import (
	"context"
	"encoding/json"
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
func (c *Context) ReadJSON(v interface{}) error {
	return json.NewDecoder(c.Request.Body).Decode(v)
}
func (c *Context) WriteJSON(v interface{}, statusCode ...int) error {
	if len(statusCode) != 0 {
		c.Writer.WriteHeader(statusCode[0])
	}
	c.Writer.Header().Set("content-type", "application/json; charset=utf-8")
	return json.NewEncoder(c.Writer).Encode(v)
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
