# Migration Guide: Generic Events Deprecation

## Overview

As of **October 2025**, the Brokle platform has deprecated the generic `event_type: "event"` in favor of OTEL-native structured observability using `Span.Type = "event"`.

---

## What Changed

### ❌ Deprecated (Removed)
- `TelemetryEventType.EVENT` constant
- Generic event processing path
- `TelemetryEvent` domain entity
- Event queue infrastructure
- `/v1/ingest/batch` endpoint no longer accepts `event_type: "event"`

### ✅ New Pattern
- Use `event_type: "span"` with `payload.type: "event"`
- Events are now first-class spans in the OTEL model
- Unified storage in `spans` table (ClickHouse)
- Full trace hierarchy support

---

## Migration Path

### Backend API

#### Before (Deprecated)
```bash
curl -X POST http://localhost:8080/v1/ingest/batch \
  -H "X-API-Key: bk_your_secret" \
  -H "Content-Type: application/json" \
  -d '{
    "events": [{
      "event_id": "01HQXYZ...",
      "event_type": "event",
      "payload": {
        "name": "user_click",
        "button": "checkout",
        "page": "/cart"
      }
    }]
  }'
```

**Result**: ❌ `400 Bad Request` - "event" is not a valid event type

#### After (OTEL-Native)
```bash
curl -X POST http://localhost:8080/v1/ingest/batch \
  -H "X-API-Key: bk_your_secret" \
  -H "Content-Type: application/json" \
  -d '{
    "events": [{
      "event_id": "01HQXYZ...",
      "event_type": "span",
      "payload": {
        "type": "event",
        "name": "user_click",
        "input": {
          "button": "checkout",
          "page": "/cart"
        },
        "metadata": {
          "event.domain": "user_interaction",
          "event.category": "engagement"
        }
      }
    }]
  }'
```

**Result**: ✅ `200 OK` - Event stored as span with `type = "event"`

---

### Python SDK (Legacy)

#### Before (Deprecated)
```python
from brokle_legacy.types.telemetry import TelemetryEventType

client.submit_batch([
    {
        "event_type": TelemetryEventType.EVENT,  # ❌ Removed from enum
        "payload": {
            "action": "user_login",
            "user_id": "123"
        }
    }
])
```

**Result**: ❌ `AttributeError: TelemetryEventType has no attribute 'EVENT'`

#### After (OTEL-Native)
```python
from brokle_legacy.types.telemetry import TelemetryEventType

client.submit_batch([
    {
        "event_type": TelemetryEventType.SPAN,  # ✅ Use SPAN
        "payload": {
            "type": "event",  # ✅ Set span type to "event"
            "name": "user_login",
            "input": {"user_id": "123"},
            "metadata": {
                "event.domain": "user_lifecycle",
                "event.action": "login"
            }
        }
    }
])
```

**Result**: ✅ Event stored as span with `type = "event"`

---

### Python SDK (New - Recommended)

The new Python SDK already uses the correct pattern:

```python
from brokle import Brokle

brokle = Brokle(api_key="bk_your_secret")

# Events as spans (zero-duration)
with brokle.start_as_current_span(
    name="user_click",
    as_type="event",  # ✅ Uses SpanType.EVENT internally
    input={"button": "checkout", "page": "/cart"},
    metadata={"event.domain": "user_interaction"}
):
    pass  # Event recorded immediately
```

---

## Event Attribute Conventions

To help categorize and query events, use these attribute conventions:

```json
{
  "metadata": {
    "event.domain": "user_interaction|user_lifecycle|system|business",
    "event.category": "engagement|conversion|error|info",
    "event.action": "click|submit|view|login|logout",
    "event.source": "frontend|backend|mobile|api",
    "event.severity": "info|warning|error"
  }
}
```

### Example: User Interaction Event
```python
with brokle.start_as_current_span(
    name="button_click",
    as_type="event",
    input={"button_id": "checkout_btn", "button_text": "Checkout"},
    metadata={
        "event.domain": "user_interaction",
        "event.category": "engagement",
        "event.action": "click",
        "event.source": "frontend",
        "page.url": "/cart"
    }
):
    pass
```

### Example: Business Event
```python
with brokle.start_as_current_span(
    name="user_signup",
    as_type="event",
    input={"email": "user@example.com", "method": "google_oauth"},
    metadata={
        "event.domain": "user_lifecycle",
        "event.category": "conversion",
        "event.action": "signup",
        "event.source": "backend"
    }
):
    pass
```

---

## Querying Events

### ClickHouse Queries

**Get all events for a project:**
```sql
SELECT *
FROM spans
WHERE project_id = 'proj_123'
  AND type = 'event'
  AND start_time >= now() - INTERVAL 7 DAY
ORDER BY start_time DESC;
```

**Analytics: Count events by name:**
```sql
SELECT
    name,
    COUNT(*) as event_count,
    AVG(duration_ms) as avg_duration
FROM spans
WHERE type = 'event'
  AND start_time >= now() - INTERVAL 1 DAY
GROUP BY name
ORDER BY event_count DESC;
```

**Get all events in a trace (mixed with spans/generations):**
```sql
SELECT *
FROM spans
WHERE trace_id = 'trace_xyz'
ORDER BY start_time ASC;
```

---

## Benefits of New Pattern

### ✅ OTEL-Native Compatibility
- Compatible with OpenTelemetry standard
- Works with OTEL collectors and exporters
- Industry-standard observability

### ✅ Unified Storage
- All spans in one table
- Efficient querying across all span types
- Consistent event structure

### ✅ Hierarchical Events
- Events can have parent-child relationships
- Events can be part of traces
- Full context propagation

### ✅ Better Analytics
- Filter by `type = 'event'` for event-specific analytics
- Query across all span types
- Unified metrics and dashboards

---

## Breaking Changes

### API Breaking Changes
- ❌ `/v1/ingest/batch` rejects `event_type: "event"` → returns `400 Bad Request`
- ❌ `/v1/telemetry/validate` rejects "event" type → returns validation error
- ✅ Valid types: `["trace", "span", "quality_score", "session"]`

### SDK Breaking Changes
- ❌ Python Legacy SDK: `TelemetryEventType.EVENT` removed from enum
- ❌ Old code using `event_type=TelemetryEventType.EVENT` will raise `AttributeError`
- ✅ Migration: Use `TelemetryEventType.SPAN` with `payload.type="event"`

### No User Impact
- ✅ Early development phase - no production users affected
- ✅ No database migration needed (events table never existed)
- ✅ Clear migration path documented

---

## Rollback Plan

If you need to temporarily rollback (not recommended):

1. **Backend**: Revert commits from this cleanup
2. **SDK**: Pin to previous version in `requirements.txt`
3. **Long-term**: Migrate to new pattern (old pattern will not be maintained)

---

## Support

For questions or issues:
- GitHub Issues: https://github.com/brokle/brokle/issues
- Documentation: https://docs.brokle.com
- Community: https://discord.gg/brokle

---

**Migration Date**: October 29, 2025
**Affected Versions**: All versions after this migration
**Status**: ✅ Complete - No backward compatibility maintained
