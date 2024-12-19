//go:build !jsoniter
// +build !jsoniter

package goweb

import "encoding/json"

func (c *Context) ReadJSON(v any) error {
	defer c.Request.Body.Close()
	return json.NewDecoder(c.Request.Body).Decode(v)
}

func (c *Context) WriteJSON(v any, statusCode ...int) error {
	if len(statusCode) != 0 {
		c.Writer.WriteHeader(statusCode[0])
	}
	c.Writer.Header().Set("content-type", "application/json; charset=utf-8")
	return json.NewEncoder(c.Writer).Encode(v)
}
