package otel

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/consts"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

func init() {
	otel.SetTracerProvider(otelProvider)
}

const (
	environment = "production"
	id          = 1
)

// otelProvider jaeger provider
var otelProvider *tracesdk.TracerProvider

func OtelProvider() *tracesdk.TracerProvider {
	return otelProvider
}

// 重构：改成OtelCollector来收集
var otelCollectorURL = "http://otel-collector:4318/v1/traces"

func init() {
	if consts.IsTest {
		otelProvider, _ = tracerProvider("http://192.168.31.74:4318/v1/traces")
	} else if consts.IsCluster {
		otelProvider, _ = tracerProvider(otelCollectorURL)
	} else if consts.IsCompose {
		otelProvider, _ = tracerProvider(otelCollectorURL)
	}
	BetaGoOtelTracer = otelProvider.Tracer("command-handler")
	LarkRobotOtelTracer = otelProvider.Tracer("larkrobot-handler")
}

// BetaGoOtelTracer a
var (
	BetaGoOtelTracer    trace.Tracer
	LarkRobotOtelTracer trace.Tracer
)

// tracerProvider returns an OpenTelemetry TracerProvider configured to use
// the Jaeger exporter that will send spans to the provided url. The returned
// TracerProvider will also use a Resource configured with all the information
// about the application.
func tracerProvider(url string) (*tracesdk.TracerProvider, error) {
	// Create the Jaeger exporter
	ctx := context.Background()
	exp, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(url))
	if err != nil {
		return nil, err
	}

	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceName(consts.BotIdentifier),
				attribute.String("environment", environment),
				attribute.Int64("ID", id),
			)),
	)
	return tp, nil
}
