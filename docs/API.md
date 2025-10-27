# 🌐 API Documentation

## Overview

The Brokle API provides comprehensive access to AI infrastructure management, observability, and analytics. It features both REST API endpoints and real-time WebSocket connections for live updates.

## Base URLs

- **Development**: `http://localhost:8080/api`
- **Production**: `https://api.brokle.com/api`
- **WebSocket**: `ws://localhost:8080/ws` (dev) / `wss://api.brokle.com/ws` (prod)

## Authentication

The Brokle API implements a **dual route architecture** with separate authentication mechanisms for SDK and Dashboard access.

### Dual Route Architecture

#### SDK Routes (`/v1/*`) - API Key Authentication
**Authentication**: Industry-standard API keys (`bk_{40_char_random}`)
**Rate Limiting**: API key-based rate limiting
**Target Users**: SDK integration, programmatic access

**Endpoints:**
- `POST /v1/chat/completions` - OpenAI-compatible chat completions
- `POST /v1/completions` - OpenAI-compatible text completions
- `POST /v1/embeddings` - OpenAI-compatible embeddings
- `GET /v1/models` - Available AI models
- `POST /v1/ingest/batch` - Unified telemetry batch processing
- `POST /v1/route` - AI routing decisions

**Example:**
```bash
curl -H "X-API-Key: bk_..." https://api.brokle.com/v1/models
# Alternative: Authorization: Bearer bk_...
```

#### Dashboard Routes (`/api/v1/*`) - JWT Authentication
**Authentication**: Bearer JWT tokens with session management
**Rate Limiting**: IP-based and user-based rate limiting
**Target Users**: Web dashboard, administrative access

**Endpoints:**
- `/api/v1/auth/*` - Authentication & session management
- `/api/v1/users/*` - User profile management
- `/api/v1/organizations/*` - Organization management with RBAC
- `/api/v1/projects/*` - Project and API key management
- `/api/v1/analytics/*` - Metrics & reporting
- `/api/v1/billing/*` - Usage & billing management

**Example:**
```bash
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." https://api.brokle.com/api/v1/users/me
```

### API Key Format

Brokle uses industry-standard API keys (following GitHub/Stripe/OpenAI patterns):

```
bk_{40_character_random_secret}
```

**Features:**
- **Prefix**: `bk_` (Brokle identifier)
- **Secret**: 40 characters of cryptographically secure random data
- **Security**: SHA-256 hashed storage with O(1) validation
- **Preview**: `bk_AbCd...yym0` (first 4 + last 4 chars for display)

**Example:**
```bash
curl -H "X-API-Key: bk_AbCdEfGhIjKlMnOpQrStUvWxYz0123456789AbCd" \
  https://api.brokle.com/v1/models
```

### JWT Token Authentication

Dashboard routes use Bearer JWT tokens:

```bash
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  https://api.brokle.com/api/v1/users/me
```

### Authentication Flow

1. **Login** - Get JWT token with credentials (`POST /api/v1/auth/login`)
2. **API Key** - Generate project-scoped API keys for SDK access
3. **Session** - Maintain session with refresh tokens

📖 **See Also:** [PATTERNS.md](development/PATTERNS.md) for detailed authentication patterns

## Standard Response Format

All API endpoints return responses in this format:

```json
{
  "success": true,
  "data": {
    // Response data here
  },
  "error": null,
  "meta": {
    "timestamp": "2024-01-01T12:00:00Z",
    "request_id": "req_123abc"
  }
}
```

### Error Response Format

```json
{
  "success": false,
  "data": null,
  "error": {
    "type": "VALIDATION_ERROR",
    "code": "INVALID_INPUT",
    "message": "Email is required",
    "details": {
      "field": "email",
      "value": ""
    }
  },
  "meta": {
    "timestamp": "2024-01-01T12:00:00Z",
    "request_id": "req_123abc"
  }
}
```

## Authentication Endpoints

### POST /auth/login

