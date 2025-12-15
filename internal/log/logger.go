package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var globalLogger *zap.Logger

// Init initializes the global logger
func Init(level string, jsonFormat bool) error {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	var config zap.Config
	if jsonFormat {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config.Level = zap.NewAtomicLevelAt(zapLevel)
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	logger, err := config.Build(
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return err
	}

	globalLogger = logger
	return nil
}

// Close flushes any buffered log entries
func Close() {
	if globalLogger != nil {
		_ = globalLogger.Sync()
	}
}

// WithFields returns a logger with additional context fields
func WithFields(fields ...zap.Field) *zap.Logger {
	if globalLogger == nil {
		// Fallback logger if not initialized
		globalLogger, _ = zap.NewProduction()
	}
	return globalLogger.With(fields...)
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	if globalLogger == nil {
		globalLogger, _ = zap.NewProduction()
	}
	globalLogger.Debug(msg, fields...)
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	if globalLogger == nil {
		globalLogger, _ = zap.NewProduction()
	}
	globalLogger.Info(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	if globalLogger == nil {
		globalLogger, _ = zap.NewProduction()
	}
	globalLogger.Warn(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	if globalLogger == nil {
		globalLogger, _ = zap.NewProduction()
	}
	globalLogger.Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	if globalLogger == nil {
		globalLogger, _ = zap.NewProduction()
	}
	globalLogger.Fatal(msg, fields...)
	os.Exit(1)
}

// Module returns a logger with module context
func Module(module string) *zap.Logger {
	return WithFields(zap.String("module", module))
}

// Session returns a logger with session context
func Session(sessionID string) *zap.Logger {
	return WithFields(zap.String("session_id", sessionID))
}

// Action returns a logger with action context
func Action(action string) *zap.Logger {
	return WithFields(zap.String("action", action))
}