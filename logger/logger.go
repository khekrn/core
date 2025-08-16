// Package logger provides structured logging with context support,
// multiple output formats, and production-ready configuration using Zap.
//
// This package offers context-aware logging, automatic request ID inclusion,
// and different configurations for development and production environments.
//
// Example usage:
//
//	// Initialize logger
//	logger.InitLogger("info", "production")
//	defer logger.Sync()
//
//	// Basic logging
//	logger.Info("Application started")
//	logger.Error("Error occurred", zap.String("error", "connection failed"))
//
//	// Context-aware logging
//	ctx := context.WithValue(context.Background(), "RequestID", "req-123")
//	log := logger.FromContext(ctx)
//	log.Info("Processing request") // Automatically includes request_id
package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey int

const (
	loggerKey contextKey = iota
)

// Logger is the global logger instance
var Logger *zap.Logger

// InitLogger initializes the global logger with the specified log level and environment.
//
// logLevel: The minimum log level (debug, info, warn, error, fatal, panic)
// env: The environment type (development, production) - affects output format and features
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

// WithContext creates a new context with the specified logger instance
func WithContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext extracts a logger from the context. If no logger is found,
// it returns the global logger. If a RequestID is present in the context,
// it automatically adds it as a field to the logger.
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

// Info logs an info level message using the global logger
func Info(message string, fields ...zap.Field) {
	Logger.Info(message, fields...)
}

// Error logs an error level message using the global logger
func Error(message string, fields ...zap.Field) {
	Logger.Error(message, fields...)
}

// Debug logs a debug level message using the global logger
func Debug(message string, fields ...zap.Field) {
	Logger.Debug(message, fields...)
}

// Warn logs a warning level message using the global logger
func Warn(message string, fields ...zap.Field) {
	Logger.Warn(message, fields...)
}

// Fatal logs a fatal level message using the global logger and exits the program
func Fatal(message string, fields ...zap.Field) {
	Logger.Fatal(message, fields...)
}

// Sync flushes any buffered log entries. Should be called before program exit.
func Sync() error {
	return Logger.Sync()
}
