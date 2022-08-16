package buf

import (
	"bytes"
	"github.com/Rehtt/Kit/vt/color"
)

func (buf *Buf) WriteColor(b interface{}, colors ...color.Color) *Buf {
	color.NewColors(colors...).Fprint(buf.buf, b)
	return buf
}

func (buf *Buf) ToColorString(colors []color.Color, free ...bool) string {
	if len(free) != 0 && free[0] {
		defer buf.Free()
	}
	return color.NewColors(colors...).Sprint(buf.buf.String())
}
func (buf *Buf) ToColorBytes(colors []color.Color, free ...bool) []byte {
	if len(free) != 0 && free[0] {
		defer buf.Free()
	}
	tmp := bufPool.Get().(*bytes.Buffer)
	defer bufPool.Put(tmp)

	color.NewColors(colors...).Fprint(tmp, buf.buf.String())
	return tmp.Bytes()
}
