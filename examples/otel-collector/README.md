# OpenTelemetry Collector Examples for Brokle Integration

This directory contains production-ready OTEL Collector configurations for integrating with Brokle in various scenarios.

---

## Quick Decision Matrix

**Which config should I use?**

| Your Scenario | Recommended Config | Why |
|---------------|-------------------|-----|
| Just getting started | `01-basic.yaml` | Simplest setup, single backend |
| Evaluating Brokle alongside Datadog/Jaeger | `02-multi-backend.yaml` | A/B testing, gradual migration |
| High traffic (>100K spans/day) | `03-tail-sampling.yaml` | 95% cost reduction |
| GDPR/HIPAA/PCI compliance | `04-pii-scrubbing.yaml` | Removes PII before export |
| Multiple requirements | Combine configs | Mix sampling + PII + multi-backend |

---

## Configuration Files

### 01-basic.yaml - Simple Setup

**Use when:**
- First time using OTEL Collector with Brokle
- Development or proof-of-concept
- Small to medium traffic (<100K spans/day)
- Single observability backend

**Key features:**
- ✅ Minimal configuration
- ✅ Single backend (Brokle only)
- ✅ Basic batching (100 spans / 1s)
- ✅ Health check endpoint
- ✅ Auto-retry on failures

**Expected performance:**
- Latency: 50-100ms (p99)
- Throughput: ~10K spans/sec
- Memory: ~50MB

**Quick start:**
```bash
export BROKLE_API_KEY="bk_your_key_here"
otelcol --config=01-basic.yaml
```

---

### 02-multi-backend.yaml - Fan-Out Pattern

**Use when:**
- Migrating from existing observability platform
- Need to send to multiple backends (Brokle + Datadog + Jaeger)
- A/B testing Brokle vs current solution
- Compliance requires data in multiple locations

**Key features:**
- ✅ Parallel export to 3+ backends
- ✅ Independent retry logic per backend
- ✅ Single collector instance
- ✅ Resource attribute enrichment

**Example use case:**
```
Your App → Collector → ├─> Brokle (AI observability)
                        ├─> Datadog (APM monitoring)
                        └─> Jaeger (distributed tracing)
```

**Expected performance:**
- Latency: Max of all backend latencies
- Throughput: Limited by slowest backend
- Memory: ~100MB (buffering for all backends)

**Cost consideration:**
- ⚠️ Each span sent to ALL backends (3x ingestion costs)
- 💡 Use tail sampling to reduce volume if needed

**Quick start:**
```bash
export BROKLE_API_KEY="bk_your_key"
export DATADOG_API_KEY="your_dd_key"
export DEPLOYMENT_ENV="production"
otelcol --config=02-multi-backend.yaml
```

---

### 03-tail-sampling.yaml - Cost Optimization

**Use when:**
- High traffic (>100K spans/day)
- Concerned about ingestion costs
- Want to keep all errors but sample success
- Need intelligent trace selection

**Key features:**
- ✅ Keeps 100% of errors
- ✅ Keeps 100% of slow traces (>1s)
- ✅ Keeps 100% of VIP/enterprise user traces
- ✅ Samples 1% of remaining traces
- ✅ 95-98% cost reduction

**Sampling decision:**
```
Input: 1,000,000 spans/day
  ├─> Errors (1%):           10,000 spans → KEEP ALL
  ├─> Slow (2%):             20,000 spans → KEEP ALL
  ├─> VIP users (0.1%):       1,000 spans → KEEP ALL
  └─> Remaining (96.9%):    969,000 spans → SAMPLE 1% (9,690 spans)

Output: 40,690 spans/day (96% reduction)
Cost: $100/day → $4/day (96% savings)
```

**Trade-offs:**
- ⚠️ 10-second latency (buffering for tail sampling decision)
- ⚠️ Higher memory usage (~1GB for 100K trace buffer)
- ⚠️ May miss incomplete traces if they take >10s to complete

**Expected performance:**
- Latency: 1-5 seconds (decision_wait time)
- Throughput: 1M+ spans/sec input, ~20K spans/sec output
- Memory: ~1GB (100K traces × 10KB avg)

**Quick start:**
```bash
export BROKLE_API_KEY="bk_your_key"
otelcol --config=03-tail-sampling.yaml
```

