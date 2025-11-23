# OTEL Events Support - Future Implementation Guide

**Status**: Deferred (Not Currently Implemented)
**Decision Date**: November 19, 2025
**Rationale**: Current attribute-based approach sufficient for 99% of use cases

---

## What Are OTEL Events?

OpenTelemetry Events are **timestamped occurrences within a span**. They represent discrete moments during span execution, similar to structured logs.

**Event Structure**:
```json
{
  "timestamp": "2025-11-19T10:00:00.123Z",
  "name": "gen_ai.content.prompt",
  "attributes": {
    "gen_ai.prompt": [
      {"role": "user", "content": "Hello"}
    ]
  }
}
```

---

## OTEL GenAI Recommendation

The OTEL GenAI Working Group **recommends Events for LLM input/output**:

> "The LLM Working Group has recommended capturing details on events instead of span attributes because many backend systems can struggle with those often large payloads."

**Official Event Types**:
- `gen_ai.content.prompt` - LLM input messages
- `gen_ai.content.completion` - LLM output messages

**Source**: https://opentelemetry.io/docs/specs/semconv/gen-ai/gen-ai-events/

---

## Why Brokle Deferred Events

### Timestamp Relevance Test

**Events should be used when timestamps matter**. Examples:

✅ **Good use cases for Events**:
- System calls with precise timing
- Payment gateway callbacks
- Database query execution points
- Iterative loop steps
- Streaming chunk arrivals

❌ **Poor use cases for Events**:
- LLM input (timestamp = span start)
- LLM output (timestamp = span end)
- Function arguments (timestamp not meaningful)
- Return values (timestamp = span end)

**For LLM observability**: Span start/end times already capture what matters. Events add complexity without benefit.

---

## Current Architecture (Attribute-Based)

### ClickHouse Schema

**Spans table already supports Events** (zero-cost future-proofing):

```sql
CREATE TABLE spans (
    -- Standard fields
    span_id String,
    trace_id String,
    start_time DateTime64(3),

    -- Attributes (current implementation)
    span_attributes JSON CODEC(ZSTD(1)),

    -- Events (preserved but unused)
    events_timestamp Array(DateTime64(3)) CODEC(ZSTD(1)),
    events_name Array(LowCardinality(String)) CODEC(ZSTD(1)),
    events_attributes Array(String) CODEC(ZSTD(1)),
    events_dropped_attributes_count Array(UInt32) CODEC(ZSTD(1)),

    -- ...
) ENGINE = MergeTree();
```

**Note**: Events columns exist but are not populated. OTLP converter reads Events from wire protocol but doesn't process them.

---

## Performance Comparison

| Aspect | Span Attributes (Current) | Events (Future) |
|--------|--------------------------|-----------------|
| **Storage** | Single column (co-located) | Nested arrays in same row |
| **Query** | Simple JSON extraction | Array unnesting (arrayJoin) |
| **Compression** | ZSTD on full column | ZSTD on arrays |
| **Timestamps** | Span start/end | Per-event precision |
| **Size limit** | 1MB (configurable) | Same 1MB limit |
| **Performance** | ✅ Faster (one query) | ⚠️ Slower (JOIN required) |

**Brokle Analysis**:
- 78% cheaper than S3 with ZSTD compression
- Typical LLM calls <100KB (well under 1MB limit)
- Attributes approach proven at scale

---

## When to Implement Events

**Trigger Conditions** (any of):

1. **Size Issues**: Regular spans >1MB causing backend rejections
2. **Timestamp Granularity**: Need sub-span timing (e.g., streaming chunks)
3. **Multimodal Content**: Images/audio requiring separate retrieval
4. **Compliance Requirement**: Must match OTEL GenAI spec exactly
5. **Performance Degradation**: Attribute-based queries become bottleneck

**Current Status**: None of these conditions met

---

## Implementation Roadmap (When Needed)

### Phase 1: Backend Event Support

**File**: `internal/core/services/observability/otlp_converter.go`

**Add extraction logic**:
```go
// Extract OTLP Events from spans
if len(span.Events) > 0 {
    eventsTimestamp := make([]string, len(span.Events))
    eventsName := make([]string, len(span.Events))
    eventsAttributes := make([]string, len(span.Events))

    for i, event := range span.Events {
        eventsTimestamp[i] = convertUnixNano(event.TimeUnixNano).Format(time.RFC3339Nano)
        eventsName[i] = event.Name

        // Extract gen_ai.content.prompt or gen_ai.content.completion
        eventAttrs := extractAttributesFromKeyValues(event.Attributes)
        eventsAttributes[i] = marshalAttributes(eventAttrs)
    }

    payload["events_timestamp"] = eventsTimestamp
    payload["events_name"] = eventsName
    payload["events_attributes"] = eventsAttributes
}
```

**Add to `createTraceEvent()`**:
```go
// Extract input from gen_ai.content.prompt event
for i, eventName := range payload["events_name"].([]string) {
    if eventName == "gen_ai.content.prompt" {
        eventAttrs := payload["events_attributes"].([]string)[i]
        // Parse and extract gen_ai.prompt
        payload["input"] = extractPromptFromEvent(eventAttrs)
        break
    }
}
```

### Phase 2: SDK Event Support

