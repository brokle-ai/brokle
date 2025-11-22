# OpenTelemetry Collector Integration

Brokle is fully compatible with the OpenTelemetry Collector. This guide shows you how to integrate Brokle with your existing OTEL infrastructure.

---

## When to Use OTEL Collector

### ✅ **Use Collector When:**

- **You already run a collector** for other backends (Datadog, Jaeger, etc.)
- **High traffic** (>100K spans/day) requiring tail-based sampling
- **Compliance requires** PII scrubbing before data leaves your network
- **Multi-backend** strategy (send to Brokle + other platforms simultaneously)
- **Advanced processing** needed (filtering, enrichment, transformation)

### ❌ **Don't Use Collector When:**

- **Simple deployment** - Use [Brokle SDK](./direct-sdk.md) instead (lower latency)
- **Low volume** (<100K spans/day) - Direct integration is simpler
- **Lowest latency required** - Collector adds 50-100ms overhead
- **First time with OpenTelemetry** - Start with SDK, add collector later if needed

---

## Quick Start (5 Minutes)

### 1. Get Your Brokle API Key

```bash
# Sign up at https://app.brokle.com
# Navigate to Settings → API Keys
# Copy your API key (starts with "bk_")

export BROKLE_API_KEY="bk_your_api_key_here"
```

### 2. Download OTEL Collector

```bash
# Linux
wget https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v0.91.0/otelcol-contrib_0.91.0_linux_amd64.tar.gz
tar -xzf otelcol-contrib_0.91.0_linux_amd64.tar.gz

# macOS (Intel)
wget https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v0.91.0/otelcol-contrib_0.91.0_darwin_amd64.tar.gz
tar -xzf otelcol-contrib_0.91.0_darwin_amd64.tar.gz

# macOS (M1/M2)
wget https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v0.91.0/otelcol-contrib_0.91.0_darwin_arm64.tar.gz
tar -xzf otelcol-contrib_0.91.0_darwin_arm64.tar.gz
```

### 3. Create Configuration File

```yaml
# config.yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    timeout: 1s
    send_batch_size: 100

exporters:
  otlphttp/brokle:
    endpoint: https://api.brokle.com/v1/traces
    headers:
      X-API-Key: ${BROKLE_API_KEY}
    compression: gzip

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlphttp/brokle]
```

### 4. Start Collector

```bash
./otelcol-contrib --config=config.yaml
```

### 5. Configure Your Application

```bash
# Point your app's OTEL exporter to the collector
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318

# Run your application
python your_app.py  # Or node, go, java, etc.
```

### 6. Verify in Brokle Dashboard

Visit https://app.brokle.com → Traces

You should see traces appearing within seconds.

---

## Architecture

### Integration Pattern

```
┌─────────────────┐
│  Your App       │
│  (OTEL SDK)     │  Sends OTLP traces
└────────┬────────┘
         │ OTLP/HTTP or gRPC
         ↓
┌─────────────────┐
│ OTEL Collector  │  Batches, samples, processes
│  (Your infra)   │  Port 4318 (HTTP) or 4317 (gRPC)
└────────┬────────┘
         │ OTLP/HTTP + gzip
         ↓
┌─────────────────┐
│  Brokle API     │  Receives OTLP natively
│  api.brokle.com │  Port 443 (HTTPS)
└────────┬────────┘
         │
         ├──→ Redis Streams (async queue)
         │
         └──→ ClickHouse (analytics storage)
```

**Key points:**
- ✅ Collector runs in YOUR infrastructure (you control it)
- ✅ Brokle receives OTLP natively (no special setup)
- ✅ Data processed asynchronously for reliability

---

## Configuration Examples

We provide 4 production-ready configurations:

### [01-basic.yaml](../../examples/otel-collector/01-basic.yaml) - Simple Setup

**Copy-paste ready** configuration for getting started.

```yaml
# Basic setup
exporters:
  otlphttp/brokle:
    endpoint: https://api.brokle.com/v1/traces
    headers:
      X-API-Key: ${BROKLE_API_KEY}
```

**Use when**: First time, POC, simple deployments

---

### [02-multi-backend.yaml](../../examples/otel-collector/02-multi-backend.yaml) - Fan-Out

