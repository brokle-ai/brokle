# Migration Guide: Custom Batch API to OTLP

## Overview

As of **October 2025**, the Brokle platform has fully migrated to **OpenTelemetry Protocol (OTLP)** for all telemetry ingestion. The custom batch API (`/v1/ingest/batch`) has been removed in favor of industry-standard OTLP endpoints.

This migration provides:
- ✅ **Industry Standard**: 100% OpenTelemetry spec compliant
- ✅ **Ecosystem Integration**: Works with all OTLP-compatible tools
- ✅ **Better Performance**: Native protobuf encoding, no custom envelope overhead
- ✅ **Vendor Neutral**: Use standard OpenTelemetry SDKs
- ✅ **Future Proof**: Full OTLP ecosystem support (traces, metrics, logs)

---

## What Changed

### ❌ Removed (Custom Batch API)
- `POST /v1/ingest/batch` endpoint
- `POST /v1/telemetry/validate` endpoint
- Custom envelope format (`TelemetryBatchRequest`)
- Batch-specific validation
- `ingestion_batches` tracking table
- Custom batch examples in SDKs

### ✅ Now Available (OTLP Standard)
- `POST /v1/otlp/traces` - OTLP traces ingestion (Protobuf + JSON)
- `POST /v1/traces` - Alias for `/otlp/traces`
- Standard OpenTelemetry SDK integration
- W3C Trace Context propagation
- Gzip compression support
- Future: `/v1/otlp/metrics`, `/v1/otlp/logs`

---

## Migration Path

### Python: OpenTelemetry SDK

#### Before (Custom Batch API)
```python
from brokle_legacy.types.telemetry import TelemetryEventType

client.submit_batch([
    {
        "event_type": TelemetryEventType.TRACE,
        "payload": {
            "name": "my-operation",
            "user_id": "user-123",
            "duration_ms": 150
        }
    }
])
```

#### After (OpenTelemetry Standard)
```python
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter

# Configure Brokle OTLP exporter
exporter = OTLPSpanExporter(
    endpoint="http://localhost:8080/v1/otlp/traces",
    headers={"X-API-Key": "bk_your_secret"}
)

# Setup tracer provider
provider = TracerProvider()
provider.add_span_processor(BatchSpanProcessor(exporter))
trace.set_tracer_provider(provider)

# Use standard OpenTelemetry API
tracer = trace.get_tracer(__name__)

with tracer.start_as_current_span("my-operation") as span:
    span.set_attribute("user_id", "user-123")
    span.set_attribute("custom.attribute", "value")
    # Operation code here
    # Span automatically ends with duration
```

### JavaScript/TypeScript: OpenTelemetry SDK

#### Setup
```bash
npm install @opentelemetry/sdk-node \
            @opentelemetry/auto-instrumentations-node \
            @opentelemetry/exporter-trace-otlp-http
```

#### Implementation
```typescript
import { NodeSDK } from '@opentelemetry/sdk-node';
import { getNodeAutoInstrumentations } from '@opentelemetry/auto-instrumentations-node';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';

// Configure Brokle OTLP exporter
const traceExporter = new OTLPTraceExporter({
  url: 'http://localhost:8080/v1/otlp/traces',
  headers: {
    'X-API-Key': 'bk_your_secret'
  }
});

// Initialize SDK with auto-instrumentation
const sdk = new NodeSDK({
  traceExporter,
  instrumentations: [getNodeAutoInstrumentations()]
});

sdk.start();

// Use standard OpenTelemetry API
import { trace } from '@opentelemetry/api';

const tracer = trace.getTracer('my-service');

tracer.startActiveSpan('my-operation', (span) => {
  span.setAttribute('user_id', 'user-123');
  // Operation code here
  span.end();
});
```

### Go: OpenTelemetry SDK

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
    "go.opentelemetry.io/otel/sdk/trace"
)

// Configure Brokle OTLP exporter
exporter, err := otlptracehttp.New(
    context.Background(),
    otlptracehttp.WithEndpoint("localhost:8080"),
    otlptracehttp.WithURLPath("/v1/otlp/traces"),
    otlptracehttp.WithHeaders(map[string]string{
        "X-API-Key": "bk_your_secret",
    }),
    otlptracehttp.WithInsecure(), // Use WithTLSConfig for production
)

// Setup tracer provider
tp := trace.NewTracerProvider(
    trace.WithBatcher(exporter),
)
otel.SetTracerProvider(tp)

