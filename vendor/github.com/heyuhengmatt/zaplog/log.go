package zaplog

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// // init logger
// func init() {
// 	Logger = NewLogger()
// 	SugarLogger = NewSugarLogger()
// }

// ZapLogger is a wrapper of zap.Logger
type ZapLogger struct {
	logger *zap.Logger
	level  zapcore.Level
}

// 全局logger
var (
	Logger      *ZapLogger
	SugarLogger *ZapLoggerSugar
)

// Level 日志级别别名
type Level = zapcore.Level

var (
	// DebugLevel debug level
	DebugLevel = zap.DebugLevel

	// InfoLevel info level
	InfoLevel = zap.InfoLevel

	// WarnLevel warn level
	WarnLevel = zap.WarnLevel

	// ErrorLevel error level
	ErrorLevel = zap.ErrorLevel

	// DPanicLevel dpanic level
	DPanicLevel = zap.DPanicLevel

	// PanicLevel panic level
	PanicLevel = zap.PanicLevel

	// FatalLevel fatal level
	FatalLevel = zap.FatalLevel
)

// 类型映射
var (
	String  = zap.String
	Strings = zap.Strings
	Stringp = zap.Stringp

	Byte        = zap.Binary
	ByteString  = zap.ByteString
	ByteStrings = zap.ByteStrings

	Int    = zap.Int
	Ints   = zap.Ints
	Int16  = zap.Int16
	Int16s = zap.Int16s
	Int32  = zap.Int32
	Int32s = zap.Int32s
	Int64  = zap.Int64
	Int64s = zap.Int64s
	Uint   = zap.Uint
	Uints  = zap.Uints

	Float64  = zap.Float64
	Float64s = zap.Float64s

	Bool  = zap.Bool
	Bools = zap.Bools

	Time  = zap.Time
	Times = zap.Times

	Duration  = zap.Duration
	Durations = zap.Durations

	Any = zap.Any

	Skip = zap.Skip

	Error = zap.Error
)

// Debug  debug
//
//	@receiver l
//	@param msg
//	@param fields
func (l *ZapLogger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

// Info info
//
//	@receiver l
//	@param msg
//	@param fields
func (l *ZapLogger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

// Warn warn
//
//	@receiver l
//	@param msg
//	@param fields
func (l *ZapLogger) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}

// Error error
//
//	@receiver l
//	@param msg
//	@param fields
func (l *ZapLogger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

// DPanic dpanic
//
//	@receiver l
//	@param msg
//	@param fields
func (l *ZapLogger) DPanic(msg string, fields ...zap.Field) {
	l.logger.DPanic(msg, fields...)
}

// Panic panic
//
//	@receiver l
//	@param msg
//	@param fields
func (l *ZapLogger) Panic(msg string, fields ...zap.Field) {
	l.logger.Panic(msg, fields...)
}

// Fatal  fatal
//
//	@receiver l *ZapLogger
//	@param msg string
//	@param fields ...zap.Field
//	@author kevinmatthe
func (l *ZapLogger) Fatal(msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, fields...)
}

// LogLevelSetter  set log level
//
//	@return map
func LogLevelSetter() map[string]Level {
	return map[string]zapcore.Level{
		"DEBUG":  DebugLevel,
		"INFO":   InfoLevel,
		"WARN":   WarnLevel,
		"ERROR":  ErrorLevel,
		"DPANIC": DPanicLevel,
		"PANIC":  PanicLevel,
		"FATAL":  FatalLevel,
	}
}

// NewLogger new logger
//
//	@param logBasePath
//	@return zapLogger
func NewLogger(logBasePath string) (zapLogger *ZapLogger) {
	if logBasePath == "" {
		logBasePath = "/data/logs/zaplog-default"
	}
	zapLogger = new(ZapLogger)
	zapLogger.logger = new(zap.Logger)
	defer zapLogger.Sync()
	var (
		logFileWriter, panicLogWriter, errLogWriter zapcore.WriteSyncer

		logLevel Level
		ok       bool
		cfg      zap.Config
	)
	if logLevel, ok = LogLevelSetter()[strings.ToUpper(os.Getenv("LOG_LEVEL"))]; !ok {
		// 未设置、不匹配则取默认值Info
		logLevel = InfoLevel
	}
	// 普通日志目录Rotation
	logFileWriter = zapcore.AddSync(&lumberjack.Logger{
		Filename:   logBasePath + ".log",
		MaxSize:    20, // 日志体积/MB
		MaxBackups: 3,
		MaxAge:     7, //日志保留天数
		Compress:   false,
	})
	// 错误日志目录Rotation
	errLogWriter = zapcore.AddSync(&lumberjack.Logger{
		Filename:   logBasePath + "_error.log",
		MaxSize:    20, // 日志体积/MB
		MaxBackups: 3,
		MaxAge:     7, //日志保留天数
		Compress:   false,
	})
	panicLogWriter = zapcore.AddSync(&lumberjack.Logger{
		Filename:   logBasePath + "_panic.log",
		MaxSize:    20, // 日志体积/MB
		MaxBackups: 3,
		MaxAge:     7, //日志保留天数
		Compress:   false,
	})
	if logLevel == DebugLevel {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}
	cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	cfg.EncoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	cfg.EncoderConfig.ConsoleSeparator = "  "
	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewConsoleEncoder(cfg.EncoderConfig), zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), logFileWriter), logLevel),
		//错误日志/Panic日志输出到err.log
		zapcore.NewCore(zapcore.NewConsoleEncoder(cfg.EncoderConfig), zapcore.NewMultiWriteSyncer(errLogWriter), ErrorLevel),
		//Panic日志输出到panic.log
		zapcore.NewCore(zapcore.NewConsoleEncoder(cfg.EncoderConfig), zapcore.NewMultiWriteSyncer(panicLogWriter), DPanicLevel),
	)
	zapLogger.logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(PanicLevel))
	return
}

// Sync sync
//
//	@receiver logger
//	@return error
func (l *ZapLogger) Sync() error {
	return l.logger.Sync()
}
