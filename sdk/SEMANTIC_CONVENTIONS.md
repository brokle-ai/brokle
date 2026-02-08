# Brokle Semantic Conventions

**Version**: 1.0.0
**Date**: November 19, 2025
**Compliance**: OTLP 1.38+ with OpenInference Extensions

---

## Overview

Brokle follows **OpenTelemetry (OTEL) semantic conventions** with extensions from the **OpenInference specification** for AI/ML workloads. This document defines all attributes used across SDKs and backend.

**Standards Followed**:
- ✅ OTEL Trace API 1.38+
- ✅ OTEL GenAI Semantic Conventions 1.28+/1.38+
- ✅ OpenInference Semantic Conventions
- ✅ Industry consensus patterns (validated against 7 OTEL-native platforms)

---

## Attribute Namespaces

### 1. OTEL Standard (`gen_ai.*`)

**For LLM-specific data** following official OTEL GenAI conventions:

| Attribute | Type | Description | Example |
|-----------|------|-------------|---------|
| `gen_ai.provider.name` | String | LLM provider | `"openai"`, `"anthropic"` |
| `gen_ai.request.model` | String | Requested model | `"gpt-4"`, `"claude-3-opus"` |
| `gen_ai.response.model` | String | Actual model used | `"gpt-4-0613"` |
| `gen_ai.input.messages` | JSON String | Input messages (ChatML) | `[{"role":"user","content":"..."}]` |
| `gen_ai.output.messages` | JSON String | Output messages (ChatML) | `[{"role":"assistant","content":"..."}]` |
| `gen_ai.usage.input_tokens` | Integer | Input token count | `2450` |
| `gen_ai.usage.output_tokens` | Integer | Output token count | `892` |

**Reference**: https://opentelemetry.io/docs/specs/semconv/gen-ai/

---

### 2. OpenInference Standard (`input.value`, `output.value`)

**For generic (non-LLM) data** following OpenInference pattern:

| Attribute | Type | Description | Example |
|-----------|------|-------------|---------|
| `input.value` | String | Generic input data | `{"query":"weather","location":"Bangalore"}` |
| `input.mime_type` | String | Content type | `"application/json"`, `"text/plain"` |
| `output.value` | String | Generic output data | `{"temp":25,"status":"sunny"}` |
| `output.mime_type` | String | Content type | `"application/json"`, `"text/plain"` |

**Supported MIME Types**:
- `application/json` - Structured data (objects, arrays)
- `text/plain` - Unstructured text

**Reference**: https://github.com/Arize-ai/openinference/blob/main/spec/semantic_conventions.md

---

### 3. Brokle Custom (`brokle.*`)

**For Brokle-specific features**:

#### Trace Management
| Attribute | Type | Description | Storage Location |
|-----------|------|-------------|------------------|
| `brokle.trace.tags` | JSON Array | Filterable tags | `traces.tags` |
| `brokle.trace.metadata` | JSON Object | Custom metadata | `traces.metadata` |
| `brokle.release` | String | Global app version (e.g., "v2.1.24") | `traces.metadata.brokle.release` → materialized column `traces.release` |
| `brokle.span.version` | String | Per-span experiment/version (e.g., "experiment-A") | `span_attributes['brokle.span.version']` → materialized column `span_version` |
| `brokle.environment` | String | Environment (production/staging/dev) | `traces.environment` |

#### Span Categorization
| Attribute | Type | Description | Values | Storage Location |
|-----------|------|-------------|--------|------------------|
| `brokle.span.type` | String | Operation type | `generation`, `span`, `tool`, `agent`, `chain`, `retrieval` | `spans.attributes.brokle.span.type` → materialized column `spans.span_type` |
| `brokle.span.level` | String | Importance level | `DEBUG`, `DEFAULT`, `WARNING`, `ERROR` | `spans.attributes.brokle.span.level` |


#### LLM Analytics (Extracted by Backend)
| Attribute | Type | Description |
|-----------|------|-------------|
| `brokle.llm.message_count` | Integer | Total messages |
| `brokle.llm.user_message_count` | Integer | User messages count |
| `brokle.llm.assistant_message_count` | Integer | Assistant messages count |
| `brokle.llm.system_message_count` | Integer | System messages count |
| `brokle.llm.tool_message_count` | Integer | Tool messages count |
| `brokle.llm.first_role` | String | First message role |
| `brokle.llm.last_role` | String | Last message role |
| `brokle.llm.has_tool_calls` | Boolean | Contains tool invocations |

