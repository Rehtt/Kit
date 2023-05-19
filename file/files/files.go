package files

import (
	"bytes"
	"io"
	"os"
)

type Files struct {
	list  []string
	index int
	buf   bytes.Buffer
	fn    func(r io.Reader, w io.Writer) error
}

func NewReader(fileList []string) *Files {
	return &Files{
		list:  fileList,
		index: -1,
	}
}

func (f *Files) Read(b []byte) (n int, err error) {
	if err = f.init(); err != nil {
		return 0, err
	}
	n, err = f.buf.Read(b)
	if err != nil {
		if err != io.EOF {
			return 0, err
		}
		if len(b) > n {
			m, err := f.Read(b[n:])
			if err != nil {
				return 0, err
			}
			n += m
		}
	}
	return
}
func (f *Files) AfterReadFile(fn func(r io.Reader, w io.Writer) error) {
	f.fn = fn
}
func (f *Files) init() error {
	if f.buf.Len() == 0 {
		f.index += 1
		if f.index >= len(f.list) {
			return io.EOF
		}
		if f.list[f.index] == "" {
			return f.init()
		}
		file, err := os.Open(f.list[f.index])
		if err != nil {
			return err
		}
		defer file.Close()
		f.buf.Reset()
		if f.fn != nil {
			if err = f.fn(file, &f.buf); err != nil {
				return err
			}
		} else {
			if _, err = f.buf.ReadFrom(file); err != nil {
				return err
			}
		}

	}
	return nil
}
