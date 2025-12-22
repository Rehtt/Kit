package bytes

import "unsafe"

// UnsafeToString 高性能转换
// Deprecated: 请使用 UnsafeBytesToString
// 有 GC 风险，uintptr是整数可能导致 GC 不知道这是指针，最后 GC 回收了底层数组
func UnsafeToString(b []byte) string {
	x := (*[3]uintptr)(unsafe.Pointer(&b))
	h := [2]uintptr{x[0], x[1]}
	return *(*string)(unsafe.Pointer(&h))
}

func UnsafeBytesToString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(unsafe.SliceData(b), len(b))
}
