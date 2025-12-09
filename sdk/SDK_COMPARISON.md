# Brokle SDK Competitive Analysis

**Version**: 2.0.0
**Date**: December 2025
**Competitors Analyzed**: OpenLIT, OpenLLMetry (Traceloop), Braintrust, Langfuse, LangSmith, Optik

---

## Executive Summary

This document provides a comprehensive comparison of Brokle's SDK against six leading LLM observability SDKs. The competitors fall into two categories:

**OTEL-Native SDKs**:
- **Brokle** - Explicit control with platform integration
- **OpenLIT** - Auto-magic with comprehensive coverage
- **OpenLLMetry** - Structured workflows with modularity
- **Langfuse** - OTEL-native with prompt management

**Custom Tracing SDKs**:
- **Braintrust** - Logging-first, eval-centric
- **LangSmith** - Hierarchical traces with framework integration
- **Optik** - Decorator-first with evaluation focus

### Quick Comparison Matrix

| Aspect | **Brokle** | **OpenLIT** | **OpenLLMetry** | **Braintrust** | **Langfuse** | **LangSmith** | **Optik** |
|--------|------------|-------------|-----------------|----------------|--------------|---------------|-----------|
| **Architecture** | OTEL-native | OTEL-native | OTEL-native | Custom + OTEL bridge | OTEL-native | Custom tracing | Custom tracing |
| **Philosophy** | Explicit control | Auto-magic | Workflow hierarchy | Eval-centric | Prompt-first | Framework-integrated | Decorator-first |
| **Provider Coverage** | 2 | 48+ | 32+ | 12+ | 15+ | 10+ | 15+ |
| **Unique Value** | Routing, caching, scoring | GPU, guardrails | Datasets, experiments | Eval framework | Prompt versioning | LangChain native | Prompt optimization |
| **Best For** | Platform-integrated observability | Quick setup, broad coverage | Complex AI pipelines | Evaluation workflows | Prompt-managed apps | LangChain users | Evaluation & optimization |

---

## Table of Contents

