package postgres

import (
	"context"
	"s3MediaStreamer/app/model"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// GetTracer returns an OpenTelemetry tracer based on the app information.
func GetTracer(ctx context.Context) trace.Tracer {
	// Retrieve appInfo from the context.
	appsInfo, _ := ctx.Value("appInfo").(*model.AppInfo)
	// Fallback in case appInfo is not found in the context.
	if appsInfo == nil {
		return otel.Tracer("default")
	}
	// Create a tracer using appName and version.
	return otel.Tracer(appsInfo.AppName + "/" + appsInfo.Version + "/repository")
}
