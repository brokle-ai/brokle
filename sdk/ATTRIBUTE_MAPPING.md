# Brokle SDK Attribute Mapping - OTEL 1.38+ Specification

**Date**: November 13, 2025
**Schema Version**: v2.0 (OTEL 1.38+ Native)
**Applies To**: Python SDK v2.0.0+, JavaScript SDK v1.3.0+

---

## Overview

This document defines the **canonical attribute naming conventions** for all Brokle SDKs across all programming languages. These conventions ensure consistency with OpenTelemetry 1.38+ semantic conventions and enable backend materialized column extraction.

---

## Namespace Rules

### ✅ **OTEL Standard Namespaces** (Required)

All OTEL-defined attributes MUST use official namespace conventions:

```
gen_ai.*          - GenAI semantic conventions (OTEL 1.28+)
user.*            - User context (OTEL standard)
session.*         - Session tracking (OTEL standard)
service.*         - Service metadata (OTEL resource attributes)
deployment.*      - Deployment context (OTEL resource attributes)
```

### ✅ **Brokle Custom Namespaces** (Dot Notation ONLY)

All Brokle custom attributes MUST use **dot notation** for hierarchical organization:

```
brokle.span.*     - Span metadata
brokle.cost.*     - Cost tracking (backend-set only)
brokle.prompt.*   - Prompt management
brokle.routing.*  - Intelligent routing
brokle.score.*    - Quality scores
```

### ❌ **FORBIDDEN Patterns**

```
❌ brokle.span_type     - Underscore notation (inconsistent)
❌ brokle_span_type     - No hierarchy (flat namespace)
❌ brokle-span-type     - Hyphen notation (invalid)
```

---

## Attribute Reference

### Gen AI Attributes (OTEL 1.38+)

| Attribute | Type | Example | Description |
|-----------|------|---------|-------------|
| `gen_ai.operation.name` | string | `"chat"` | Operation type |
| `gen_ai.provider.name` | string | `"openai"` | LLM provider |
| `gen_ai.request.model` | string | `"gpt-4"` | Requested model |
| `gen_ai.response.model` | string | `"gpt-4-0613"` | Actual model used |
| `gen_ai.request.temperature` | number | `0.7` | Temperature parameter |
| `gen_ai.request.max_tokens` | number | `1000` | Max tokens parameter |
| `gen_ai.request.top_p` | number | `0.9` | Top-p parameter |
| `gen_ai.usage.input_tokens` | **string** | `"1234"` | Input tokens (CRITICAL: string!) |
| `gen_ai.usage.output_tokens` | **string** | `"5678"` | Output tokens (CRITICAL: string!) |
| `gen_ai.input.messages` | JSON string | `'[{...}]'` | Input messages array |
| `gen_ai.output.messages` | JSON string | `'[{...}]'` | Output messages array |
| `gen_ai.response.id` | string | `"chatcmpl-123"` | Provider response ID |
| `gen_ai.response.finish_reasons` | JSON array | `["stop"]` | Finish reasons |

---

### Brokle Custom Attributes

#### Span Management

| Attribute | Type | Example | SDK Sets? | Backend Extracts? |
|-----------|------|---------|-----------|-------------------|
| `brokle.span.type` | string | `"generation"` | ✅ SDK | ✅ Materialized column |
| `brokle.span.level` | string | `"DEFAULT"` | ✅ SDK | ✅ Materialized column |

**Valid Span Types**:
- `generation` - LLM generation
- `span` - Generic span
- `event` - Event marker
- `tool` - Tool/function call
- `agent` - Agent execution
- `chain` - Chain of operations
- `retrieval` - Vector/document retrieval
- `embedding` - Embedding generation

**Valid Span Levels**:
- `DEBUG`, `DEFAULT`, `INFO`, `WARNING`, `ERROR`

---

#### Cost Tracking (Backend-Only)

| Attribute | Type | Example | SDK Sets? | Backend Sets? |
|-----------|------|---------|-----------|---------------|
| `brokle.cost.input` | **string** | `"0.004500000"` | ❌ NO | ✅ YES |
| `brokle.cost.output` | **string** | `"0.018000000"` | ❌ NO | ✅ YES |
| `brokle.cost.total` | **string** | `"0.022500000"` | ❌ NO | ✅ YES |

**CRITICAL RULES**:
1. ❌ **SDKs MUST NOT set cost attributes**
2. ✅ **Backend calculates** costs from usage tokens + model pricing
3. ✅ **Backend formats** as strings with 9 decimal precision
4. ✅ **ClickHouse extracts** to Decimal(18,9) materialized columns

**Why Backend-Only**:
- Prevents billing errors from client-side estimates
- Centralized model pricing updates
- Exact precision with Decimal(18,9)
- Consistent cost calculation across all SDKs

---

#### Usage Metrics

| Attribute | Type | Example | Description |
|-----------|------|---------|-------------|
| `brokle.usage.total_tokens` | **string** | `"6912"` | Total tokens (convenience) |
| `brokle.usage.latency_ms` | number | `1234` | Response latency |

---

#### Prompt Management

| Attribute | Type | Example | SDK Sets? | Backend Extracts? |
|-----------|------|---------|-----------|-------------------|
| `brokle.prompt.id` | string | `"prompt-123"` | ✅ Optional | ✅ Materialized |
| `brokle.prompt.name` | string | `"chat-v1"` | ✅ Optional | ✅ Materialized |
| `brokle.prompt.version` | string | `"2"` | ✅ Optional | ✅ Materialized |

---

#### Environment & Metadata

