//go:build jsoniter
// +build jsoniter

package web

import (
	jsoniter "github.com/json-iterator/go"
)

func (c *Context) ReadJSON(v any) error {
	defer c.Request.Body.Close()
	return jsoniter.NewDecoder(c.Request.Body).Decode(v)
}

func (c *Context) WriteJSON(v any, statusCode ...int) error {
	if len(statusCode) != 0 {
		c.Writer.WriteHeader(statusCode[0])
	}
	c.Writer.Header().Set("content-type", "application/json; charset=utf-8")
	return jsoniter.NewEncoder(c.Writer).Encode(v)
}
