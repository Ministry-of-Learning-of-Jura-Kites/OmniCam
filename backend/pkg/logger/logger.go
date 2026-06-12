package logger

import (
	"context"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger(isTest bool) *zap.Logger {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	encoderConfig.TimeKey = "ts"
	encoderConfig.MessageKey = "msg"
	encoderConfig.LevelKey = "level"
	encoderConfig.CallerKey = "caller"

	// stdout core — keeps logs visible in docker logs
	stdoutCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zap.DebugLevel,
	)

	// OTel core — sends logs to collector → Loki
	otelCore := otelzap.NewCore("omnicam-backend",
		otelzap.WithLoggerProvider(global.GetLoggerProvider()),
	)

	// tee to both
	core := zapcore.NewTee(stdoutCore, otelCore)
	logger := zap.New(core, zap.AddCaller())

	if isTest {
		logger = logger.With(zap.String("source", "test"))
	}

	return logger
}
func WithTraceID(ctx context.Context, logger *zap.Logger) *zap.Logger {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return logger
	}
	return logger.With(
		zap.String("trace_id", span.SpanContext().TraceID().String()),
		zap.String("span_id", span.SpanContext().SpanID().String()),
	)
}
