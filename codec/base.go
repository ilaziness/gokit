package codec

import (
	"math"
	"strings"

	"github.com/ilaziness/gokit/str"
)

// base62Chars 常量，包含数字、小写字母和大写字母
const (
	base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	base        = 62
)

// DecimalToBase62 将十进制数字转换为62进制字符串
func DecimalToBase62(num int64) string {
	if num == 0 {
		return string(base62Chars[0])
	}
	// 处理符号位
	sign := ""
	if num < 0 {
		sign = "-"
		num = -num
	}

	var result []byte
	for num > 0 {
		remainder := num % base
		result = append(result, base62Chars[remainder])
		num = num / base
	}

	return sign + str.ReverseString(string(result))
}

// Base62ToDecimal 将62进制字符串转换为十进制数字
func Base62ToDecimal(str string) int64 {
	// 处理符号位
	sign := int64(1)
	if len(str) > 0 && str[0] == '-' {
		sign = -1
		str = str[1:]
	}

	var num int64
	length := len(str)
	for i, char := range str {
		index := strings.IndexRune(base62Chars, char)
		if index == -1 {
			return 0
		}
		exponent := float64(length - 1 - i)
		num += int64(index) * int64(math.Pow(base, exponent))
	}
	return num * sign
}