---

### 04-pii-scrubbing.yaml - Data Governance

**Use when:**
- GDPR/HIPAA/PCI DSS compliance required
- Regulated industry (healthcare, finance, government)
- Data must be scrubbed before leaving network
- Privacy-by-design architecture

**Key features:**
- ✅ Removes emails, phones, SSNs, credit cards
- ✅ Hashes IP addresses (preserves analytics)
- ✅ Redacts PII from span names and messages
- ✅ Deletes authentication tokens
- ✅ Regex-based pattern matching

**PII removal examples:**
```
Before scrubbing:
  span.name: "User john@example.com logged in"
  user.email: "john@example.com"
  client.ip: "192.168.1.1"
  user.phone: "555-123-4567"
  payment.card: "4532-1234-5678-9010"

After scrubbing:
  span.name: "User [EMAIL_REDACTED] logged in"
  user.email: DELETED
  client.ip: "hash_a3f8b2c1" (hashed for analytics)
  user.phone: DELETED
  payment.card: "[CARD_REDACTED]"
```

**Compliance:**
- ✅ GDPR Article 32 (Data minimization)
- ✅ HIPAA Privacy Rule (De-identification)
- ✅ PCI DSS Requirement 3.4 (Masking PAN)
- ✅ CCPA (Data minimization)

**Expected performance:**
- Latency: 50-100ms (minimal processing overhead)
- Throughput: ~10K spans/sec
- Memory: ~50MB

**Quick start:**
```bash
export BROKLE_API_KEY="bk_your_key"
otelcol --config=04-pii-scrubbing.yaml
```

---

## Combining Configurations

### Example: Multi-Backend + PII Scrubbing

```yaml
# Combine processors from both configs
processors:
  attributes:
    # ... PII removal from 04-pii-scrubbing.yaml

  transform:
    # ... Regex scrubbing from 04-pii-scrubbing.yaml

  batch:
    # ... Batching config

exporters:
  otlphttp/brokle:
    # ... Brokle config

  otlp/datadog:
    # ... Datadog config

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [attributes, transform, batch]  # PII removed first
      exporters: [otlphttp/brokle, otlp/datadog]   # Then fan out
```

### Example: Tail Sampling + PII Scrubbing

```yaml
processors:
  attributes:
    # ... PII removal

  tail_sampling:
    # ... Sampling policies

  batch:
    # ... Batching

service:
  pipelines:
    traces:
      receivers: [otlp]
      # Order: PII first, then sample, then batch
      processors: [attributes, tail_sampling, batch]
      exporters: [otlphttp/brokle]
```

---

## Feature Comparison

| Feature | Basic | Multi-Backend | Tail Sampling | PII Scrubbing |
|---------|-------|---------------|---------------|---------------|
| **Backends** | 1 (Brokle) | 3+ | 1 (Brokle) | 1 (Brokle) |
| **Latency** | 50-100ms | 50-100ms | 1-5s | 50-100ms |
| **Memory** | 50MB | 100MB | 1GB | 50MB |
| **Cost Reduction** | 0% | 0% | 95-98% | 0% |
| **Compliance** | No | No | No | GDPR/HIPAA/PCI |
| **Use Case** | Simple | A/B testing | High volume | Regulated |

---

## Performance Characteristics

### Latency Comparison

| Config | p50 | p95 | p99 | Notes |
|--------|-----|-----|-----|-------|
| Basic | 20ms | 50ms | 100ms | Single hop |
| Multi-Backend | 30ms | 80ms | 150ms | Max of all backends |
| Tail Sampling | 1.5s | 3s | 5s | Buffering for decision |
| PII Scrubbing | 25ms | 60ms | 120ms | Attribute processing |

### Throughput Limits

| Config | Max Throughput | Bottleneck |
|--------|---------------|------------|
| Basic | ~10K spans/sec | ClickHouse writes |
| Multi-Backend | ~5K spans/sec | Slowest backend |
| Tail Sampling | 1M+ input, 20K output | Sampling ratio |
| PII Scrubbing | ~10K spans/sec | Regex processing |

---

## Environment Variables

All configs support these environment variables:

### Required:
```bash
# Brokle API authentication
export BROKLE_API_KEY="bk_your_api_key_here"
```