Authenticate user with email and password.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "ulid",
      "email": "user@example.com",
      "name": "John Doe"
    },
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "rt_123abc",
    "expires_in": 3600
  }
}
```

### POST /auth/register

Register a new user account.

**Request:**
```json
{
  "email": "user@example.com",
  "name": "John Doe",
  "password": "password123",
  "organization_name": "Acme Corp"
}
```

### POST /auth/refresh

Refresh an expired JWT token.

**Request:**
```json
{
  "refresh_token": "rt_123abc"
}
```

### POST /auth/logout

Logout and invalidate session.

## User Management

### GET /users/me

Get current user profile.

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "ulid",
    "email": "user@example.com",
    "name": "John Doe",
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

### PATCH /users/me

Update current user profile.

**Request:**
```json
{
  "name": "John Smith",
  "avatar_url": "https://example.com/avatar.jpg"
}
```

## Organization Management

### GET /organizations

List user's organizations.

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "ulid",
      "name": "Acme Corp",
      "role": "owner",
      "created_at": "2024-01-01T12:00:00Z"
    }
  ]
}
```

### POST /organizations

Create a new organization.

**Request:**
```json
{
  "name": "New Corp",
  "description": "My new organization"
}
```

### GET /organizations/{id}

Get organization details.

### PATCH /organizations/{id}

Update organization details.

## Project Management

### GET /organizations/{org_id}/projects

List projects in an organization.

**Query Parameters:**
- `page` - Page number (default: 1)
- `limit` - Items per page (default: 20, max: 100)
- `search` - Search query

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "ulid",
      "name": "Production API",
      "description": "Main production environment",
      "organization_id": "ulid",
      "created_at": "2024-01-01T12:00:00Z"
    }
  ],
  "meta": {
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 45,
      "total_page": 3
    }
  }
}
```

### POST /organizations/{org_id}/projects

Create a new project.

### GET /projects/{id}

Get project details.

### PATCH /projects/{id}

Update project details.

## API Key Management

### GET /api-keys

List user's API keys.

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "ulid",
      "name": "Production API Key",
      "key_prefix": "bk_live_123...",
      "scopes": ["read:metrics", "write:routing"],
      "last_used_at": "2024-01-01T11:30:00Z",
      "created_at": "2024-01-01T10:00:00Z"
    }
  ]
}
```

### POST /api-keys

Create a new API key.

**Request:**
```json
{
  "name": "New API Key",
  "scopes": ["read:metrics", "write:routing"],
  "expires_at": "2025-01-01T00:00:00Z"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "ulid",
    "name": "New API Key",
    "key": "bk_live_1234567890abcdef...",
    "key_prefix": "bk_live_123...",
    "scopes": ["read:metrics", "write:routing"],
    "expires_at": "2025-01-01T00:00:00Z"
  }
}
```

### DELETE /api-keys/{id}

Revoke an API key.

## AI Routing & Providers

### GET /routing/providers

List available AI providers.

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "openai",
      "name": "OpenAI",
      "status": "active",
      "models": ["gpt-4", "gpt-3.5-turbo"],
      "health_score": 0.98,
      "avg_latency_ms": 1200,
      "success_rate": 0.995
    }
  ]
}
```

### POST /routing/providers

Add a new AI provider configuration.

### GET /routing/providers/{id}/health

Get provider health status.

### POST /routing/decisions

Get routing recommendation for a request.

**Request:**
```json
{
  "model": "gpt-4",
  "prompt_length": 1500,
  "max_tokens": 500,
  "priority": "balanced"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "recommended_provider": "openai",
    "confidence_score": 0.92,
    "estimated_cost": 0.045,
    "estimated_latency_ms": 1200,
    "alternatives": [
      {
        "provider": "anthropic",
        "confidence_score": 0.88,
        "estimated_cost": 0.052
      }
    ]
  }
}
```

## OpenAI-Compatible Endpoints

### POST /v1/chat/completions

OpenAI-compatible chat completions endpoint.

**Request:**
```json
{
  "model": "gpt-4",
  "messages": [
    {
      "role": "user",
      "content": "Hello, world!"
    }
  ],
  "temperature": 0.7,
  "max_tokens": 100
}
```

**Response:** (OpenAI-compatible format with Brokle extensions)
```json
{
  "id": "chatcmpl-123",
  "object": "chat.completion",
  "created": 1677652288,
  "model": "gpt-4",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Hello! How can I help you today?"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 9,
    "completion_tokens": 12,
    "total_tokens": 21
  },
  "brokle": {
    "provider": "openai",
    "routing_decision_id": "rd_123",
    "cost": 0.00042,
    "latency_ms": 1150,
    "quality_score": 0.94
  }
}
```

### POST /v1/completions

OpenAI-compatible text completions endpoint.

### POST /v1/embeddings

OpenAI-compatible embeddings endpoint.

### GET /v1/models

List available models across all providers.

## Analytics & Observability

### GET /analytics/metrics

Get platform metrics.

**Query Parameters:**
- `start_time` - ISO timestamp (required)
- `end_time` - ISO timestamp (required) 
- `granularity` - minute, hour, day (default: hour)
- `metrics` - Comma-separated metric names
- `filters[provider]` - Filter by provider
- `filters[model]` - Filter by model

**Response:**
```json
{
  "success": true,
  "data": {
    "metrics": [
      {
        "name": "requests_total",
        "values": [
          {
            "timestamp": "2024-01-01T12:00:00Z",
            "value": 1250
          }
        ]
      },
      {
        "name": "avg_latency_ms",
        "values": [
          {
            "timestamp": "2024-01-01T12:00:00Z", 
            "value": 1180
          }
        ]
      }
    ]
  }
}
```

### GET /analytics/events

Get platform events.

**Query Parameters:**
- `start_time` - ISO timestamp
- `end_time` - ISO timestamp
- `event_type` - Filter by event type
- `page` - Page number
- `limit` - Items per page

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "ulid",
      "type": "routing.decision",
      "timestamp": "2024-01-01T12:00:00Z",
      "data": {
        "provider": "openai",
        "model": "gpt-4",
        "confidence": 0.92
      }
    }
  ]
}
```

