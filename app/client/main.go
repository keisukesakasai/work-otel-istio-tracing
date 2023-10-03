package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	logging "client/internal/log"
	tracing "client/internal/trace"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	Service          = "otel-istio-tracing-client"
	SleepEnv         = "SLEEP"
	ServerAddressEnv = "SERVER_ADDRESS"
)

var tracer = otel.GetTracerProvider().Tracer("")

func main() {
	// Start Tracing
	tp, err := tracing.InitTracer()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	url := os.Getenv(ServerAddressEnv)
	sleep, _ := strconv.Atoi(os.Getenv(SleepEnv))

	// Start main task.
	for range time.Tick(time.Duration(sleep) * time.Second) {
		ctx := context.Background()
		err = Request(ctx, url)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func Request(ctx context.Context, url string) (err error) {
	logger := logging.GetLoggerFromCtx(ctx)

	client := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	err = func(ctx context.Context) error {
		ctx, span := tracer.Start(
			ctx,
			"say hello",
			trace.WithAttributes(semconv.PeerService(Service)),
		)
		defer span.End()

		req, _ := http.NewRequestWithContext(
			ctx,
			"GET",
			url,
			nil,
		)

		logger.Infof("Sending Request...\n")
		res, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		body, err := io.ReadAll(res.Body)
		_ = res.Body.Close()

		logger.Infof("Response Received: %s\n\n\n", body)

		return err
	}(ctx)

	return err
}
