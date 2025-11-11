package otel

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"golang.org/x/sync/errgroup"
)

type ShutdownFunc func(context.Context) error

func InitOtelSdk(
	ctx context.Context,
	serviceName string,
) (shutdownFuncs []ShutdownFunc, err error) {
	propagator := GetTextMapPropagator()
	otel.SetTextMapPropagator(propagator)

	res, err := resource.New(
		ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithContainer(),
		resource.WithTelemetrySDK(),
		resource.WithOS(),
		resource.WithAttributes(semconv.ServiceName(serviceName)),
	)
	if err != nil {
		err = fmt.Errorf("failed initializing otel tracerProvider with error=%w", err)
		return nil, err
	}

	tracerProvider, err := InitTracerProvider(
		ctx,
		fmt.Sprintf("%s:%d", "otel-collector", 4317),
		serviceName,
		res,
	)
	if err != nil {
		err = fmt.Errorf("failed initializing otel tracerProvider with error=%w", err)
		return nil, err
	}
	otel.SetTracerProvider(tracerProvider)
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)

	metricEndpoint := fmt.Sprintf("%s:%d", "otel-collector", 4317)
	meterProvider, err := InitMetricProvider(ctx, metricEndpoint, res)
	if err != nil {
		err = fmt.Errorf("failed initializing otel meterProvider with error=%w", err)
		return shutdownFuncs, err
	}
	otel.SetMeterProvider(meterProvider)
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)

	return shutdownFuncs, nil
}

func ShutdownOtel(ctx context.Context, shutdownFuncs []ShutdownFunc) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	erg, ctx := errgroup.WithContext(ctx)
	for _, shutdown := range shutdownFuncs {
		erg.Go(func() error { return shutdown(ctx) })
	}
	return erg.Wait()
}

func GetTextMapPropagator() propagation.TextMapPropagator {
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	return propagator
}
