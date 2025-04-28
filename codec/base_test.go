package codec

import (
	"testing"
)

func TestDecimalToBase62(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{"Zero", 0, "0"},
		{"SingleDigit", 10, "a"},
		{"MultipleDigits", 62, "10"},
		{"LargeNumber", 3844, "100"},
		{"VeryLargeNumber", 238327, "ZZZ"},
		{"MixedCase", 238328, "1000"},
		{"NegativeNumber", -1, "-1"},
		{"NegativeLargeNumber", -238328, "-1000"},
		{"NegativeLargeNumber", 136745, "zzz"},
		// 新增负数测试用例
		{"NegativeSingleDigit", -10, "-a"},
		{"NegativeMultipleDigits", -62, "-10"},
		{"NegativeVeryLargeNumber", -238327, "-ZZZ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DecimalToBase62(tt.input)
			if result != tt.expected {
				t.Errorf("DecimalToBase62(%d) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBase62ToDecimal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{"Zero", "0", 0},
		{"SingleDigit", "a", 10},
		{"MultipleDigits", "10", 62},
		{"LargeNumber", "100", 3844},
		{"VeryLargeNumber", "zzz", 136745},
		{"InvalidCharacter", "zz@", 0},
		{"MixedCase", "1000", 238328},
		{"EmptyString", "", 0},
		{"LeadingZeros", "0001", 1},
		{"NegativeNumber", "-1", -1},
		{"NegativeLargeNumber", "-1000", -238328},
		// 新增负数测试用例
		{"NegativeSingleDigit", "-a", -10},
		{"NegativeMultipleDigits", "-10", -62},
		{"NegativeVeryLargeNumber", "-zzz", -136745},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Base62ToDecimal(tt.input)
			if result != tt.expected {
				t.Errorf("Base62ToDecimal(%s) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}
