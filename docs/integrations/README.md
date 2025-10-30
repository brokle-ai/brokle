# Brokle Integration Guide

This directory contains integration guides for connecting your applications to Brokle.

---

## Quick Decision: Which Integration Method?

### Decision Tree

```
START: How do you want to integrate with Brokle?
│
├─> "Simplest setup, lowest latency"
│   └─> ✅ Use Brokle SDK
│       📖 Guide: direct-sdk.md (coming soon)
│
├─> "I already run an OTEL Collector"
│   └─> ✅ Add Brokle as OTLP exporter
│       📖 Guide: opentelemetry-collector.md
│
├─> "I want vendor-agnostic integration"
│   └─> ✅ Use OpenTelemetry SDK directly
│       📖 Guide: direct-otlp.md (coming soon)
│
└─> "I need multiple backends (Brokle + Datadog)"
    └─> ✅ Use OTEL Collector with fan-out
        📖 Guide: opentelemetry-collector.md
```

---

## Integration Methods Comparison

| Method | Latency | Complexity | Vendor Lock-in | Advanced Features | Best For |
|--------|---------|------------|----------------|-------------------|----------|
| **Brokle SDK** | 5-10ms | Low | Yes (Brokle) | ✅ Built-in batching/retry | Most users |
| **OTEL Collector** | 50-100ms | Medium | No (OTLP standard) | ✅ Sampling, PII, multi-backend | Enterprises |
| **Direct OTLP** | 5-10ms | Low | No (OTLP standard) | ❌ Manual batching/retry | OTEL-native teams |

---

## Method 1: Brokle SDK (Recommended)

**What it is**: Official Brokle SDK with built-in batching, retry, and error handling.

**Available for:**
- Python (`pip install brokle`)
- JavaScript/TypeScript (`npm install brokle`)

**Pros:**
- ✅ **Lowest latency** (~5-10ms end-to-end)
- ✅ **Simplest setup** (3 lines of code)
- ✅ **Built-in best practices** (batching, retry, error handling)
- ✅ **Automatic instrumentation** for popular frameworks
- ✅ **Type-safe** (TypeScript, Python type hints)

**Cons:**
- ❌ Vendor lock-in (Brokle-specific SDK)
- ❌ No multi-backend support (Brokle only)

**When to use:**
- First time using Brokle
- Want lowest latency
- Don't need multi-backend
- Prefer simplicity over flexibility

**Quick start:**
```python
# Python
from brokle import Brokle

brokle = Brokle(api_key="bk_your_key")
brokle.trace("my-operation", lambda: do_work())
```

```typescript
// TypeScript
import { Brokle } from 'brokle';

const brokle = new Brokle({ apiKey: 'bk_your_key' });
await brokle.trace('my-operation', async () => await doWork());
```

**Documentation**: `direct-sdk.md` (coming soon)

---

## Method 2: OpenTelemetry Collector (Enterprise)

**What it is**: Industry-standard OTEL Collector that forwards traces to Brokle.

**Pros:**
- ✅ **Vendor-agnostic** (standard OTLP protocol)
- ✅ **Multi-backend** (send to Brokle + Datadog + Jaeger simultaneously)
- ✅ **Advanced processing** (tail sampling, PII scrubbing, filtering)
- ✅ **Cost optimization** (95% reduction with tail sampling)
- ✅ **Compliance** (remove PII before data leaves network)
- ✅ **Already familiar** (if you use OTEL ecosystem)

**Cons:**
- ❌ Higher latency (50-100ms vs 5-10ms)
- ❌ More complex setup (collector deployment and config)
- ❌ Another service to manage

**When to use:**
- Already run OTEL Collector for other backends
- Need tail-based sampling (high volume >100K spans/day)
- Compliance requires PII scrubbing
- Want to send to multiple backends (Brokle + others)
- Evaluating Brokle alongside existing tools

**Quick start:**
```yaml
# collector-config.yaml
receivers:
  otlp:
    protocols:
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    timeout: 1s

exporters:
  otlphttp/brokle:
    endpoint: https://api.brokle.com/v1/traces
    headers:
      X-API-Key: ${BROKLE_API_KEY}

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlphttp/brokle]
```

**Documentation**: [opentelemetry-collector.md](./opentelemetry-collector.md)

---

## Method 3: Direct OTLP (Vendor-Agnostic)

**What it is**: Use OpenTelemetry SDK directly, export OTLP to Brokle API.

