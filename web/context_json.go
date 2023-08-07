package goweb

import "encoding/json"

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
