// Copyright (c) 2025 Rehtt
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// @Author: Rehtt dsreshiram@gmail.com
// @Date: 2025/11/21

package util

import "errors"

type Error[T any] struct {
	error
}

func (e *Error[T]) Error() string {
	return e.error.Error()
}

func (e *Error[T]) Unwrap() error {
	return e.error
}

func MarkAsError[T any](err error) error {
	if err == nil {
		return nil
	}
	if IsError[T](err) {
		return err
	}
	return &Error[T]{err}
}

// AsError 递归穿透识别
func AsError[T any](err error) bool {
	var target *Error[T]
	return errors.As(err, &target)
}

// IsError 只识别顶层
func IsError[T any](err error) bool {
	_, ok := err.(*Error[T])
	return ok
}

// UnwrapError 只解包顶层
func UnwrapError[T any](err error) error {
	if e, ok := err.(*Error[T]); ok {
		return e.error
	}
	return err
}
