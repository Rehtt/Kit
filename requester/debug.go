// Copyright (c) 2025 Rehtt
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// @Author: Rehtt dsreshiram@gmail.com
// @Date: 2025/11/25

package requester

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"text/template"

	"github.com/Rehtt/Kit/bytes"
)

func (h *Requester) Debug(debug bool) *Requester {
	h.debug = debug
	return h
}

func (h *Requester) printRequestDebug(req *http.Request) {
	if !h.debug {
		return
	}
	templ := `

Request
	url:	{{.URL}}
	method:	{{.Method}}
	{{- range $k,$v := .Header }}
		{{- range $v }}
	header:	{{$k}}:{{.}}
		{{- end }}
	{{- end }}
	body:	{{if eq .Body ""}}<nil>
	{{- else}}
	--- body start ---
	{{.Body}}
	--- body end ---
	{{- end}}
	`
	data := struct {
		*http.Request
		Body string
	}{
		Request: req,
	}
	if req.Body != nil {
		raw, _ := io.ReadAll(req.Body)
		buf := bytes.MakeByteBuffer(raw)
		data.Body = buf.String()
		req.Body.Close()
		req.Body = &buf
	}
	err := template.Must(template.New("debug").Parse(templ)).Execute(os.Stdout, data)
	if err != nil {
		fmt.Println(err)
	}
}

func (h *Requester) printResponseDebug(resp *http.Response) {
	if !h.debug {
		return
	}
	templ := `
Response
	{{- range $k,$v := .Header }}
		{{- range $v }}
	header:	{{$k}}:{{.}}
		{{- end}}
	{{- end }}
	body:	{{if eq .Body ""}}<nil>
	{{- else}}
	--- body start ---
	{{.Body}}
	--- body end ---
	{{- end}}

	`
	data := struct {
		*http.Response
		Body string
	}{
		Response: resp,
	}
	if resp.Body != nil {
		raw, _ := io.ReadAll(resp.Body)
		buf := bytes.MakeByteBuffer(raw)
		data.Body = buf.String()
		resp.Body.Close()
		resp.Body = &buf
	}

	err := template.Must(template.New("debug").Parse(templ)).Execute(os.Stdout, data)
	if err != nil {
		fmt.Println(err)
	}
}
