package slice

import (
	"slices"
	"testing"
)

func cmpSlice[E comparable](a, b [][]E) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if len(a[i]) != len(b[i]) {
			return false
		}

		if !slices.Equal(a[i], b[i]) {
			return false
		}
	}
	return true
}

func TestSplit(t *testing.T) {
	type args struct {
		data []int
		n    int
	}
	tests := []struct {
		name string
		args args
		want [][]int
	}{
		{
			name: "test1",
			args: args{
				data: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				n:    3,
			},
			want: [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}, {10}},
		},
		{
			name: "test2",
			args: args{
				data: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				n:    2,
			},
			want: [][]int{{1, 2}, {3, 4}, {5, 6}, {7, 8}, {9, 10}},
		},
		{
			name: "test3",
			args: args{
				data: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				n:    1,
			},
			want: [][]int{{1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}, {10}},
		},
		{
			name: "test4",
			args: args{
				data: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				n:    0,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Split(tt.args.data, tt.args.n); !cmpSlice(got, tt.want) {
				t.Errorf("Split() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIterSplit(t *testing.T) {
	type args struct {
		data []int
		n    int
	}
	tests := []struct {
		name string
		args args
		want [][]int
	}{
		{
			name: "test1",
			args: args{
				data: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				n:    3,
			},
			want: [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}, {10}},
		},
		{
			name: "test2",
			args: args{
				data: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				n:    2,
			},
			want: [][]int{{1, 2}, {3, 4}, {5, 6}, {7, 8}, {9, 10}},
		},
		{
			name: "test3",
			args: args{
				data: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				n:    1,
			},
			want: [][]int{{1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}, {10}},
		},
		{
			name: "test4",
			args: args{
				data: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				n:    0,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got [][]int
			IterSplit(tt.args.data, tt.args.n)(func(v []int) bool {
				got = append(got, v)
				return true
			})
			if !cmpSlice(got, tt.want) {
				t.Errorf("IterSplit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIterSplit2(t *testing.T) {
	type args struct {
		data []int
		n    int
	}
	tests := []struct {
		name string
		args args
		want [][]int
	}{
		{
			name: "test1",
			args: args{
				data: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				n:    3,
			},
			want: [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}, {10}},
		},
		{
			name: "test2",
			args: args{
				data: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				n:    2,
			},
			want: [][]int{{1, 2}, {3, 4}, {5, 6}, {7, 8}, {9, 10}},
		},
		{
			name: "test3",
			args: args{
				data: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				n:    1,
			},
			want: [][]int{{1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}, {10}},
		},
		{
			name: "test4",
			args: args{
				data: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				n:    0,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got [][]int
			IterSplit2(tt.args.data, tt.args.n)(func(v []int, i int) bool {
				got = append(got, v)
				return true
			})
			if !cmpSlice(got, tt.want) {
				t.Errorf("IterSplit2() = %v, want %v", got, tt.want)
			}
		})
	}
}