| Attribute | Type | Example | Description |
|-----------|------|---------|-------------|
| `brokle.environment` | string | `"production"` | Environment tag |
| `brokle.version` | string | `"v1.2.3"` | Application version |
| `brokle.release` | string | `"release-456"` | Release identifier |
| `brokle.streaming` | boolean | `true` | Streaming response flag |
| `brokle.cached` | boolean | `true` | Response from cache |

---

#### OTEL Resource Attributes (Trace-Level)

| Attribute | Type | Example | Description |
|-----------|------|---------|-------------|
| `user.id` | string | `"user-123"` | OTEL standard user ID |
| `session.id` | string | `"session-456"` | OTEL standard session ID |

---

## Data Type Rules

### Strings vs Numbers

**Always Strings**:
- ✅ `gen_ai.usage.input_tokens` → `"1234"`
- ✅ `gen_ai.usage.output_tokens` → `"5678"`
- ✅ `brokle.usage.total_tokens` → `"6912"`
- ✅ `brokle.cost.*` → `"0.004500000"` (backend-only)

**Numbers**:
- ✅ `gen_ai.request.temperature` → `0.7`
- ✅ `gen_ai.request.max_tokens` → `1000`
- ✅ `brokle.usage.latency_ms` → `1234`

**JSON Strings**:
- ✅ `gen_ai.input.messages` → `'[{"role":"user","content":"hi"}]'`
- ✅ `gen_ai.output.messages` → `'[{"role":"assistant","content":"hello"}]'`

---

## SDK Implementation Examples

### Python SDK v2.0.0

```python
from brokle import get_client
from brokle.types.attributes import BrokleOtelSpanAttributes as Attrs

client = get_client()

with client.start_as_current_span("llm-call", as_type="generation") as span:
    # ✅ CORRECT: Use constants (recommended)
    span.set_attribute(Attrs.BROKLE_SPAN_TYPE, "generation")
    span.set_attribute(Attrs.GEN_AI_PROVIDER_NAME, "openai")
    span.set_attribute(Attrs.GEN_AI_REQUEST_MODEL, "gpt-4")

    # ✅ CORRECT: Tokens as strings
    span.set_attribute(Attrs.GEN_AI_USAGE_INPUT_TOKENS, "1234")
    span.set_attribute(Attrs.GEN_AI_USAGE_OUTPUT_TOKENS, "5678")

    # ❌ WRONG: Don't set costs (backend-only)
    # span.set_attribute("brokle.cost.total", "0.022")  # Don't do this!

    span.update(output="response text")

client.flush()
```

### JavaScript SDK v1.3.0+

```typescript
import { Brokle, Attrs } from '@brokle/sdk'

const client = new Brokle({ apiKey: 'bk_...' })

await client.traced('llm-call', async (span) => {
  // ✅ CORRECT: Use constants
  span.setAttribute(Attrs.BROKLE_SPAN_TYPE, 'generation')
  span.setAttribute(Attrs.GEN_AI_PROVIDER_NAME, 'openai')

  // ✅ CORRECT: Tokens as strings
  span.setAttribute(Attrs.GEN_AI_USAGE_INPUT_TOKENS, '1234')
  span.setAttribute(Attrs.GEN_AI_USAGE_OUTPUT_TOKENS, '5678')

  // ❌ WRONG: Don't set costs
  // span.setAttribute('brokle.cost.total', '0.022')

  return response
})
```

---

## Backend Behavior (For Reference)

### Materialized Column Extraction

Backend extracts attributes from JSON to typed materialized columns:

```sql
-- Materialized column definitions (ClickHouse)
gen_ai_request_model String MATERIALIZED span_attributes.`gen_ai.request.model`
gen_ai_usage_input_tokens UInt32 MATERIALIZED toUInt32OrNull(span_attributes.`gen_ai.usage.input_tokens`)
brokle_span_type String MATERIALIZED span_attributes.`brokle.span.type`
brokle_cost_total Decimal(18,9) MATERIALIZED toDecimal64OrNull(span_attributes.`brokle.cost.total`, 9)
```

**Performance**: Queries on materialized columns are 10-25x faster than JSON extraction.

---

## Migration Support

### Testing Compatibility

```python
# Test attribute extraction
import json

def verify_attributes(span):
    """Verify attributes are OTEL 1.38+ compliant"""
    attrs = span.attributes

    # Check namespace
    assert "brokle.span.type" in attrs, "Should use dot notation"
    assert "brokle.span_type" not in attrs, "Should not use underscore"

    # Check token format
    input_tokens = attrs.get("gen_ai.usage.input_tokens")
    assert isinstance(input_tokens, str), "Tokens should be strings"

    # Verify no cost attributes from SDK
    assert "brokle.cost.total" not in attrs or "backend-set", "SDKs shouldn't set costs"

    print("✅ Attributes are OTEL 1.38+ compliant")
```

---

## Questions?

**Q: Why remove `brokle.span_type` (underscore)?**
A: Inconsistent with OTEL conventions. All OTEL attributes use dot notation for hierarchical namespaces (e.g., `gen_ai.request.model`). Consistency enables better tooling support.

**Q: Why tokens as strings?**
A: OTEL 1.38+ best practice for consistency across platforms. Prevents numeric type mismatches between languages (JavaScript Number vs Python int vs Go uint32).

**Q: Can I still send costs from SDK?**
A: Not recommended. Backend calculates costs from usage + model pricing for accuracy. If you have custom pricing, send usage tokens and configure model pricing in backend.

**Q: Will v1.x SDKs break?**
A: No, but they won't benefit from backend optimizations. The `brokle.span_type` attribute will be stored but not extracted to materialized columns, resulting in slower queries.

**Q: Do I need to update my code?**
A: Only if you hardcoded `"brokle.span_type"` string literals. If using `Attrs.BROKLE_SPAN_TYPE` constants, no changes needed.
