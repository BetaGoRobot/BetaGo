// internal/logging/logger.go
package logging

import (
	"os"

	"github.com/rs/zerolog"
)

var Logger = NewLogger()

// NewLogger 创建一个配置了 OtelHook 的 zerolog.Logger 实例。
func NewLogger() zerolog.Logger {
	// 设置全局日志级别
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// 创建一个 OtelHook 实例
	otelHook := SpanEventHook{}

	// 创建 logger 并注册 hook
	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger().
		Hook(otelHook)

	return logger
}
