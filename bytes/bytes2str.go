package bytes

import "unsafe"

// ToString 高性能转换
func ToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