### GET /analytics/traces/{trace_id}

Get distributed trace details.

### GET /analytics/dashboard

Get dashboard summary data.

**Response:**
```json
{
  "success": true,
  "data": {
    "summary": {
      "total_requests_today": 15420,
      "total_cost_today": 45.23,
      "avg_latency_ms": 1150,
      "success_rate": 0.998
    },
    "providers": [
      {
        "id": "openai",
        "requests": 8500,
        "cost": 28.50,
        "avg_latency": 1200,
        "success_rate": 0.995
      }
    ],
    "top_models": [
      {
        "model": "gpt-4",
        "requests": 6200,
        "cost": 35.20
      }
    ]
  }
}
```

## Billing & Usage

### GET /billing/usage

Get usage statistics.

**Query Parameters:**
- `period` - current, last_30_days, last_90_days
- `granularity` - day, week, month

**Response:**
```json
{
  "success": true,
  "data": {
    "current_period": {
      "start_date": "2024-01-01",
      "end_date": "2024-01-31", 
      "total_requests": 45000,
      "total_cost": 127.50,
      "breakdown": {
        "by_provider": {
          "openai": { "requests": 30000, "cost": 85.20 },
          "anthropic": { "requests": 15000, "cost": 42.30 }
        },
        "by_model": {
          "gpt-4": { "requests": 20000, "cost": 75.50 },
          "gpt-3.5-turbo": { "requests": 25000, "cost": 52.00 }
        }
      }
    }
  }
}
```

### GET /billing/invoices

Get billing invoices.

### GET /billing/subscriptions

Get subscription details.

## Configuration Management

### GET /config

Get application configuration.

**Response:**
```json
{
  "success": true,
  "data": {
    "features": {
      "semantic_caching": true,
      "real_time_metrics": true,
      "custom_models": false
    },
    "limits": {
      "max_requests_per_minute": 1000,
      "max_prompt_length": 32000
    },
    "providers": {
      "openai": {
        "enabled": true,
        "models": ["gpt-4", "gpt-3.5-turbo"]
      }
    }
  }
}
```

### PATCH /config

Update configuration settings.

## WebSocket Events

### Connection

Connect to WebSocket for real-time updates:

```javascript
const ws = new WebSocket('ws://localhost:8080/ws')

ws.onopen = () => {
  // Send authentication
  ws.send(JSON.stringify({
    type: 'auth',
    token: 'your-jwt-token'
  }))
  
  // Subscribe to events
  ws.send(JSON.stringify({
    type: 'subscribe',
    events: ['metrics.updated', 'routing.decision']
  }))
}
```

### Event Types

#### metrics.updated
```json
{
  "type": "metrics.updated",
  "timestamp": "2024-01-01T12:00:00Z",
  "data": {
    "metric": "requests_per_minute",
    "value": 850,
    "change": "+12%"
  }
}
```

