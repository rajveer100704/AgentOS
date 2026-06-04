package telemetry

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var tracer trace.Tracer

// Init initialises OpenTelemetry with the requested exporter and wires up the
// TraceStore as a secondary span processor so the in-process dashboard can
// query recent traces without talking to Jaeger.
//
// Supported exporter values: "stdout" | "jaeger" | "otlp"
// endpoint is the OTLP gRPC address, e.g. "localhost:4317".
func Init(serviceName, exporter, endpoint string, store *TraceStore) (func(), error) {
	var sp sdktrace.SpanExporter
	var err error

	switch exporter {
	case "jaeger", "otlp":
		// Use OTLP gRPC to export to Jaeger (>= v1.35 natively accepts OTLP).
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(endpoint),
		}
		// Jaeger in dev mode typically listens on plain text.
		opts = append(opts, otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))

		client := otlptracegrpc.NewClient(opts...)
		sp, err = otlptrace.New(context.Background(), client)
		if err != nil {
			log.Printf("[telemetry] OTLP exporter failed, falling back to stdout: %v", err)
			sp, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
		}
	default: // "stdout" and anything unrecognised
		sp, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
	}

	if err != nil {
		return nil, err
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			attribute.String("service.version", "v0.8.0"),
			attribute.String("deployment.environment", "production"),
		),
	)
	if err != nil {
		return nil, err
	}

	// Build the provider with:
	//   1. Batching exporter → Jaeger / stdout
	//   2. TraceStore processor (if provided) → in-process ring buffer
	processors := []sdktrace.TracerProviderOption{
		sdktrace.WithBatcher(sp),
		sdktrace.WithResource(res),
	}
	if store != nil {
		processors = append(processors, sdktrace.WithSpanProcessor(store.NewProcessor()))
	}

	tp := sdktrace.NewTracerProvider(processors...)
	otel.SetTracerProvider(tp)
	tracer = tp.Tracer(serviceName)

	log.Printf("[telemetry] initialized (exporter: %s, endpoint: %s)", exporter, endpoint)

	shutdown := func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("[telemetry] shutdown error: %v", err)
		}
	}

	return shutdown, nil
}

// Tracer returns the global AegisFlow tracer.
// Safe to call before Init — falls back to the no-op global tracer.
func Tracer() trace.Tracer {
	if tracer == nil {
		return otel.Tracer("aegisflow")
	}
	return tracer
}
