package logger

import (
	"sync"

	"github.com/nghiavan0610/btaskee-quiz-service/config"
	"go.uber.org/zap"
)

type Logger struct {
	zapLogger *zap.Logger
}

var (
	logger     *Logger
	loggerOnce sync.Once
)

func NewLogger(zapLogger *zap.Logger) *Logger {
	return &Logger{zapLogger}
}

func ProvideLogger(config *config.Config) *Logger {
	loggerOnce.Do(func() {
		isProd := config.Server.GoEnv == "production"

		var zapLogger *zap.Logger
		var err error
		if isProd {
			cfg := zap.NewProductionConfig()
			cfg.DisableCaller = true
			cfg.DisableStacktrace = true // Disable stack traces in production
			zapLogger, err = cfg.Build()
		} else {
			cfg := zap.NewDevelopmentConfig()
			cfg.DisableCaller = true
			cfg.DisableStacktrace = true                     // Disable stack traces in development
			cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel) // Enable debug level for development
			zapLogger, err = cfg.Build()
		}
		if err != nil {
			panic(err)
		}

		logger = NewLogger(zapLogger)
	})

	return logger
}

func (l *Logger) Info(message string, params ...interface{}) {
	if len(params) > 0 {
		l.zapLogger.Info(message, zap.Any("params:", params))
		return
	}
	l.zapLogger.Info(message)
}

func (l *Logger) Error(message string, err interface{}) {
	if errMsg, ok := err.(string); ok {
		l.zapLogger.Error(message, zap.String("err", errMsg))
	} else if errObj, ok := err.(error); ok {
		l.zapLogger.Error(message, zap.String("err:", errObj.Error()))
	}
}

func (l *Logger) Errors(message string, err ...interface{}) {
	if len(err) > 0 {
		l.zapLogger.Info(message, zap.Any("errors: ", err))
		return
	}
	l.zapLogger.Info(message)
}

func (l *Logger) Warn(message string, params ...interface{}) {
	if len(params) > 0 {
		l.zapLogger.Warn(message, zap.Any("params:", params))
		return
	}
	l.zapLogger.Warn(message)
}

// WarnFields logs a warning with structured fields (no stack trace)
func (l *Logger) WarnFields(message string, fields map[string]interface{}) {
	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}
	l.zapLogger.Warn(message, zapFields...)
}

// ErrorFields logs an error with structured fields
func (l *Logger) ErrorFields(message string, fields map[string]interface{}) {
	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}
	l.zapLogger.Error(message, zapFields...)
}

// Debug logs a debug message
func (l *Logger) Debug(message string, params ...interface{}) {
	if len(params) > 0 {
		l.zapLogger.Debug(message, zap.Any("params:", params))
		return
	}
	l.zapLogger.Debug(message)
}

// DebugFields logs debug information with structured fields
func (l *Logger) DebugFields(message string, fields map[string]interface{}) {
	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}
	l.zapLogger.Debug(message, zapFields...)
}

func (l *Logger) InfoWithMask(message, secret string) {
	len := len(secret)
	if len > 0 && len <= 8 {
		secret = "*************"
	} else if len > 8 {
		secret = "*************" + secret[len-4:]
	}
	l.zapLogger.Info(message, zap.String("data", secret))
}