Send traces to **Brokle + Datadog + Jaeger** simultaneously.

```yaml
exporters:
  otlphttp/brokle:
    endpoint: https://api.brokle.com/v1/traces
  otlp/datadog:
    endpoint: https://api.datadoghq.com
  otlp/jaeger:
    endpoint: jaeger:4317

service:
  pipelines:
    traces:
      exporters: [otlphttp/brokle, otlp/datadog, otlp/jaeger]
```

**Use when**: Evaluating Brokle, gradual migration, multi-backend compliance

**Migration path**:
1. Week 1: Add Brokle exporter alongside current tools
2. Week 2-4: Build dashboards in Brokle, train team
3. Week 5+: Remove old exporters, keep Brokle

---

### [03-tail-sampling.yaml](../../examples/otel-collector/03-tail-sampling.yaml) - Cost Optimization

**Reduce costs by 95%** while keeping all errors and slow traces.

```yaml
processors:
  tail_sampling:
    policies:
      - name: errors
        type: status_code
        status_code: {status_codes: [ERROR]}  # Keep ALL errors

      - name: slow
        type: latency
        latency: {threshold_ms: 1000}  # Keep ALL slow traces

      - name: sample-rest
        type: probabilistic
        probabilistic: {sampling_percentage: 1}  # Sample 1% of rest
```

**Results**:
- Input: 1M spans/day
- Output: 40K spans/day (96% reduction)
- **Cost**: $100/day → $4/day

**Use when**: High volume, cost concerns

---

### [04-pii-scrubbing.yaml](../../examples/otel-collector/04-pii-scrubbing.yaml) - Compliance

Remove sensitive data **before** sending to Brokle.

```yaml
processors:
  attributes:
    actions:
      - key: user.email
        action: delete
      - key: client.ip
        action: hash
      - key: payment.card
        action: delete
```

**Removes**:
- ✅ Email addresses
- ✅ IP addresses (hashed for analytics)
- ✅ Credit card numbers
- ✅ Phone numbers
- ✅ SSNs, auth tokens

**Compliance**: GDPR Article 32, HIPAA Privacy Rule, PCI DSS 3.4

**Use when**: GDPR/HIPAA/PCI requirements

---

## Supported Protocols

| Protocol | Port | Status | Notes |
|----------|------|--------|-------|
| **OTLP/gRPC** | 4317 | ✅ | Binary protobuf, most efficient |
| **OTLP/HTTP** | 4318 | ✅ | HTTP with protobuf or JSON |
| **Gzip compression** | - | ✅ | Recommended (80% bandwidth reduction) |

**Brokle endpoint:**
- Production: `https://api.brokle.com/v1/traces` (OpenTelemetry standard)

Supports Protobuf and JSON formats with gzip compression.

---

## Authentication

Brokle uses **API key authentication** via HTTP header.

### In Collector Config:

```yaml
exporters:
  otlphttp/brokle:
    endpoint: https://api.brokle.com/v1/traces
    headers:
      X-API-Key: ${BROKLE_API_KEY}  # Set via environment variable
```

### Get Your API Key:

1. Log in to https://app.brokle.com
2. Navigate to **Settings → API Keys**
3. Click **Create API Key**
4. Copy the key (starts with `bk_`)
5. Set environment variable:
   ```bash
   export BROKLE_API_KEY="bk_your_key_here"
   ```

---

## Deployment Patterns

### Docker Deployment

```bash
docker run -d \
  --name otel-collector \
  -p 4317:4317 \
  -p 4318:4318 \
  -e BROKLE_API_KEY="bk_your_key" \
  -v $(pwd)/config.yaml:/etc/otel-collector-config.yaml \
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
      - ./config.yaml:/etc/otel-collector-config.yaml
    ports:
      - "4317:4317"
      - "4318:4318"
    environment:
      - BROKLE_API_KEY=${BROKLE_API_KEY}
```

### Kubernetes DaemonSet

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: otel-collector
  namespace: observability
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
        - containerPort: 4317
          name: otlp-grpc
        - containerPort: 4318
          name: otlp-http
        env:
        - name: BROKLE_API_KEY
          valueFrom:
            secretKeyRef:
              name: brokle-api-key
              key: api-key
        volumeMounts:
        - name: config
          mountPath: /conf
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
      volumes:
      - name: config
        configMap:
          name: otel-collector-config
