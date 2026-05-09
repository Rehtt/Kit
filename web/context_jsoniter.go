//go:build jsoniter && !sonic
// +build jsoniter,!sonic

package web

import (
	jsoniter "github.com/json-iterator/go"
)

var JSON = jsoniter.ConfigDefault

func (c *Context) ReadJSON(v any) error {
	defer c.Request.Body.Close()

	return JSON.NewDecoder(c.Request.Body).Decode(v)
}

func (c *Context) WriteJSON(v any, statusCode ...int) error {
	c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	if len(statusCode) != 0 {
		c.Writer.WriteHeader(statusCode[0])
	}
	return JSON.NewEncoder(c.Writer).Encode(v)
}
