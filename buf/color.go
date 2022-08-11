package buf

import (
	"bytes"
	"github.com/fatih/color"
)

func (buf *Buf) WriteColor(b interface{}, colors ...color.Attribute) *Buf {
	color.New(colors...).Fprint(buf.buf, b)
	return buf
}

func (buf *Buf) ToColorString(colors []color.Attribute, free ...bool) string {
	if len(free) != 0 && free[0] {
		defer buf.Free()
	}
	return color.New(colors...).Sprint(buf.buf.String())
}
func (buf *Buf) ToColorBytes(colors []color.Attribute, free ...bool) []byte {
	if len(free) != 0 && free[0] {
		defer buf.Free()
	}
	tmp := bufPool.Get().(*bytes.Buffer)
	defer bufPool.Put(tmp)

	color.New(colors...).Fprint(tmp, buf.buf.String())
	return tmp.Bytes()
}
