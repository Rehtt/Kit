package http

import (
	"net/http"
)

func SetHeader(header http.Header, key, value string) {
	if header.Get(key) == "" {
		header.Set(key, value)
	}
}