**Note**: LLM analytics attributes are **auto-extracted by backend** from ChatML data in `input.value` or `gen_ai.input.messages`. SDKs do NOT set these directly.

#### Cost Tracking (Backend-Calculated, Map Storage)

**ARCHITECTURE**: Costs are calculated by backend and stored in dedicated ClickHouse Map columns (NOT in attributes).

**Storage Locations**:
- `spans.cost_details` Map - Per-usage-type cost breakdown: `{"input": "0.001", "output": "0.002", "total": "0.003"}`
- `spans.pricing_snapshot` Map - Historical pricing audit: `{"input_price_per_million": "5.00", "output_price_per_million": "15.00"}`
- `spans.total_cost` Decimal(18,12) - Pre-computed total for fast SUM() aggregation

**Query Examples**:
```sql
-- Cost breakdown
SELECT
    cost_details['input'] as input_cost,
    cost_details['output'] as output_cost,
    total_cost
FROM spans WHERE span_id = '...'

-- Aggregate total spend
SELECT SUM(total_cost) as total_spend
FROM spans WHERE project_id = '...' AND start_time >= '2025-01-01'
```

**Note**: SDKs should NOT set cost attributes. Backend calculates from usage tokens + provider pricing tables.

---

## Priority Order for Input/Output

When multiple attributes are present, backend uses this priority:

### For LLM Spans:
```
1. gen_ai.input.messages     (PRIORITY 1 - OTLP standard)
2. input.value               (PRIORITY 2 - OpenInference fallback)
```

### For Generic Spans:
```
1. input.value               (PRIORITY 1 - OpenInference standard)
2. gen_ai.input.messages     (PRIORITY 2 - Also supported)
```

**Auto-detection in SDKs**:
- ChatML format (`[{"role":"user",...}]`) → Sets `gen_ai.input.messages`
- Generic data → Sets `input.value` + `input.mime_type`

---

## Release & Version Fields: Two Distinct Concepts

Brokle supports two version-related fields for different use cases:

### Field Comparison Table

| Field | Set By | SDK Support | Namespace Key | Storage Location | Use Case |
|-------|--------|-------------|---------------|------------------|----------|
| **release** | User (global app version) | `Brokle(release="v2.1.24")` | `brokle.release` | Resource attribute → `traces.release` | Track which app version generated this trace |
| **version** (span) | User (per-span version) | `start_span(..., version="prompt-v3")` | `brokle.span.version` | `span_attributes['brokle.span.version']` → `spans.span_version` | Per-span A/B testing and experiment tracking |

**Note**: `version` is exclusively a **per-span parameter**. It is NOT available as a constructor/config option. Use `release` for global app versioning.

### Usage Examples

**Python - Setting Release and Per-Span Version**:
```python
from brokle import Brokle

# Initialize with release (global app version)
client = Brokle(
    release="v2.1.24",              # Global app version (resource attribute)
)

# Start span with per-span version (for A/B testing)
with client.start_as_current_span(
    "llm-generation",
    version="prompt-v3"  # Per-span version
) as span:
    # Do work
    pass
```

**JavaScript - Setting Release and Per-Span Version**:
```typescript
import { Brokle } from '@brokle/sdk';

// Initialize with release (global app version)
const client = new Brokle({
  release: 'v2.1.24',              // Global app version (resource attribute)
});

// Start span with per-span version (for A/B testing)
await client.startActiveSpan('llm-generation', async (span) => {
  // Do work
}, undefined, {
  version: 'prompt-v3'  // Per-span version
});
```

**Storage Behavior**:
- `release="v2.1.24"` → Resource attribute (`brokle.release`) → Materialized to `traces.release` column
- `version="prompt-v3"` (span) → Span attribute (`brokle.span.version`) → Materialized to `spans.span_version` column

### Query Examples

**Filter traces by release**:
```sql
SELECT trace_id, name, release
FROM traces
WHERE release = 'v2.1.24'
ORDER BY start_time DESC;
```

**Filter spans by version**:
```sql
SELECT span_id, span_name, version
FROM spans
WHERE version = 'prompt-v3'
ORDER BY start_time DESC;
```

---

## SDK Usage Patterns

