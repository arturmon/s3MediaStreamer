package otel

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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
	logger   *logs.Logger
	Disabled bool
}

func InitializeTracer(ctx context.Context, cfg *model.Config, logger *logs.Logger, appName, version string) (*Provider, error) {
	config := ProviderConfig{
		JaegerEndpoint: cfg.AppConfig.OpenTelemetry.JaegerEndpoint + "/api/traces",
		ServiceName:    appName,
		ServiceVersion: version,
		Environment:    cfg.AppConfig.OpenTelemetry.Environment,
		Cfg:            cfg,
		Logger:         logger,
		Disabled:       !cfg.AppConfig.OpenTelemetry.TracingEnabled,
	}
	tracer, err := initProvider(ctx, config)
	if err != nil {
		return &Provider{}, err
	}
	return tracer, nil
}

func initProvider(ctx context.Context, config ProviderConfig) (*Provider, error) {
	if config.Disabled {
		return &Provider{provider: trace.NewNoopTracerProvider(), logger: config.Logger, Disabled: true}, nil //nolint:staticcheck // SA1019 NewNoopTracerProvider() is deprecated
	}
	/*
		if config.Disabled {
			return &Provider{provider: sdktrace.NewTracerProvider(), logger: config.Logger, disabled: true}, nil
		}

	*/
	tp, tpErr := jaegerTraceProvider(ctx, config)
	if tpErr != nil {
		config.Logger.Fatal(tpErr.Error())
		return nil, fmt.Errorf("failed to create jaeger trace provider: %w", tpErr)
	}
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{}))

	return &Provider{provider: tp, logger: config.Logger, Disabled: false}, nil
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

// LogWithTrace logs messages with TraceID and SpanID from the context.
func (p *Provider) LogWithTrace(ctx context.Context, msg string) {
	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	p.logger.Info(msg, slog.String("trace_id", traceID), slog.String("span_id", spanID))
}

func GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	return hostname
}

func (p *Provider) TraceAndLog(ctx context.Context, span trace.Span, message string, attributes ...attribute.KeyValue) {
	if !p.Disabled {
		p.LogWithTrace(ctx, message)
		span.SetAttributes(attributes...)
	}
}
