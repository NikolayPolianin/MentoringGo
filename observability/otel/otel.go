package otel

import (
	"checkgo/observability/otel/tracer"
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
)

type Config struct {
	AppName    string  `envconfig:"APP_NAME"`
	AppVersion string  `envconfig:"APP_VERSION"`
	Endpoint   string  `envconfig:"OTEL_ENDPOINT"`
	Namespace  string  `envconfig:"OTEL_NAMESPACE"`
	InstanceID string  `envconfig:"OTEL_INSTANCE_ID"`
	Ratio      float64 `envconfig:"OTEL_RATIO" default:"1.0"`
}

func InitTracer(ctx context.Context) error {
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpointURL("http://localhost:4317/v1/traces"))
	if err != nil {
		return err
	}

	res, err := resource.New(ctx,

		resource.WithAttributes(
			semconv.ServiceNameKey.String("OneService"),
			semconv.ServiceVersionKey.String("0"),
			semconv.ServiceInstanceIDKey.String("2"),
			semconv.ServiceNamespaceKey.String("dev")))
	if err != nil {
		_ = exporter.Shutdown(ctx)
		return err
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
		trace.WithSampler(
			trace.ParentBased(trace.TraceIDRatioBased(1.0)),
		))
	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{}))

	tracer.Init(otel.Tracer(""))

	return nil
}
