package util

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestReadLastNLines(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		n            int
		wantLines    []string
		expectErr    bool
		noChangeSeek bool
	}{
		{
			name:      "基本情况: 读取最后2行",
			content:   "line1\nline2\nline3\nline4",
			n:         2,
			wantLines: []string{"line3", "line4"},
		},
		{
			name:      "行数不足: 请求5行实际3行",
			content:   "line1\nline2\nline3",
			n:         5,
			wantLines: []string{"line1", "line2", "line3"},
		},
		{
			name:      "精确匹配: 请求3行实际3行",
			content:   "a\nb\nc",
			n:         3,
			wantLines: []string{"a", "b", "c"},
		},
		{
			name:      "末尾有换行符",
			content:   "line1\nline2\n",
			n:         1,
			wantLines: []string{"line2"},
		},
		{
			name:      "单行无换行符",
			content:   "onlyoneline",
			n:         1,
			wantLines: []string{"onlyoneline"},
		},
		{
			name:      "空文件",
			content:   "",
			n:         5,
			wantLines: []string{},
		},
		{
			name:      "N为0",
			content:   "abc",
			n:         0,
			wantLines: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 使用 strings.NewReader，因为它实现了 io.ReadSeeker
			reader := strings.NewReader(tt.content)

			// 故意先移动一下指针，测试 noChangeSeek 是否能恢复到这里
			initialOffset := int64(1)
			if len(tt.content) > 0 {
				reader.Seek(initialOffset, io.SeekStart)
			} else {
				initialOffset = 0
			}

			got, err := ReadLastNLines(reader, tt.n, tt.noChangeSeek)

			if (err != nil) != tt.expectErr {
				t.Errorf("ReadLastNLines() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantLines) {
				t.Errorf("ReadLastNLines() = %v, want %v", got, tt.wantLines)
			}

			// 验证 Seek 位置恢复
			if tt.noChangeSeek {
				curr, _ := reader.Seek(0, io.SeekCurrent)
				if curr != initialOffset {
					t.Errorf("Seek position not restored. Got %d, want %d", curr, initialOffset)
				}
			}
		})
	}
}
