package str

import (
	"crypto/rand"
)

// ReverseString 翻转输入字符串
func ReverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// RandomString 生成指定长度的随机字符串
func RandomString(length int) string {
	if length <= 0 {
		return ""
	}

	// 使用 crypto/rand 生成加密安全的随机字节
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(err) // 处理随机数生成失败的情况
	}

	// 将随机字节映射到字符集
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}
