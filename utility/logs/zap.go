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
	return logger.Ctx(context.Background()) // 默认都要搞一个context出来
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
	stdoutLogger := zap.New(stdoutCore, zap.AddCaller(), zap.AddCallerSkip(1))

	// 组合
	logger = NewContextualLogger(stdoutLogger, otelLogger)
}

// ContextualLogger ：双写封装
type ContextualLogger struct {
	stdout     *zap.Logger
	otel       *zap.Logger
	withFields []zap.Field
}

func NewContextualLogger(stdoutLogger, otelLogger *zap.Logger) *ContextualLogger {
	return &ContextualLogger{stdout: stdoutLogger, otel: otelLogger, withFields: make([]zap.Field, 0)}
}

func (l *ContextualLogger) Ctx(ctx context.Context) *ContextualLogger {
	if ctx == nil {
		return l
	}
	var traceID, spanID string
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasSpanID() && spanCtx.HasTraceID() {
		traceID = spanCtx.TraceID().String()
		spanID = spanCtx.SpanID().String()
	}
	stdoutWithFields := append([]zap.Field{
		zap.String("trace_id", traceID),
		zap.String("span_id", spanID),
	}, l.withFields...)
	// stdout 不带 context，只带 trace/span id
	stdoutWithTrace := l.stdout.With(stdoutWithFields...)

	// otel 需要 context
	otelWithFields := append([]zap.Field{zap.Any("context", ctx)}, l.withFields...)
	otelWithCtx := l.otel.With(otelWithFields...)

	return &ContextualLogger{
		stdout:     stdoutWithTrace,
		otel:       otelWithCtx,
		withFields: l.withFields,
	}
}

func (l *ContextualLogger) With(fields ...zap.Field) *ContextualLogger {
	return &ContextualLogger{
		stdout:     l.stdout,
		otel:       l.otel,
		withFields: fields,
	}
}

func (l *ContextualLogger) Debug(msg string, fields ...zap.Field) {
	l.stdout.Debug(msg, append(l.withFields, fields...)...)
	l.otel.Debug(msg, fields...)
}

func (l *ContextualLogger) Info(msg string, fields ...zap.Field) {
	fields = append(fields, zap.String("level", "INFO"))
	l.stdout.Info(msg, fields...)
	l.otel.Info(msg, fields...)
}

func (l *ContextualLogger) Error(msg string, fields ...zap.Field) {
	fields = append(fields, zap.String("level", "ERROR"))
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
