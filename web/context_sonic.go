//go:build sonic && !jsoniter
// +build sonic,!jsoniter

package web

import (
	"github.com/bytedance/sonic"
)

var JSON = sonic.ConfigStd

// 追求至极速度使用sonic ConfigFastest不验证json struct
func (c *Context) ReadJSON(v any) error {
	defer c.Request.Body.Close()
	return JSON.NewDecoder(c.Request.Body).Decode(v)
}

// 追求至极速度使用sonic ConfigFastest不验证json struct
func (c *Context) WriteJSON(v any, statusCode ...int) error {
	c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	if len(statusCode) != 0 {
		c.Writer.WriteHeader(statusCode[0])
	}
	return JSON.NewEncoder(c.Writer).Encode(v)
}
