package str

import (
	"strconv"
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

// 测试随机字符串生成功能
func TestRandomString(t *testing.T) {
	testCases := []struct {
		length   int
		expected bool // 是否期望生成非空字符串
	}{
		{0, false},
		{5, true},
		{10, true},
		{-1, false}, // 负数长度应返回空字符串
	}

	for _, tc := range testCases {
		t.Run("length="+strconv.Itoa(tc.length), func(t *testing.T) {
			result := RandomString(tc.length)
			if tc.expected && len(result) != tc.length {
				t.Errorf("RandomString(%d) generated string of length %d, expected %d", tc.length, len(result), tc.length)
			}
			if !tc.expected && result != "" {
				t.Errorf("RandomString(%d) generated non-empty string: %q", tc.length, result)
			}

			// 校验生成的字符串是否只包含预定义字符集中的字符
			const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
			for _, c := range result {
				if !contains(charset, byte(c)) {
					t.Errorf("RandomString(%d) contains invalid character: %c", tc.length, c)
				}
			}

			// 校验多次调用生成的结果是否唯一
			if tc.expected {
				results := make(map[string]bool)
				for i := 0; i < 10; i++ { // 多次调用以增加概率检测
					str := RandomString(tc.length)
					if results[str] {
						t.Errorf("RandomString(%d) generated duplicate string: %q", tc.length, str)
					}
					results[str] = true
				}
			}
		})
	}
}

// 辅助函数：判断字符是否在字符串中
func contains(s string, c byte) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return true
		}
	}
	return false
}