#### routing.decision
```json
{
  "type": "routing.decision", 
  "timestamp": "2024-01-01T12:00:00Z",
  "data": {
    "request_id": "req_123",
    "provider": "openai",
    "model": "gpt-4",
    "confidence": 0.92,
    "latency_ms": 1150
  }
}
```

#### usage.threshold
```json
{
  "type": "usage.threshold",
  "timestamp": "2024-01-01T12:00:00Z", 
  "data": {
    "threshold": "monthly_limit",
    "current": 8500,
    "limit": 10000,
    "percentage": 85
  }
}
```

#### system.alert
```json
{
  "type": "system.alert",
  "timestamp": "2024-01-01T12:00:00Z",
  "data": {
    "level": "warning",
    "message": "Provider latency above threshold",
    "provider": "anthropic",
    "metric": "avg_latency_ms",
    "value": 3500,
    "threshold": 3000
  }
}
```

### Subscription Management

```javascript
// Subscribe to specific events
ws.send(JSON.stringify({
  type: 'subscribe',
  events: ['metrics.updated']
}))

// Unsubscribe from events
ws.send(JSON.stringify({
  type: 'unsubscribe', 
  events: ['routing.decision']
}))

// Subscribe with filters
ws.send(JSON.stringify({
  type: 'subscribe',
  events: ['metrics.updated'],
  filters: {
    provider: 'openai',
    metric: 'requests_per_minute'
  }
}))
```

## Rate Limiting

All endpoints are subject to rate limiting:

- **Free Tier**: 100 requests/minute
- **Pro Tier**: 1,000 requests/minute  
- **Business Tier**: 10,000 requests/minute
- **Enterprise**: Custom limits

Rate limit headers:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 995
X-RateLimit-Reset: 1640995200
```

## Error Codes

### HTTP Status Codes

- `200` - Success
- `201` - Created
- `204` - No Content
- `400` - Bad Request
- `401` - Unauthorized  
- `403` - Forbidden
- `404` - Not Found
- `409` - Conflict
- `429` - Too Many Requests
- `500` - Internal Server Error

### Custom Error Types

- `VALIDATION_ERROR` - Invalid input data
- `NOT_FOUND` - Resource not found
- `UNAUTHORIZED` - Authentication required
- `FORBIDDEN` - Insufficient permissions
- `RATE_LIMITED` - Rate limit exceeded
- `PROVIDER_ERROR` - AI provider error
- `QUOTA_EXCEEDED` - Usage quota exceeded
- `INTERNAL_ERROR` - Server error

## SDKs and Libraries

### Official SDKs

- **JavaScript/TypeScript**: `npm install @brokle/sdk-js`
- **Python**: `pip install brokle-python`
- **Go**: `go get github.com/brokle-ai/brokle-go`

### Community SDKs

- **PHP**: `composer require brokle/php-sdk`
- **Ruby**: `gem install brokle`
- **Java**: Maven/Gradle packages available

### OpenAI SDK Compatibility

Brokle is compatible with existing OpenAI SDKs:

```javascript
// JavaScript - just change the base URL
const openai = new OpenAI({
  apiKey: 'your-brokle-api-key',
  baseURL: 'https://api.brokle.com/v1'
})
```

```python
# Python
import openai

openai.api_base = "https://api.brokle.com/v1"
openai.api_key = "your-brokle-api-key"
```

## Webhooks

### Configuration

Configure webhooks to receive events:

```json
{
  "url": "https://your-app.com/webhooks/brokle",
  "events": ["usage.threshold", "system.alert"],
  "secret": "your-webhook-secret"
}
```

### Event Delivery

Webhook payload format:

```json
{
  "id": "evt_123abc",
  "type": "usage.threshold",
  "timestamp": "2024-01-01T12:00:00Z",
  "data": {
    // Event-specific data
  },
  "organization_id": "ulid",
  "project_id": "ulid"
}
```

### Verification

Verify webhook signatures using HMAC-SHA256:

```javascript
const crypto = require('crypto')

function verifyWebhook(payload, signature, secret) {
  const expectedSignature = crypto
    .createHmac('sha256', secret)
    .update(payload)
    .digest('hex')
  
  return signature === `sha256=${expectedSignature}`
}
```

---

This API documentation provides comprehensive coverage of all Brokle platform endpoints, real-time features, and integration patterns. For additional examples and SDKs, visit our [developer portal](https://developers.brokle.com).