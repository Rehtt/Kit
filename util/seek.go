package util

import (
	"bufio"
	"errors"
	"io"
)

// ReadLastNLines 读取 io.ReadSeeker 的最后 N 行
func ReadLastNLines(data io.ReadSeeker, n int, noChangeSeek bool, setChunkSize ...int64) ([]string, error) {
	if data == nil {
		return nil, errors.New("data is nil")
	}
	if n <= 0 {
		return []string{}, nil
	}

	// 当前位置
	originalPos, err := data.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	dataSize, err := data.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}

	var chunkSize int64 = 1024
	if len(setChunkSize) > 0 {
		if setChunkSize[0] < 1 {
			return nil, errors.New("chunkSize must be greater than 0")
		}
		chunkSize = setChunkSize[0]
	}

	var (
		currentPos           = dataSize
		newlinesFound        = 0
		chunk                = make([]byte, chunkSize)
		finalReadStart int64 = 0
	)

	// 如果文件很小，直接从头读；否则从尾部开始分块向前读
	for currentPos > 0 {
		// 确定本次回退的步长
		step := min(currentPos, chunkSize)

		// 移动游标
		currentPos -= step
		_, err = data.Seek(currentPos, io.SeekStart)
		if err != nil {
			return nil, err
		}

		// 读取块
		_, err = data.Read(chunk[:step])
		if err != nil {
			return nil, err
		}

		// 在块内倒序扫描换行符
		for i := int(step) - 1; i >= 0; i-- {
			isLastByteOfFile := (currentPos + int64(i)) == dataSize-1
			if chunk[i] == '\n' {
				if !isLastByteOfFile {
					newlinesFound++
				}
				if newlinesFound >= n {
					finalReadStart = currentPos + int64(i) + 1
					goto FoundStart
				}
			}
		}
	}
	// 如果循环结束还没找到足够的换行符，说明文件行数 < N，从头开始读
	finalReadStart = 0

FoundStart:
	_, err = data.Seek(finalReadStart, io.SeekStart)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(data)
	lines := make([]string, 0, n)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}

	if noChangeSeek {
		_, err = data.Seek(originalPos, io.SeekStart)
		if err != nil {
			return lines, err
		}
	}

	return lines, nil
}
