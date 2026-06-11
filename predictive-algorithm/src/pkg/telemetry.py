import os
from opentelemetry import trace, propagate
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.resources import Resource, SERVICE_NAME
from opentelemetry.trace.propagation.tracecontext import TraceContextTextMapPropagator


def init_telemetry():
    resource = Resource(
        attributes={
            SERVICE_NAME: os.getenv("OTEL_SERVICE_NAME", "omnicam-algo"),
            "deployment.environment": os.getenv("OTEL_ENVIRONMENT", "production"),
            "service.version": os.getenv("OTEL_SERVICE_VERSION", "latest"),
        }
    )

    exporter = OTLPSpanExporter(
        endpoint=os.getenv(
            "OTEL_EXPORTER_OTLP_ENDPOINT", "http://otelcol:4318/v1/traces"
        ),
    )

    provider = TracerProvider(resource=resource)
    provider.add_span_processor(BatchSpanProcessor(exporter))
    trace.set_tracer_provider(provider)
    propagate.set_global_textmap(TraceContextTextMapPropagator())

    return trace.get_tracer("omnicam-algo")