```

**Create secret:**
```bash
kubectl create secret generic brokle-api-key \
  --from-literal=api-key="bk_your_key_here" \
  -n observability
```

**Create ConfigMap:**
```bash
kubectl create configmap otel-collector-config \
  --from-file=otel-collector-config.yaml=01-basic.yaml \
  -n observability
```

---

## Performance & Scaling

### Expected Performance

| Deployment | Throughput | Latency (p99) | Memory |
|------------|-----------|---------------|--------|
| Single collector | ~10K spans/sec | <100ms | 50-100MB |
| Multiple collectors | ~100K spans/sec | <100ms | 50-100MB each |
| With tail sampling | 1M+ input | 1-5s | 500MB-1GB |

### Scaling Strategies

#### Horizontal Scaling (Multiple Collectors)

```
App Instances → Load Balancer → Collector Pool
                                    ↓
                              [Brokle API]
```

Deploy multiple collector instances and load balance across them.

#### Vertical Scaling (Larger Instance)

Increase collector resources:
- CPU: 2-4 cores for high throughput
- Memory: 1-2GB for tail sampling
- Network: 1Gbps+ for high volume

---

## Troubleshooting

### No traces appearing in Brokle

**Check 1: Collector health**
```bash
curl http://localhost:13133
# Should return: {"status":"Server available"}
```

**Check 2: Collector logs**
```bash
# Look for export errors
docker logs otel-collector | grep -i "export\|error"

# Common errors:
# - "401 Unauthorized" → Invalid API key
# - "Connection refused" → Network issue
# - "Timeout" → Brokle API unreachable
```

**Check 3: Verify API key**
```bash
echo $BROKLE_API_KEY
# Should start with "bk_"

# Test API key manually
curl -X POST https://api.brokle.com/v1/traces \
  -H "X-API-Key: $BROKLE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{}'
# Should return 400 (bad request), not 401 (proves auth works)
```

**Check 4: Application sending to collector**
```bash
# Verify app OTLP endpoint
echo $OTEL_EXPORTER_OTLP_ENDPOINT
# Should be: http://localhost:4318 (or collector hostname)

# Test with curl
curl -X POST http://localhost:4318/v1/traces \
  -H "Content-Type: application/json" \
  -d '{}'
# Should return 200 (collector accepts request)
```

---

### High latency

**Symptom**: Traces appear in Brokle but with 5-10 second delay

**Common causes:**

1. **Tail sampling enabled** → Expected (decision_wait time)
   - Solution: Use head-based sampling instead

2. **Batch timeout too long** → Buffering in collector
   ```yaml
   processors:
     batch:
       timeout: 500ms  # Reduce from 1s
   ```

3. **Network latency** → Slow connection to Brokle
   - Solution: Check network, consider regional deployment

---

### Collector memory usage high

**Symptom**: Collector using >1GB memory

**Common causes:**

1. **Tail sampling buffer too large**
   ```yaml
   tail_sampling:
     num_traces: 50000  # Reduce from 100,000
   ```

2. **Too many backends** → Buffering for all exporters
   - Solution: Remove unused exporters

3. **Large batch sizes** → More data in memory
   ```yaml
   batch:
     send_batch_max_size: 1000  # Reduce from 5000
   ```

---

### Traces incomplete or missing spans

**Symptom**: Parent spans without children, or vice versa

**Common causes:**

1. **decision_wait too short** (tail sampling)
   ```yaml
   tail_sampling:
     decision_wait: 30s  # Increase from 10s for slow traces
   ```

2. **Application exporting spans out of order**
   - Solution: Ensure app flushes all spans before shutdown

3. **Collector restarts** → Buffered data lost
   - Solution: Configure graceful shutdown

---

## Advanced Configuration

### Custom Sampling Policies

Keep all traces for specific users:

```yaml
processors:
  tail_sampling:
    policies:
      - name: vip-users
        type: string_attribute
        string_attribute:
          key: user.tier
          values: [vip, enterprise, premium]

      - name: sample-rest
        type: probabilistic
        probabilistic: {sampling_percentage: 1}
