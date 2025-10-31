#!/bin/bash
set -e

echo "========================================"
echo "OTEL Collector -> Brokle Integration Test"
echo "========================================"
echo ""

# Wait for services to stabilize
echo "Waiting for services to start..."
sleep 10

# Check collector health
echo "[1/4] Checking OTEL Collector health..."
if curl -f http://localhost:13133 > /dev/null 2>&1; then
    echo "  OK: Collector is healthy"
else
    echo "  FAIL: Collector health check failed"
    echo "  Try: docker compose -f docker-compose.test.yml logs otel-collector"
    exit 1
fi

# Check Brokle API health
echo "[2/4] Checking Brokle API health..."
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    echo "  OK: Brokle API is healthy"
else
    echo "  FAIL: Brokle API health check failed"
    echo "  Try: docker compose -f docker-compose.test.yml logs brokle-backend"
    exit 1
fi

# Check ClickHouse connectivity
echo "[3/4] Checking ClickHouse connectivity..."
if docker compose -f docker-compose.test.yml exec -T clickhouse \
    clickhouse-client --user brokle --password test --query "SELECT 1" > /dev/null 2>&1; then
    echo "  OK: ClickHouse is reachable"
else
    echo "  FAIL: ClickHouse connectivity failed"
    echo "  Try: docker compose -f docker-compose.test.yml logs clickhouse"
    exit 1
fi

# Wait for traces to arrive
echo "[4/4] Waiting for traces to arrive in ClickHouse..."
sleep 5

# Count traces
COUNT=$(docker compose -f docker-compose.test.yml exec -T clickhouse \
    clickhouse-client --user brokle --password test --query \
    "SELECT COUNT(*) FROM traces WHERE start_time > now() - INTERVAL 1 MINUTE" 2>/dev/null || echo "0")

echo "  Traces found: $COUNT"

if [ "$COUNT" -gt 0 ]; then
    echo ""
    echo "========================================"
    echo "SUCCESS: Integration verified!"
    echo "========================================"
    echo ""
    echo "Metrics:"
    echo "  - Traces received: $COUNT"
    echo "  - Collector endpoint: http://localhost:4318"
    echo "  - Brokle API: http://localhost:8080"
    echo ""
    echo "Next steps:"
    echo "  - View traces: docker compose -f docker-compose.test.yml exec clickhouse clickhouse-client --user brokle --password test --query \"SELECT * FROM traces LIMIT 5\""
    echo "  - Check logs: docker compose -f docker-compose.test.yml logs"
    echo ""
    exit 0
else
    echo ""
    echo "========================================"
    echo "FAIL: No traces found in ClickHouse"
    echo "========================================"
    echo ""
    echo "Troubleshooting steps:"
    echo ""
    echo "1. Check collector logs:"
    echo "   docker compose -f docker-compose.test.yml logs otel-collector"
    echo "   Look for: 'Exporting failed' or '401/403 errors'"
    echo ""
    echo "2. Check API logs:"
    echo "   docker compose -f docker-compose.test.yml logs brokle-backend"
    echo "   Look for: 'OTLP' or 'trace' related messages"
    echo ""
    echo "3. Check Redis streams:"
    echo "   docker compose -f docker-compose.test.yml exec redis redis-cli KEYS 'telemetry:*'"
    echo ""
    echo "4. Check ClickHouse tables:"
    echo "   docker compose -f docker-compose.test.yml exec clickhouse clickhouse-client --user brokle --password test --query \"SHOW TABLES\""
    echo ""
    exit 1
fi
