package jaeger_client

import (
	"github.com/BetaGoRobot/BetaGo/betagovar"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

func init() {
	otel.SetTracerProvider(betaGoJaegerProvider)
}

const (
	service     = "BetaGo"
	environment = "production"
	id          = 1
)

// betaGoJaegerProvider jaeger provider
var betaGoJaegerProvider *tracesdk.TracerProvider

func init() {
	if betagovar.IsTest {
		betaGoJaegerProvider, _ = tracerProvider("http://192.168.31.74:14268/api/traces")
	} else {
		betaGoJaegerProvider, _ = tracerProvider("http://jaeger-all-in-one-ix-chart.ix-jaeger-all-in-one:14268/api/traces")
	}
	BetaGoCommandTracer = betaGoJaegerProvider.Tracer("command-handler")
}

// BetaGoCommandTracer a
var BetaGoCommandTracer trace.Tracer

// tracerProvider returns an OpenTelemetry TracerProvider configured to use
// the Jaeger exporter that will send spans to the provided url. The returned
// TracerProvider will also use a Resource configured with all the information
// about the application.
func tracerProvider(url string) (*tracesdk.TracerProvider, error) {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(service),
			attribute.String("environment", environment),
			attribute.Int64("ID", id),
		)),
	)
	return tp, nil
}
