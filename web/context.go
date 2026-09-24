/**
 * @Author: dsreshiram@gmail.com
 * @Date: 2022/7/16 下午 08:50
 */

package web

import (
	"context"
	"io"
	"net/http"
	"time"

	kitbytes "github.com/Rehtt/Kit/bytes"
	kitstrings "github.com/Rehtt/Kit/strings"
)

const (
	contextParamBufferSize = 8
	maxContextParamCap     = 64
)

type pathParam struct {
	key   string
	value string
}

type Context struct {
	Request *http.Request
	// Writer 默认指向 &c.rw；中间件可替换为自定义 wrapper。
	Writer http.ResponseWriter
	rw     responseWriter

	// params 是当前请求的内部参数表。它优先使用 paramBuf，超过固定缓冲
	// 后才扩容；公开 API 只在被调用时创建独立 map。
	paramBuf [contextParamBufferSize]pathParam
	params   []pathParam

	// handlers 指向不可变 routeSnapshot 中的预编译执行链，不在请求结束时
	// 清空底层元素，以免修改仍可能被其它请求使用的快照切片。
	handlers []HandlerFunc
	index    int

	// Context: 请求生命周期的可取消 ctx，parent 是 request.Context()。
	// values: GOweb 全局 value chain，仅作为 Value() 查找的 fallback，
	// 不作为 WithCancel 的 parent，避免 stdlib 走未知类型 fallback 起 goroutine。
	context.Context
	cancel context.CancelFunc
	values context.Context
}

type (
	HandlerFunc func(ctx *Context)
	HandlerOpt  struct {
		Description string
	}
)

// Value 先查请求 ctx，未命中再回退到 GOweb.SetValue 写入的全局值。
func (c *Context) Value(key any) any {
	if c.Context != nil {
		if v := c.Context.Value(key); v != nil {
			return v
		}
	}
	if c.values != nil {
		return c.values.Value(key)
	}
	return nil
}

func (c *Context) AllUrlPathParam() map[string]string {
	if len(c.params) == 0 {
		return nil
	}
	params := make(map[string]string, len(c.params))
	for _, param := range c.params {
		params[param.key] = param.value
	}
	return params
}

func (c *Context) GetUrlPathParam(key string) string {
	for i := len(c.params) - 1; i >= 0; i-- {
		if c.params[i].key == key {
			return c.params[i].value
		}
	}
	return ""
}

func (c *Context) GetContextValue(key any) any {
	return c.Value(key)
}

func (c *Context) SetContextValue(key any, value any) {
	c.Context = context.WithValue(c.Context, key, value)
}

func (c *Context) WriteString(s string) (int, error) {
	if sw, ok := c.Writer.(io.StringWriter); ok {
		return sw.WriteString(s)
	}
	return c.Writer.Write(kitstrings.UnsafeStringToBytes(s))
}

func (c *Context) Write(b []byte) (int, error) {
	return c.Writer.Write(b)
}

func (c *Context) ReadFrom(src io.Reader) (int64, error) {
	if rw, ok := c.Writer.(io.ReaderFrom); ok {
		return rw.ReadFrom(src)
	}
	return io.Copy(c.Writer, src)
}

func (c *Context) Read(b []byte) (int, error) {
	return c.Request.Body.Read(b)
}

func (c *Context) ReadAll() ([]byte, error) {
	return io.ReadAll(c.Request.Body)
}

func (c *Context) ReadString() (string, error) {
	data, err := c.ReadAll()
	if err != nil {
		return "", err
	}
	return kitbytes.UnsafeBytesToString(data), nil
}

func (c *Context) Stop() {
	c.cancel()
}

// Next 显式调用则先把后续跑完再回到调用点（用于尾置逻辑）。
// ctx.Stop() / 客户端断开后下次循环立即短路。
func (c *Context) Next() {
	c.index++
	for c.index < len(c.handlers) {
		select {
		case <-c.Done():
			c.index = len(c.handlers)
			return
		default:
		}
		c.handlers[c.index](c)
		c.index++
	}
}

func (c *Context) SetReadDeadline(deadline time.Time) error {
	return http.NewResponseController(c.Writer).SetReadDeadline(deadline)
}

func (c *Context) SetWriteDeadline(deadline time.Time) error {
	return http.NewResponseController(c.Writer).SetWriteDeadline(deadline)
}
