package env

import "os"

// GetEnvWithDefault 带默认值的获取方法
//
//	@param s
//	@param d
//	@return r
func GetEnvWithDefault(s, d string) (r string) {
	if r = os.Getenv(s); r != "" {
		return r
	}
	return d
}

// GetEnvWithDefaultGenerics 带默认值的获取方法
//
//	@param s string
//	@param d string
//	@return r string
//	@author heyuhengmatt
//	@update 2024-08-12 03:21:17
func GetEnvWithDefaultGenerics[T any](key string, defaultValue T, parseFunc func(string) T) (r T) {
	if rs := os.Getenv(key); rs != "" {
		return parseFunc(rs)
	}
	return defaultValue
}
