package tracing

import (
	"context"
	"fmt"

	"github.com/weeb-vip/list-service/config"
	"github.com/weeb-vip/go-tracing-lib/providers"
	"github.com/weeb-vip/go-tracing-lib/tracing"
	"go.opentelemetry.io/otel/trace"
)

var (
	shutdownFunc func(context.Context) error
	appCtx       context.Context
)

// InitTracing initializes the tracing system with Tempo/Grafana provider
func InitTracing(ctx context.Context) (context.Context, error) {
	cfg := config.LoadConfigOrPanic()

	// Get tracing configuration
	tracingCfg := GetTracingConfig()

	// Setup Tempo provider configuration
	providerConfig := providers.ProviderConfig{
		ServiceName:    cfg.AppConfig.APPName,
		ServiceVersion: tracingCfg.ServiceVersion,
	}

	// Create Tempo provider with custom endpoint
	tracerProvider, shutdown, err := NewTempoProvider(ctx, providerConfig, tracingCfg.Endpoint, tracingCfg.Insecure, cfg.AppConfig.Env)
	if err != nil {
		return ctx, fmt.Errorf("failed to create Tempo provider: %w", err)
	}

	provider := tracing.Provider{
		TracerProvider: tracerProvider,
		Shutdown:       shutdown,
	}

	tracingConfig := tracing.TracingConfig{
		Provider:    provider,
		ServiceName: cfg.AppConfig.APPName,
	}

	// Setup OpenTelemetry SDK
	shutdownFn, tracedCtx, err := tracing.SetupOTelSDK(ctx, tracingConfig)
	if err != nil {
		return ctx, fmt.Errorf("failed to setup tracing: %w", err)
	}

	// Store shutdown function globally for cleanup
	shutdownFunc = shutdownFn
	appCtx = tracedCtx

	return tracedCtx, nil
}

// Shutdown gracefully shuts down the tracing system
func Shutdown(ctx context.Context) error {
	if shutdownFunc != nil {
		return shutdownFunc(ctx)
	}
	return nil
}

// GetTracer returns a tracer from the context or creates a new one
func GetTracer(ctx context.Context) trace.Tracer {
	return tracing.TracerFromContext(ctx)
}

// GetServiceName returns the service name from the context
func GetServiceName(ctx context.Context) string {
	return tracing.GetServiceName(ctx)
}