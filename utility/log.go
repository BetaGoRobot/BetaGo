package utility

import "github.com/heyuhengmatt/zaplog"

var (
	ZapLogger   *zaplog.ZapLogger
	SugerLogger *zaplog.ZapLoggerSugar
)

// InitLogger 初始化日志
func InitLogger() {
	ZapLogger = zaplog.NewLogger("/data/log/betago")
	SugerLogger = zaplog.NewSugarLogger(ZapLogger)
}
