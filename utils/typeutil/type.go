package typeutil

// SliceMap 遍历处理slice的每个元素，结果替换原slice对应的元素
func SliceMap[T any](slice []T, fn func(T) T) []T {
	for i, v := range slice {
		slice[i] = fn(v)
	}
	return slice
}

// Slice2Map 遍历处理slice的每个元素，结果作为map的key，原slice的元素作为value
func Slice2Map[T any, K comparable](slice []T, fn func(T) K) map[K]T {
	m := make(map[K]T)
	for _, v := range slice {
		m[fn(v)] = v
	}
	return m
}
