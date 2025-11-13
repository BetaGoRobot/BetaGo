package otel

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/consts"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

func init() {
	otel.SetTracerProvider(tracerProvider)
}

const (
	environment = "production"
	id          = 1
)

// tracerProvider jaeger provider
var (
	tracerProvider *tracesdk.TracerProvider
	loggerProvider *log.LoggerProvider
)

func OtelProvider() *tracesdk.TracerProvider {
	return tracerProvider
}

func LoggerProvider() *log.LoggerProvider {
	return loggerProvider
}

func init() {
	otelCollectorEP := "otel-collector:4317"
	if consts.IsTest {
		otelCollectorEP = "192.168.31.74:4317"
	}
	tracerProvider, _ = newTracerProvider(otelCollectorEP)
	loggerProvider, _ = newLoggerProvider(otelCollectorEP)
	BetaGoOtelTracer = tracerProvider.Tracer("command-handler")
	LarkRobotOtelTracer = tracerProvider.Tracer("larkrobot-handler")
}

// BetaGoOtelTracer a
var (
	BetaGoOtelTracer    trace.Tracer
	LarkRobotOtelTracer trace.Tracer
)

// newTracerProvider returns an OpenTelemetry TracerProvider configured to use
// the Jaeger exporter that will send spans to the provided url. The returned
// TracerProvider will also use a Resource configured with all the information
// about the application.
func newTracerProvider(url string) (*tracesdk.TracerProvider, error) {
	// Create the Jaeger exporter
	ctx := context.Background()
	exp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(url), otlptracegrpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(newResource()),
	)
	return tp, nil
}

func newResource() *resource.Resource {
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(consts.BotIdentifier),
			attribute.String("environment", environment),
			attribute.Int64("ID", id),
		),
	)
	if err != nil {
		panic(err)
	}
	return res
}

func newLoggerProvider(ep string) (*log.LoggerProvider, error) {
	ctx := context.Background()
	exporter, err := otlploggrpc.New(
		ctx, otlploggrpc.WithEndpoint(ep), otlploggrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}
	processor := log.NewBatchProcessor(exporter)
	return log.NewLoggerProvider(
		log.WithResource(newResource()),
		log.WithProcessor(processor),
	), nil
}
