package metrics

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"time"
)

type Metrics struct {
	Request interface{}
}

func InitMetrics(ctx context.Context) error {
	exporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint("localhost:4318"),
		otlpmetrichttp.WithURLPath("/v1/metrics"),
		otlpmetrichttp.WithInsecure())
	if err != nil {
		return err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("OneService")))
	if err != nil {
		_ = exporter.Shutdown(ctx)
		return err
	}
	reader := metric.NewPeriodicReader(exporter, metric.WithInterval(10*time.Second))

	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(reader))

	otel.SetMeterProvider(mp)
	return nil
}
