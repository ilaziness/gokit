package str

import (
	"testing"
)

// 测试字符串翻转功能
func TestReverseString(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"hello", "olleh"},
		{"", ""},
		{"中文", "文中"},
		{"12345", "54321"},
		{"aBcDef", "feDcBa"},
	}

	for _, tc := range testCases {
		t.Run("input="+tc.input, func(t *testing.T) {
			result := ReverseString(tc.input)
			if result != tc.expected {
				t.Errorf("ReverseString(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}
