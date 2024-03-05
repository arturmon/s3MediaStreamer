package otel

import (
	"context"
	"s3MediaStreamer/app/internal/config"
	"s3MediaStreamer/app/pkg/logging"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

type ProviderConfig struct {
	JaegerEndpoint string
	ServiceName    string
	ServiceVersion string
	Environment    string // Examples: 'staging', 'production'
	Cfg            *config.Config
	Logger         *logging.Logger
	// Set this to `true` if you want to disable tracing completly.
	Disabled bool
}
type Provider struct {
	provider trace.TracerProvider
}

func InitProvider(ctx context.Context, config ProviderConfig) (Provider, error) {
	if config.Disabled {
		return Provider{provider: trace.NewNoopTracerProvider()}, nil
	}

	tp, tpErr := JaegerTraceProvider(ctx, config)
	if tpErr != nil {
		config.Logger.Fatal(tpErr)
	}
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{}))

	return Provider{provider: tp}, nil
}

func JaegerTraceProvider(ctx context.Context, config ProviderConfig) (*sdktrace.TracerProvider, error) {
	exp, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithEndpointURL(config.JaegerEndpoint),
		otlptracehttp.WithURLPath("/v1/traces"),
	)
	if err != nil {
		return nil, err
	}

	serviceName := config.ServiceName + "-" + GetHostname()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(config.ServiceVersion),
			semconv.DeploymentEnvironmentKey.String(config.Environment),
		)),
	)
	return tp, nil
}

func (p Provider) Close(ctx context.Context) error {
	if prv, ok := p.provider.(*sdktrace.TracerProvider); ok {
		return prv.Shutdown(ctx)
	}

	return nil
}
