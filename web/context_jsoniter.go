//go:build jsoniter
// +build jsoniter

package goweb

import (
	jsoniter "github.com/json-iterator/go"
)

func (c *Context) ReadJSON(v interface{}) error {
	return jsoniter.NewDecoder(c.Request.Body).Decode(v)
}
func (c *Context) WriteJSON(v interface{}, statusCode ...int) error {
	if len(statusCode) != 0 {
		c.Writer.WriteHeader(statusCode[0])
	}
	c.Writer.Header().Set("content-type", "application/json; charset=utf-8")
	return jsoniter.NewEncoder(c.Writer).Encode(v)
}
