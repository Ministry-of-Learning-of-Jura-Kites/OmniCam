import logging
import os

from opentelemetry import trace, propagate
from opentelemetry.sdk.resources import Resource, SERVICE_NAME

from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor

from opentelemetry.exporter.otlp.proto.http.trace_exporter import (
    OTLPSpanExporter,
)

from opentelemetry.sdk._logs import (
    LoggerProvider,
    LoggingHandler,
)

from opentelemetry.sdk._logs.export import (
    BatchLogRecordProcessor,
)

from opentelemetry.exporter.otlp.proto.http._log_exporter import (
    OTLPLogExporter,
)

from opentelemetry.trace.propagation.tracecontext import (
    TraceContextTextMapPropagator,
)


def init_telemetry():
    service_name = os.getenv("OTEL_SERVICE_NAME", "omnicam-algo")

    resource = Resource(
        attributes={
            SERVICE_NAME: service_name,
            "deployment.environment": os.getenv("OTEL_ENVIRONMENT", "production"),
            "service.version": os.getenv("OTEL_SERVICE_VERSION", "latest"),
        }
    )

    exporter = OTLPSpanExporter(
        endpoint=os.getenv(
            "OTEL_EXPORTER_OTLP_ENDPOINT", "http://otelcol:4318/v1/traces"
        ),
        timeout=10,
    )

    provider = TracerProvider(resource=resource)
    provider.add_span_processor(
        BatchSpanProcessor(
            exporter,
            max_queue_size=2048,
            max_export_batch_size=512,
        )
    )

    trace.set_tracer_provider(provider)

    log_provider = LoggerProvider(resource=resource)

    log_exporter = OTLPLogExporter(
        endpoint=os.getenv(
            "OTEL_EXPORTER_OTLP_LOGS_ENDPOINT",
            "http://otelcol:4318/v1/logs",
        ),
        timeout=10,
    )

    log_provider.add_log_record_processor(BatchLogRecordProcessor(log_exporter))

    otel_handler = LoggingHandler(
        logger_provider=log_provider,
        level=logging.INFO,
    )

    root_logger = logging.getLogger()

    # Prevent duplicate handlers during reloads/tests
    if not any(isinstance(h, LoggingHandler) for h in root_logger.handlers):
        root_logger.addHandler(otel_handler)

    # REQUIRED: ensures tracecontext propagation across NATS / HTTP
    propagator = TraceContextTextMapPropagator()
    propagate.set_global_textmap(propagator)

    tracer = trace.get_tracer(service_name)

    return tracer
