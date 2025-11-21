// Copyright (c) 2025 Rehtt
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// @Author: Rehtt dsreshiram@gmail.com
// @Date: 2025/11/21

// Generate by Gemini3-pro

package util

import (
	"errors"
	"fmt"
	"testing"
)

// 定义两个不同的类型用于测试泛型隔离
type (
	DatabaseTag struct{}
	NetworkTag  struct{}
)

func TestMarkAsError(t *testing.T) {
	baseErr := errors.New("original error")

	t.Run("基本标记功能", func(t *testing.T) {
		// 标记为 DatabaseTag 错误
		err := MarkAsError[DatabaseTag](baseErr)

		// 验证不为 nil
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		// 验证 AsError 返回 true
		if !AsError[DatabaseTag](err) {
			t.Error("AsError[DatabaseTag] should be true")
		}

		// 验证错误信息保留
		if err.Error() != "original error" {
			t.Errorf("expected 'original error', got '%s'", err.Error())
		}
	})

	t.Run("Nil安全检查", func(t *testing.T) {
		err := MarkAsError[DatabaseTag](nil)
		if err != nil {
			t.Error("expected nil, got error")
		}
	})

	t.Run("避免重复包装 (优化验证)", func(t *testing.T) {
		err1 := MarkAsError[DatabaseTag](baseErr)
		// 再次标记同一个错误
		err2 := MarkAsError[DatabaseTag](err1)

		// 关键点：指针地址应该完全相同，说明没有创建新的 Error 结构体
		if err1 != err2 {
			t.Errorf("MarkAsError should return the original object if type matches. Got %p and %p", err1, err2)
		}
	})
}

func TestTypeIsolation(t *testing.T) {
	baseErr := errors.New("boom")

	// 标记为 NetworkTag
	netErr := MarkAsError[NetworkTag](baseErr)

	// 验证它是 NetworkTag
	if !AsError[NetworkTag](netErr) {
		t.Error("Should be NetworkTag")
	}

	// 验证它不是 DatabaseTag
	if AsError[DatabaseTag](netErr) {
		t.Error("Should NOT be DatabaseTag")
	}
}

func TestUnwrapError(t *testing.T) {
	baseErr := errors.New("root cause")
	markedErr := MarkAsError[DatabaseTag](baseErr)

	t.Run("解包顶层错误", func(t *testing.T) {
		unwrapped := UnwrapError[DatabaseTag](markedErr)
		if unwrapped != baseErr {
			t.Errorf("Expected original error, got %v", unwrapped)
		}
	})

	t.Run("解包非匹配类型", func(t *testing.T) {
		// 尝试用 NetworkTag 解包 DatabaseTag 错误
		unwrapped := UnwrapError[NetworkTag](markedErr)
		// 应该原样返回，不做解包
		if unwrapped != markedErr {
			t.Error("UnwrapError should return original error if type mismatch")
		}
	})

	t.Run("解包普通错误", func(t *testing.T) {
		unwrapped := UnwrapError[DatabaseTag](baseErr)
		if unwrapped != baseErr {
			t.Error("UnwrapError should return original error if it's not wrapped")
		}
	})
}

func TestNestedErrors(t *testing.T) {
	baseErr := errors.New("db down")
	markedErr := MarkAsError[DatabaseTag](baseErr)

	// 模拟标准库 wrap： fmt.Errorf 包装了 markedErr
	wrappedErr := fmt.Errorf("request failed: %w", markedErr)

	t.Run("AsError 能够穿透标准库 Wrap", func(t *testing.T) {
		// 即使外面套了一层 fmt.Errorf，AsError 依然应该能利用 errors.As 找到内部的 DatabaseTag
		if !AsError[DatabaseTag](wrappedErr) {
			t.Error("AsError should find deeply nested generic errors")
		}
	})

	t.Run("UnwrapError 仅解包顶层", func(t *testing.T) {
		// UnwrapError 只检查最外层。因为最外层是 fmt.wrapError，不是 Error[T]，所以不应该解包
		result := UnwrapError[DatabaseTag](wrappedErr)
		if result != wrappedErr {
			t.Error("UnwrapError should NOT unwrap nested errors, only top-level")
		}
	})
}

// --- 性能测试 ---

// 定义一个全局变量，防止编译器优化掉返回值
var globalErr error

func BenchmarkMarkAsError_New(b *testing.B) {
	err := errors.New("base")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 每次都分配新包装
		globalErr = MarkAsError[DatabaseTag](err)
	}
}

func BenchmarkMarkAsError_Existing(b *testing.B) {
	err := errors.New("base")
	marked := MarkAsError[DatabaseTag](err)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 命中类型断言优化：不分配内存，O(1) 返回
		globalErr = MarkAsError[DatabaseTag](marked)
	}
}

func BenchmarkAsError(b *testing.B) {
	err := MarkAsError[DatabaseTag](errors.New("base"))
	// 模拟稍微深一点的嵌套
	wrapped := fmt.Errorf("w: %w", fmt.Errorf("w: %w", err))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AsError[DatabaseTag](wrapped)
	}
}

func BenchmarkUnwrapError_Hit(b *testing.B) {
	err := MarkAsError[DatabaseTag](errors.New("base"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		globalErr = UnwrapError[DatabaseTag](err)
	}
}
