package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	sugar *zap.SugaredLogger
}

func NewLogger() *Logger {
	config := zap.NewProductionConfig()

	// Customize Time format
	config.EncoderConfig.TimeKey = "time"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Customize Duration format to milliseconds (e.g. "3.5ms")
	config.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder

	// Disable Caller
	config.DisableCaller = true

	logger, _ := config.Build()

	// Create a SugaredLogger to support printf-style logging
	sugar := logger.Sugar()

	return &Logger{
		sugar: sugar,
	}
}

// GetZapLogger returns the raw zap logger instance (for middleware usage)
func (l *Logger) GetZapLogger() *zap.Logger {
	return l.sugar.Desugar()
}

func (l *Logger) Info(v ...any) {
	l.sugar.Info(v...)
}

func (l *Logger) Error(v ...any) {
	l.sugar.Error(v...)
}

func (l *Logger) Warn(v ...any) {
	l.sugar.Warn(v...)
}

func (l *Logger) Debug(v ...any) {
	l.sugar.Debug(v...)
}

func (l *Logger) Fatal(v ...any) {
	l.sugar.Fatal(v...)
}

func (l *Logger) Infof(format string, v ...any) {
	l.sugar.Infof(format, v...)
}

func (l *Logger) Errorf(format string, v ...any) {
	l.sugar.Errorf(format, v...)
}

func (l *Logger) Warnf(format string, v ...any) {
	l.sugar.Warnf(format, v...)
}

func (l *Logger) Debugf(format string, v ...any) {
	l.sugar.Debugf(format, v...)
}

func (l *Logger) Fatalf(format string, v ...any) {
	l.sugar.Fatalf(format, v...)
}
