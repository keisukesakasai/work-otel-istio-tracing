package tracing

import (
	"context"
	"fmt"
	"os"

	logging "client/internal/log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	OtelCollectorAddressEnv = "OTEL_COLLECTOR_ADDRESS"
)

func InitTracer() (*sdktrace.TracerProvider, error) {
	ctx := context.Background()
	logger := logging.GetLoggerFromCtx(ctx)

	otelCollectorAddress := os.Getenv(OtelCollectorAddressEnv)

	conn, err := grpc.DialContext(ctx,
		otelCollectorAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logger.Info("Initialize Trace Provider ...: ", otelCollectorAddress)

	return tp, nil
}
