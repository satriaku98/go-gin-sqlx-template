package logger

import (
	"context"

	"go.opentelemetry.io/otel/trace"
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

// withTraceID adds trace_id field to logger if available in context
func (l *Logger) withTraceID(ctx context.Context) *zap.SugaredLogger {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		traceID := span.SpanContext().TraceID().String()
		return l.sugar.With("trace_id", traceID)
	}
	return l.sugar
}

// WithFields creates a new logger with additional fields
// This is useful for adding context-specific fields without affecting the base logger
func (l *Logger) WithFields(ctx context.Context, fields map[string]any) *Logger {
	logger := l.withTraceID(ctx)

	// Convert map to key-value pairs for With()
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}

	return &Logger{
		sugar: logger.With(args...),
	}
}

func (l *Logger) Info(ctx context.Context, v ...any) {
	l.withTraceID(ctx).Info(v...)
}

func (l *Logger) Error(ctx context.Context, v ...any) {
	l.withTraceID(ctx).Error(v...)
}

func (l *Logger) Warn(ctx context.Context, v ...any) {
	l.withTraceID(ctx).Warn(v...)
}

func (l *Logger) Debug(ctx context.Context, v ...any) {
	l.withTraceID(ctx).Debug(v...)
}

func (l *Logger) Fatal(ctx context.Context, v ...any) {
	l.withTraceID(ctx).Fatal(v...)
}

func (l *Logger) Infof(ctx context.Context, format string, v ...any) {
	l.withTraceID(ctx).Infof(format, v...)
}

func (l *Logger) Errorf(ctx context.Context, format string, v ...any) {
	l.withTraceID(ctx).Errorf(format, v...)
}

func (l *Logger) Warnf(ctx context.Context, format string, v ...any) {
	l.withTraceID(ctx).Warnf(format, v...)
}

func (l *Logger) Debugf(ctx context.Context, format string, v ...any) {
	l.withTraceID(ctx).Debugf(format, v...)
}

func (l *Logger) Fatalf(ctx context.Context, format string, v ...any) {
	l.withTraceID(ctx).Fatalf(format, v...)
}
