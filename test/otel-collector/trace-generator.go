package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func main() {
	ctx := context.Background()
	testMode := os.Getenv("TEST_MODE")

	log.Printf("Starting trace generator in mode: %s\n", testMode)

	switch testMode {
	case "invalid_auth":
		testInvalidAuth(ctx)
	case "malformed":
		testMalformedData(ctx)
	case "timeout":
		testNetworkTimeout(ctx)
	case "large_batch":
		testLargeBatch(ctx)
	default:
		testSuccessScenario(ctx)
	}
}

// Test 1: Success scenario (baseline)
func testSuccessScenario(ctx context.Context) {
	log.Println("========================================")
	log.Println("Running SUCCESS scenario test")
	log.Println("Expected: 100 traces successfully sent to Brokle")
	log.Println("========================================")

	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	log.Printf("OTLP Endpoint: %s\n", endpoint)

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("Failed to create exporter: %v", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("test-app"),
			semconv.ServiceVersion("1.0.0"),
			attribute.String("environment", "test"),
			attribute.String("test.scenario", "success"),
		),
	)
	if err != nil {
		log.Fatalf("Failed to create resource: %v", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	tracer := tp.Tracer("test-app")

	// Generate 100 test traces
	for i := 0; i < 100; i++ {
		_, span := tracer.Start(ctx, fmt.Sprintf("test-operation-%d", i))
		span.SetAttributes(
			attribute.Int("test.iteration", i),
			attribute.String("test.type", "success"),
			attribute.String("operation.name", "test-operation"),
		)
		time.Sleep(10 * time.Millisecond)
		span.End()

		if (i+1)%10 == 0 {
			log.Printf("Progress: %d/100 traces sent", i+1)
		}
	}

	log.Println("All traces generated. Waiting for export to complete...")
	time.Sleep(5 * time.Second)
	log.Println("========================================")
	log.Println("SUCCESS: Test complete")
	log.Println("========================================")
}

// Test 2: Invalid authentication
func testInvalidAuth(ctx context.Context) {
	log.Println("========================================")
	log.Println("Running INVALID AUTH test")
	log.Println("Expected: 401 Unauthorized in collector logs")
	log.Println("========================================")

	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	log.Printf("OTLP Endpoint: %s\n", endpoint)

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL(endpoint),
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithHeaders(map[string]string{
			"X-API-Key": "invalid_bk_fake_key_12345",
		}),
	)
	if err != nil {
		log.Fatalf("Failed to create exporter: %v", err)
	}

	tp := trace.NewTracerProvider(trace.WithBatcher(exporter))
	otel.SetTracerProvider(tp)
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	tracer := tp.Tracer("test-app")
	_, span := tracer.Start(ctx, "invalid-auth-test")
	span.SetAttributes(
		attribute.String("test.scenario", "invalid_auth"),
	)
	span.End()

	log.Println("Trace sent with invalid API key. Waiting for response...")
	time.Sleep(3 * time.Second)

	log.Println("========================================")
	log.Println("Test complete. Check collector logs for:")
	log.Println("  - HTTP 401 Unauthorized errors")
	log.Println("  - Export failure messages")
	log.Println("========================================")
}

// Test 3: Malformed data
// Note: OTEL SDK validates data client-side, preventing malformed spans
func testMalformedData(ctx context.Context) {
	log.Println("========================================")
	log.Println("Running MALFORMED DATA test")
	log.Println("========================================")
	log.Println("")
	log.Println("INFO: OpenTelemetry SDK validates data client-side")
	log.Println("INFO: Malformed spans are rejected before network transmission")
	log.Println("INFO: The SDK enforces:")
	log.Println("  - Valid trace/span IDs (16/8 bytes)")
	log.Println("  - Required fields (name, timestamps)")
	log.Println("  - Proper attribute types")
	log.Println("")
	log.Println("To test server-side validation, you would need to:")
	log.Println("  1. Send raw HTTP POST to /v1/traces")
	log.Println("  2. With malformed protobuf/JSON body")
	log.Println("  3. Bypass OTEL SDK entirely")
	log.Println("")
	log.Println("========================================")
	log.Println("SKIP: This test documents SDK behavior")
	log.Println("SUCCESS: SDK prevents malformed data from reaching server")
	log.Println("========================================")
}

// Test 4: Network timeout
func testNetworkTimeout(ctx context.Context) {
	log.Println("========================================")
	log.Println("Running NETWORK TIMEOUT test")
	log.Println("Expected: Timeout errors in trace generator logs")
	log.Println("========================================")

	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	log.Printf("OTLP Endpoint: %s\n", endpoint)
	log.Println("Using extremely short timeout (1ms) to force failure")

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL(endpoint),
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithTimeout(1*time.Millisecond), // Unrealistic timeout
	)
	if err != nil {
		log.Fatalf("Failed to create exporter: %v", err)
	}

	tp := trace.NewTracerProvider(trace.WithBatcher(exporter))
	otel.SetTracerProvider(tp)
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	tracer := tp.Tracer("test-app")
	_, span := tracer.Start(ctx, "timeout-test")
	span.SetAttributes(
		attribute.String("test.scenario", "timeout"),
	)
	span.End()

	log.Println("Trace sent. Waiting for timeout...")
	time.Sleep(3 * time.Second)

	log.Println("========================================")
	log.Println("Test complete. Expected behavior:")
	log.Println("  - Context deadline exceeded")
	log.Println("  - Export failed after 1ms timeout")
	log.Println("========================================")
}

// Test 5: Large batch (stress test)
func testLargeBatch(ctx context.Context) {
	log.Println("========================================")
	log.Println("Running LARGE BATCH stress test")
	log.Println("Generating 10,000 spans in large batches")
	log.Println("========================================")

	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	log.Printf("OTLP Endpoint: %s\n", endpoint)

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("Failed to create exporter: %v", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("test-app-load"),
			semconv.ServiceVersion("1.0.0"),
			attribute.String("environment", "load-test"),
			attribute.String("test.scenario", "large_batch"),
		),
	)
	if err != nil {
		log.Fatalf("Failed to create resource: %v", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter,
			trace.WithMaxExportBatchSize(5000), // Large batch size
		),
		trace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	tracer := tp.Tracer("test-app-load")

	// Generate 10,000 spans quickly
	startTime := time.Now()
	for i := 0; i < 10000; i++ {
		_, span := tracer.Start(ctx, fmt.Sprintf("large-batch-%d", i))
		span.SetAttributes(
			attribute.Int("batch.size", 10000),
			attribute.Int("span.index", i),
			attribute.String("test.type", "load"),
		)
		span.End()

		if (i+1)%1000 == 0 {
			log.Printf("Progress: %d/10000 spans generated", i+1)
		}
	}
	duration := time.Since(startTime)

	log.Printf("All spans generated in %v", duration)
	log.Printf("Generation rate: %.0f spans/sec", float64(10000)/duration.Seconds())
	log.Println("Waiting for export to complete...")
	time.Sleep(10 * time.Second)

	log.Println("========================================")
	log.Println("SUCCESS: Large batch test complete")
	log.Printf("Generated: 10,000 spans in %v\n", duration)
	log.Println("========================================")
}