1. [Overview](#1-overview)
2. [Architecture Comparison](#2-architecture-comparison)
3. [Package Structure](#3-package-structure)
4. [Integration Patterns](#4-integration-patterns)
5. [Provider Support](#5-provider-support)
6. [Semantic Conventions](#6-semantic-conventions)
7. [Configuration](#7-configuration)
8. [Unique Features](#8-unique-features)
9. [Evaluation Frameworks](#9-evaluation-frameworks)
10. [Strengths & Weaknesses](#10-strengths--weaknesses)
11. [Recommendations](#11-recommendations)

---

## 1. Overview

### 1.1 Project Information

| Aspect | **Brokle** | **OpenLIT** | **OpenLLMetry** | **Braintrust** | **Langfuse** | **LangSmith** | **Optik** |
|--------|------------|-------------|-----------------|----------------|--------------|---------------|-----------|
| **Maintainer** | Brokle AI | OpenLIT.io | Traceloop | Braintrust Inc | Langfuse | LangChain | Comet ML |
| **License** | MIT | Apache 2.0 | Apache 2.0 | MIT | MIT | MIT | Apache 2.0 |
| **Python Package** | `brokle` | `openlit` | `traceloop-sdk` | `braintrust` | `langfuse` | `langsmith` | `opik` |
| **JS Package** | `brokle` (monorepo) | `@openlit/sdk` | `@traceloop/node-server-sdk` | `braintrust` | `langfuse` | `langsmith` | `opik` |
| **Primary Focus** | Observability + Routing | Auto-instrumentation | Pipeline tracing | Evaluations | Prompt management | LangChain ecosystem | Eval optimization |

### 1.2 Design Philosophy

**Brokle**: *"Explicit control with platform integration"*
- Three distinct integration patterns for different use cases
- Platform features (routing, caching, scoring) integrated into SDK
- Backend-calculated costs and analytics
- OpenInference compatibility for multi-platform interop

**OpenLIT**: *"One-liner auto-magic with comprehensive coverage"*
- Single `init()` call instruments everything automatically
- Widest provider coverage (48+ integrations)
- Client-side cost calculation with pricing JSON
- GPU monitoring and guardrails built-in

**OpenLLMetry**: *"Structured workflows with modular architecture"*
- Decorator-based workflow/task hierarchy
- Separate npm/pip packages per provider
- Focus on AI pipeline orchestration
- Dataset management and experiments

**Braintrust**: *"Logging-first, evaluation-centric"*
- Custom span/tracing system (NOT OTEL-native)
- Evals are primary use case with comprehensive Eval() DSL
- Provider wrappers + callback handlers
- Optional OTEL bridge for compatibility

**Langfuse**: *"OTEL-native with prompt management"*
- Built on OpenTelemetry standards
- Rich observation type system (span, generation, agent, tool, etc.)
- Integrated prompt versioning and templating
- Datasets and scoring for evaluations

**LangSmith**: *"Framework-integrated hierarchical tracing"*
- Custom hierarchical RunTree model (NOT OTEL-native)
- Deep LangChain integration via callback handlers
- Comprehensive evaluation framework with pytest plugin
- Prompt Hub for centralized prompt management

**Optik**: *"Decorator-first with evaluation focus"*
- Custom tracing with `@track` decorator
- Non-blocking background queue system
- Separate `opik-optimizer` package for prompt optimization
- 15+ framework integrations

---

## 2. Architecture Comparison

### 2.1 Tracing Architecture

| SDK | **Tracing Model** | **OTEL Compatibility** | **Export Protocol** |
|-----|-------------------|------------------------|---------------------|
| **Brokle** | OTEL TracerProvider | Native | OTLP/HTTP (Protobuf + Gzip) |
| **OpenLIT** | OTEL TracerProvider | Native | OTLP/HTTP or gRPC |
| **OpenLLMetry** | OTEL TracerProvider | Native | OTLP/HTTP or gRPC |
| **Braintrust** | Custom spans + queue | Optional bridge | Custom HTTP API |
| **Langfuse** | OTEL TracerProvider | Native | OTLP/HTTP |
| **LangSmith** | Custom RunTree | Optional OTEL extra | Custom HTTP API |
| **Optik** | Custom spans + queue | Limited (Vercel AI only) | Custom REST API |

### 2.2 Architecture Diagrams

#### OTEL-Native SDKs (Brokle, OpenLIT, OpenLLMetry, Langfuse)
```
┌─────────────────────────────────────────────────────────────┐
│                     Application Code                         │
│  ┌──────────────────┐  ┌──────────────────────┐            │
│  │ @observe/track   │  │  Wrapped Clients     │            │
│  │ Decorator        │  │  (OpenAI, Anthropic) │            │
│  └──────────────────┘  └──────────────────────┘            │
└────────────┬───────────────────────────────────────────────┘
             │
┌────────────▼───────────────────────────────────────────────┐
│  OpenTelemetry SDK                                          │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ TracerProvider                                       │   │
│  │  └─ BatchSpanProcessor                               │   │
│  │      └─ OTLPSpanExporter (Protobuf + Gzip)          │   │
│  └─────────────────────────────────────────────────────┘   │
└────────────┬───────────────────────────────────────────────┘
             │ OTLP/HTTP
             ▼
┌───────────────────────────────────────────────────────────┐
│   Backend (Brokle/OpenLIT/Traceloop/Langfuse/Grafana)     │
└───────────────────────────────────────────────────────────┘
```

#### Custom Tracing SDKs (Braintrust, LangSmith, Optik)
```
┌─────────────────────────────────────────────────────────────┐
│                     Application Code                         │
│  ┌──────────────────┐  ┌──────────────────────┐            │
│  │ @traceable/track │  │  Wrapped Clients     │            │
│  │ Decorator        │  │  (OpenAI, Anthropic) │            │
│  └──────────────────┘  └──────────────────────┘            │
└────────────┬───────────────────────────────────────────────┘
             │
┌────────────▼───────────────────────────────────────────────┐
│  Custom Tracing System                                      │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ RunTree / Span Manager                               │   │
│  │  └─ Background Queue (non-blocking)                  │   │
│  │      └─ HTTP Client (JSON + Compression)            │   │
│  └─────────────────────────────────────────────────────┘   │
└────────────┬───────────────────────────────────────────────┘
             │ Custom HTTP API
             ▼
┌───────────────────────────────────────────────────────────┐
│   Platform Backend (Braintrust/LangSmith/Optik)           │
└───────────────────────────────────────────────────────────┘
```

### 2.3 OTEL Compatibility Analysis

| SDK | **OTEL Export to Any Backend** | **Standard Semantic Conventions** | **Span Context Propagation** |
|-----|--------------------------------|-----------------------------------|------------------------------|
| **Brokle** | ✅ Full | ✅ GenAI 1.38+ | ✅ Native |
| **OpenLIT** | ✅ Full | ✅ GenAI 1.28+ | ✅ Native |
| **OpenLLMetry** | ✅ Full | ✅ GenAI conventions | ✅ Native |
| **Braintrust** | ⚠️ Via bridge only | ⚠️ Custom + bridge | ⚠️ Custom |
| **Langfuse** | ✅ Full | ✅ Custom OTEL attrs | ✅ Native |
| **LangSmith** | ⚠️ Optional extra | ⚠️ Custom RunTree | ⚠️ Custom |
| **Optik** | ⚠️ Vercel AI only | ⚠️ Custom | ⚠️ Headers-based |

**Analysis**: Brokle, OpenLIT, OpenLLMetry, and Langfuse are fully OTEL-native, enabling export to any OTEL-compatible backend (Grafana, Datadog, Jaeger, etc.). Braintrust, LangSmith, and Optik use custom tracing with limited OTEL bridges.

---

## 3. Package Structure

### 3.1 Python SDK Structure Comparison

| SDK | **Architecture** | **Core Files** | **Instrumentor Location** |
|-----|------------------|----------------|---------------------------|
| **Brokle** | Single package | client.py, decorators.py, exporter.py | `wrappers/` |
| **OpenLIT** | Monolithic | __init__.py, _instrumentors.py | `instrumentation/` (48 embedded) |
| **OpenLLMetry** | Modular monorepo | 33+ separate packages | Individual packages |
| **Braintrust** | Single package | logger.py (207KB), framework.py (57KB) | `wrappers/` |
| **Langfuse** | Single package | client.py (119KB), span.py (65KB) | `openai.py`, `langchain/` |
| **LangSmith** | Single package | client.py (334KB), run_helpers.py (75KB) | `wrappers/` |
| **Optik** | Single package | opik_client.py, tracker.py | `integrations/` |

### 3.2 JavaScript/TypeScript SDK Structure

| SDK | **Architecture** | **Build Output** | **Framework** |
|-----|------------------|------------------|---------------|
| **Brokle** | pnpm monorepo (4 packages) | ESM + CJS via tsup | Node.js |
| **OpenLIT** | Single package | ESM via esbuild | Node.js |
| **OpenLLMetry** | Nx + Lerna (15+ packages) | ESM + CJS via tsup | Node.js |
| **Braintrust** | Single package | ESM + CJS | Isomorphic |
| **Langfuse** | pnpm monorepo (6 packages) | ESM + CJS | Node.js 20+ |
| **LangSmith** | Single package | ESM + CJS | Isomorphic |
| **Optik** | Single package | ESM + CJS | Node.js |

### 3.3 Install Size & Dependencies

| SDK | **Python Deps** | **Install Footprint** | **Optional Deps** |
|-----|-----------------|----------------------|-------------------|
| **Brokle** | opentelemetry-sdk, httpx | Small | - |
| **OpenLIT** | opentelemetry-*, tiktoken | Medium-Large (all instrumentors) | GPU libs |
| **OpenLLMetry** | Variable (pick packages) | Variable | Provider-specific |
| **Braintrust** | requests, GitPython, wrapt | Medium | OTEL packages |
| **Langfuse** | opentelemetry-*, httpx, backoff | Medium | openai, langchain |
| **LangSmith** | httpx, orjson, zstandard | Medium | OTEL, pytest |
| **Optik** | pydantic, httpx, litellm | Medium | Provider-specific |

---

## 4. Integration Patterns

### 4.1 Pattern Comparison Matrix

| Pattern | **Brokle** | **OpenLIT** | **OpenLLMetry** | **Braintrust** | **Langfuse** | **LangSmith** | **Optik** |
|---------|------------|-------------|-----------------|----------------|--------------|---------------|-----------|
| **One-liner Init** | `get_client()` | `openlit.init()` | `Traceloop.init()` | `init()` | `Langfuse()` | `Client()` | `opik.configure()` |
| **Decorator** | `@observe()` | `@openlit.trace` | `@workflow/@task` | N/A | `@observe()` | `@traceable()` | `@track()` |
| **Wrapper Functions** | `wrap_openai()` | Auto | Auto | `wrap_openai()` | OpenAI module | `wrap_openai()` | `track_openai()` |
| **Context Manager** | `start_as_current_span()` | `start_trace()` | N/A | `start_span()` | `start_generation()` | N/A | Context-based |
| **Callback Handler** | ⏳ | Auto | Auto | LangChain | LangChain | LangChain | LangChain |
| **Manual API** | `Brokle()` class | N/A | N/A | `RunTree` | `Langfuse()` | `RunTree` | `Opik()` |

### 4.2 Code Examples

#### Brokle - Three Integration Patterns
```python
from brokle import Brokle, observe, get_client
from brokle.wrappers import wrap_openai

# Pattern 1: Wrapper (Recommended)
openai_client = wrap_openai(OpenAI())
response = openai_client.chat.completions.create(...)

# Pattern 2: Decorator
@observe(name="analyze", as_type="generation")
def analyze_data(query: str) -> str:
    return openai_client.chat.completions.create(...)

# Pattern 3: Native SDK (Maximum Control)
with client.start_as_current_generation(name="chat", model="gpt-4") as span:
    response = openai_client.chat.completions.create(...)
    span.set_attribute(Attrs.GEN_AI_USAGE_INPUT_TOKENS, 150)
```

#### OpenLIT - Auto-Magic
```python
import openlit
openlit.init()  # That's it - all providers auto-instrumented

client = OpenAI()
response = client.chat.completions.create(...)  # Auto-traced
```

#### OpenLLMetry - Workflow Hierarchy
```python
from traceloop.sdk.decorators import workflow, task, agent

@workflow("document_processing")
def process_document(doc):
    return summarize(extract_text(doc))

@task("extract")
def extract_text(doc): ...

@task("summarize")
def summarize(text): ...
```

#### Braintrust - Eval-Centric
```python
from braintrust import Eval, wrap_openai

client = wrap_openai(OpenAI())

# Primary use case: Evaluations
Eval("my-project",
    data=lambda: [{"input": "test", "expected": "result"}],
    task=lambda input: client.chat.completions.create(...),
    scores=[exact_match, relevance]
)
```

#### Langfuse - Prompt-First
```python
from langfuse import Langfuse
from langfuse.openai import openai  # Drop-in replacement

langfuse = Langfuse()

# Prompt management
prompt = langfuse.get_prompt("chat-assistant", version=5)

@langfuse.observe()
def my_function(query):
    return openai.chat.completions.create(...)
```

#### LangSmith - Framework-Integrated
```python
from langsmith import traceable, wrap_openai

client = wrap_openai(OpenAI())

@traceable(name="my_function", run_type="chain")
def process_query(query: str):
    return client.chat.completions.create(...)

# Deep LangChain integration via callbacks
from langchain.callbacks import LangChainCallbackHandler
```

#### Optik - Decorator-First
```python
import opik
from opik.integrations.openai import track_openai

opik.configure(api_key="...")
client = track_openai(OpenAI())

@opik.track
def my_function(x):
    opik.set_tags(["production"])
    return client.chat.completions.create(...)
```

### 4.3 Integration Pattern Analysis

| Criteria | **Brokle** | **OpenLIT** | **OpenLLMetry** | **Braintrust** | **Langfuse** | **LangSmith** | **Optik** |
|----------|------------|-------------|-----------------|----------------|--------------|---------------|-----------|
| **Learning Curve** | Medium | Low | Medium | Medium | Medium | Medium | Low |
| **Flexibility** | ✅ High | ❌ Limited | ⚠️ Medium | ⚠️ Medium | ✅ High | ✅ High | ⚠️ Medium |
| **Explicit Control** | ✅ Full | ❌ Minimal | ⚠️ Partial | ⚠️ Partial | ✅ Full | ✅ Full | ⚠️ Partial |
| **Code Changes** | Minimal | None | Decorators | Wrappers | Minimal | Decorators | Decorators |

---

## 5. Provider Support

### 5.1 LLM Providers

| Provider | **Brokle** | **OpenLIT** | **OpenLLMetry** | **Braintrust** | **Langfuse** | **LangSmith** | **Optik** |
|----------|------------|-------------|-----------------|----------------|--------------|---------------|-----------|
| OpenAI | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Anthropic | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Azure OpenAI | ⏳ | ✅ | ✅ | ⚠️ | ✅ | ✅ | ⚠️ |
| Cohere | ⏳ | ✅ | ✅ | ⏳ | Via LangChain | ⏳ | ⏳ |
| Mistral AI | ⏳ | ✅ | ✅ | ⏳ | Via LangChain | ⏳ | ⏳ |
| AWS Bedrock | ⏳ | ✅ | ✅ | ⏳ | Via LangChain | ⏳ | ✅ |
| Google Vertex AI | ⏳ | ✅ | ✅ | ✅ | Via LangChain | ⏳ | ✅ |
| Ollama | ⏳ | ✅ | ✅ | ⏳ | Via LangChain | ⏳ | ⏳ |
| Groq | ⏳ | ✅ | ✅ | ⏳ | Via LangChain | ⏳ | ✅ |
| LiteLLM | ⏳ | ✅ | ✅ | ✅ | ⏳ | ⏳ | ✅ |

### 5.2 Framework Integrations

| Framework | **Brokle** | **OpenLIT** | **OpenLLMetry** | **Braintrust** | **Langfuse** | **LangSmith** | **Optik** |
|-----------|------------|-------------|-----------------|----------------|--------------|---------------|-----------|
| LangChain | ⏳ | ✅ | ✅ | ✅ | ✅ | ✅ Native | ✅ |
| LlamaIndex | ⏳ | ✅ | ✅ | ⏳ | ⏳ | ⏳ | ✅ |
| Haystack | ⏳ | ✅ | ✅ | ⏳ | ⏳ | ⏳ | ✅ |
| CrewAI | ⏳ | ✅ | ⏳ | ⏳ | ⏳ | ⏳ | ✅ |
| DSPy | ⏳ | ✅ | ⏳ | ✅ | ⏳ | ⏳ | ✅ |
| Vercel AI SDK | ⏳ | ⏳ | ⏳ | ✅ | ⏳ | ✅ | ✅ OTEL |
| OpenAI Agents | ⏳ | ✅ | ⏳ | ✅ | ⏳ | ✅ | ⏳ |

### 5.3 Vector Databases

| Vector DB | **Brokle** | **OpenLIT** | **OpenLLMetry** | **Braintrust** | **Langfuse** | **LangSmith** | **Optik** |
|-----------|------------|-------------|-----------------|----------------|--------------|---------------|-----------|
| ChromaDB | ⏳ | ✅ | ✅ | ⏳ | ⏳ | ⏳ | ⏳ |
| Pinecone | ⏳ | ✅ | ✅ | ⏳ | ⏳ | ⏳ | ⏳ |
| Qdrant | ⏳ | ✅ | ✅ | ⏳ | ⏳ | ⏳ | ⏳ |
| Weaviate | ⏳ | ⏳ | ✅ | ⏳ | ⏳ | ⏳ | ⏳ |

### 5.4 Coverage Summary

| Metric | **Brokle** | **OpenLIT** | **OpenLLMetry** | **Braintrust** | **Langfuse** | **LangSmith** | **Optik** |
|--------|------------|-------------|-----------------|----------------|--------------|---------------|-----------|
| **LLM Providers** | 2 | 23+ | 17+ | 12+ | 15+ (via integrations) | 10+ | 15+ |
| **Frameworks** | 0 | 10+ | 8+ | 8+ | 5+ | 5+ | 10+ |
| **Vector DBs** | 0 | 5 | 7 | 0 | 0 | 0 | 0 |
| **Total Direct** | **2** | **48+** | **32+** | **12+** | **15+** | **10+** | **15+** |

---

## 6. Semantic Conventions

### 6.1 Standards Compliance

| Standard | **Brokle** | **OpenLIT** | **OpenLLMetry** | **Braintrust** | **Langfuse** | **LangSmith** | **Optik** |
|----------|------------|-------------|-----------------|----------------|--------------|---------------|-----------|
| OTEL GenAI 1.28+ | ✅ | ✅ | ✅ | ⚠️ Bridge | ✅ | ❌ Custom | ❌ Custom |
| OTEL GenAI 1.38+ | ✅ | ⏳ | ⏳ | ❌ | ✅ | ❌ | ❌ |
| OpenInference | ✅ | ❌ | ⚠️ Partial | ❌ | ❌ | ❌ | ❌ |
| Custom Namespace | `brokle.*` | `openlit.*` | `traceloop.*` | N/A | `langfuse.*` | Custom RunTree | `opik.*` |

### 6.2 Core Attribute Support

| Attribute | **Brokle** | **OpenLIT** | **OpenLLMetry** | **Braintrust** | **Langfuse** | **LangSmith** | **Optik** |
|-----------|------------|-------------|-----------------|----------------|--------------|---------------|-----------|
| `gen_ai.provider.name` | ✅ | ✅ | ✅ | Custom | ✅ | Custom | Custom |
| `gen_ai.request.model` | ✅ | ✅ | ✅ | Custom | ✅ | Custom | Custom |
| `gen_ai.usage.input_tokens` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `gen_ai.usage.output_tokens` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `gen_ai.input.messages` | ✅ JSON | ✅ JSON | ✅ JSON | Custom | ✅ JSON | Custom | Custom |
| `input.value` (OpenInference) | ✅ | ❌ | ⚠️ | ❌ | ❌ | ❌ | ❌ |
| `input.mime_type` | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |

### 6.3 Custom Namespace Comparison

**Brokle** (`brokle.*`):
```python
brokle.span.type           # generation|span|tool|agent|chain|retrieval
brokle.release             # App deployment version
brokle.version             # A/B testing version
brokle.routing.strategy    # Routing decision (unique)
brokle.routing.cache_hit   # Semantic cache hit (unique)
brokle.score.name          # Quality score (unique)
```

**Langfuse** (`langfuse.*`):
```python
langfuse.observation.type           # span|generation|agent|tool|chain
langfuse.observation.model.name     # Model identifier
langfuse.observation.prompt.name    # Prompt template reference
langfuse.trace.user_id              # User tracking
langfuse.trace.session_id           # Session grouping
```

**LangSmith** (Custom RunTree):
```python
# Not OTEL attributes - custom data model
run.run_type      # llm|chain|tool|retriever|embedding|prompt|parser
run.inputs        # Input dict
run.outputs       # Output dict
run.tags          # String tags
run.metadata      # Custom metadata
```

**Optik** (Custom):
```python
span.type         # general|tool|llm|guardrail
span.input        # Captured inputs
span.output       # Captured outputs
span.metadata     # Custom metadata
span.tags         # String tags
```

---

## 7. Configuration

### 7.1 Environment Variables

| Variable | **Brokle** | **OpenLIT** | **Langfuse** | **LangSmith** | **Braintrust** | **Optik** |
|----------|------------|-------------|--------------|---------------|----------------|-----------|
| API Key | `BROKLE_API_KEY` | `OTEL_*_HEADERS` | `LANGFUSE_*_KEY` | `LANGSMITH_API_KEY` | `BRAINTRUST_API_KEY` | `OPIK_API_KEY` |
| Endpoint | `BROKLE_HOST` | `OTEL_*_ENDPOINT` | `LANGFUSE_HOST` | `LANGSMITH_ENDPOINT` | `BRAINTRUST_API_URL` | `OPIK_URL_OVERRIDE` |
| Tracing Toggle | `BROKLE_ENABLED` | Via init() | `LANGFUSE_TRACING_ENABLED` | `LANGSMITH_TRACING` | Via init() | Via configure() |
| Sample Rate | Via init() | ⏳ | `LANGFUSE_SAMPLE_RATE` | `LANGSMITH_SAMPLE_RATE` | ⏳ | ⏳ |

### 7.2 Programmatic Configuration

#### Brokle
```python
client = Brokle(
    api_key="bk_...",
    environment="production",
    release="v2.1.24",
    sample_rate=0.1,
    mask=lambda data: mask_pii(data),
)
```

#### Langfuse
```python
langfuse = Langfuse(
    public_key="pk_...",
    secret_key="sk_...",
    environment="production",
    release="v2.1",
    sample_rate=0.5,
    mask=custom_mask_fn,
)
```

#### LangSmith
```python
from langsmith import Client
client = Client(
    api_key="ls_...",
    # Custom retry, SSL, etc.
)
```

### 7.3 Configuration Feature Matrix

| Feature | **Brokle** | **OpenLIT** | **Langfuse** | **LangSmith** | **Braintrust** | **Optik** |
|---------|------------|-------------|--------------|---------------|----------------|-----------|
| **Sample Rate** | ✅ | ⏳ | ✅ | ✅ | ⏳ | ⏳ |
| **Content Capture Toggle** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Custom Masking** | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ |
| **Tracing Kill Switch** | ✅ | ⏳ | ✅ | ✅ | ⏳ | ✅ |
| **Release/Version Tracking** | ✅ | ⏳ | ✅ | ⏳ | ✅ Git | ⏳ |
| **Batch Control** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

---

## 8. Unique Features

### 8.1 Feature Matrix

| Feature | **Brokle** | **OpenLIT** | **OpenLLMetry** | **Braintrust** | **Langfuse** | **LangSmith** | **Optik** |
|---------|------------|-------------|-----------------|----------------|--------------|---------------|-----------|
| **Intelligent Routing** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Semantic Caching** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Quality Scoring** | ✅ | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ |
| **Backend Cost Calc** | ✅ | ❌ Client | ❌ | ✅ | ✅ | ✅ | ✅ |
| **OpenInference** | ✅ | ❌ | ⚠️ | ❌ | ❌ | ❌ | ❌ |
| **GPU Monitoring** | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Guardrails** | ⏳ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Prompt Management** | ⏳ | ✅ | ✅ | ✅ | ✅ | ✅ Hub | ⏳ |
| **Prompt Optimization** | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| **Dataset Management** | ⏳ | ⏳ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Workflow Hierarchy** | ⚠️ Nesting | ⚠️ | ✅ | ⚠️ | ⚠️ | ⚠️ | ⚠️ |
| **MCP Protocol** | ⏳ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Public Trace Sharing** | ⏳ | ⏳ | ⏳ | ⏳ | ⏳ | ✅ | ⏳ |
| **Git Metadata Capture** | ⏳ | ⏳ | ⏳ | ✅ | ⏳ | ⏳ | ⏳ |
| **pytest Plugin** | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ |

### 8.2 Brokle-Exclusive Features

#### Intelligent Routing
```python
# SDK tracks routing decisions
span.set_attribute(Attrs.BROKLE_ROUTING_STRATEGY, "cost_optimized")
span.set_attribute(Attrs.BROKLE_ROUTING_PROVIDER_SELECTED, "anthropic")
```

#### Semantic Caching
```python
span.set_attribute(Attrs.BROKLE_ROUTING_CACHE_HIT, True)
```

#### OpenInference Compatibility
```python
# Auto-detects format and sets appropriate attributes
input={"query": "weather"}  # → input.value + input.mime_type
```

### 8.3 Competitor-Exclusive Features

#### OpenLIT: GPU Monitoring
```python
openlit.init(collect_gpu_stats=True)
# Captures: gpu.utilization, gpu.memory.used, gpu.temperature
```

#### Braintrust: Eval Framework
```python
Eval("project", data=dataset, task=my_func, scores=[accuracy, relevance])
```

#### Langfuse: Prompt Versioning
```python
prompt = langfuse.get_prompt("chat-assistant", version=5)
```

#### LangSmith: pytest Plugin
```python
@langsmith.test
def test_my_chain():
    result = my_chain("input")
    expect(result).to_contain("expected")
```

#### Optik: Prompt Optimization
```python
from opik_optimizer import EvolutionaryOptimizer
optimizer = EvolutionaryOptimizer(dataset="tiny-test")
optimized = optimizer.optimize(my_prompt)
```

---

## 9. Evaluation Frameworks

### 9.1 Evaluation Comparison

| Aspect | **Brokle** | **Braintrust** | **Langfuse** | **LangSmith** | **Optik** |
|--------|------------|----------------|--------------|---------------|-----------|
| **Eval Framework** | ⏳ | ✅ Comprehensive | ✅ Basic | ✅ Comprehensive | ✅ Good |
| **Built-in Metrics** | ⏳ | ✅ | ⚠️ Custom only | ✅ | ✅ |
| **LLM-as-Judge** | ⏳ | ✅ | ⚠️ Custom | ✅ | ✅ |
| **Dataset Management** | ⏳ | ✅ | ✅ | ✅ | ✅ |
| **Experiment Comparison** | ⏳ | ✅ | ✅ | ✅ | ✅ |
| **Regression Testing** | ⏳ | ✅ Base exp | ⏳ | ✅ | ⏳ |

### 9.2 Braintrust Eval Example
```python
from braintrust import Eval

Eval("my-project",
    data=lambda: [
        {"input": "What is 2+2?", "expected": "4"},
        {"input": "Capital of France?", "expected": "Paris"},
    ],
    task=lambda input: llm.generate(input),
    scores=[
        ExactMatch(),
        LLMClassifier("relevance", template="..."),
    ]
)
```

### 9.3 LangSmith Eval Example
```python
from langsmith import Client
from langsmith.evaluation import evaluate

client = Client()

def my_evaluator(run, example):
    return {"score": 1.0 if run.outputs == example.outputs else 0.0}

evaluate(
    lambda inputs: my_chain(inputs),
    data=client.list_examples(dataset_name="my-dataset"),
    evaluators=[my_evaluator]
)
```

### 9.4 Optik Eval Example
```python
from opik.evaluation import evaluate
from opik.evaluation.metrics import ExactMatch, IsJson

result = evaluate(
    experiment_name="test-run",
    dataset=opik.get_dataset("my-dataset"),
    task=my_function,
    scoring_metrics=[ExactMatch(), IsJson()]
)
```

---

## 10. Strengths & Weaknesses

### 10.1 Brokle SDK

| Strengths | Weaknesses |
|-----------|------------|
| ✅ **Explicit Control** - Three integration patterns | ⏳ **Limited Providers** - Only 2 vs 48+ |
| ✅ **OTEL-Native** - Standard compliance | ⏳ **No GPU Monitoring** |
| ✅ **Backend Cost Calc** - No client-side pricing | ⏳ **No Eval Framework** |
| ✅ **OpenInference** - Multi-platform interop | ⏳ **No Prompt Management** |
| ✅ **Platform Features** - Routing, caching, scoring | ⏳ **No Guardrails** |
| ✅ **Custom Masking** - Privacy protection | |

### 10.2 OpenLIT SDK

| Strengths | Weaknesses |
|-----------|------------|
| ✅ **48+ Providers** - Widest coverage | ❌ **Magic Over Control** - Limited explicit options |
| ✅ **One-liner Setup** - `init()` instruments all | ❌ **Client-Side Costs** - Ships pricing data |
| ✅ **GPU Monitoring** - NVIDIA/AMD | ❌ **Monolithic Bundle** - All instrumentors included |
| ✅ **Guardrails** - PII, injection detection | ❌ **No OpenInference** |
| ✅ **MCP Support** | |

### 10.3 OpenLLMetry SDK

| Strengths | Weaknesses |
|-----------|------------|
| ✅ **Most Modular** - Pick packages you need | ❌ **Complex Dependencies** - Many packages |
| ✅ **Workflow Hierarchy** - `@workflow/@task/@agent` | ❌ **No Cost Tracking** |
| ✅ **Dataset Management** | ❌ **No OpenInference** |
| ✅ **Strong TypeScript** | |

### 10.4 Braintrust SDK

| Strengths | Weaknesses |
|-----------|------------|
| ✅ **Best Eval Framework** - Comprehensive DSL | ❌ **NOT OTEL-Native** - Custom tracing |
| ✅ **Git Integration** - Compliance metadata | ❌ **Limited OTEL Compat** - Bridge only |
| ✅ **Multi-Trial Evals** - Variance analysis | ❌ **Platform Lock-in** |
| ✅ **Background Processing** - Non-blocking | |

### 10.5 Langfuse SDK

| Strengths | Weaknesses |
|-----------|------------|
| ✅ **OTEL-Native** - Standard compliance | ❌ **No Cost Optimization** - Observability only |
| ✅ **Prompt Versioning** - Unique feature | ❌ **No Provider Routing** |
| ✅ **Rich Observation Types** - 10+ types | |
| ✅ **Sampling Support** | |
| ✅ **Media Handling** | |

### 10.6 LangSmith SDK

| Strengths | Weaknesses |
|-----------|------------|
| ✅ **LangChain Native** - Deep integration | ❌ **NOT OTEL-Native** - Custom RunTree |
| ✅ **pytest Plugin** - Test integration | ❌ **Framework Dependency** |
| ✅ **Prompt Hub** - Centralized management | ❌ **Complex Setup** for non-LangChain |
| ✅ **Public Trace Sharing** | |

### 10.7 Optik SDK

| Strengths | Weaknesses |
|-----------|------------|
| ✅ **Prompt Optimization** - Unique feature | ❌ **NOT OTEL-Native** - Custom tracing |
| ✅ **15+ Integrations** - Good coverage | ❌ **Limited OTEL Support** - Vercel AI only |
| ✅ **Non-blocking Design** - Production ready | |
| ✅ **Evaluation Metrics** - Built-in | |
| ✅ **Dynamic Tracing Control** | |

---

## 11. Recommendations

### 11.1 High Priority (Should Implement)

#### 1. Expand Provider Coverage
**What**: Add support for top 10 LLM providers
**Why**: OpenLIT has 48+ providers; this is a competitive gap
**Providers to Add**:
1. Azure OpenAI
2. Cohere
3. AWS Bedrock
4. Google Vertex AI
5. Mistral AI
6. Groq
7. Ollama
8. LiteLLM

#### 2. Add Framework Integrations
**What**: LangChain, LlamaIndex callback handlers
**Why**: Critical for AI pipeline observability
**Approach**: Create callback handlers that emit Brokle spans

#### 3. Add Evaluation Framework
**What**: Basic eval framework with datasets
**Why**: Every major competitor has this
**MVP Features**:
- Dataset management
- Custom evaluators
- Score aggregation

### 11.2 Medium Priority (Nice to Have)

#### 4. Add Workflow Decorators
**What**: `@workflow`, `@task`, `@agent` decorators
**Why**: OpenLLMetry's structured hierarchy is popular

#### 5. Add Prompt Management
**What**: `brokle.get_prompt(key, version)` API
**Why**: Both Langfuse and LangSmith have this

#### 6. Add Vector DB Support
**What**: ChromaDB, Pinecone instrumentation
**Why**: RAG pipelines are increasingly common

### 11.3 Lower Priority (Future Consideration)

#### 7. GPU Monitoring
**What**: NVIDIA/AMD GPU metrics
**Why**: OpenLIT has this; useful for on-prem

#### 8. Guardrails Integration
**What**: PII detection, prompt injection detection
**Why**: Safety/compliance requirements

#### 9. Prompt Optimization (like Optik)
**What**: Automated prompt improvement algorithms
**Why**: Unique differentiator if implemented well

### 11.4 Maintain as Differentiators

| Feature | Status | Competitors Lacking |
|---------|--------|---------------------|
| **Backend Cost Calculation** | ✅ Keep | OpenLIT (client-side) |
| **OpenInference Compatibility** | ✅ Keep | ALL competitors |
| **Three Integration Patterns** | ✅ Keep | Most have 1-2 |
| **Custom Masking Function** | ✅ Keep | Most competitors |
| **Intelligent Routing** | ✅ Keep | ALL competitors |
| **Semantic Caching** | ✅ Keep | ALL competitors |
| **Trace-Level Sampling** | ✅ Keep | Some competitors |

---

## Appendix A: Architecture Decision Matrix

### When to Choose Each SDK

| Use Case | **Recommended SDK** | **Reason** |
|----------|---------------------|------------|
| Need routing + caching + scoring | **Brokle** | Only platform with these features |
| Quick setup, broad provider coverage | **OpenLIT** | One-liner, 48+ providers |
| Complex AI pipelines with workflow hierarchy | **OpenLLMetry** | Best workflow decorators |
| Heavy evaluation workloads | **Braintrust** | Best eval framework |
| Prompt versioning critical | **Langfuse** | Native prompt management |
| LangChain-based application | **LangSmith** | Native LangChain integration |
| Need prompt optimization | **Optik** | Unique optimizer package |
| Export to any OTEL backend | **Brokle, OpenLIT, OpenLLMetry, Langfuse** | OTEL-native |
| GPU monitoring needed | **OpenLIT** | Only SDK with GPU support |

---

## Document History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | December 2025 | Initial analysis (OpenLIT, OpenLLMetry) |
| 2.0.0 | December 2025 | Added Braintrust, Langfuse, LangSmith, Optik |

---

**End of Document**
