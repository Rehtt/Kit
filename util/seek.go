package util

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

func ReadLastNLines(data io.ReadSeeker, n int, whence int, setChunkSize ...int64) ([]string, error) {
	if data == nil {
		return nil, errors.New("data is nil")
	}
	if whence != io.SeekEnd && whence != io.SeekCurrent {
		return nil, errors.New("whence must be SeekEnd or SeekCurrent")
	}
	if n <= 0 {
		return []string{}, nil
	}

	// 结束边界
	endPos, err := data.Seek(0, whence)
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
		currentPos           = endPos
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
		searchBuf := chunk[:step]
		for {
			idx := bytes.LastIndexByte(searchBuf, '\n')
			if idx == -1 {
				break
			}
			if g := currentPos + int64(idx); g != endPos-1 {
				newlinesFound++
				if newlinesFound >= n {
					finalReadStart = g + 1
					goto FoundStart
				}
			}
			searchBuf = searchBuf[:idx]
		}
	}
	// 如果循环结束还没找到足够的换行符，说明文件行数 < N，从头开始读
	finalReadStart = 0

FoundStart:
	_, err = data.Seek(finalReadStart, io.SeekStart)
	if err != nil {
		return nil, err
	}
	readLength := endPos - finalReadStart
	limitReader := io.LimitReader(data, readLength)

	scanner := bufio.NewScanner(limitReader)
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

	return lines, nil
}
