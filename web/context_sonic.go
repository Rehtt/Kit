//go:build sonic
// +build sonic

package web

import (
	"github.com/bytedance/sonic"
)

// 追求至极速度使用sonic ConfigFastest不验证json struct
func (c *Context) ReadJSON(v any) error {
	defer c.Request.Body.Close()
	return sonic.ConfigFastest.NewDecoder(c.Request.Body).Decode(v)
}

// 追求至极速度使用sonic ConfigFastest不验证json struct
func (c *Context) WriteJSON(v any, statusCode ...int) error {
	if len(statusCode) != 0 {
		c.Writer.WriteHeader(statusCode[0])
	}
	c.Writer.Header().Set("content-type", "application/json; charset=utf-8")
	return sonic.ConfigFastest.NewEncoder(c.Writer).Encode(v)
}
