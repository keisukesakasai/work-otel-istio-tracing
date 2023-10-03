package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	tracing "server/internal/trace"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
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

	r := gin.New()
	r.Use(
		otelgin.Middleware("otel-istio-tracing-server"),
	)

	r.GET("/", Handler)
	r.Run(":8080")
}

func Handler(c *gin.Context) {
	randomWord := generateRandomWord(c)
	c.String(http.StatusOK, randomWord)
}

func generateRandomWord(c *gin.Context) string {
	_, span := tracer.Start(c.Request.Context(), "generateRandomWord")
	defer span.End()

	words := []string{"apple", "banana", "cherry", "date", "elderberry"}

	seed := time.Now().UnixNano()
	rand.New(rand.NewSource(seed))
	return words[rand.Intn(len(words))]
}
