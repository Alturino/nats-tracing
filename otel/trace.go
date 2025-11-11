package otel

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

func InitTracerProvider(
	ctx context.Context,
	endpoint, serviceName string,
	res *resource.Resource,
) (*sdktrace.TracerProvider, error) {
	traceExporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		err = fmt.Errorf("failed creating traceExporter with error: %w", err)
		return nil, err
	}

	traceRes, err := resource.Merge(
		resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceName(serviceName)),
		res,
	)
	if err != nil {
		err = fmt.Errorf("failed creating traceRes with error: %w", err)
		return nil, err
	}

	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter, sdktrace.WithBatchTimeout(5*time.Second)),
		sdktrace.WithResource(traceRes),
	)

	return traceProvider, nil
}
