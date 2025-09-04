# Brokle Enterprise Edition Documentation

## Table of Contents
- [Overview](#overview)
- [Open-Core Business Model](#open-core-business-model)
- [License Tiers](#license-tiers)
- [Enterprise Features](#enterprise-features)
- [Architecture](#architecture)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [API Integration](#api-integration)
- [Deployment](#deployment)
- [Support](#support)

## Overview

Brokle Enterprise Edition extends the open-source Brokle AI Infrastructure Platform with advanced features for production-scale AI operations. Built on an open-core model, Brokle EE provides enterprise-grade security, compliance, analytics, and support while maintaining compatibility with the OSS version.

### Key Value Propositions

- **Complete AI Infrastructure**: Unified platform for gateway, observability, caching, and optimization
- **Enterprise Security**: SSO integration, advanced RBAC, compliance controls
- **Advanced Analytics**: Predictive insights, custom dashboards, ML-powered optimization
- **Production Scale**: Unlimited requests, multi-organization support, dedicated infrastructure
- **Enterprise Support**: SLA guarantees, dedicated success managers, priority support

## Open-Core Business Model

### OSS vs Enterprise Split

Brokle follows a sustainable open-core model that provides substantial value in both versions:

#### Open Source Features (Forever Free)
- **Core AI Gateway**: OpenAI-compatible proxy, provider routing, rate limiting
- **Basic Observability**: Request/response logging, basic metrics, error tracking
- **Semantic Caching**: Vector-based response caching for cost optimization
- **Multi-tenancy**: Organization → Project → Environment hierarchy
- **REST API**: Complete programmatic access to core features
- **Basic Analytics**: Cost tracking, usage metrics, basic dashboards
- **Community Support**: GitHub issues, community forums

#### Enterprise Features (License Required)
- **Advanced Security**: SSO (SAML/OIDC), custom RBAC, audit logging
- **Compliance**: SOC2/HIPAA/GDPR controls, data retention policies, PII anonymization
- **Predictive Analytics**: ML-powered insights, anomaly detection, cost forecasting
- **Custom Dashboards**: Advanced visualization, custom metrics, executive reporting
- **Advanced Integrations**: Enterprise connectors, webhook automation, workflow orchestration
- **Premium Support**: SLA guarantees, dedicated account management, priority response
- **On-Premise Deployment**: Air-gapped installations, custom compliance requirements

### Business Model Strategy

The open-core model ensures:

1. **Strong OSS Foundation**: Core platform remains fully functional and actively developed
2. **Clear Value Differentiation**: Enterprise features solve real business problems at scale
3. **Smooth Upgrade Path**: No migration complexity when upgrading from OSS to Enterprise
4. **Sustainable Development**: Enterprise revenue funds continued OSS development

## License Tiers

### Free Tier (Open Source)
- **Cost**: $0/month
- **Requests**: 10,000/month
- **Users**: 5 users
- **Projects**: 2 projects
- **Features**: All OSS features
- **Support**: Community support
- **Data Retention**: 7 days

### Pro Tier
- **Cost**: $29/month
- **Requests**: 100,000/month
- **Users**: 10 users
- **Projects**: 10 projects
- **Features**: Advanced RBAC
- **Support**: Email support (48h response)
- **Data Retention**: 30 days

### Business Tier
- **Cost**: $99/month
- **Requests**: 1,000,000/month
- **Users**: 50 users
- **Projects**: 100 projects
- **Features**: SSO, Compliance, Predictive Analytics, Custom Dashboards
- **Support**: Priority support (24h response)
- **Data Retention**: 180 days

### Enterprise Tier
- **Cost**: Custom pricing
- **Requests**: Unlimited
- **Users**: Unlimited
- **Projects**: Unlimited
- **Features**: All features + On-premise deployment
- **Support**: Dedicated success manager (4h response, 24/7 on-call)
- **Data Retention**: Custom (up to 7 years for compliance)

## Enterprise Features

### 1. Advanced Security & Authentication

#### Single Sign-On (SSO)
- **SAML 2.0**: Enterprise identity provider integration
- **OIDC/OAuth2**: Modern authentication protocols
- **Automatic Provisioning**: User accounts created on first login
- **Role Mapping**: Map IdP groups to Brokle roles automatically
- **Session Management**: Configurable session timeouts, SSO logout

#### Advanced Role-Based Access Control (RBAC)
- **Custom Roles**: Define roles beyond the basic owner/admin/developer/viewer
- **Granular Permissions**: Fine-grained control over API access, resource management
- **Hierarchical Scopes**: Organization → Project → Environment permission inheritance
- **Dynamic Permissions**: Context-aware permissions based on resource ownership
- **Audit Trails**: Complete access logging for compliance requirements

### 2. Compliance & Data Governance

#### Compliance Standards
- **SOC2 Type II**: Security controls for availability, confidentiality, integrity
- **HIPAA**: Health information privacy and security controls
- **GDPR**: Data protection and privacy controls for EU residents
- **Custom Frameworks**: Configurable compliance rules and validation

#### Data Management
- **PII Anonymization**: Automatic detection and anonymization of personal data
- **Data Retention**: Configurable retention policies (7 days to 7 years)
- **Audit Logging**: Immutable audit trails for all system interactions
- **Data Export**: Compliance-ready data exports in multiple formats
- **Right to Deletion**: GDPR-compliant data deletion workflows

### 3. Advanced Analytics & Intelligence

#### Predictive Insights
- **Cost Forecasting**: ML-powered monthly and quarterly cost predictions
- **Usage Trend Analysis**: Detect patterns in API usage and performance
- **Anomaly Detection**: Automatic detection of unusual patterns or security threats
- **Capacity Planning**: Recommendations for scaling infrastructure
- **Provider Optimization**: AI-powered suggestions for optimal provider routing

#### Custom Dashboards
- **Drag-and-Drop Builder**: Create custom visualizations without code
- **Executive Reporting**: High-level dashboards for business stakeholders
- **Real-time Alerts**: Custom alerts based on business metrics
- **Export Capabilities**: Scheduled reports via email, Slack, webhook
- **Embedded Analytics**: White-label dashboards for customer-facing applications

#### ML-Powered Optimization
- **Intelligent Routing**: ML models optimize provider selection for cost and performance
- **Quality Scoring**: Automatic evaluation of AI response quality
- **Cost Optimization**: Real-time recommendations to reduce AI spending by 30-50%
- **Performance Tuning**: Automatic optimization of caching and routing parameters

### 4. Enterprise Support & Services

#### Support Tiers
- **Standard**: Business hours support, 48-hour response SLA
- **Priority**: 24/7 support, 24-hour response SLA, phone support
- **Dedicated**: Dedicated success manager, 4-hour response SLA, 24/7 on-call

#### Professional Services
- **Migration Assistance**: Help migrating from existing AI infrastructure
- **Custom Integration**: Built custom connectors and integrations
- **Training & Onboarding**: Team training on platform best practices
- **Architecture Review**: Expert review of AI infrastructure architecture

### 5. Enterprise Deployment Options

#### Cloud Deployment
- **Multi-Region**: Deploy across multiple AWS/GCP/Azure regions
- **High Availability**: Auto-scaling, load balancing, failover
- **Managed Service**: Fully managed infrastructure with SLA guarantees
- **Security**: VPC deployment, private networking, encryption at rest/transit

#### On-Premise Deployment
- **Air-Gapped**: Completely isolated deployment for maximum security
- **Kubernetes**: Deploy on existing Kubernetes infrastructure
- **Docker Compose**: Simplified deployment for smaller environments
- **Custom Infrastructure**: Support for unique deployment requirements

## Architecture

### Build System Architecture

Brokle uses Go build tags to cleanly separate OSS and Enterprise features:

```go
// OSS build (default)
go build -o brokle ./cmd/server

// Enterprise build  
go build -tags="enterprise" -o brokle-enterprise ./cmd/server
```

### Code Organization

```
internal/
├── config/
│   ├── config.go          # Core configuration
│   ├── ee.go              # Enterprise config (build tag: enterprise)
│   ├── ee_stub.go         # Stub config (build tag: !enterprise)
│   └── license.go         # License management wrapper
├── ee/                    # Enterprise features
│   ├── sso/               # Single Sign-On
│   ├── rbac/              # Advanced RBAC
│   ├── compliance/        # Compliance features
│   └── analytics/         # Advanced analytics
├── errors/
│   └── enterprise.go      # Professional error responses
├── middleware/
│   └── enterprise.go      # Feature gating middleware
└── services/
    └── license_service.go # License validation service
```

### Feature Gating

Enterprise features are gated using middleware that checks license entitlements:

```go
// Require enterprise feature
router.Use(middleware.EnterpriseFeature("advanced_rbac", licenseService, logger))

// Require enterprise license
router.Use(middleware.RequireEnterpriseLicense(licenseService, logger))

// Check usage limits
router.Use(middleware.CheckUsageLimit("requests", licenseService, logger))
```

### Error Handling

Professional error responses include:
- Standard error codes and HTTP status codes
- Upgrade paths and pricing information
- UTM tracking for conversion optimization
- Support contact information

```json
{
  "error": {
    "code": "FEATURE_NOT_AVAILABLE",
    "message": "Advanced Role-Based Access Control requires Business tier or higher. You're currently on Free tier.",
    "feature": "advanced_rbac",
    "current_tier": "free",
    "required_tier": "business",
    "actions": [
      {
        "type": "upgrade",
        "label": "Upgrade to Business",
        "url": "https://brokle.ai/pricing?tier=business&utm_source=api&utm_campaign=feature_upgrade&utm_content=advanced_rbac",
        "primary": true
      }
    ]
  }
}
```

## Getting Started

### 1. License Activation

#### Online License Activation
```bash
# Set environment variables
export BROKLE_ENTERPRISE_LICENSE_KEY="your-license-key"
export BROKLE_ENTERPRISE_LICENSE_TYPE="business"

# Start enterprise version
./brokle-enterprise
```

#### Offline License Activation (Air-gapped)
```bash
# Enable offline mode
export BROKLE_ENTERPRISE_LICENSE_OFFLINE_MODE="true"
export BROKLE_ENTERPRISE_LICENSE_KEY="offline-license-jwt"
export BROKLE_ENTERPRISE_LICENSE_VALID_UNTIL="2025-12-31T23:59:59Z"
```

### 2. Feature Configuration

```yaml
# config.yaml
enterprise:
  license:
    type: "business"
    features:
      - "advanced_rbac"
      - "sso_integration"
      - "custom_compliance"
      - "predictive_insights"
  
  sso:
    enabled: true
    provider: "saml"
    metadata_url: "https://your-idp.com/metadata"
    entity_id: "brokle-ai-platform"
  
  rbac:
    enabled: true
    custom_roles:
      - name: "ai_architect"
        permissions: ["models:deploy", "analytics:advanced"]
        scopes: ["org", "project"]
  
  compliance:
    enabled: true
    audit_retention: "2160h"  # 90 days
    pii_anonymization: true
    soc2_compliance: true
  
  analytics:
    enabled: true
    predictive_insights: true
    custom_dashboards: true
    ml_models: true
```

### 3. User Management

#### SSO User Provisioning
Users are automatically created when they first log in via SSO. Role mapping is configured in the SSO provider:

```yaml
enterprise:
  sso:
    attributes:
      role: "http://schemas.microsoft.com/ws/2008/06/identity/claims/role"
      email: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
      name: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name"
```

#### Custom Role Assignment
```bash
# Create custom role
curl -X POST /api/v1/rbac/roles \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "ml_engineer",
    "permissions": ["models:read", "models:deploy", "analytics:read"],
    "scopes": ["project"]
  }'

# Assign role to user
curl -X POST /api/v1/rbac/users/user123/roles \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"role_id": "ml_engineer", "scope": "project:ai-chatbot"}'
```

## Configuration

### Environment Variables

```bash
# License Configuration
BROKLE_ENTERPRISE_LICENSE_KEY="your-enterprise-license-key"
BROKLE_ENTERPRISE_LICENSE_TYPE="business"
BROKLE_ENTERPRISE_LICENSE_OFFLINE_MODE="false"

# SSO Configuration  
BROKLE_ENTERPRISE_SSO_ENABLED="true"
BROKLE_ENTERPRISE_SSO_PROVIDER="saml"
BROKLE_ENTERPRISE_SSO_METADATA_URL="https://idp.example.com/metadata"
BROKLE_ENTERPRISE_SSO_ENTITY_ID="brokle-platform"

# RBAC Configuration
BROKLE_ENTERPRISE_RBAC_ENABLED="true"

# Compliance Configuration
BROKLE_ENTERPRISE_COMPLIANCE_ENABLED="true"
BROKLE_ENTERPRISE_COMPLIANCE_AUDIT_RETENTION="2160h"
BROKLE_ENTERPRISE_COMPLIANCE_SOC2_COMPLIANCE="true"
BROKLE_ENTERPRISE_COMPLIANCE_PII_ANONYMIZATION="true"

# Analytics Configuration
BROKLE_ENTERPRISE_ANALYTICS_ENABLED="true"
BROKLE_ENTERPRISE_ANALYTICS_PREDICTIVE_INSIGHTS="true"
BROKLE_ENTERPRISE_ANALYTICS_CUSTOM_DASHBOARDS="true"

# Support Configuration
BROKLE_ENTERPRISE_SUPPORT_LEVEL="priority"
BROKLE_ENTERPRISE_SUPPORT_SLA="99.95%"
BROKLE_ENTERPRISE_SUPPORT_DEDICATED_MANAGER="true"
```

### Configuration File Example

```yaml
# config/enterprise.yaml
app:
  name: "Brokle Enterprise Platform"
  version: "1.0.0"

enterprise:
  license:
    key: "${BROKLE_ENTERPRISE_LICENSE_KEY}"
    type: "enterprise"
    max_requests: 10000000
    max_users: 1000
    max_projects: 1000
    features:
      - "advanced_rbac"
      - "sso_integration"
      - "custom_compliance"
      - "predictive_insights"
      - "custom_dashboards"
      - "on_premise_deployment"
      - "dedicated_support"
    
  sso:
    enabled: true
    provider: "saml"
    metadata_url: "https://idp.company.com/federationmetadata/2007-06/federationmetadata.xml"
    entity_id: "brokle-enterprise"
    certificate: "/etc/brokle/saml-cert.pem"
    attributes:
      role: "http://schemas.microsoft.com/ws/2008/06/identity/claims/role"
      email: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
      
  rbac:
    enabled: true
    custom_roles:
      - name: "platform_admin"
        permissions: ["*:*"]
        scopes: ["org"]
      - name: "ml_engineer"  
        permissions: ["models:*", "analytics:read", "projects:read"]
        scopes: ["project", "environment"]
      - name: "business_analyst"
        permissions: ["analytics:*", "dashboards:*"]
        scopes: ["org", "project"]
        
  compliance:
    enabled: true
    audit_retention: "61320h"  # 7 years
    data_retention: "8760h"    # 1 year
    pii_anonymization: true
    soc2_compliance: true
    hipaa_compliance: true
    gdpr_compliance: true
    
  analytics:
    enabled: true
    predictive_insights: true
    custom_dashboards: true
    ml_models: true
    export_formats: ["json", "csv", "parquet", "pdf"]
    
  support:
    level: "dedicated"
    sla: "99.99%"
    dedicated_manager: true
    on_call_support: true
```

## API Integration

### License Validation API

```bash
# Check current license status
curl -X GET /api/v1/enterprise/license \
  -H "Authorization: Bearer $TOKEN"

# Response
{
  "license": {
    "type": "business",
    "valid_until": "2025-12-31T23:59:59Z",
    "max_requests": 1000000,
    "max_users": 50,
    "max_projects": 100,
    "features": ["advanced_rbac", "sso_integration", "custom_compliance"],
    "is_valid": true,
    "last_validated": "2024-09-02T15:30:00Z"
  },
  "usage": {
    "requests": 245000,
    "users": 12,
    "projects": 8,
    "last_updated": "2024-09-02T15:30:00Z"
  },
  "is_valid": true
}
```

### Feature Entitlement API

```bash
# Check specific feature availability
curl -X GET "/api/v1/enterprise/features/advanced_rbac" \
  -H "Authorization: Bearer $TOKEN"

# Response  
{
  "feature": "advanced_rbac",
  "available": true,
  "required_tier": "business",
  "current_tier": "business"
}
```

### Usage Limits API

```bash
# Check usage against limits
curl -X GET /api/v1/enterprise/usage \
  -H "Authorization: Bearer $TOKEN"

# Response
{
  "limits": {
    "requests": {
      "limit": 1000000,
      "used": 245000,
      "remaining": 755000,
      "reset_date": "2024-10-01T00:00:00Z"
    },
    "users": {
      "limit": 50,
      "used": 12,
      "remaining": 38
    }
  }
}
```

## Deployment

### Docker Deployment

#### Enterprise Docker Build
```dockerfile
# Dockerfile.enterprise
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .

# Build with enterprise tags
RUN CGO_ENABLED=0 GOOS=linux go build -tags="enterprise" -ldflags="-w -s" -o brokle-enterprise cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/brokle-enterprise .
COPY --from=builder /app/configs ./configs

CMD ["./brokle-enterprise"]
```

#### Docker Compose
```yaml
# docker-compose.enterprise.yml
version: '3.8'

services:
  brokle-enterprise:
    build:
      context: .
      dockerfile: Dockerfile.enterprise
    ports:
      - "8080:8080"
    environment:
      - BROKLE_ENTERPRISE_LICENSE_KEY=${LICENSE_KEY}
      - BROKLE_ENTERPRISE_LICENSE_TYPE=enterprise
      - BROKLE_ENTERPRISE_SSO_ENABLED=true
      - BROKLE_DATABASE_HOST=postgres
      - BROKLE_REDIS_HOST=redis
    depends_on:
      - postgres
      - redis
      - clickhouse
      
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: brokle_enterprise
      POSTGRES_USER: brokle
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      
  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}
    
  clickhouse:
    image: clickhouse/clickhouse-server:23
    environment:
      CLICKHOUSE_USER: brokle
      CLICKHOUSE_PASSWORD: ${CLICKHOUSE_PASSWORD}
    volumes:
      - clickhouse_data:/var/lib/clickhouse

volumes:
  postgres_data:
  clickhouse_data:
```

### Kubernetes Deployment

```yaml
# k8s/enterprise-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: brokle-enterprise
  labels:
    app: brokle-enterprise
spec:
  replicas: 3
  selector:
    matchLabels:
      app: brokle-enterprise
  template:
    metadata:
      labels:
        app: brokle-enterprise
    spec:
      containers:
      - name: brokle-enterprise
        image: brokle/enterprise:latest
        ports:
        - containerPort: 8080
        env:
        - name: BROKLE_ENTERPRISE_LICENSE_KEY
          valueFrom:
            secretKeyRef:
              name: brokle-license
              key: license-key
        - name: BROKLE_ENTERPRISE_SSO_ENABLED
          value: "true"
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
```

## Support

### Getting Help

#### Business Tier Support
- **Email**: support@brokle.ai
- **Response Time**: 24 hours
- **Hours**: Business hours (9 AM - 5 PM EST)
- **Channels**: Email, documentation, community

#### Enterprise Tier Support
- **Dedicated Manager**: Your assigned success manager
- **Response Time**: 4 hours (critical), 24 hours (standard)
- **Hours**: 24/7 on-call for critical issues
- **Channels**: Phone, email, Slack connect, video calls
- **Phone**: +1 (555) 123-BROKLE

### Professional Services

Contact our professional services team for:
- **Migration Planning**: Strategy for moving from existing AI infrastructure
- **Custom Development**: Built features specific to your requirements
- **Integration Support**: Help integrating with existing enterprise systems
- **Training Programs**: Team training on platform capabilities and best practices
- **Architecture Reviews**: Expert analysis of your AI infrastructure design

### Resources

- **Documentation**: https://docs.brokle.ai/enterprise
- **API Reference**: https://docs.brokle.ai/api
- **Status Page**: https://status.brokle.ai
- **Community**: https://community.brokle.ai
- **Security**: https://brokle.ai/security
- **Compliance**: https://brokle.ai/compliance

---

For additional information or to discuss enterprise requirements, contact our sales team at sales@brokle.ai or visit https://brokle.ai/contact.