// Use standard OpenTelemetry API
tracer := otel.Tracer("my-service")

ctx, span := tracer.Start(ctx, "my-operation")
defer span.End()

span.SetAttributes(
    attribute.String("user_id", "user-123"),
    attribute.String("custom.attr", "value"),
)
```

---

## Advanced Configuration

### Compression (Gzip)

OTLP supports gzip compression for reduced bandwidth:

```python
# Python
from opentelemetry.exporter.otlp.proto.http.trace_exporter import Compression

exporter = OTLPSpanExporter(
    endpoint="http://localhost:8080/v1/otlp/traces",
    headers={"X-API-Key": "bk_your_secret"},
    compression=Compression.Gzip  # Enable gzip compression
)
```

```typescript
// JavaScript
const traceExporter = new OTLPTraceExporter({
  url: 'http://localhost:8080/v1/otlp/traces',
  headers: { 'X-API-Key': 'bk_your_secret' },
  compression: 'gzip'  // Enable gzip
});
```

### Batch Configuration

```python
# Python - Configure batching behavior
from opentelemetry.sdk.trace.export import BatchSpanProcessor

processor = BatchSpanProcessor(
    exporter,
    max_queue_size=2048,        # Max spans in queue
    schedule_delay_millis=5000,  # Export every 5 seconds
    max_export_batch_size=512,   # Max spans per batch
)

provider = TracerProvider()
provider.add_span_processor(processor)
```

### Sampling

```python
# Python - Sample 10% of traces (cost optimization)
from opentelemetry.sdk.trace.sampling import TraceIdRatioBased

provider = TracerProvider(
    sampler=TraceIdRatioBased(0.1)  # Sample 10%
)
```

---

## Brokle-Specific Attributes

### Project Context

The Brokle backend automatically extracts `project_id` from your API key. No additional configuration needed!

### Custom Attributes (Brokle Extensions)

Use semantic attribute conventions:

```python
span.set_attribute("brokle.environment", "production")
span.set_attribute("brokle.version", "v1.2.3")
span.set_attribute("brokle.user_id", "user-123")
span.set_attribute("brokle.session_id", "session-456")

# LLM-specific attributes
span.set_attribute("gen_ai.system", "openai")
span.set_attribute("gen_ai.request.model", "gpt-4")
span.set_attribute("gen_ai.request.temperature", 0.7)
span.set_attribute("gen_ai.usage.input_tokens", 100)
span.set_attribute("gen_ai.usage.output_tokens", 50)
```

### Span Types

Brokle extends OTLP with span type classification:

```python
# Set span type via attribute
span.set_attribute("brokle.span_type", "generation")  # LLM generation
span.set_attribute("brokle.span_type", "span")        # Generic span
span.set_attribute("brokle.span_type", "tool")        # Tool call
span.set_attribute("brokle.span_type", "agent")       # Agent execution
span.set_attribute("brokle.span_type", "retrieval")   # RAG retrieval
span.set_attribute("brokle.span_type", "event")       # Generic event
```

---

## Deduplication

### How It Works

Brokle automatically deduplicates OTLP spans using deterministic ULID generation:

- **Trace Deduplication**: Derived from OTLP `trace_id`
- **Span Deduplication**: Derived from OTLP `span_id`
- **TTL**: 24 hours (configurable)
- **Storage**: Redis (atomic claims)

### Retry Safety

If you retry a failed OTLP export, Brokle automatically detects and skips duplicate spans:

```python
# Automatic retry with deduplication
processor = BatchSpanProcessor(
    exporter,
    # Retries handled by OTLP exporter
    # Brokle deduplication prevents duplicates
)
```

Response for duplicate spans:
```json
{
  "status": "all_duplicates",
  "duplicates": 5
}
```

---

## Breaking Changes

### API Changes

| Old Endpoint | New Endpoint | Status |
|--------------|--------------|--------|
| `POST /v1/ingest/batch` | `POST /v1/otlp/traces` | **REPLACED** |
| `POST /v1/telemetry/validate` | N/A (OTLP schema validation) | **REMOVED** |
| `/v1/telemetry/health` | `/v1/observability/health` (future) | **MIGRATED** |

### SDK Changes

| Old SDK | New SDK | Migration |
|---------|---------|-----------|
| `brokle_legacy` (Python) | OpenTelemetry SDK | See examples above |
| Custom batch methods | Standard OTLP TracerProvider | Industry standard |
| `TelemetryEventType` enum | OTLP span attributes | Semantic conventions |

---

## Migration Checklist

### For Application Developers

- [ ] Install OpenTelemetry SDK for your language
- [ ] Configure OTLP exporter with Brokle endpoint
- [ ] Add `X-API-Key` header to exporter
- [ ] Replace custom batch calls with OpenTelemetry API
- [ ] Test with Brokle backend (`/v1/otlp/traces`)
- [ ] Verify traces appear in Brokle dashboard

### For Platform Operators

- [ ] Update API documentation (remove batch endpoints)
- [ ] Update SDK examples (add OTLP integration)
- [ ] Monitor OTLP endpoint usage
- [ ] Verify deduplication working (check Redis)
- [ ] Check ClickHouse for traces/spans data

---

## Troubleshooting

### OTLP Export Fails

**Error**: "Failed to export spans"

**Solutions**:
1. Verify endpoint: `http://localhost:8080/v1/otlp/traces`
2. Check API key header: `X-API-Key: bk_your_secret`
3. Test manually:
   ```bash
   curl -X POST http://localhost:8080/v1/otlp/traces \
     -H "X-API-Key: bk_your_secret" \
     -H "Content-Type: application/json" \
     -d '{"resourceSpans": []}'
   ```

