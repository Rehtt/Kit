package completion

import (
	"testing"
)

func TestNormalizeCompletionFunc(t *testing.T) {
	tests := []struct {
		name     string
		fn       CompletionFunc
		input    string
		expected []CompletionItem
	}{
		{
			name: "func(string) []string",
			fn: func(s string) []string {
				return []string{"test1", "test2"}
			},
			input: "test",
			expected: []CompletionItem{
				{Value: "test1"},
				{Value: "test2"},
			},
		},
		{
			name: "func(string) []CompletionItem",
			fn: func(s string) []CompletionItem {
				return []CompletionItem{
					{Value: "item1", Description: "desc1"},
					{Value: "item2", Description: "desc2"},
				}
			},
			input: "item",
			expected: []CompletionItem{
				{Value: "item1", Description: "desc1"},
				{Value: "item2", Description: "desc2"},
			},
		},
		{
			name:     "invalid function type",
			fn:       "invalid",
			input:    "test",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalized := normalizeCompletionFunc(tt.fn)
			result := normalized(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d items, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i].Value != expected.Value || result[i].Description != expected.Description {
					t.Errorf("expected %+v, got %+v", expected, result[i])
				}
			}
		})
	}
}
