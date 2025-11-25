// Copyright (c) 2025 Rehtt
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// @Author: Rehtt dsreshiram@gmail.com
// @Date: 2025/11/25

package requester

import (
	"io"
	"net/http"
	"sync"
)

type Requester struct {
	url      string
	m        string
	header   http.Header
	body     io.Reader
	response *http.Response

	err   error
	debug bool
}

var requesterPool = sync.Pool{
	New: func() any {
		return &Requester{
			header: make(http.Header),
		}
	},
}

type StreamResult[T any] struct {
	Data T
	Err  error
}

type StreamResultBytes StreamResult[[]byte]

func (s StreamResultBytes) String() string {
	if len(s.Data) == 0 {
		return ""
	}
	return string(s.Data)
}