### Authentication Errors

**Error**: 401 Unauthorized

**Solutions**:
- Verify API key is valid
- Check header format: `X-API-Key` (not `Authorization`)
- Test key validation:
  ```bash
  curl -X POST http://localhost:8080/v1/auth/validate-key \
    -H "Content-Type: application/json" \
    -d '{"api_key": "bk_your_secret"}'
  ```

### No Data Appearing

**Checklist**:
1. ✅ OTLP exporter configured correctly?
2. ✅ Tracer provider registered (`otel.SetTracerProvider`)?
3. ✅ Spans actually created?
4. ✅ Exporter flushed (important for short-lived apps)?
5. ✅ Check Brokle logs for errors

---

## Benefits of OTLP Migration

### Industry Standard
- Works with Jaeger, Zipkin, Datadog, New Relic, etc.
- Standard semantic conventions
- Vendor-neutral observability

### Better Performance
- Native protobuf encoding (~40% smaller than JSON)
- Efficient batch processing
- Built-in compression support

### Rich Ecosystem
- Auto-instrumentation libraries
- Framework integrations (Express, Flask, Gin, etc.)
- Database/HTTP/gRPC automatic tracing

### Future-Proof
- OTLP Metrics Protocol (coming soon)
- OTLP Logs Protocol (roadmap)
- OpenTelemetry is CNCF graduated project

---

## Resources

### OpenTelemetry Documentation
- **Official Docs**: https://opentelemetry.io/docs/
- **Python SDK**: https://opentelemetry.io/docs/languages/python/
- **JavaScript SDK**: https://opentelemetry.io/docs/languages/js/
- **Go SDK**: https://opentelemetry.io/docs/languages/go/

### Brokle-Specific
- **API Reference**: https://docs.brokle.com/api
- **OTLP Integration Guide**: https://docs.brokle.com/integrations/otlp
- **GitHub Issues**: https://github.com/brokle/brokle/issues

### Example Repositories
- **Python OTLP Examples**: `sdk/python/examples/otlp_*`
- **Brokle Test Suite**: `test/otel-collector/` (OTLP Collector integration)

---

## FAQ

### Q: Can I still use the custom batch API?
**A**: No, it has been completely removed. All applications must migrate to OTLP.

### Q: Will my existing data be affected?
**A**: No. All existing trace/span/score data remains in ClickHouse. Only the ingestion method changes.

### Q: Do I need to change my ClickHouse queries?
**A**: No. Data is still stored in the same `traces`, `spans`, and `quality_scores` tables with the same schema.

### Q: Is deduplication still supported?
**A**: Yes! Deduplication works identically with OTLP span IDs (deterministic ULID generation).

### Q: What about performance?
**A**: OTLP is actually **faster** than the custom batch API due to native protobuf encoding and no envelope overhead.

### Q: Can I use both OTLP and custom wrappers?
**A**: Use OpenTelemetry SDK only. The Brokle wrappers (`wrap_openai`, etc.) may be deprecated in favor of OTLP auto-instrumentation.

---

**Migration Date**: October 29, 2025
**Status**: ✅ Complete - OTLP-Only Architecture
**No Backward Compatibility**: Hard cutover (no users affected in early development)