```

### Multi-Region Deployment

Send to closest Brokle region:

```yaml
exporters:
  otlphttp/brokle-us:
    endpoint: https://us.api.brokle.com/v1/traces
    headers:
      X-API-Key: ${BROKLE_API_KEY}

  otlphttp/brokle-eu:
    endpoint: https://eu.api.brokle.com/v1/traces
    headers:
      X-API-Key: ${BROKLE_API_KEY}

service:
  pipelines:
    traces-us:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlphttp/brokle-us]  # US traffic

    traces-eu:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlphttp/brokle-eu]  # EU traffic
```

*Note: Multi-region endpoints are enterprise-only. Contact sales.*

---

## Best Practices

### 1. Always Enable Compression

```yaml
exporters:
  otlphttp/brokle:
    compression: gzip  # Reduces bandwidth by ~80%
```

### 2. Configure Retries

```yaml
exporters:
  otlphttp/brokle:
    retry_on_failure:
      enabled: true
      initial_interval: 1s
      max_interval: 30s
      max_elapsed_time: 300s  # Retry for 5 minutes
```

### 3. Enable Health Checks

```yaml
extensions:
  health_check:
    endpoint: 0.0.0.0:13133

service:
  extensions: [health_check]
```

**Then monitor:**
```bash
curl http://localhost:13133
```

### 4. Set Resource Attributes

```yaml
processors:
  resource:
    attributes:
      - key: deployment.environment
        value: ${DEPLOYMENT_ENV}
        action: upsert
      - key: service.namespace
        value: ${SERVICE_NAMESPACE}
        action: upsert
```

**Benefits**: Better filtering and grouping in Brokle dashboard

---

## Security

### API Key Management

**✅ Do:**
- Store API keys in secrets manager (AWS Secrets Manager, HashiCorp Vault)
- Use environment variables (not hardcoded in config)
- Rotate keys regularly (every 90 days)
- Use different keys per environment (dev, staging, prod)

**❌ Don't:**
- Commit API keys to git
- Share keys across teams
- Use same key for all environments
- Log API keys in collector output

### Example with Kubernetes Secrets:

```yaml
# Create secret
apiVersion: v1
kind: Secret
metadata:
  name: brokle-api-key
type: Opaque
stringData:
  api-key: "bk_your_actual_key_here"
```

```yaml
# Reference in collector
env:
- name: BROKLE_API_KEY
  valueFrom:
    secretKeyRef:
      name: brokle-api-key
      key: api-key
```

---

## Monitoring Collector

### Health Check Endpoint

```bash
curl http://localhost:13133

# Response:
{
  "status": "Server available",
  "upSince": "2024-10-28T08:00:00Z",
  "uptime": "3600s"
}
```

### Metrics (if Prometheus enabled)

```yaml
# Add to config
service:
  telemetry:
    metrics:
      address: 0.0.0.0:8888  # Prometheus metrics endpoint
```

**Key metrics to monitor:**
- `otelcol_exporter_sent_spans` - Spans exported to Brokle
- `otelcol_exporter_send_failed_spans` - Failed exports
- `otelcol_processor_batch_batch_send_size` - Batch sizes
- `otelcol_receiver_accepted_spans` - Spans received from app

---

## Cost Optimization

### 1. Use Tail Sampling (Highest Impact)

**Before**: 1M spans/day × $0.10/1K = $100/day
**After**: 40K spans/day × $0.10/1K = $4/day
**Savings**: **$2,880/month** (96% reduction)

See [03-tail-sampling.yaml](../../examples/otel-collector/03-tail-sampling.yaml)

### 2. Enable Compression

```yaml
exporters:
  otlphttp/brokle:
    compression: gzip  # 80% bandwidth reduction
```

**Savings**: Network transfer costs reduced

### 3. Adjust Batch Sizes

```yaml
processors:
  batch:
    send_batch_size: 1000  # Larger batches = fewer requests
```

**Savings**: Fewer HTTP requests, lower processing overhead

---

## Migration from Other Platforms

### From Datadog

```yaml
# Step 1: Add Brokle alongside Datadog
exporters:
  otlp/datadog:
    endpoint: https://api.datadoghq.com
  otlphttp/brokle:
    endpoint: https://api.brokle.com/v1/traces

service:
  pipelines:
    traces:
      exporters: [otlp/datadog, otlphttp/brokle]  # Both backends