### Pattern 1: Decorator (Generic Input/Output)

**Python**:
```python
from brokle import observe

@observe(capture_input=True, capture_output=True)
def get_weather(location: str, units: str = "celsius"):
    return {"temp": 25, "location": location}

result = get_weather("Bangalore")
```

**Attributes Set**:
```json
{
  "input.value": "{\"location\":\"Bangalore\",\"units\":\"celsius\"}",
  "input.mime_type": "application/json",
  "output.value": "{\"temp\":25,\"location\":\"Bangalore\",\"units\":\"celsius\"}",
  "output.mime_type": "application/json"
}
```

**JavaScript**:
```typescript
// Note: @observe decorator not yet implemented in JS SDK
// Use traced() method instead
```

---

### Pattern 2: Manual Span (Generic Data)

**Python**:
```python
from brokle import get_client

client = get_client()

with client.start_as_current_span(
    "api-request",
    input={"endpoint": "/weather", "query": "Bangalore"},
    output={"status": 200, "data": {"temp": 25}}
) as span:
    # Do work
    pass
```

**JavaScript**:
```typescript
const client = getClient();

await client.traced('api-request', async (span) => {
  // Do work
  return result;
}, undefined, {
  input: { endpoint: '/weather', query: 'Bangalore' },
  output: { status: 200, data: { temp: 25 } }
});
```

**Attributes Set**:
```json
{
  "input.value": "{\"endpoint\":\"/weather\",\"query\":\"Bangalore\"}",
  "input.mime_type": "application/json",
  "output.value": "{\"status\":200,\"data\":{\"temp\":25}}",
  "output.mime_type": "application/json"
}
```

---

### Pattern 3: LLM Generation (ChatML Auto-Detected)

**Python**:
```python
with client.start_as_current_span(
    "llm-conversation",
    input=[{"role": "user", "content": "What's the weather?"}],
    output=[{"role": "assistant", "content": "It's 25°C."}]
) as span:
    pass
```

**JavaScript**:
```typescript
await client.traced('llm-conversation', async (span) => {
  // LLM call
}, undefined, {
  input: [{ role: 'user', content: "What's the weather?" }],
  output: [{ role: 'assistant', content: "It's 25°C." }]
});
```

**Attributes Set** (auto-detected as ChatML):
```json
{
  "gen_ai.input.messages": "[{\"role\":\"user\",\"content\":\"What's the weather?\"}]",
  "gen_ai.output.messages": "[{\"role\":\"assistant\",\"content\":\"It's 25°C.\"}]"
}
```

**Backend Extracts** → LLM metadata:
```json
{
  "brokle.llm.message_count": 1,
  "brokle.llm.user_message_count": 1,
  "brokle.llm.first_role": "user",
  "brokle.llm.last_role": "user",
  "brokle.llm.has_tool_calls": false
}
```

---

### Pattern 4: Explicit LLM Generation

**Python**:
```python
with client.start_as_current_generation(
    name="chat",
    model="gpt-4",
    provider="openai",
    input_messages=[{"role": "user", "content": "Hello"}]
) as gen:
    # Make LLM call
    gen.set_attribute(
        Attrs.GEN_AI_OUTPUT_MESSAGES,
        json.dumps([{"role": "assistant", "content": "Hi!"}])
    )
```

**JavaScript**:
```typescript
await client.generation('chat', 'gpt-4', 'openai', async (span) => {
  const response = await openai.chat.completions.create({...});
  span.setAttribute(Attrs.GEN_AI_OUTPUT_MESSAGES, JSON.stringify([...]));
  return response;
});
```

---

## Data Flow

### SDK → Backend → ClickHouse

