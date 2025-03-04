package log

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey int

const (
	loggerKey contextKey = iota
)

var Logger *zap.Logger

// InitLogger initializes the logger with the specified log level.
func InitLogger(logLevel, env string) {
	var err error

	// Set default log level to InfoLevel
	level := zapcore.InfoLevel

	if logLevel != "" {
		var lvl zapcore.Level
		if err := lvl.UnmarshalText([]byte(logLevel)); err == nil {
			level = lvl
		}
	}

	zapCfg := zap.Config{
		Level:             zap.NewAtomicLevelAt(level),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: env == "production",
		Encoding:          "console",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:          "timestamp",
			LevelKey:         "level",
			NameKey:          "logger",
			CallerKey:        "caller",
			MessageKey:       "message",
			StacktraceKey:    "stacktrace",
			LineEnding:       zapcore.DefaultLineEnding,
			EncodeLevel:      zapcore.CapitalColorLevelEncoder,
			EncodeTime:       zapcore.ISO8601TimeEncoder,
			EncodeDuration:   zapcore.StringDurationEncoder,
			EncodeCaller:     zapcore.ShortCallerEncoder,
			ConsoleSeparator: " | ",
		},
		OutputPaths:      []string{"stdout", "/tmp/logs"},
		ErrorOutputPaths: []string{"stderr"},
	}

	Logger, err = zapCfg.Build(zap.AddCaller(), zap.AddCallerSkip(1))
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
}

// WithContext creates a new context with the logger
func WithContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext extracts the logger from context
func FromContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return Logger
	}

	// First check for logger directly in context
	if logger, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
		return logger
	}

	// Fallback to adding request ID if available
	if requestID, ok := ctx.Value("RequestID").(string); ok && requestID != "" {
		return Logger.With(zap.String("request_id", requestID))
	}

	return Logger
}

// Info logs an info message with additional context fields.
func Info(message string, fields ...zap.Field) {
	Logger.Info(message, fields...)
}

// Error logs an error message with additional context fields.
func Error(message string, fields ...zap.Field) {
	Logger.Error(message, fields...)
}

// Debug logs a debug message with additional context fields.
func Debug(message string, fields ...zap.Field) {
	Logger.Debug(message, fields...)
}

// Warn logs a warning message with additional context fields.
func Warn(message string, fields ...zap.Field) {
	Logger.Warn(message, fields...)
}

// Fatal logs a fatal message with additional context fields.
func Fatal(message string, fields ...zap.Field) {
	Logger.Fatal(message, fields...)
}

// Sync flushes any buffered log entries.
func Sync() error {
	return Logger.Sync()
}
