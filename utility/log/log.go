package log

import "github.com/kevinmatthe/zaplog"

// var  ZapLog日志变量
var (
	Zlog *zaplog.ZapLogger
	SLog *zaplog.ZapLoggerSugar
)

// InitLogger 初始化日志
func init() {
	if Zlog == nil {
		Zlog = zaplog.NewLogger("/data/log/betago")
	}
	if SLog == nil {
		SLog = zaplog.NewSugarLogger(Zlog)
	}
}
