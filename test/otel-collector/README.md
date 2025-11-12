# OpenTelemetry Collector Integration Tests

This directory contains integration tests to validate that Brokle works correctly with OpenTelemetry Collectors.

## Setup

Before running tests, you need to set up the database schema and seed test data.

### 1. Start the test stack

```bash
cd test/otel-collector
docker compose -f docker-compose.test.yml up -d
```

Wait for services to be healthy (~30-60 seconds).

### 2. Run database migrations

```bash
# Run migrations inside the container
docker compose -f docker-compose.test.yml exec brokle-api /migrate -db all up
```

### 3. Seed test data

**From project root**, run the seed command against the test database:

```bash
# Navigate to project root
cd ../..

# Set environment variables for test database
export DATABASE_URL=postgres://brokle:test@localhost:5433/brokle_test?sslmode=disable
export CLICKHOUSE_URL=clickhouse://brokle:test@localhost:9001/default
export REDIS_URL=redis://localhost:6379/0

# Run seed command (creates org, user, project, API key)
make seed-dev
```

**Note**: The docker-compose exposes database ports (5433, 9001, 8124) specifically for seeding from the host.

---

## Quick Start

After setup is complete:

```bash
cd test/otel-collector

# Run verification
./verify-ingestion.sh
```

Expected output:
```
========================================
SUCCESS: Integration verified!
========================================

Metrics:
  - Traces received: 100
  - Collector endpoint: http://localhost:4318
  - Brokle API: http://localhost:8080
```

## Architecture

```
┌─────────────┐
│ Trace       │
│ Generator   │  Generates OTLP spans
└──────┬──────┘
       │ OTLP/HTTP
       ↓
┌──────────────┐
│ OTEL         │  Receives, batches, exports
│ Collector    │  Port 4318 (HTTP), 4317 (gRPC)
└──────┬───────┘
       │ OTLP/HTTP + gzip
       ↓
┌──────────────┐
│ Brokle API   │  Processes OTLP traces
└──────┬───────┘  Port 8080
       │
       ├──→ Redis Streams (async queue)
       │
       └──→ ClickHouse (analytics storage)
```

## Test Scenarios

### 1. Success Scenario (default)

Sends 100 valid traces through the collector to Brokle.

```bash
docker compose -f docker-compose.test.yml up trace-generator
```

**Expected result**: 100 traces in ClickHouse

---

### 2. Invalid Authentication

Tests how the system handles invalid API keys.

```bash
TEST_MODE=invalid_auth docker compose -f docker-compose.test.yml up trace-generator
```

**Expected result**: 401 Unauthorized in collector logs

---

### 3. Malformed Data

Documents that the OTEL SDK validates data client-side.

```bash
TEST_MODE=malformed docker compose -f docker-compose.test.yml up trace-generator
```

**Expected result**: Educational output (no server request sent)

---

### 4. Network Timeout

Tests timeout handling with an unrealistic 1ms timeout.

```bash
TEST_MODE=timeout docker compose -f docker-compose.test.yml up trace-generator
```

**Expected result**: Timeout errors in trace generator logs

---

### 5. Large Batch Stress Test

Generates 10,000 spans to test high-volume scenarios.

```bash
TEST_MODE=large_batch docker compose -f docker-compose.test.yml up trace-generator
```

**Expected result**: 10,000 traces successfully processed

---

## Tested Collector Versions

✅ **Tested and working:**
- `otel/opentelemetry-collector-contrib:0.91.0` (recommended)
- `otel/opentelemetry-collector-contrib:0.90.0`
- `otel/opentelemetry-collector-contrib:0.89.0`

⚠️  **Not tested:**
- `otel/opentelemetry-collector` (core, not contrib) - use contrib version
- Versions <0.85.0 (OTLP spec changes may cause issues)

To test with a different version:
```yaml
# Edit docker-compose.test.yml
otel-collector:
  image: otel/opentelemetry-collector-contrib:0.90.0  # Change version
```

---

## OTLP Protocol Support

| Protocol | Status | Port | Notes |
|----------|--------|------|-------|
| OTLP/gRPC (protobuf) | ✅ | 4317 | Binary format, most efficient |
| OTLP/HTTP (protobuf) | ✅ | 4318 | HTTP with binary protobuf |
| OTLP/HTTP (JSON) | ✅ | 4318 | HTTP with JSON (less efficient) |
| Gzip compression | ✅ | All | Reduces bandwidth by ~80% |
| Zstandard compression | ❌ | - | Not yet supported |

---

## Health Checks

All services expose health check endpoints:

### OTEL Collector
```bash
curl http://localhost:13133
# {"status":"Server available","upSince":"2024-10-28T01:00:00Z"}
```

### Brokle API
```bash
curl http://localhost:8080/health
# {"status":"healthy","timestamp":"2024-10-28T01:00:00Z"}
```

### ClickHouse
```bash
docker compose -f docker-compose.test.yml exec clickhouse \
  clickhouse-client --user brokle --password test --query "SELECT 1"
# 1
```

### Redis
```bash
docker compose -f docker-compose.test.yml exec redis redis-cli ping
# PONG
```

---

## Troubleshooting

### No traces in ClickHouse

**Symptom**: `verify-ingestion.sh` reports 0 traces

**Diagnosis steps:**

