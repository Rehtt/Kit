package strings

import "unsafe"

// UnsafeToBytes 高性能转换
// Deprecated: 请使用 UnsafeStringToBytes
// 有 GC 风险，uintptr是整数可能导致 GC 不知道这是指针，最后 GC 回收了底层数组
func UnsafeToBytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func UnsafeStringToBytes(str string) []byte {
	if str == "" {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(str), len(str))
}
