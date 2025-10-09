package bytes

import "unsafe"

// UnsafeToString 高性能转换
func UnsafeToString(b []byte) string {
	x := (*[3]uintptr)(unsafe.Pointer(&b))
	h := [2]uintptr{x[0], x[1]}
	return *(*string)(unsafe.Pointer(&h))
}
