package logs

import (
	"context"
	"os"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *ContextualLogger

func L() *ContextualLogger {
	return logger
}

func init() {
	otelCore := otelzap.NewCore(consts.BotIdentifier, otelzap.WithLoggerProvider(otel.LoggerProvider()))
	otelLogger := zap.New(otelCore, zap.AddCaller())

	// Stdout logger
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.TimeKey = "time"
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(encCfg)
	stdoutCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zap.InfoLevel)
	stdoutLogger := zap.New(stdoutCore, zap.AddCaller())

	// 组合
	logger = NewContextualLogger(stdoutLogger, otelLogger)
}

// ContextualLogger ：双写封装
type ContextualLogger struct {
	stdout *zap.Logger
	otel   *zap.Logger
}

func NewContextualLogger(stdoutLogger, otelLogger *zap.Logger) *ContextualLogger {
	return &ContextualLogger{stdout: stdoutLogger, otel: otelLogger}
}

func (l *ContextualLogger) Ctx(ctx context.Context) *ContextualLogger {
	if ctx == nil {
		return l
	}

	spanCtx := trace.SpanContextFromContext(ctx)
	traceID := spanCtx.TraceID().String()
	spanID := spanCtx.SpanID().String()

	// stdout 不带 context，只带 trace/span id
	stdoutWithTrace := l.stdout.With(
		zap.String("trace_id", traceID),
		zap.String("span_id", spanID),
	)

	// otel 需要 context
	otelWithCtx := l.otel.With(zap.Any("context", ctx))

	return &ContextualLogger{
		stdout: stdoutWithTrace,
		otel:   otelWithCtx,
	}
}

func (l *ContextualLogger) Debug(msg string, fields ...zap.Field) {
	l.stdout.Debug(msg, fields...)
	l.otel.Debug(msg, fields...)
}

func (l *ContextualLogger) Info(msg string, fields ...zap.Field) {
	l.stdout.Info(msg, fields...)
	l.otel.Info(msg, fields...)
}

func (l *ContextualLogger) Error(msg string, fields ...zap.Field) {
	l.stdout.Error(msg, fields...)
	l.otel.Error(msg, fields...)
}

func (l *ContextualLogger) Warn(msg string, fields ...zap.Field) {
	l.stdout.Warn(msg, fields...)
	l.otel.Warn(msg, fields...)
}

func (l *ContextualLogger) Panic(msg string, fields ...zap.Field) {
	l.stdout.Panic(msg, fields...)
	l.otel.Panic(msg, fields...)
}

func (l *ContextualLogger) DPanic(msg string, fields ...zap.Field) {
	l.stdout.DPanic(msg, fields...)
	l.otel.DPanic(msg, fields...)
}

func (l *ContextualLogger) Fatal(msg string, fields ...zap.Field) {
	l.stdout.Fatal(msg, fields...)
	l.otel.Fatal(msg, fields...)
}

func (l *ContextualLogger) Sync() {
	_ = l.stdout.Sync()
	_ = l.otel.Sync()
}