### Optional (multi-backend):
```bash
# Datadog (if using 02-multi-backend.yaml)
export DATADOG_API_KEY="your_datadog_api_key"

# Deployment metadata
export DEPLOYMENT_ENV="production"  # dev, staging, production
```

### Testing locally:
```bash
# For local testing, use localhost endpoints
export BROKLE_API_ENDPOINT="http://localhost:8080/v1/traces"
```

---

## Installation

### Download OTEL Collector

```bash
# Linux (x86_64)
wget https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v0.91.0/otelcol-contrib_0.91.0_linux_amd64.tar.gz
tar -xzf otelcol-contrib_0.91.0_linux_amd64.tar.gz
chmod +x otelcol-contrib

# macOS (x86_64)
wget https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v0.91.0/otelcol-contrib_0.91.0_darwin_amd64.tar.gz
tar -xzf otelcol-contrib_0.91.0_darwin_amd64.tar.gz
chmod +x otelcol-contrib

# macOS (ARM64 / M1/M2)
wget https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v0.91.0/otelcol-contrib_0.91.0_darwin_arm64.tar.gz
tar -xzf otelcol-contrib_0.91.0_darwin_arm64.tar.gz
chmod +x otelcol-contrib
```

### Run Collector

```bash
# Basic example
./otelcol-contrib --config=01-basic.yaml

# With environment variables
BROKLE_API_KEY="bk_..." ./otelcol-contrib --config=02-multi-backend.yaml
```

---

## Docker Deployment

### Docker Run

```bash
docker run -d \
  --name otel-collector \
  -p 4317:4317 \
  -p 4318:4318 \
  -p 13133:13133 \
  -e BROKLE_API_KEY="bk_your_key" \
  -v $(pwd)/01-basic.yaml:/etc/otel-collector-config.yaml \
  otel/opentelemetry-collector-contrib:0.91.0 \
  --config=/etc/otel-collector-config.yaml
```

### Docker Compose

```yaml
services:
  otel-collector:
    image: otel/opentelemetry-collector-contrib:0.91.0
    command: ["--config=/etc/otel-collector-config.yaml"]
    volumes:
      - ./01-basic.yaml:/etc/otel-collector-config.yaml
    ports:
      - "4317:4317"
      - "4318:4318"
      - "13133:13133"
    environment:
      - BROKLE_API_KEY=${BROKLE_API_KEY}
```

---

## Kubernetes Deployment

### DaemonSet Pattern (Node-level collection)

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: otel-collector
spec:
  selector:
    matchLabels:
      app: otel-collector
  template:
    metadata:
      labels:
        app: otel-collector
    spec:
      containers:
      - name: otel-collector
        image: otel/opentelemetry-collector-contrib:0.91.0
        args: ["--config=/conf/otel-collector-config.yaml"]
        ports:
        - containerPort: 4317  # OTLP gRPC
        - containerPort: 4318  # OTLP HTTP
        - containerPort: 13133 # Health check
        env:
        - name: BROKLE_API_KEY
          valueFrom:
            secretKeyRef:
              name: brokle-credentials
              key: api-key
        volumeMounts:
        - name: config
          mountPath: /conf
      volumes:
      - name: config
        configMap:
          name: otel-collector-config
```

### Sidecar Pattern (Pod-level collection)

```yaml
# Add to your existing Pod spec
spec:
  containers:
  # Your application container
  - name: app
    image: your-app:latest
    env:
    - name: OTEL_EXPORTER_OTLP_ENDPOINT
      value: "http://localhost:4318"  # Send to local collector

  # OTEL Collector sidecar
  - name: otel-collector
    image: otel/opentelemetry-collector-contrib:0.91.0
    args: ["--config=/conf/otel-collector-config.yaml"]
    ports:
    - containerPort: 4317
    - containerPort: 4318
    - containerPort: 13133
    env:
    - name: BROKLE_API_KEY
      valueFrom:
        secretKeyRef:
          name: brokle-credentials
          key: api-key
    volumeMounts:
    - name: config
      mountPath: /conf
  volumes:
  - name: config
    configMap:
      name: otel-collector-config
```

---

## Testing Configurations

### 1. Validate Config Syntax

```bash
# Dry-run (validates syntax without starting)
otelcol --config=01-basic.yaml --dry-run

