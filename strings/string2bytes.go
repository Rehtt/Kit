package strings

import "unsafe"

// ToBytes 高性能转换
func ToBytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func ToString(b []byte) string {
	x := (*[3]uintptr)(unsafe.Pointer(&b))
	h := [2]uintptr{x[0], x[1]}
	return *(*string)(unsafe.Pointer(&h))
}
