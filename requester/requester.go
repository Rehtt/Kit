package requester

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"text/template"

	"github.com/Rehtt/Kit/bytes"
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

func NewRequester() *Requester {
	return requesterPool.Get().(*Requester).Clear()
}

func (h *Requester) Get(u string) *Requester {
	h.url = u
	h.m = http.MethodGet
	return h
}

func (h *Requester) RequestJSON(method string, u string, obj any) *Requester {
	h.m = method
	h.url = u
	h.SetHead("content-type", "application/json")
	if obj != nil {
		var buf bytes.ByteBuffer
		switch obj.(type) {
		case string:
			buf.Write([]byte(obj.(string)))
		case []byte:
			buf.Write(obj.([]byte))
		default:
			h.err = json.NewEncoder(&buf).Encode(obj)
		}
		h.body = &buf

	}
	return h
}

func (h *Requester) Post(u string, body io.Reader) *Requester {
	h.m = http.MethodPost
	h.url = u
	h.body = body
	return h
}

func (h *Requester) PostJSON(u string, obj any) *Requester {
	return h.RequestJSON(http.MethodPost, u, obj)
}

func (h *Requester) Put(u string, body io.Reader) *Requester {
	h.m = http.MethodPut
	h.url = u
	h.body = body
	return h
}

func (h *Requester) PutJSON(u string, obj any) *Requester {
	return h.RequestJSON(http.MethodPut, u, obj)
}

func (h *Requester) Delete(u string, body io.Reader) *Requester {
	h.m = http.MethodDelete
	h.url = u
	h.body = body
	return h
}

func (h *Requester) DeleteJSON(u string, obj any) *Requester {
	return h.RequestJSON(http.MethodDelete, u, obj)
}

func (h *Requester) AddHead(key, value string) *Requester {
	h.header.Add(key, value)
	return h
}

func (h *Requester) SetHead(key, value string) *Requester {
	h.header.Set(key, value)
	return h
}

func (h *Requester) Response(ctx context.Context) (*http.Response, error) {
	if h.err != nil {
		return nil, h.err
	}

	req, err := http.NewRequestWithContext(ctx, h.m, h.url, h.body)
	if err != nil {
		return nil, err
	}
	req.Header = h.header.Clone()

	h.printRequestDebug(req)

	h.response, err = http.DefaultClient.Do(req)

	h.printResponseDebug(h.response)

	return h.response, nil
}

func (h *Requester) AsBytes(ctx context.Context) []byte {
	resp, err := h.Response(ctx)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return data
}

func (h *Requester) AsString(ctx context.Context) string {
	return bytes.ToString(h.AsBytes(ctx))
}

func (h *Requester) AsJSON(ctx context.Context, obj any) error {
	resp, err := h.Response(ctx)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(obj)
}

func (h *Requester) Clear() *Requester {
	if len(h.header) > 0 {
		h.header = make(http.Header)
	}
	h.err = nil
	h.url = ""
	h.m = ""
	h.body = nil
	h.response = nil
	return h
}

func (h *Requester) Clone() *Requester {
	new := NewRequester()
	new.header = h.header.Clone()
	new.url = h.url
	new.m = h.m
	new.body = h.body
	return new
}

func (h *Requester) Close() {
	if h.response != nil {
		h.response.Body.Close()
	}
	requesterPool.Put(h)
}

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
