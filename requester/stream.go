// Copyright (c) 2025 Rehtt
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// @Author: Rehtt dsreshiram@gmail.com
// @Date: 2025/11/25

package requester

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// HandleStream 流处理
// blankLine 可选参数，返回空行
func (h *Requester) HandleStream(ctx context.Context, f func(response *http.Response, data []byte) error, blankLine ...bool) error {
	resp, err := h.Response(ctx)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			raw := scanner.Bytes()
			if len(raw) == 0 && !(len(blankLine) > 0 && blankLine[0]) {
				continue
			}
			if err := f(resp, raw); err != nil {
				return err
			}
		}
	}
	if err := scanner.Err(); err != nil {
		if err != context.Canceled {
			return err
		}
	}

	return nil
}

// HandleEventStream 以阻塞处理SSE
func (h *Requester) HandleEventStream(ctx context.Context, f func(data []byte) error, blankLine ...bool) error {
	h.SetHead("Accept", "text/event-stream")
	return h.HandleStream(ctx, func(response *http.Response, data []byte) error {
		ct := response.Header.Get("Content-Type")
		if !strings.Contains(ct, "text/event-stream") {
			return fmt.Errorf("response header content-type not text/event-stream, got: %s", ct)
		}
		return f(data)
	}, blankLine...)
}

// AsEventStream 以channel方式处理
func (h *Requester) AsEventStream(ctx context.Context, blankLine ...bool) <-chan StreamResultBytes {
	ch := make(chan StreamResultBytes, 30)
	go func(ctx context.Context, ch chan StreamResultBytes) {
		defer close(ch)
		if err := h.HandleEventStream(ctx, func(raw []byte) error {
			data := make([]byte, len(raw))
			copy(data, raw)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case ch <- StreamResultBytes{Data: data}:
			}
			return nil
		}, blankLine...); err != nil {
			select {
			case <-ctx.Done():
			case ch <- StreamResultBytes{Err: err}:
			}
		}
	}(ctx, ch)
	return ch
}

// HandleJSONStream 处理NDJSON/SSE JSON
func HandleJSONStream[T any](ctx context.Context, r *Requester, f func(*T) error) error {
	r.SetHead("Accept", "application/x-ndjson, application/jsonl, application/json, text/event-stream")
	var buf bytes.Buffer
	var ctype string
	return r.HandleStream(ctx, func(response *http.Response, raw []byte) error {
		if ctype == "" {
			ct := response.Header.Get("Content-Type")
			isNDJSON := strings.Contains(ct, "application/x-ndjson") ||
				strings.Contains(ct, "application/jsonl") ||
				strings.Contains(ct, "application/stream+json") ||
				strings.Contains(ct, "application/json") // 兜底
			if isNDJSON {
				ctype = "ndjson"
			} else if strings.Contains(ct, "text/event-stream") {
				ctype = "event-stream"
			} else {
				return fmt.Errorf("response header content-type not application/x-ndjson, application/jsonl, application/json, text/event-stream, got: %s", ct)
			}
		}

		switch ctype {
		case "ndjson":
			if len(raw) == 0 {
				return nil
			}
			out := new(T)
			if err := json.Unmarshal(raw, out); err != nil {
				return err
			}
			return f(out)
		case "event-stream":
			if len(raw) == 0 {
				if buf.Len() > 0 {
					out := new(T)
					if err := json.Unmarshal(buf.Bytes(), out); err != nil {
						return err
					}
					buf.Reset()
					return f(out)
				}
				return nil
			}
			data, ok := bytes.CutPrefix(raw, []byte("data: "))
			if ok {
				buf.Write(data)
				buf.WriteByte('\n')
			}
			return nil
		}
		return nil
	}, true)
}

func AsJSONStream[T any](ctx context.Context, r *Requester) <-chan StreamResult[*T] {
	ch := make(chan StreamResult[*T], 30)
	go func(ctx context.Context, ch chan StreamResult[*T]) {
		defer close(ch)
		if err := HandleJSONStream(ctx, r, func(data *T) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case ch <- StreamResult[*T]{Data: data}:
			}
			return nil
		}); err != nil {
			select {
			case <-ctx.Done():
			case ch <- StreamResult[*T]{Err: err}:
			}
		}
	}(ctx, ch)
	return ch
}