```
┌─────────────────────────────────────────────────────────────┐
│ SDK (Python/JavaScript)                                     │
├─────────────────────────────────────────────────────────────┤
│ 1. Detect data type (ChatML vs generic)                    │
│ 2. Set appropriate attributes:                             │
│    - ChatML → gen_ai.input.messages                        │
│    - Generic → input.value + input.mime_type               │
│ 3. Send via OTLP to /v1/traces                             │
└─────────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────────┐
│ Backend (Go)                                                │
├─────────────────────────────────────────────────────────────┤
│ 1. Parse OTLP request                                       │
│ 2. Extract attributes (priority order):                    │
│    - Try gen_ai.input.messages first                       │
│    - Fallback to input.value                               │
│ 3. Validate/auto-detect MIME type                          │
│ 4. Truncate if >1MB                                        │
│ 5. Extract LLM metadata if ChatML                          │
│ 6. Convert to Brokle events                                │
│ 7. Publish to Redis Streams                                │
└─────────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────────┐
│ Worker (Go)                                                 │
├─────────────────────────────────────────────────────────────┤
│ 1. Consume from Redis Streams                              │
│ 2. Route to TraceService/SpanService                       │
│ 3. Write to ClickHouse                                     │
└─────────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────────┐
│ ClickHouse (Database)                                       │
├─────────────────────────────────────────────────────────────┤
│ traces table:                                               │
│   - input String (generic OR LLM)                          │
│   - output String (generic OR LLM)                         │
│   - input_mime_type String                                 │
│   - output_mime_type String                                │
│                                                             │
│ spans table:                                                │
│   - span_attributes JSON (contains brokle.llm.*)           │
│   - Materialized columns for analytics                     │
└─────────────────────────────────────────────────────────────┘
```

---

## Best Practices

### 1. Use Auto-Detection

**✅ DO**:
```python
# Let SDK auto-detect format
with client.start_as_current_span("trace", input=data):
    pass
```

**❌ DON'T**:
```python
# Manual detection unnecessary
if is_chatml(data):
    span.set_attribute("gen_ai.input.messages", ...)
else:
    span.set_attribute("input.value", ...)
```

### 2. Let Backend Extract Metadata

**✅ DO**:
```python
# Send raw ChatML, backend extracts metadata
input = [{"role": "user", "content": "Hello"}]
span.set_attribute("gen_ai.input.messages", json.dumps(input))
```

**❌ DON'T**:
```python
# Don't manually count messages
span.set_attribute("brokle.llm.message_count", len(input))
```

### 3. Trust MIME Types

**✅ DO**:
```typescript
// Frontend: Trust MIME type from backend
if (inputMimeType === 'application/json') {
  return <JSONViewer data={JSON.parse(input)} />;
}
```

**❌ DON'T**:
```typescript
// Don't re-detect on frontend
if (looksLikeJSON(input)) { ... }
```

---

## Size Limits & Truncation

### Backend Limits

| Limit | Value | Behavior |
|-------|-------|----------|
| **Attribute Value** | 1MB | Truncated with `...[truncated]` suffix |
| **OTLP Batch** | 10MB | Request rejected with HTTP 413 |
| **ClickHouse Row** | Unlimited | ZSTD compression applied |

**Truncation Indicators**:
- `input_truncated: true` - Input was truncated
- `output_truncated: true` - Output was truncated

**Recommendation**: Keep input/output <100KB for optimal performance.

---

## Extended Token Usage (Cache, Audio, Multi-modal)

The backend supports 10+ token types for flexible cost tracking. SDKs can send any combination based on provider capabilities.

### Cache Tokens (Anthropic, OpenAI)

**Python - Prompt Caching**:
```python
from brokle.types.attributes import Attrs

span.set_attribute(Attrs.GEN_AI_USAGE_INPUT_TOKENS, 1000)
span.set_attribute(Attrs.GEN_AI_USAGE_INPUT_TOKENS_CACHE_READ, 500)  # Cache hit
span.set_attribute(Attrs.GEN_AI_USAGE_OUTPUT_TOKENS, 200)

# Backend stores in usage_details map:
# {"input": 1000, "cache_read_input_tokens": 500, "output": 200, "total": 1200}
```

**JavaScript - Prompt Caching**:
```typescript
import { Attrs } from '@brokle/sdk';

span.setAttribute(Attrs.GEN_AI_USAGE_INPUT_TOKENS, 1000);
span.setAttribute(Attrs.GEN_AI_USAGE_INPUT_TOKENS_CACHE_READ, 500);  // Cache hit
span.setAttribute(Attrs.GEN_AI_USAGE_INPUT_TOKENS_CACHE_CREATION, 100);  // Cache write
span.setAttribute(Attrs.GEN_AI_USAGE_OUTPUT_TOKENS, 200);
```

### Audio Tokens (OpenAI Whisper, TTS, Realtime API)

