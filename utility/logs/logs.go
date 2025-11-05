package logs

import (
	"context"

	"github.com/kevinmatthe/zaplog"
)

var L *Logger

func init() {
	L = &Logger{
		zlog: zaplog.NewLogger("/data/log/betago"),
	}
}

type Logger struct {
	zlog *zaplog.ZapLogger
}

func (l *Logger) Info(ctx context.Context, msg string, keysAndValues ...interface{}) {
	if len(keysAndValues) > 0 {
		l.zlog.Infow(msg, keysAndValues...)
	} else {
		l.zlog.Info(msg)
	}
}

func (l *Logger) Error(ctx context.Context, msg string, keysAndValues ...interface{}) {
	if len(keysAndValues) > 0 {
		l.zlog.Errorw(msg, keysAndValues...)
	} else {
		l.zlog.Error(msg)
	}
}

func (l *Logger) Debug(ctx context.Context, msg string, keysAndValues ...interface{}) {
	if len(keysAndValues) > 0 {
		l.zlog.Debugw(msg, keysAndValues...)
	} else {
		l.zlog.Debug(msg)
	}
}

func (l *Logger) Warn(ctx context.Context, msg string, keysAndValues ...interface{}) {
	if len(keysAndValues) > 0 {
		l.zlog.Warnw(msg, keysAndValues...)
	} else {
		l.zlog.Warn(msg)
	}
}

func (l *Logger) Fatal(ctx context.Context, msg string, keysAndValues ...interface{}) {
	if len(keysAndValues) > 0 {
		l.zlog.Fatalw(msg, keysAndValues...)
	} else {
		l.zlog.Fatal(msg)
	}
}