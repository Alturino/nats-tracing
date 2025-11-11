package otel

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

func InitMetricProvider(
	c context.Context,
	endpoint string,
	res *resource.Resource,
) (*metric.MeterProvider, error) {
	otlpMetricExporter, err := otlpmetricgrpc.New(
		c,
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		err = fmt.Errorf("failed to initializing otlpMetricExporter with error: %w", err)
		return nil, err
	}

	promExporter, err := prometheus.New(prometheus.WithNamespace("bloodhound"))
	if err != nil {
		err = fmt.Errorf("failed to initializing promExporter with error: %w", err)
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(
			metric.NewPeriodicReader(otlpMetricExporter, metric.WithInterval(2*time.Second)),
		),
		metric.WithReader(promExporter),
	)

	return meterProvider, nil
}
