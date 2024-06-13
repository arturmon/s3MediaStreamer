package otel

import (
	"context"
	"os"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

type Repository interface {
	Close(ctx context.Context) error
}

type ProviderConfig struct {
	JaegerEndpoint string
	ServiceName    string
	ServiceVersion string
	Environment    string // Examples: 'staging', 'production'
	Cfg            *model.Config
	Logger         *logs.Logger
	// Set this to `true` if you want to disable tracing completly.
	Disabled bool
}
type Provider struct {
	provider trace.TracerProvider
}

func InitializeTracer(ctx context.Context, cfg *model.Config, logger *logs.Logger, appName, version string) (*Provider, error) {
	config := ProviderConfig{
		JaegerEndpoint: cfg.AppConfig.OpenTelemetry.JaegerEndpoint + "/api/traces",
		ServiceName:    appName,
		ServiceVersion: version,
		Environment:    cfg.AppConfig.OpenTelemetry.Environment,
		Cfg:            cfg,
		Logger:         logger,
		Disabled:       cfg.AppConfig.OpenTelemetry.TracingEnabled,
	}
	tracer, err := initProvider(ctx, config)
	if err != nil {
		return &Provider{}, err
	}
	return tracer, nil
}

func initProvider(ctx context.Context, config ProviderConfig) (*Provider, error) {
	if config.Disabled {
		return &Provider{provider: trace.NewNoopTracerProvider()}, nil //nolint:staticcheck // SA1019 NewNoopTracerProvider() is deprecated
	}

	tp, tpErr := jaegerTraceProvider(ctx, config)
	if tpErr != nil {
		config.Logger.Fatal(tpErr)
		return nil, tpErr
	}
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{}))

	return &Provider{provider: tp}, nil
}

func jaegerTraceProvider(ctx context.Context, config ProviderConfig) (*sdktrace.TracerProvider, error) {
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

func GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	return hostname
}
