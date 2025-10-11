package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new logger instance
func New(level string, environment string) (*zap.Logger, error) {
	var zapConfig zap.Config

	if environment == "production" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Parse log level
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}
	zapConfig.Level = zap.NewAtomicLevelAt(zapLevel)

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}
