# 📊🚀 Brokle - AI Observability & Gateway Platform

**Production-grade observability and routing for LLM apps** – Your first step toward complete AI infrastructure.

*Built for transparency, extensibility, and control — with the flexibility of open source.*

## 🎯 Current Capabilities
- **Advanced Observability** - 40+ AI-specific metrics with real-time insights  
- **Intelligent Gateway** - OpenAI-compatible proxy with multi-provider smart routing
- **Request Tracing** - End-to-end visibility with correlation IDs
- **Cost Analytics** - Real-time cost tracking and optimization insights

## 🗺️ Our Vision

Brokle is starting with **observability and gateway** as the foundation.  
Here's what we're exploring next (no strict timeline):

- 🔄 **Semantic Caching & Advanced Optimization** – Reduce latency and costs
- 🚀 **Model Hosting & Multi-modal APIs** – Expand beyond text LLMs  
- 🌐 **Unified AI Infrastructure Platform** – Bring it all together

Our long-term goal is the unified AI infrastructure platform — starting with what production teams need most: visibility and control.

## 🏗️ Architecture

### Backend (Go Monolith)
- **Single binary** with HTTP + WebSocket support
- **Multi-database** - PostgreSQL + ClickHouse + Redis
- **Real-time features** - WebSocket connections and events
- **Background processing** - Async job workers

### Frontend (Next.js SSR)
- **Server-side rendering** for performance
- **Real-time dashboard** with WebSocket integration
- **Heavy interactions** with complex state management
- **Mobile-responsive** design

## 🚀 Quick Start

### Prerequisites
- Go 1.24+
- Node.js 18+
- PostgreSQL 16+
- ClickHouse 24+
- Redis 7+

### Development Setup

```bash
# Clone the repository
git clone https://github.com/brokle-ai/brokle-platform.git
cd brokle-platform

# Setup development environment
make setup

# Start development servers
make dev
```

This will start:
- Go API server on `http://localhost:8080`
- Next.js dashboard on `http://localhost:3000`
- Databases initialized automatically for local dev

### Production Deployment

```bash
# Build for production
make build-prod

# Deploy with Docker
docker-compose up -d

# Or deploy to Kubernetes
kubectl apply -f deployments/kubernetes/
```

## 📚 Documentation

- [**Architecture Overview**](docs/ARCHITECTURE.md) - System design and data flow
- [**Development Guide**](docs/DEVELOPMENT.md) - Local setup and workflow
- [**API Documentation**](docs/API.md) - REST API and WebSocket events
- [**Deployment Guide**](docs/DEPLOYMENT.md) - Production deployment
- [**Coding Standards**](docs/CODING_STANDARDS.md) - Development patterns

## 🛠️ Development Commands

```bash
# Development
make dev              # Start full stack (Go + Next.js)
make dev-backend      # Go API server only
make dev-frontend     # Next.js dashboard only

# Database Operations
make migrate-up       # Run database migrations
make migrate-down     # Rollback migrations
make seed            # Seed databases with sample data
make db-reset        # Reset all databases

# Build & Test
make build           # Build backend and frontend
make test            # Run all tests
make lint            # Run linters

# Docker
make docker-build    # Build Docker images
make docker-dev      # Start with Docker Compose
```

## 🌟 Key Features

### Advanced Observability
- **Real-time Metrics** - 40+ AI-specific performance indicators
- **Request Tracing** - End-to-end visibility into AI requests
- **Quality Scoring** - Automated response quality assessment
- **Cost Analytics** - Detailed cost breakdown and optimization

### AI Gateway & Routing
- **Intelligent Provider Selection** - ML-powered routing decisions
- **Multi-provider Smart Routing** - Seamless switching between providers
- **Health Monitoring** - Automatic failover and recovery
- **OpenAI Compatibility** - Drop-in replacement for existing code

### Production Scale
- **High Availability** - Multi-region deployment support
- **Auto-scaling** - Handle millions of requests per minute
- **Security** - Enterprise-grade authentication and authorization
- **Enterprise-ready foundations** - HA, scaling, security with compliance readiness in roadmap

### Why Brokle
- **Built for transparency, extensibility, and control** — with the flexibility of open source
- **Production-ready architecture** - Scalable monolith with microservices patterns
- **Complete visibility** - Comprehensive monitoring from day one  
- **Developer-first** - OpenAI-compatible with extensive customization

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](docs/CONTRIBUTING.md) for details.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Links

- **Website**: [https://brokle.com](https://brokle.com)
- **Documentation**: [https://docs.brokle.com](https://docs.brokle.com)
- **Community**: [Discord Server](https://discord.gg/brokle)
- **Twitter**: [@BrokleAI](https://twitter.com/BrokleAI)

---

**Built with ❤️ by the Brokle team. Making AI infrastructure simple and powerful.**
