package tracing

import (
	"context"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// Init bira koji će tracer da inicijalizuje na osnovu env varijable.
// Vraća TracerProvider i shutdown funkciju.
func Init(serviceName string) (*sdktrace.TracerProvider, func(context.Context) error, error) {
	url := os.Getenv("JAEGER_ENDPOINT")
	var tp *sdktrace.TracerProvider
	var err error

	if url != "" {
		tp, err = initJaegerTracer(url, serviceName)
	} else {
		tp, err = initFileTracer()
	}

	if err != nil {
		return nil, nil, err
	}

	// Registruj kao globalnog provajdera
	otel.SetTracerProvider(tp)

	// Shutdown funkcija za elegantno gašenje
	shutdown := func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return tp.Shutdown(ctx)
	}

	return tp, shutdown, nil
}

// initFileTracer inicijalizuje provider koji upisuje trejsove u traces.json
func initFileTracer() (*sdktrace.TracerProvider, error) {
	log.Println("Initializing tracing to traces.json")
	f, err := os.Create("traces.json")
	if err != nil {
		return nil, err
	}

	exp, err := stdouttrace.New(
		stdouttrace.WithWriter(f),
		stdouttrace.WithPrettyPrint(),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	return tp, nil
}

// initJaegerTracer inicijalizuje provider koji šalje trejsove na Jaeger kolektor
func initJaegerTracer(url string, serviceName string) (*sdktrace.TracerProvider, error) {
	log.Printf("Initializing tracing to Jaeger at %s\n", url)
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)
	return tp, nil
}
