package log

import "github.com/kevinmatthe/zaplog"

// var  ZapLog日志变量
var (
	ZapLogger   *zaplog.ZapLogger
	SugerLogger *zaplog.ZapLoggerSugar
)

// InitLogger 初始化日志
func init() {
	if ZapLogger == nil {
		ZapLogger = zaplog.NewLogger("/data/log/betago")
	}
	if SugerLogger == nil {
		SugerLogger = zaplog.NewSugarLogger(ZapLogger)
	}
}