**Python - Audio Usage**:
```python
# Whisper transcription
span.set_attribute(Attrs.GEN_AI_USAGE_INPUT_AUDIO_TOKENS, 10000)

# TTS synthesis
span.set_attribute(Attrs.GEN_AI_USAGE_OUTPUT_AUDIO_TOKENS, 5000)

# Realtime API (both directions)
span.set_attribute(Attrs.GEN_AI_USAGE_INPUT_AUDIO_TOKENS, 8000)
span.set_attribute(Attrs.GEN_AI_USAGE_OUTPUT_AUDIO_TOKENS, 6000)
```

**JavaScript - Audio Usage**:
```typescript
// Whisper
span.setAttribute(Attrs.GEN_AI_USAGE_INPUT_AUDIO_TOKENS, 10000);

// TTS
span.setAttribute(Attrs.GEN_AI_USAGE_OUTPUT_AUDIO_TOKENS, 5000);
```

### Reasoning Tokens (OpenAI o1 Models)

**Python - o1 Models with Internal Reasoning**:
```python
# OpenAI o1-preview/o1-mini
span.set_attribute(Attrs.GEN_AI_USAGE_INPUT_TOKENS, 500)
span.set_attribute(Attrs.GEN_AI_USAGE_REASONING_TOKENS, 15000)  # Internal CoT
span.set_attribute(Attrs.GEN_AI_USAGE_OUTPUT_TOKENS, 200)

# Backend calculates costs separately for reasoning tokens
# pricing_snapshot: {"input_price_per_million": 2.50, "reasoning_price_per_million": 10.00, ...}
```

**JavaScript - o1 Models**:
```typescript
span.setAttribute(Attrs.GEN_AI_USAGE_INPUT_TOKENS, 500);
span.setAttribute(Attrs.GEN_AI_USAGE_REASONING_TOKENS, 15000);
span.setAttribute(Attrs.GEN_AI_USAGE_OUTPUT_TOKENS, 200);
```

### Image Tokens (Vision Models)

**Python - GPT-4V, Claude Vision**:
```python
# Vision model usage
span.set_attribute(Attrs.GEN_AI_USAGE_INPUT_TOKENS, 300)
span.set_attribute(Attrs.GEN_AI_USAGE_IMAGE_TOKENS, 1500)  # Image processing
span.set_attribute(Attrs.GEN_AI_USAGE_OUTPUT_TOKENS, 400)
```

**JavaScript - Vision Models**:
```typescript
span.setAttribute(Attrs.GEN_AI_USAGE_IMAGE_TOKENS, 1500);
```

### Flexible Backend Storage

**Backend stores all token types in `usage_details` Map:**
```sql
-- ClickHouse query
SELECT
  usage_details,  -- {"input": 1000, "cache_read_input_tokens": 500, "output": 200, "audio_input": 5000}
  cost_details,   -- {"input": "0.002500", "cache_read": "0.000625", "output": "0.002000", "audio_input": "0.500000"}
  total_cost      -- Pre-computed sum for fast aggregation
FROM spans
WHERE model_name = 'gpt-4o-realtime'
```

**Benefits:**
- ✅ No schema changes needed for new token types
- ✅ Infinite flexibility (video, batch, future modalities)
- ✅ Per-usage-type cost breakdown
- ✅ Audit trail via pricing_snapshot

---

## Backward Compatibility

### Migration from Old Attributes

**Deprecated** (removed Nov 2025):
- ❌ `brokle.trace.input`
- ❌ `brokle.trace.output`

**Current Standard**:
- ✅ `input.value` (generic data)
- ✅ `gen_ai.input.messages` (LLM data)

**No migration needed**: Zero users at time of change.

---

## Query Examples

### ClickHouse Queries

**Filter by LLM message count**:
```sql
SELECT trace_id, brokle_llm_message_count
FROM spans
WHERE JSONExtractInt(span_attributes, 'brokle.llm.message_count') > 5
ORDER BY start_time DESC;
```

**Find traces with tool calls**:
```sql
SELECT DISTINCT trace_id, span_name
FROM spans
WHERE JSONExtractBool(span_attributes, 'brokle.llm.has_tool_calls') = true;
```

**Aggregate by message role distribution**:
```sql
SELECT
    AVG(JSONExtractInt(span_attributes, 'brokle.llm.user_message_count')) as avg_user,
    AVG(JSONExtractInt(span_attributes, 'brokle.llm.assistant_message_count')) as avg_assistant
FROM spans
WHERE brokle_span_type = 'generation'
AND start_time > now() - INTERVAL 7 DAY;
```

