package http

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// SendFile 分段传输文件
func SendFile(writer http.ResponseWriter, request *http.Request, f *os.File, buf_n ...int) {
	info, err := f.Stat()
	if err != nil {
		log.Println("sendFile1", err.Error())
		http.NotFound(writer, request)
		return
	}
	SetHeader(writer.Header(), "Accept-Ranges", "bytes")
	SetHeader(writer.Header(), "Content-Disposition", "attachment; filename="+info.Name())

	etag := sha1.New()
	etag.Write([]byte(strconv.FormatInt(info.ModTime().UnixNano(), 10)))
	SetHeader(writer.Header(), "ETag", fmt.Sprintf("%x", etag.Sum(nil)))
	var start, end int64
	IfRange := request.Header.Get("If-Range")
	// fmt.Println(request.Header,"\n")
	if r := request.Header.Get("Range"); r != "" && (IfRange == fmt.Sprintf("%x", etag.Sum(nil)) || IfRange == "") {
		if strings.Contains(r, "bytes=") && strings.Contains(r, "-") {

			fmt.Sscanf(r, "bytes=%d-%d", &start, &end)
			if end == 0 {
				end = info.Size() - 1
			}
			if start > end || start < 0 || end < 0 || end >= info.Size() {
				writer.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
				log.Println("sendFile2 start:", start, "end:", end, "size:", info.Size())
				return
			}
			writer.Header().Set("Content-Length", strconv.FormatInt(end-start+1, 10))
			writer.Header().Set("Content-Range", fmt.Sprintf("bytes %v-%v/%v", start, end, info.Size()))
			writer.WriteHeader(http.StatusPartialContent)
		} else {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
	} else {
		writer.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
		start = 0
		end = info.Size() - 1
	}
	_, err = f.Seek(start, 0)
	if err != nil {
		log.Println("sendFile3", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	bufSize := 512
	if len(buf_n) != 0 {
		bufSize = buf_n[0]
	}
	buf := make([]byte, bufSize)
	for {
		if end-start+1 < int64(bufSize) {
			bufSize = int(end - start + 1)
		}
		_, err := f.Read(buf[:bufSize])
		if err != nil {
			log.Println("1:", err)
			if err != io.EOF {
				log.Println("error:", err)
			}
			return
		}
		err = nil
		_, err = writer.Write(buf[:bufSize])
		if err != nil {
			// log.Println(err, start, end, info.Size(), n)
			return
		}
		start += int64(bufSize)
		if start >= end+1 {
			return
		}
	}
}
