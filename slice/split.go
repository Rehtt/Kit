package slice

// Split 将一个切片划分为多个大小为len的切片
func Split[Slice ~[]E, E any](data Slice, n int) (out []Slice) {
	if n < 1 {
		return nil
	}
	out = make([]Slice, 0, len(data)/n+1)
	for i := 0; i < len(data); i += n {
		end := min(n, len(data[i:]))
		out = append(out, data[i:i+end:i+end])
	}
	return
}

// IterSplit 将一个切片划分为多个大小为len的切片
// 并且通过一个函数迭代切片
func IterSplit[Slice ~[]E, E any](data Slice, n int) func(yield func(Slice) bool) {
	return func(yield func(Slice) bool) {
		if n < 1 {
			return
		}
		for i := 0; i < len(data); i += n {
			end := min(n, len(data[i:]))
			if !yield(data[i : i+end : i+end]) {
				return
			}
		}
	}
}

func IterSplit2[Slice ~[]E, E any](data Slice, n int) func(yield func(Slice, int) bool) {
	return func(yield func(Slice, int) bool) {
		if n < 1 {
			return
		}
		for i := 0; i < len(data); i += n {
			end := min(n, len(data[i:]))
			if !yield(data[i:i+end:i+end], i) {
				return
			}
		}
	}
}
