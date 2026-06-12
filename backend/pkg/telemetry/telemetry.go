package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

func InitTelemetry(ctx context.Context, logger *zap.Logger) func() {
	traceExporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		logger.Fatal("Failed to create OTLP trace exporter", zap.Error(err))
	}

	res, err := sdkresource.New(ctx,
		sdkresource.WithFromEnv(), // reads OTEL_SERVICE_NAME, OTEL_RESOURCE_ATTRIBUTES
		sdkresource.WithProcess(),
		sdkresource.WithOS(),
		sdkresource.WithHost(),
	)
	if err != nil {
		logger.Warn("Failed to create OTel resource, using default", zap.Error(err))
		res = sdkresource.Default()
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logExporter, err := otlploggrpc.New(ctx)
	if err != nil {
		logger.Warn("Failed to create OTLP log exporter", zap.Error(err))
	} else {
		lp := sdklog.NewLoggerProvider(
			sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
			sdklog.WithResource(res),
		)
		global.SetLoggerProvider(lp)
	}

	logger.Info("Telemetry initialized")

	return func() {
		if err := tp.Shutdown(ctx); err != nil {
			logger.Error("Failed to shutdown tracer provider", zap.Error(err))
		}
	}
}
