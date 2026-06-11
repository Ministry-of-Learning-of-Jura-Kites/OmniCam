import { defineNuxtPlugin, useRuntimeConfig } from "#app";
import {
  SpanKind,
  SpanStatusCode,
  trace,
  type Attributes,
} from "@opentelemetry/api";
import { resourceFromAttributes } from "@opentelemetry/resources";
import { WebTracerProvider } from "@opentelemetry/sdk-trace-web";
import { BatchSpanProcessor } from "@opentelemetry/sdk-trace-base";
import { OTLPTraceExporter } from "@opentelemetry/exporter-trace-otlp-http";
import { registerInstrumentations } from "@opentelemetry/instrumentation";
import { FetchInstrumentation } from "@opentelemetry/instrumentation-fetch";

declare global {
  interface Window {
    __otelInitialized?: boolean;
  }
}

function createExceptionAttributes(err: unknown): Attributes {
  const message = err instanceof Error ? err.message : String(err);

  return {
    "exception.message": message,
    ...(err instanceof Error
      ? {
          "exception.type": err.name,
          "exception.stacktrace": err.stack ?? "",
        }
      : {}),
  };
}

export default defineNuxtPlugin(() => {
  const tracer = trace.getTracer("omnicam-frontend");

  const otel = {
    traceWsSend(
      ws: WebSocket | undefined,
      operation: string,
      payload?: Uint8Array | ArrayBuffer,
    ) {
      if (!ws) {
        return;
      }

      tracer.startActiveSpan(
        `websocket.send.${operation}`,
        {
          kind: SpanKind.CLIENT,
          attributes: {
            "websocket.url": ws.url,
            "websocket.operation": operation,
            "message.size_bytes": payload?.byteLength ?? 0,
          },
        },
        (span) => {
          try {
            span.setStatus({
              code: SpanStatusCode.OK,
            });
          } catch (err: unknown) {
            const message = err instanceof Error ? err.message : String(err);

            span.addEvent("exception", createExceptionAttributes(err));

            span.setStatus({
              code: SpanStatusCode.ERROR,
              message,
            });

            throw err;
          } finally {
            span.end();
          }
        },
      );
    },

    traceWsReceive<T>(
      websocketUrl: string,
      operation: string,
      handler: (data: T) => void | Promise<void>,
    ) {
      return async (data: T) => {
        tracer.startActiveSpan(
          `websocket.receive.${operation}`,
          {
            kind: SpanKind.CLIENT,
            attributes: {
              "websocket.url": websocketUrl,
              "websocket.operation": operation,
            },
          },
          async (span) => {
            try {
              await handler(data);

              span.setStatus({
                code: SpanStatusCode.OK,
              });
            } catch (err: unknown) {
              const message = err instanceof Error ? err.message : String(err);

              span.addEvent("exception", createExceptionAttributes(err));

              span.setStatus({
                code: SpanStatusCode.ERROR,
                message,
              });

              throw err;
            } finally {
              span.end();
            }
          },
        );
      };
    },
  };

  if (import.meta.client && !window.__otelInitialized) {
    window.__otelInitialized = true;

    try {
      const config = useRuntimeConfig();

      const exporter = new OTLPTraceExporter({
        url: `${config.public.otelEndpoint}/v1/traces`,
      });

      const provider = new WebTracerProvider({
        resource: resourceFromAttributes({
          "service.name": config.public.otelServiceName,
          "service.version": config.public.otelServiceVersion,
          "deployment.environment": config.public.deployEnv,
        }),
        spanProcessors: [
          new BatchSpanProcessor(exporter, {
            maxQueueSize: 100,
            maxExportBatchSize: 20,
            scheduledDelayMillis: 5000,
          }),
        ],
      });

      provider.register();

      registerInstrumentations({
        instrumentations: [
          new FetchInstrumentation({
            ignoreUrls: [
              /blob:/,
              /\/_nuxt\//,
              /\/__nuxt/,
              /hot-update/,
              /localhost:4318/,
              /\.glb$/,
              /\.gltf$/,
              /\/assets\//,
            ],
            propagateTraceHeaderCorsUrls: [/localhost:8080/],
          }),
        ],
      });
    } catch (err) {
      console.error("[OTEL] initialization failed", err);
    }
  }

  return {
    provide: {
      tracer,
      otel,
    },
  };
});