```

**Timeline**: Run both for 2-4 weeks, then remove Datadog exporter

### From Jaeger

```yaml
# Step 1: Add Brokle alongside Jaeger
exporters:
  otlp/jaeger:
    endpoint: jaeger:4317
  otlphttp/brokle:
    endpoint: https://api.brokle.com/v1/traces

service:
  pipelines:
    traces:
      exporters: [otlp/jaeger, otlphttp/brokle]
```

**Timeline**: Validate Brokle, then remove Jaeger exporter

### From New Relic

```yaml
# Step 1: Add Brokle alongside New Relic
exporters:
  otlphttp/newrelic:
    endpoint: https://otlp.nr-data.net:443
    headers:
      api-key: ${NEW_RELIC_LICENSE_KEY}
  otlphttp/brokle:
    endpoint: https://api.brokle.com/v1/traces
    headers:
      X-API-Key: ${BROKLE_API_KEY}

service:
  pipelines:
    traces:
      exporters: [otlphttp/newrelic, otlphttp/brokle]
```

---

## FAQ

### Q: Do I need the contrib or core collector?

**A:** Use **contrib** (recommended). It includes all processors (tail_sampling, attributes, transform) needed for advanced configs.

```bash
# ✅ Recommended
otel/opentelemetry-collector-contrib:0.91.0

# ❌ Not recommended (missing processors)
otel/opentelemetry-collector:0.91.0
```

---

### Q: What collector versions are supported?

**A:** Tested with:
- ✅ v0.91.0 (recommended)
- ✅ v0.90.0
- ✅ v0.89.0

Versions <0.85.0 may have OTLP spec compatibility issues.

---

### Q: Can I use Brokle without a collector?

**A:** Yes! Three integration options:

1. **Brokle SDK** (recommended) - Lowest latency
2. **OTEL Collector** (this guide) - Advanced processing
3. **Direct OTLP** - Vendor-agnostic

See [Integration Comparison](./README.md) for details.

---

### Q: Does collector add latency?

**A:** Yes, typically 50-100ms for batching and network hop.

- Without collector: App → Brokle (5-10ms)
- With collector: App → Collector → Brokle (50-100ms)

**Trade-off**: Accept higher latency for advanced features (sampling, PII scrubbing, multi-backend).

---

### Q: Can I send to Brokle AND other platforms?

**A:** Yes! See [02-multi-backend.yaml](../../examples/otel-collector/02-multi-backend.yaml)

```yaml
service:
  pipelines:
    traces:
      exporters: [otlphttp/brokle, otlp/datadog, otlp/jaeger]
```

Traces fan out to all backends in parallel.

---

### Q: How do I reduce costs for high-volume apps?

**A:** Use tail-based sampling. See [03-tail-sampling.yaml](../../examples/otel-collector/03-tail-sampling.yaml)

Keeps all errors but samples 1% of success → **95% cost reduction**

---

### Q: Is PII automatically removed?

**A:** No. You must configure PII scrubbing in the collector.

See [04-pii-scrubbing.yaml](../../examples/otel-collector/04-pii-scrubbing.yaml) for GDPR/HIPAA compliant setup.

---

## Next Steps

1. **Choose a configuration** from [examples](../../examples/otel-collector/)
2. **Set your API key**: `export BROKLE_API_KEY="bk_..."`
3. **Start the collector**: `otelcol --config=01-basic.yaml`
4. **Configure your app**: `export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318`
5. **Verify in Brokle**: Check https://app.brokle.com for traces

---

## Additional Resources

- [Example Configurations](../../examples/otel-collector/) - 4 production-ready configs
- [Integration Tests](../../test/otel-collector/) - Test infrastructure
- [Brokle Architecture](../ARCHITECTURE.md) - System architecture
- [API Documentation](../API.md) - OTLP endpoint details
- [OTEL Collector Docs](https://opentelemetry.io/docs/collector/) - Official documentation

---

## Support

Need help?
- [Troubleshooting Guide](#troubleshooting) (this page)
- [Example Configs](../../examples/otel-collector/README.md)
- [GitHub Issues](https://github.com/brokle/brokle/issues)
- Email: support@brokle.com
- Slack: [Community Slack](https://brokle.com/slack)
