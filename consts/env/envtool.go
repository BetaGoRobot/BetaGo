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
