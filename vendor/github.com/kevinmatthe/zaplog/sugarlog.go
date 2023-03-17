package zaplog

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLoggerSugar is a wrapper of zap.SugaredLogger
type ZapLoggerSugar struct {
	logger *zap.SugaredLogger
	level  zapcore.Level
}

// Debug  debug
//  @receiver l
//  @param msg
//  @param fields
func (l *ZapLoggerSugar) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(fields)
}

// Debugf debugf
//  @receiver l
//  @param msg
//  @param fields
func (l *ZapLoggerSugar) Debugf(template string, args ...interface{}) {
	l.logger.Debugf(template, args...)
}

// Info info
//  @receiver l
//  @param msg
//  @param fields
func (l *ZapLoggerSugar) Info(msg string, fields ...zap.Field) {
	l.logger.Info(fields)
}

// Infof infof
//  @receiver l
//  @param msg
//  @param fields
func (l *ZapLoggerSugar) Infof(template string, args ...interface{}) {
	l.logger.Infof(template, args...)
}

// Warning warn
//  @receiver l
//  @param msg
//  @param fields
func (l *ZapLoggerSugar) Warning(msg string, fields ...zap.Field) {
	l.logger.Warn(fields)
}

// Warningf warnf
//  @receiver l
//  @param msg
//  @param fields
func (l *ZapLoggerSugar) Warningf(template string, args ...interface{}) {
	l.logger.Warnf(template, args...)
}

// Error error
//  @receiver l
//  @param msg
//  @param fields
func (l *ZapLoggerSugar) Error(msg string, fields ...zap.Field) {
	l.logger.Error(fields)
}

// Errorf errorf
//  @receiver l
//  @param template
//  @param args
func (l *ZapLoggerSugar) Errorf(template string, args ...interface{}) {
	l.logger.Errorf(template, args...)
}

// DPanic dpanic
//  @receiver l
//  @param msg
//  @param fields
func (l *ZapLoggerSugar) DPanic(msg string, fields ...zap.Field) {
	l.logger.DPanic(fields)
}

// DPanicf panic
//  @receiver l
//  @param msg
//  @param fields
func (l *ZapLoggerSugar) DPanicf(template string, args ...interface{}) {
	l.logger.DPanicf(template, args...)
}

// Panic panic
//  @receiver l
//  @param msg
//  @param fields
func (l *ZapLoggerSugar) Panic(msg string, fields ...zap.Field) {
	l.logger.Panic(fields)
}

// Panicf panic
//  @receiver l
//  @param msg
//  @param fields
func (l *ZapLoggerSugar) Panicf(template string, args ...interface{}) {
	l.logger.Panicf(template, args...)
}

// Emergencyf 兼容性方法，不建议在新的模块中使用
//  @receiver l *ZapLogger
//  @param format string
//  @param args ...interface{}
func (l *ZapLoggerSugar) Emergencyf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

// NewSugarLogger new logger
//  @return ZapLogger
func NewSugarLogger(Logger *ZapLogger) (zapLogger *ZapLoggerSugar) {
	return &ZapLoggerSugar{
		logger: Logger.logger.Sugar(),
		level:  Logger.level,
	}
}

// Sync sync1
//  @receiver l *ZapLoggerSugar
func (l *ZapLoggerSugar) Sync() {
	l.logger.Sync()
}