**Pros:**
- ✅ **Vendor-agnostic** (standard OTLP, easy to switch)
- ✅ **Low latency** (~5-10ms, same as Brokle SDK)
- ✅ **No collector needed** (direct export)
- ✅ **Widely supported** (OTEL SDKs for 11+ languages)

**Cons:**
- ❌ Manual batching/retry configuration
- ❌ No multi-backend (without collector)
- ❌ More boilerplate than Brokle SDK

**When to use:**
- Want OTEL-native instrumentation
- Concerned about vendor lock-in
- Already using OTEL SDKs
- Don't need advanced processing (sampling, PII)

**Quick start:**
```python
# Python
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor

exporter = OTLPSpanExporter(
    endpoint="https://api.brokle.com/v1/traces",
    headers={"X-API-Key": "bk_your_key"}
)
provider = TracerProvider()
provider.add_span_processor(BatchSpanProcessor(exporter))
```

**Documentation**: `direct-otlp.md` (coming soon)

---

## Feature Comparison

| Feature | Brokle SDK | OTEL Collector | Direct OTLP |
|---------|-----------|----------------|-------------|
| **Setup Time** | 5 minutes | 30 minutes | 15 minutes |
| **Latency (p99)** | 5-10ms | 50-100ms | 5-10ms |
| **Vendor Lock-in** | Yes | No | No |
| **Multi-Backend** | No | Yes | No |
| **Tail Sampling** | No | Yes | No |
| **PII Scrubbing** | No | Yes | Manual |
| **Auto-Instrumentation** | Yes | No | Partial |
| **Type Safety** | Yes | N/A | Yes |
| **Production-Ready** | Yes | Yes | Yes (with manual config) |

---

## Use Case Examples

### Scenario 1: Startup Building MVP

**Recommendation**: Brokle SDK
**Why**: Fastest time to value, lowest complexity
**Guide**: `direct-sdk.md`

---

### Scenario 2: Enterprise Migrating from Datadog

**Recommendation**: OTEL Collector (multi-backend)
**Why**: Run both platforms during evaluation period
**Guide**: `opentelemetry-collector.md` → `02-multi-backend.yaml`

**Example:**
```yaml
# Send to both Brokle and Datadog
exporters: [otlphttp/brokle, otlp/datadog]
```

**Timeline**:
- Week 1-2: Add Brokle alongside Datadog
- Week 3-4: Build dashboards in Brokle
- Week 5+: Remove Datadog exporter

---

### Scenario 3: High-Traffic Application (1M+ spans/day)

**Recommendation**: OTEL Collector (tail sampling)
**Why**: 95% cost reduction while keeping all errors
**Guide**: `opentelemetry-collector.md` → `03-tail-sampling.yaml`

**Cost impact:**
- Before: 1M spans/day × $0.10/1K = $100/day ($3,000/month)
- After: 40K spans/day × $0.10/1K = $4/day ($120/month)
- **Savings: $2,880/month** (96% reduction)

---

### Scenario 4: Healthcare/Finance (GDPR/HIPAA Compliance)

**Recommendation**: OTEL Collector (PII scrubbing)
**Why**: Remove PII before data leaves your network
**Guide**: `opentelemetry-collector.md` → `04-pii-scrubbing.yaml`

**What gets removed:**
- Email addresses
- IP addresses (hashed)
- Credit card numbers
- Phone numbers
- SSNs, authentication tokens

**Compliance**: GDPR Article 32, HIPAA Privacy Rule, PCI DSS 3.4

---

### Scenario 5: OTEL-Native Team

**Recommendation**: Direct OTLP
**Why**: Already using OTEL SDKs, want vendor-agnostic integration
**Guide**: `direct-otlp.md`

**Example:**
```go
// Go
exporter, _ := otlptrace.New(ctx,
    otlptrace.WithEndpoint("https://api.brokle.com/v1/traces"),
    otlptrace.WithHeaders(map[string]string{
        "X-API-Key": "bk_your_key",
    }),
)
```

---

## Integration Decision Matrix

| Your Situation | Recommended Method | Config/Guide |
|----------------|-------------------|--------------|
| New to Brokle, want simple setup | Brokle SDK | `direct-sdk.md` |
| Already use OTEL Collector | Add Brokle exporter | `opentelemetry-collector.md` |
| Evaluating Brokle vs Datadog | OTEL Collector (multi-backend) | `02-multi-backend.yaml` |
| High volume (>100K spans/day) | OTEL Collector (tail sampling) | `03-tail-sampling.yaml` |
| GDPR/HIPAA compliance | OTEL Collector (PII scrubbing) | `04-pii-scrubbing.yaml` |
| Want OTEL-native | Direct OTLP | `direct-otlp.md` |
| Kubernetes deployment | OTEL Collector (DaemonSet) | `opentelemetry-collector.md` |