**Python SDK**:
```python
# Add event creation helper
with client.start_as_current_span("llm-call") as span:
    # Add prompt event
    span.add_event(
        "gen_ai.content.prompt",
        attributes={
            "gen_ai.prompt": json.dumps([
                {"role": "user", "content": "Hello"}
            ])
        }
    )

    # Make LLM call

    # Add completion event
    span.add_event(
        "gen_ai.content.completion",
        attributes={
            "gen_ai.completion": json.dumps([
                {"role": "assistant", "content": "Hi!"}
            ])
        }
    )
```

**JavaScript SDK**:
```typescript
span.addEvent('gen_ai.content.prompt', {
  'gen_ai.prompt': JSON.stringify(messages)
});
```

### Phase 3: ClickHouse Query Patterns

**Extract events for analysis**:
```sql
-- Get all prompt events with their timestamps
SELECT
    span_id,
    arrayJoin(arrayZip(
        events_timestamp,
        events_name,
        events_attributes
    )) AS event_tuple,
    event_tuple.1 AS event_time,
    event_tuple.2 AS event_name,
    event_tuple.3 AS event_attrs
FROM spans
WHERE has(events_name, 'gen_ai.content.prompt')
AND event_name = 'gen_ai.content.prompt';
```

**Extract prompt content**:
```sql
-- Extract gen_ai.prompt from event attributes
SELECT
    span_id,
    JSONExtractString(event_attrs, 'gen_ai.prompt') AS prompt_content
FROM (
    SELECT
        span_id,
        arrayJoin(events_attributes) AS event_attrs,
        arrayJoin(events_name) AS event_name
    FROM spans
    WHERE has(events_name, 'gen_ai.content.prompt')
)
WHERE event_name = 'gen_ai.content.prompt';
```

### Phase 4: Frontend Timeline Visualization

Events enable rich timeline UIs:

```typescript
// Display events on span timeline
<Timeline>
  <Event time={span.start_time} name="Span Start" />
  <Event time={promptEvent.timestamp} name="Prompt Sent" />
  <Event time={completionEvent.timestamp} name="Response Received" />
  <Event time={span.end_time} name="Span End" />
</Timeline>
```

---

## Configuration (When Implemented)

### Enable Event Capture

**Environment Variable**:
```bash
# OTEL standard flag for capturing message content in events
OTEL_INSTRUMENTATION_GENAI_CAPTURE_MESSAGE_CONTENT=true
```

**SDK Configuration**:
```python
# Python
client = Brokle(
    api_key="...",
    capture_events=True,  # Enable event-based I/O
)

# JavaScript
const client = new Brokle({
  apiKey: '...',
  captureEvents: true,
});
```

---

## Migration Path

**If/when implementing Events**:

### Option A: Dual Mode (Recommended)
```python
# Support both attributes AND events
span.set_attribute("input.value", data)  # Keep for simple queries
span.add_event("gen_ai.content.prompt", {...})  # Add for timestamps
```

**Benefits**:
- Backward compatible
- Best of both worlds
- Gradual migration

### Option B: Events Only
```python
# Remove attributes, use only events
span.add_event("gen_ai.content.prompt", {...})
```

**Challenges**:
- Breaking change
- Complex queries
- Higher query latency

**Recommendation**: Option A (dual mode) if implemented

---

## Performance Considerations

### Query Latency

**Attributes** (Current):
```sql
SELECT JSONExtractString(span_attributes, 'input.value')
FROM spans
WHERE project_id = '...'
-- Query time: ~10-50ms
```

**Events** (Future):
```sql
SELECT arrayJoin(events_attributes)
FROM spans
WHERE has(events_name, 'gen_ai.content.prompt')
-- Query time: ~50-200ms (array operations slower)
```

### Storage Overhead

**Additional storage per event**:
- Timestamp: 8 bytes (DateTime64)
- Name: ~20 bytes (LowCardinality String)
- Attributes: Variable (JSON string)
- Array overhead: ~16 bytes per array

**Estimate**: +10-15% storage for typical LLM spans

---

## Testing Strategy (When Implemented)

```go
// Backend tests
func TestExtractEventsFromOTLP(t *testing.T) {
    // Verify events extracted from OTLP
}

func TestGenAIPromptEventExtraction(t *testing.T) {
    // Verify gen_ai.content.prompt event creates trace.input
}

func TestGenAICompletionEventExtraction(t *testing.T) {
    // Verify gen_ai.content.completion event creates trace.output
}

func TestEventTimestampPreserved(t *testing.T) {
    // Verify event timestamps stored correctly
}
```

---

## References

- OTEL Events Spec: https://opentelemetry.io/docs/specs/otel/trace/api/#add-events
- GenAI Events: https://opentelemetry.io/docs/specs/semconv/gen-ai/gen-ai-events/
- ClickHouse Arrays: https://clickhouse.com/docs/en/sql-reference/data-types/array
- ClickHouse arrayJoin: https://clickhouse.com/docs/en/sql-reference/functions/array-join

---

## Decision Log

**November 19, 2025**: Events support deferred
- **Reason**: Timestamps not meaningful for LLM I/O
- **Alternative**: Attribute-based approach with ZSTD compression
- **Performance**: Proven at scale (78% cheaper than S3)
- **Re-evaluation**: When size or performance issues observed

**Review Date**: Q2 2026 or when metrics indicate need