# Expected output: "Config validation succeeded"
```

### 2. Test Health Check

```bash
# Start collector
otelcol --config=01-basic.yaml &

# Wait for startup (2-3 seconds)
sleep 3

# Check health
curl http://localhost:13133

# Expected: {"status":"Server available","upSince":"..."}
```

### 3. Send Test Trace

```bash
# Using curl (raw OTLP JSON)
curl -X POST http://localhost:4318/v1/traces \
  -H "Content-Type: application/json" \
  -d '{
    "resourceSpans": [{
      "scopeSpans": [{
        "spans": [{
          "traceId": "5b8efff798038103d269b633813fc60c",
          "spanId": "eee19b7ec3c1b174",
          "name": "test-span",
          "startTimeUnixNano": "1698768000000000000",
          "endTimeUnixNano": "1698768001000000000",
          "attributes": [
            {"key": "test", "value": {"stringValue": "true"}}
          ]
        }]
      }]
    }]
  }'
```

### 4. Monitor Collector

```bash
# Watch collector logs (look for export success)
otelcol --config=01-basic.yaml 2>&1 | grep -i "export"

# Expected output (every 1-5 seconds):
# "Traces exported" {"count": 100, "backend": "otlphttp/brokle"}
```

---

## Common Issues

### Issue 1: "401 Unauthorized"

**Cause**: Invalid or missing API key

**Solution:**
```bash
# Verify API key is set
echo $BROKLE_API_KEY

# Verify it starts with "bk_"
if [[ $BROKLE_API_KEY == bk_* ]]; then
  echo "API key format is correct"
else
  echo "Invalid API key format (should start with 'bk_')"
fi

# Test API key manually
curl -X POST https://api.brokle.com/v1/traces \
  -H "X-API-Key: $BROKLE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{}'
# Should return 400 (bad request), not 401 (proves auth works)
```

---

### Issue 2: "Connection refused"

**Cause**: Brokle API endpoint unreachable

**Solution:**
```bash
# Test connectivity
curl -v https://api.brokle.com/health

# For local testing
curl -v http://localhost:8080/health

# Check firewall/network
ping api.brokle.com
```

---

### Issue 3: High memory usage (tail sampling)

**Cause**: Too many traces buffered in memory

**Solution:**
```yaml
# Reduce buffer size in 03-tail-sampling.yaml
tail_sampling:
  num_traces: 50000  # Reduce from 100,000
  decision_wait: 5s   # Reduce from 10s
```

---

### Issue 4: Traces incomplete in Brokle

**Cause**: decision_wait too short for long-running traces

**Solution:**
```yaml
# Increase decision wait time
tail_sampling:
  decision_wait: 30s  # Increase from 10s for slow traces
```

---

## Performance Tuning

### Low Latency (Real-time dashboards)

Use `01-basic.yaml` with:
```yaml
processors:
  batch:
    timeout: 500ms        # Flush faster
    send_batch_size: 50   # Smaller batches
```

### High Throughput (Analytics workloads)

Use `03-tail-sampling.yaml` with:
```yaml
processors:
  batch:
    timeout: 5s           # Less frequent flushes
    send_batch_size: 5000 # Larger batches
```

### Balanced (General purpose)

Use `01-basic.yaml` as-is (default settings)

---

## Next Steps

1. **Choose configuration** based on decision matrix above
2. **Set environment variables** (API keys, endpoints)
3. **Start collector** with chosen config
4. **Configure application** to send OTLP to collector
5. **Verify in Brokle** dashboard that traces appear
6. **Monitor performance** and adjust as needed

---

## Additional Resources

- [Test Infrastructure](../../test/otel-collector/README.md) - Integration tests
- [Brokle Documentation](../../docs/integrations/opentelemetry-collector.md) - Full integration guide
- [OTEL Collector Docs](https://opentelemetry.io/docs/collector/) - Official documentation
- [OTLP Specification](https://opentelemetry.io/docs/specs/otlp/) - Protocol details

---

## Support

For issues or questions:
1. Check troubleshooting section above
2. Review collector logs for errors
3. Test with `01-basic.yaml` first (isolate issues)
4. Open GitHub issue with full config and logs