---

## Migration Paths

### From Datadog to Brokle

1. **Phase 1**: Add OTEL Collector with multi-backend
   ```yaml
   exporters: [otlp/datadog, otlphttp/brokle]
   ```

2. **Phase 2**: Validate Brokle (2-4 weeks)
   - Build dashboards
   - Train team
   - Ensure feature parity

3. **Phase 3**: Remove Datadog exporter
   ```yaml
   exporters: [otlphttp/brokle]  # Brokle only
   ```

**Guide**: [opentelemetry-collector.md](./opentelemetry-collector.md) → Multi-Backend section

---

### From Jaeger to Brokle

Same pattern as Datadog:
1. Add Brokle exporter alongside Jaeger
2. Validate over 2-4 weeks
3. Remove Jaeger exporter

---

### From Custom Instrumentation to Brokle SDK

1. **Phase 1**: Add Brokle SDK alongside custom code
2. **Phase 2**: Gradually replace custom instrumentation
3. **Phase 3**: Remove custom code

**Benefit**: Less code to maintain, better performance

---

## Getting Help

### Documentation

- **OTEL Collector**: [opentelemetry-collector.md](./opentelemetry-collector.md)
- **API Reference**: [../API.md](../API.md)
- **Architecture**: [../ARCHITECTURE.md](../ARCHITECTURE.md)

### Example Configurations

- [Basic Setup](../../examples/otel-collector/01-basic.yaml)
- [Multi-Backend](../../examples/otel-collector/02-multi-backend.yaml)
- [Tail Sampling](../../examples/otel-collector/03-tail-sampling.yaml)
- [PII Scrubbing](../../examples/otel-collector/04-pii-scrubbing.yaml)

### Support

- GitHub Issues: https://github.com/brokle/brokle/issues
- Email: support@brokle.com
- Slack: https://brokle.com/slack

---

## Available Guides

### Current (Available Now)

- ✅ [OpenTelemetry Collector Integration](./opentelemetry-collector.md)
  - Complete guide for using OTEL Collector with Brokle
  - 4 production-ready configurations
  - Deployment patterns (Docker, Kubernetes)
  - Troubleshooting and performance tuning

### Coming Soon

- ⏳ Direct SDK Integration (`direct-sdk.md`)
  - Brokle SDK quick start
  - Language-specific guides (Python, JavaScript, TypeScript)
  - Best practices and patterns

- ⏳ Direct OTLP Integration (`direct-otlp.md`)
  - Using OpenTelemetry SDKs without Brokle SDK
  - Configuration examples for all languages
  - Comparison with Brokle SDK

---

## Quick Start by Language

### Python

```python
# Option 1: Brokle SDK (simplest)
from brokle import Brokle
brokle = Brokle(api_key="bk_your_key")

# Option 2: OTEL SDK (vendor-agnostic)
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
exporter = OTLPSpanExporter(
    endpoint="https://api.brokle.com/v1/traces",
    headers={"X-API-Key": "bk_your_key"}
)
```

### JavaScript/TypeScript

```typescript
// Option 1: Brokle SDK (simplest)
import { Brokle } from 'brokle';
const brokle = new Brokle({ apiKey: 'bk_your_key' });

// Option 2: OTEL SDK (vendor-agnostic)
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';
const exporter = new OTLPTraceExporter({
  url: 'https://api.brokle.com/v1/traces',
  headers: { 'X-API-Key': 'bk_your_key' }
});
```

### Go

```go
// Option: OTEL SDK
import "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"

exporter, _ := otlptracehttp.New(ctx,
    otlptracehttp.WithEndpointURL("https://api.brokle.com/v1/traces"),
    otlptracehttp.WithHeaders(map[string]string{
        "X-API-Key": "bk_your_key",
    }),
)
```

---

## Next Steps

1. **Choose integration method** using decision tree above
2. **Follow the guide** for your chosen method
3. **Deploy to production** and verify traces appear
4. **Optimize as needed** (add sampling, PII scrubbing, etc.)

---

## Additional Resources

- [Example Configurations](../../examples/otel-collector/) - Production-ready configs
- [Integration Tests](../../test/otel-collector/) - Test infrastructure
- [API Documentation](../API.md) - OTLP endpoint details
- [Architecture](../ARCHITECTURE.md) - System design
- [OpenTelemetry Docs](https://opentelemetry.io/docs/) - OTEL official docs
