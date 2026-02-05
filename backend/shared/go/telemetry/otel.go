package tele

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

//This SDK sets up a bunch of boilerplate about otel stuff
//It mostly creates providers and sets to be global
//Those providers will then be used by the custom telemetry types

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
// It also sets the global providers, that will at a later step be used to initialize the logger, tracer and meterer
func SetupOTelSDK(ctx context.Context, collectorAddress string, serviceName string) (func(context.Context) error, error) {
	// otlp.
	var shutdownFuncs []func(context.Context) error
	var err error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	prop := NewPropagator()
	otel.SetTextMapPropagator(prop)

	// // Set up trace provider.
	tracerProvider, err := NewTracerProvider(ctx, collectorAddress, serviceName)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// // Set up meter provider.
	// meterProvider, err := NewMeterProvider()
	// if err != nil {
	// 	handleErr(err)
	// 	return shutdown, err
	// }
	// shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	// otel.SetMeterProvider(meterProvider)

	// Set up logger provider.
	loggerProvider, err := NewLoggerProvider(ctx, collectorAddress, serviceName)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}

	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	return shutdown, err
}

func NewPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func NewMeterProvider() (*metric.MeterProvider, error) {
	metricExporter, err := stdoutmetric.New()
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			// Default is 1m. Set to 3s for demonstrative purposes.
			metric.WithInterval(time.Minute*2))), //reduce once
	)
	return meterProvider, nil
}

//fake obj for debugging the exporter, don't touch

// type debug struct {
// }

// func (e *debug) Export(ctx context.Context, records []log.Record) error {
// 	fmt.Println("Export called", ctx, records)
// 	return nil
// }
// func (*debug) ForceFlush(ctx context.Context) error {
// 	fmt.Println("force flush called", ctx)
// 	return nil
// }
// func (e *debug) Shutdown(ctx context.Context) error {
// 	fmt.Println("shutdown called", ctx)
// 	return nil
// }

func NewLoggerProvider(ctx context.Context, collectorAddress string, serviceName string) (*log.LoggerProvider, error) {

	logExporter, err := otlploggrpc.New(
		ctx,
		otlploggrpc.WithEndpointURL("dns://"+collectorAddress),
		otlploggrpc.WithInsecure(),
		otlploggrpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create log exporter: %w", err)
	}

	//TODO add service name and set up versioning?
	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(serviceName),
		semconv.ServiceVersion("my_version_TBD"),
		semconv.HostName("host_TBD"), // TODO kubernetes name or docker something
	)

	processor := log.NewBatchProcessor(logExporter)

	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(processor),
		log.WithResource(resource),
	)
	return loggerProvider, nil
}

func NewTracerProvider(ctx context.Context, collectorAddress string, serviceName string) (*trace.TracerProvider, error) {
	traceExporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpointURL("dns://"+collectorAddress),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	//TODO add service name and set up versioning?
	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(serviceName),
		semconv.ServiceVersion("my_version_TBD"),
		semconv.HostName("host_TBD"), // TODO kubernetes name or docker something
	)

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter,
			trace.WithBatchTimeout(1*time.Second)),
		trace.WithResource(resource),
	)

	return tracerProvider, nil
}
