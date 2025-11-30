# Brokle Platform Gemini Agent Context

This document provides comprehensive context for the Gemini AI agent to understand and interact with the Brokle Platform codebase.

## 1. Project Overview

The Brokle Platform is The Open-Source AI Control Plane - See Everything. Control Everything. It acts as an intelligent proxy between applications and Large Language Model (LLM) providers (e.g., OpenAI, Anthropic, Google). The platform provides observability, routing, and governance for AI in production — open source and built for scale.

### Key Differentiators:

*   **Complete Platform:** Offers an integrated solution for gateway, observability, caching, and optimization, unlike point solutions.
*   **Advanced Observability:** Provides over 40 AI-specific metrics, real-time quality scoring, and predictive analytics.
*   **Cost Optimization:** Achieves 30-50% cost reduction through intelligent routing and semantic caching.
*   **Superior Architecture:** Built on a scalable microservices architecture, contrasting with monolithic competitors.
*   **Competitive Pricing:** The Pro tier is priced 70% lower than direct competitors like Portkey.

### Strategic Focus:

The platform leads with advanced AI observability to attract users and then expands its value through its comprehensive gateway and optimization features. The long-term vision includes model hosting, multi-modal APIs, and a full AI DevOps platform.

## 2. Architecture

The platform is built on a microservices architecture with a strong emphasis on scalability, maintainability, and independent deployment.

### Core Technologies:

*   **Backend:** Go for high-performance services and Python (FastAPI) for ML/AI services.
*   **API Gateway:** Traefik with custom Lua plugins for intelligent routing and security.
*   **Inter-Service Communication:** gRPC for low-latency internal communication.
*   **Databases:** A "database-per-service" model using PostgreSQL, with ClickHouse for time-series analytics and Redis for caching.
*   **Messaging:** Kafka for asynchronous, event-driven communication between services.
*   **Containerization:** Docker and Docker Compose for development and deployment.

### Service Breakdown:

The platform consists of 11+ specialized services:

*   **Go Services:** `auth-service`, `routing-service`, `billing-service`, `analytics-service`, `cache-service`, `cost-tracking-service`, `notification-service`, `telemetry-service`, `config-service`.
*   **Python Services:** `ml-service` (for routing optimization) and `evaluation-service` (for response quality assessment).

### Data and Communication Flow:

*   **Synchronous:** gRPC is used for direct, real-time communication between services (e.g., `routing-service` calling `ml-service`).
*   **Asynchronous:** Kafka is used for event-driven data synchronization (e.g., publishing a `request.completed` event to be consumed by the `analytics-service` and `billing-service`).
*   **Authentication:** A JWT-based authentication system is used for service-level authentication, which is more performant and scalable than the previous ForwardAuth model.

## 3. Development Standards

The project enforces a strict set of development standards to ensure code quality and consistency.

### Coding Standards:

*   **Go:** Follows standard Go formatting (`gofmt`, `goimports`), `golangci-lint` for linting, and a consistent project structure (Transport Layer Pattern).
*   **Python:** Uses `black` for formatting and `ruff` for linting.
*   **Error Handling:** Utilizes a shared error library for standardized error types.
*   **Logging:** Employs structured logging with correlation IDs for traceability.

### API Design:

*   **REST:** Follows resource-based, hierarchical URL structures (e.g., `/api/v1/organizations/{orgId}/projects/{projectId}`). A standardized success and error response format is used across all services.
*   **gRPC:** Uses PascalCase for services and methods, and snake_case for field names.
*   **Versioning:** URL-based versioning for REST APIs (e.g., `/api/v1`).

### Database Standards:

*   **Schema:** Plural, lowercase, snake_case table names. Standard columns like `id`, `created_at`, `updated_at`, and `deleted_at` (for soft deletes) are used.
*   **Migrations:** `golang-migrate` for Go services and `Alembic` for Python services.

## 4. Business Context

### Business Model:

The platform operates on a tiered SaaS subscription model.

*   **Free Tier:** For individual developers and small projects (10,000 requests/month).
*   **Pro Tier ($29/month):** For growing teams and production applications (100,000 requests/month).
*   **Business Tier ($99/month):** For high-volume AI operations (1,000,000 requests/month).
*   **Enterprise Tier (Custom):** For large organizations with custom needs.

### Go-to-Market Strategy:

The strategy is to lead with a generous free tier to build a developer community and then convert users to paid tiers as their needs grow. The primary competitive advantage is offering a more comprehensive platform at a significantly lower price point.

## 5. Key Commands

The `Makefile` provides a set of commands for common development tasks:

*   `make setup`: Initializes the project (starts databases, runs migrations, generates gRPC code).
*   `make start`: Starts all services in development mode.
*   `make test`: Runs tests for all services.
*   `make proto`: Generates gRPC code from `.proto` files.
*   `make migrate-go`: Runs migrations for all Go services.
*   `make migrate-python`: Runs migrations for all Python services.
*   `make logs-service SERVICE=<service-name>`: Shows logs for a specific service.
*   `make health`: Checks the health of all running services.

## 6. AI-Specific Features

The platform includes several advanced AI intelligence features:

*   **Feedback Loop System:** Continuously learns from user interactions and response quality to improve routing decisions.
*   **A/B Testing Framework:** Allows for statistically rigorous experimentation with different prompts, models, and providers.
*   **Auto-Prompt Optimization:** Uses ML techniques (genetic algorithms, reinforcement learning) to automatically improve the effectiveness of prompts.

