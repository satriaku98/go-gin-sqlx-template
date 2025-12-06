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

func (l *Logger) Info(format string, v ...interface{}) {
	l.sugar.Infof(format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.sugar.Errorf(format, v...)
}

func (l *Logger) Debug(format string, v ...interface{}) {
	l.sugar.Debugf(format, v...)
}

func (l *Logger) Fatal(format string, v ...interface{}) {
	l.sugar.Fatalf(format, v...)
}
