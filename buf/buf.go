package buf

import (
	"bytes"
	"io"
	"sync"
)

var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

type Buf struct {
	buf *bytes.Buffer
}

func NewBuf() *Buf {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	return &Buf{buf}
}
func (buf *Buf) Free() {
	buf.buf.Reset()
	bufPool.Put(buf.buf)
}

func (buf *Buf) WriteByte(b byte) *Buf {
	buf.buf.WriteByte(b)
	return buf
}

func (buf *Buf) Write(b []byte) (n int, err error) {
	return buf.buf.Write(b)
}

func (buf *Buf) WriteRune(r rune) *Buf {
	buf.buf.WriteRune(r)
	return buf
}

func (buf *Buf) WriteBytes(b []byte) *Buf {
	buf.buf.Write(b)
	return buf
}

func (buf *Buf) WriteString(src string) *Buf {
	buf.buf.WriteString(src)
	return buf
}

func (buf *Buf) ToString(free ...bool) string {
	if len(free) != 0 && free[0] {
		defer buf.Free()
	}
	return buf.buf.String()
}

func (buf *Buf) ToBytes(free ...bool) []byte {
	if len(free) != 0 && free[0] {
		defer buf.Free()
	}
	return buf.buf.Bytes()
}

func (buf *Buf) WriteTo(w io.Writer, free ...bool) (n int64, err error) {
	if len(free) != 0 {
		defer buf.Free()
	}
	return buf.buf.WriteTo(w)
}

func (buf *Buf) ReadFrom(r io.Reader) (n int64, err error) {
	return buf.buf.ReadFrom(r)
}

func (buf *Buf) Read(b []byte) (n int, err error) {
	return buf.buf.Read(b)
}

func (buf *Buf) Reset() {
	buf.buf.Reset()
}
func (buf *Buf) Len() int {
	return buf.buf.Len()
}
func (buf *Buf) Cap() int {
	return buf.buf.Cap()
}
func (buf *Buf) Grow(n int) {
	buf.buf.Grow(n)
}