1. **Check collector logs:**
   ```bash
   docker compose -f docker-compose.test.yml logs otel-collector
   ```
   Look for:
   - `Exporting failed` - Network or auth issues
   - `401 Unauthorized` - Invalid API key
   - `Connection refused` - Brokle API not ready

2. **Check Brokle API logs:**
   ```bash
   docker compose -f docker-compose.test.yml logs brokle-api
   ```
   Look for:
   - `OTLP` or `trace` messages
   - Authentication errors
   - Database connection issues

3. **Check Redis streams:**
   ```bash
   docker compose -f docker-compose.test.yml exec redis redis-cli KEYS 'telemetry:*'
   ```
   Should show: `telemetry:batches:<project_id>`

4. **Check ClickHouse tables:**
   ```bash
   docker compose -f docker-compose.test.yml exec clickhouse \
     clickhouse-client --user brokle --password test --query "SHOW TABLES"
   ```
   Should include: `traces`, `spans`, `scores`

---

### Collector health check fails

**Symptom**: Collector container restarts or shows unhealthy

**Solutions:**

1. **Check if collector started:**
   ```bash
   docker compose -f docker-compose.test.yml logs otel-collector | head -50
   ```
   Look for: `Everything is ready` message

2. **Verify config syntax:**
   ```bash
   docker compose -f docker-compose.test.yml exec otel-collector \
     cat /etc/otel-collector-config.yaml
   ```

3. **Test health endpoint manually:**
   ```bash
   docker compose -f docker-compose.test.yml exec otel-collector \
     wget -O- http://localhost:13133
   ```

---

### Brokle API health check fails

**Symptom**: Brokle API never becomes healthy

**Solutions:**

1. **Check if API started:**
   ```bash
   docker compose -f docker-compose.test.yml logs brokle-api | grep -i "listening\|starting"
   ```

2. **Verify database connections:**
   ```bash
   docker compose -f docker-compose.test.yml logs brokle-api | grep -i "postgres\|clickhouse\|redis"
   ```

3. **Check dependencies:**
   ```bash
   docker compose -f docker-compose.test.yml ps
   ```
   All services should show `healthy` or `running`

4. **Test health endpoint:**
   ```bash
   docker compose -f docker-compose.test.yml exec brokle-api \
     wget -O- http://localhost:8080/health
   ```

---

### Large batch test hangs

**Symptom**: `large_batch` test generates spans but never completes

**Solutions:**

1. **Check collector batching:**
   - Collector may be buffering spans
   - Wait 30-60 seconds for flush

2. **Check ClickHouse write performance:**
   ```bash
   docker compose -f docker-compose.test.yml exec clickhouse \
     clickhouse-client --user brokle --password test --query \
     "SELECT COUNT(*) FROM system.processes WHERE query LIKE '%INSERT%'"
   ```

3. **Monitor Redis stream lag:**
   ```bash
   docker compose -f docker-compose.test.yml exec redis \
     redis-cli XLEN telemetry:batches:<project_id>
   ```

---

## Performance Benchmarks

Expected performance (single collector instance on development machine):

| Metric | Value | Notes |
|--------|-------|-------|
| **Max sustained throughput** | ~10,000 spans/sec | Limited by ClickHouse write speed |
| **Avg latency (p50)** | <100ms | Collector → ClickHouse |
| **Avg latency (p99)** | <500ms | Including retries |
| **Batch size** | 100 spans | Configurable in collector config |
| **Flush interval** | 1 second | Configurable in collector config |
| **Memory usage (collector)** | ~50MB | Baseline with no load |
| **Memory usage (Brokle)** | ~100MB | With Redis Streams worker |

---

## Configuration

### Collector Configuration

Key settings in `collector-config.yaml`:

```yaml
processors:
  batch:
    timeout: 1s              # Flush every 1 second
    send_batch_size: 100     # Or when 100 spans buffered
    send_batch_max_size: 1000  # Max batch size

exporters:
  otlphttp/brokle:
    timeout: 30s             # HTTP timeout
    retry_on_failure:
      max_elapsed_time: 300s # Max retry time (5 min)
```

### Brokle Configuration

Set via environment variables in `docker-compose.test.yml`:

```yaml
environment:
  - LOG_LEVEL=debug          # Verbose logging
  - ENV=test                 # Test environment
```

---

## Clean Up

```bash
# Stop all services
docker compose -f docker-compose.test.yml down

# Remove volumes (clean state)
docker compose -f docker-compose.test.yml down -v

# Remove images
docker compose -f docker-compose.test.yml down --rmi all
```

---

## Next Steps

After validating integration:

1. **Copy collector config** to `examples/otel-collector/01-basic.yaml`
2. **Add multi-backend config** for Datadog/Jaeger fanout
3. **Add tail sampling config** for high-volume scenarios
4. **Document in main docs** at `docs/integrations/opentelemetry-collector.md`

---

## Related Documentation

- [Brokle Architecture](../../docs/ARCHITECTURE.md)
- [API Documentation](../../docs/API.md)
- [OpenTelemetry Specification](https://opentelemetry.io/docs/specs/otlp/)
- [OTEL Collector Documentation](https://opentelemetry.io/docs/collector/)

---

## Support

For issues or questions:
1. Check logs using commands above
2. Review OTEL Collector documentation
3. Check Brokle API logs for errors
4. Open an issue with full logs and configuration