**Search input content** (JSON extraction):
```sql
SELECT trace_id, input
FROM traces
WHERE JSONHas(input, '$.query')  -- Check if input has 'query' field
AND JSONExtractString(input, 'query') = 'weather';
```

---

## Testing Conventions

### Backend Tests

**Test attribute extraction**:
```go
func TestExtractInputValue_GenericData(t *testing.T) {
    // Verify input.value → trace.input
}

func TestExtractGenAIMessages_LLMData(t *testing.T) {
    // Verify gen_ai.input.messages → trace.input
}

func TestLLMMetadataExtraction(t *testing.T) {
    // Verify brokle.llm.* attributes extracted
}
```

### SDK Tests

**Test serialization**:
```python
def test_serialize_with_mime_dict():
    result, mime = _serialize_with_mime({"key": "value"})
    assert mime == "application/json"

def test_is_chatml_format():
    assert _is_llm_messages_format([{"role": "user", "content": "..."}])
```

---

## References

### Official Specifications
- OTEL Semantic Conventions: https://opentelemetry.io/docs/specs/semconv/
- OTEL GenAI: https://opentelemetry.io/docs/specs/semconv/gen-ai/
- OTEL Trace API: https://opentelemetry.io/docs/specs/otel/trace/api/
- OpenInference: https://github.com/Arize-ai/openinference

### Industry Implementations
- Arize Phoenix (OpenInference reference): https://github.com/Arize-ai/phoenix
- SigNoz (OTEL ClickHouse): https://signoz.io/
- Grafana Tempo (OTEL Parquet): https://grafana.com/docs/tempo/
- OTEL Collector ClickHouse Exporter: https://github.com/open-telemetry/opentelemetry-collector-contrib

### Brokle Documentation
- ATTRIBUTE_MAPPING.md - Cross-platform attribute mapping
- EVENTS_FUTURE_SUPPORT.md - Future Events implementation guide
- ERROR_HANDLING_GUIDE.md - Error attribute patterns

---

## Changelog

### Version 1.1.0 (November 22, 2025)

**Added**:
- Extended token type constants (7 new):
  - `gen_ai.usage.input_tokens.cache_read` - Cache hit tokens (Anthropic/OpenAI)
  - `gen_ai.usage.input_tokens.cache_creation` - Cache write tokens
  - `gen_ai.usage.input_audio_tokens` - Audio input (Whisper, Realtime API)
  - `gen_ai.usage.output_audio_tokens` - Audio output (TTS, Realtime API)
  - `gen_ai.usage.reasoning_tokens` - Reasoning tokens (OpenAI o1 models)
  - `gen_ai.usage.image_tokens` - Image processing tokens (Vision models)
  - `gen_ai.usage.video_tokens` - Video tokens (future multi-modal)
- Release & Version fields (2 distinct):
  - `brokle.release` - Global app version (trace-level, in metadata)
  - `brokle.span.version` - Per-span version identifier (span-level in attributes)
- Span types: `agent`, `chain` for AI agent workflows
- Flexible backend storage via usage_details/cost_details Maps
- Pricing snapshot for billing audit trail

**Changed**:
- Backend schema: Materialized columns reduced from 16 → 3 (model_name, provider_name, span_type)
- Storage pattern: Individual columns → JSON attributes + Maps
- Cost precision: Decimal(18,9) → Decimal(18,12)
- JavaScript SDK: Release/version now sent as Resource attributes (not headers)

**Fixed**:
- JavaScript SDK critical bug: Release sent as header instead of Resource attribute
- Schema transition completed: All old field references removed

### Version 1.0.0 (November 19, 2025)

**Added**:
- OpenInference attributes: `input.value`, `output.value`, `input.mime_type`, `output.mime_type`
- LLM metadata attributes: 8 `brokle.llm.*` analytics attributes
- Auto-detection logic for ChatML vs generic data
- MIME type support with validation
- Truncation handling for large payloads (>1MB)

**Removed**:
- `brokle.trace.input` (replaced by `input.value`)
- `brokle.trace.output` (replaced by `output.value`)

**Changed**:
- Priority order: OTLP GenAI first, OpenInference fallback
- MIME type: Auto-detected if missing, validated if present

---

**Maintained by**: Brokle Platform Team
**Last Updated**: November 22, 2025
