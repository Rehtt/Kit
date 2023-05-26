package requester

import (
	"context"
	"github.com/Rehtt/Kit/bytes"
	"io"
	"net/http"
)

type Requester struct {
	url    string
	m      string
	header http.Header
	body   io.Reader
}

func NewRequester() *Requester {
	return &Requester{
		header: map[string][]string{},
	}
}
func (h *Requester) Get(u string) *Requester {
	h.url = u
	h.m = http.MethodGet
	return h
}
func (h *Requester) Post(u string, body io.Reader) *Requester {
	h.m = http.MethodPost
	h.url = u
	h.body = body
	return h
}
func (h *Requester) Put(u string, body io.Reader) *Requester {
	h.m = http.MethodPut
	h.url = u
	h.body = body
	return h
}
func (h *Requester) Delete(u string, body io.Reader) *Requester {
	h.m = http.MethodDelete
	h.url = u
	h.body = body
	return h
}
func (h *Requester) Head(key, value string) *Requester {
	h.header.Add(key, value)
	return h
}
func (h *Requester) Response(ctx context.Context) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, h.m, h.url, h.body)
	if err != nil {
		return nil, err
	}
	req.Header = h.header.Clone()
	return http.DefaultClient.Do(req)
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
