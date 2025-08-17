package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger() *zap.Logger {
	encoderConfig := zap.NewProductionEncoderConfig()

	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	// Create a console encoder (plain text)
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	writerSyncer := zapcore.AddSync(os.Stdout)

	core := zapcore.NewCore(encoder, writerSyncer, zap.DebugLevel)

	logger := zap.New(core, zap.AddCaller())

	return logger
}
