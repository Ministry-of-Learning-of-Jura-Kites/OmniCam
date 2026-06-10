package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

func InitTelemetry(ctx context.Context, logger *zap.Logger) func() {
	exporter, err := otlptracegrpc.New(ctx)
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
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logger.Info("Telemetry initialized")

	return func() {
		if err := tp.Shutdown(ctx); err != nil {
			logger.Error("Failed to shutdown tracer provider", zap.Error(err))
		}
	}
}
