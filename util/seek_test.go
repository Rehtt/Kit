package util

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestReadLastNLines(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		n             int
		wantLines     []string
		expectErr     bool
		whence        int
		preSeekOffset int64
	}{
		{
			name:      "SeekEnd: 基本情况 读取最后2行",
			content:   "line1\nline2\nline3\nline4",
			n:         2,
			wantLines: []string{"line3", "line4"},
			whence:    io.SeekEnd,
		},
		{
			name:      "SeekEnd: 行数不足 请求5行实际3行",
			content:   "line1\nline2\nline3",
			n:         5,
			wantLines: []string{"line1", "line2", "line3"},
			whence:    io.SeekEnd,
		},
		{
			name:      "SeekEnd: 末尾有换行符",
			content:   "line1\nline2\n",
			n:         1,
			wantLines: []string{"line2"},
			whence:    io.SeekEnd,
		},
		{
			name:      "SeekEnd: 空文件",
			content:   "",
			n:         5,
			wantLines: []string{},
			whence:    io.SeekEnd,
		},

		{
			name:    "SeekCurrent: 在文件中间截取",
			content: "part1\npart2\npart3\npart4",
			// 指针停在 "part2" 的末尾 (即 "part1\npart2" 之后)
			// "part1\n" 是 6 字节, "part2" 是 5 字节 = 11
			preSeekOffset: 11,
			n:             1,
			wantLines:     []string{"part2"},
			whence:        io.SeekCurrent,
		},
		{
			name:    "SeekCurrent: 在文件中间截取多行",
			content: "A\nB\nC\nD\nE",
			// 指针停在 "C" 的末尾 ("A\nB\nC" = 2+2+1 = 5 bytes)
			preSeekOffset: 5,
			n:             2,
			wantLines:     []string{"B", "C"},
			whence:        io.SeekCurrent,
		},
		{
			name:          "SeekCurrent: 指针在文件开头",
			content:       "line1\nline2",
			preSeekOffset: 0,
			n:             5,
			wantLines:     []string{}, // 前面没有数据
			whence:        io.SeekCurrent,
		},
		{
			name:          "SeekCurrent: 指针在文件中间但前面无换行",
			content:       "singlelinecontent",
			preSeekOffset: 6, // 指针在 "single" 后面
			n:             1,
			wantLines:     []string{"single"},
			whence:        io.SeekCurrent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.content)

			// 确定初始偏移量 logic
			var targetOffset int64

			if tt.whence == io.SeekCurrent {
				// 如果是 SeekCurrent，严格遵守测试用例设定的位置
				targetOffset = tt.preSeekOffset
			}

			// 执行 Seek 操作
			_, err := reader.Seek(targetOffset, io.SeekStart)
			if err != nil {
				t.Fatalf("Setup failed: could not seek to %d: %v", targetOffset, err)
			}

			// 执行被测函数
			got, err := ReadLastNLines(reader, tt.n, tt.whence)

			// 错误检查
			if (err != nil) != tt.expectErr {
				t.Errorf("ReadLastNLines() error = %v, expectErr %v", err, tt.expectErr)
				return
			}

			// 结果比对
			if !reflect.DeepEqual(got, tt.wantLines) {
				t.Errorf("ReadLastNLines() = %v, want %v", got, tt.wantLines)
			}
		})
	}
}
