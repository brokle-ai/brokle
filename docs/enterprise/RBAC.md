# Role-Based Access Control (RBAC)

## Table of Contents
- [Overview](#overview)
- [RBAC Architecture](#rbac-architecture)
- [Built-in Roles](#built-in-roles)
- [Custom Roles](#custom-roles)
- [Permissions System](#permissions-system)
- [Scope Hierarchy](#scope-hierarchy)
- [Configuration](#configuration)
- [API Integration](#api-integration)
- [SSO Integration](#sso-integration)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Overview

Brokle Enterprise provides advanced Role-Based Access Control (RBAC) that extends beyond the basic OSS roles. RBAC allows organizations to define granular permissions and custom roles aligned with their organizational structure and security policies.

### License Requirements

- **Basic RBAC**: Available in **Pro tier and above**
- **Advanced RBAC**: Available in **Business tier and above** 
- **Custom RBAC**: Available in **Enterprise tier**

### Key Features

- **Granular Permissions**: Fine-grained control over API access and resource management
- **Custom Roles**: Define roles beyond the built-in owner/admin/developer/viewer
- **Hierarchical Scopes**: Organization → Project → Environment permission inheritance
- **Dynamic Permissions**: Context-aware permissions based on resource ownership
- **SSO Integration**: Automatic role assignment from identity provider groups
- **Audit Trails**: Complete access logging for compliance requirements

### Architecture Benefits

- **Principle of Least Privilege**: Users get only the minimum necessary permissions
- **Separation of Duties**: Clear role boundaries prevent privilege escalation
- **Compliance Ready**: Audit trails and role definitions support SOX, SOC2, HIPAA
- **Scalable Governance**: Centralized role management across large organizations

## RBAC Architecture

### Permission Model

```
User → Role → Permissions → Resources
  |      |         |           |
  |      |         |           └─ API endpoints, data, features
  |      |         └─ Read, Write, Admin, Delete actions
  |      └─ Custom or built-in role definition
  └─ Individual user account
```

### Scope Hierarchy

```
Organization (Tenant)
├── Projects
│   ├── Environments (dev, staging, prod)
│   │   ├── API Keys
│   │   ├── Models
│   │   └── Analytics Data
│   └── Team Members
└── Organization Settings
```

### Permission Inheritance

```
Organization Admin
├── Can manage all projects in organization
├── Project Admin (inherited scope: specific project)
│   ├── Can manage environments in project
│   ├── Environment Admin (inherited scope: specific environment)
│   │   ├── Can manage API keys in environment
│   │   └── Developer (inherited scope: specific environment)
│   │       └── Can use APIs, view analytics
│   └── Viewer (inherited scope: project or environment)
│       └── Read-only access to resources
└── Organization Viewer
    └── Read-only access to organization resources
```

## Built-in Roles

### OSS Roles (Available in all tiers)

#### Owner
- **Scope**: Organization-wide
- **Permissions**: Full access to everything
- **Key Actions**:
  - Manage billing and subscriptions
  - Add/remove organization members
  - Delete organization
  - Assign any role to any user
  - Access all projects and environments

#### Admin
- **Scope**: Organization or Project
- **Permissions**: Administrative access within scope
- **Key Actions**:
  - Create/delete projects (org admin) or environments (project admin)
  - Manage team members within scope
  - Configure organization/project settings
  - View billing information (org admin)
  - Cannot delete organization or manage billing

#### Developer
- **Scope**: Project or Environment
- **Permissions**: Development access within scope  
- **Key Actions**:
  - Create/manage API keys
  - Deploy models and configurations
  - View analytics and logs
  - Configure environment settings
  - Cannot manage team members

#### Viewer
- **Scope**: Organization, Project, or Environment
- **Permissions**: Read-only access within scope
- **Key Actions**:
  - View projects, environments, and configurations
  - View analytics and dashboards
  - View team members
  - Cannot modify any resources

### Enterprise Roles (Business tier and above)

#### AI Architect
- **Scope**: Organization or Project
- **Permissions**: Model and infrastructure management
- **Key Actions**:
  - Design AI model architectures
  - Configure provider routing strategies
  - Set up semantic caching policies
  - Define quality scoring rules
  - Manage model deployment pipelines

#### Data Engineer  
- **Scope**: Project or Environment
- **Permissions**: Data and analytics management
- **Key Actions**:
  - Configure data pipelines
  - Manage analytics data retention
  - Set up custom dashboards
  - Export analytics data
  - Configure compliance data handling

#### Security Engineer
- **Scope**: Organization-wide
- **Permissions**: Security and compliance management
- **Key Actions**:
  - Configure SSO and authentication
  - Manage security policies
  - Set up audit logging
  - Configure compliance controls
  - Review access logs and reports

#### Business Analyst
- **Scope**: Organization or Project  
- **Permissions**: Analytics and reporting access
- **Key Actions**:
  - Create custom dashboards
  - Generate business reports
  - Access predictive analytics
  - Export business data
  - Configure business metrics

## Custom Roles

### Creating Custom Roles

Custom roles allow organizations to define permissions that match their specific workflows and organizational structure.

#### Role Definition Structure

```go
type CustomRole struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description,omitempty"`
    Permissions []string  `json:"permissions"`
    Scopes      []string  `json:"scopes"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

#### Example Custom Roles

```yaml
# AI Platform Engineer
custom_roles:
  - name: "ai_platform_engineer"
    description: "Manages AI infrastructure and platform operations"
    permissions:
      - "models:deploy"
      - "models:configure"
      - "infrastructure:manage"
      - "monitoring:configure"
      - "analytics:advanced"
    scopes: ["organization", "project"]

  # ML Operations Specialist  
  - name: "ml_ops_specialist"
    description: "Handles ML model lifecycle and operations"
    permissions:
      - "models:deploy"
      - "models:monitor"
      - "pipelines:manage"
      - "experiments:manage"
      - "analytics:read"
    scopes: ["project", "environment"]

  # AI Ethics Officer
  - name: "ai_ethics_officer"
    description: "Ensures ethical AI practices and compliance"
    permissions:
      - "compliance:audit"
      - "models:review"
      - "data:audit"
      - "policies:manage"
      - "reports:compliance"
    scopes: ["organization"]
```

## Permissions System

### Permission Categories

#### API Access Permissions
```yaml
api_permissions:
  # Core AI Gateway
  - "gateway:use"          # Make AI API requests
  - "gateway:configure"    # Configure routing rules
  - "gateway:monitor"      # View gateway metrics
  
  # Models and Providers
  - "models:read"          # View model configurations
  - "models:deploy"        # Deploy and configure models
  - "models:delete"        # Remove model configurations
  - "providers:configure"  # Manage AI provider settings
  
  # Analytics and Monitoring
  - "analytics:read"       # View basic analytics
  - "analytics:advanced"   # Access predictive insights
  - "analytics:export"     # Export analytics data
  - "dashboards:create"    # Create custom dashboards
  
  # Organization Management
  - "org:read"            # View organization info
  - "org:manage"          # Manage organization settings
  - "org:billing"         # Access billing information
  - "users:invite"        # Invite new users
  - "users:manage"        # Manage user roles
```

#### Resource Permissions
```yaml
resource_permissions:
  # Projects
  - "projects:read"        # View project details
  - "projects:create"      # Create new projects
  - "projects:manage"      # Manage project settings
  - "projects:delete"      # Delete projects
  
  # Environments
  - "environments:read"    # View environment details
  - "environments:create"  # Create new environments
  - "environments:manage"  # Manage environment settings
  - "environments:delete"  # Delete environments
  
  # API Keys
  - "apikeys:read"        # View API keys (masked)
  - "apikeys:create"      # Create new API keys
  - "apikeys:revoke"      # Revoke API keys
  - "apikeys:regenerate"  # Regenerate API keys
```

#### Enterprise Permissions
```yaml
enterprise_permissions:
  # Compliance
  - "compliance:audit"     # Generate audit reports
  - "compliance:configure" # Configure compliance settings
  - "data:anonymize"      # Anonymize PII data
  - "data:retention"      # Configure data retention
  
  # Security
  - "sso:configure"       # Configure SSO settings
  - "rbac:manage"         # Manage roles and permissions
  - "security:audit"      # View security logs
  - "policies:manage"     # Manage security policies
  
  # Advanced Analytics
  - "ml:models"           # Access ML models
  - "insights:predictive" # Access predictive insights
  - "reports:custom"      # Create custom reports
  - "data:export"         # Export enterprise data
```

### Permission Syntax

#### Basic Permissions
```
resource:action
```
Examples:
- `models:read` - Read access to models
- `analytics:export` - Export analytics data
- `users:manage` - Manage user accounts

#### Wildcard Permissions  
```
resource:*     # All actions on resource
*:action       # Action on all resources
*:*           # All actions on all resources (super admin)
```

#### Conditional Permissions
```yaml
conditional_permissions:
  # Owner-only permissions
  - permission: "projects:delete"
    condition: "resource.owner == user.id"
    
  # Time-based permissions
  - permission: "analytics:export"
    condition: "time.hour >= 9 AND time.hour <= 17"
    
  # Environment-specific permissions
  - permission: "models:deploy"
    condition: "environment != 'production' OR user.role == 'admin'"
```

## Scope Hierarchy

### Scope Types

#### Organization Scope
- **Level**: Tenant/Organization
- **Applies to**: All resources within the organization
- **Inheritance**: Permissions apply to all child projects and environments
- **Use Cases**: Organization administrators, security officers

#### Project Scope
- **Level**: Project within organization
- **Applies to**: All resources within the specific project
- **Inheritance**: Permissions apply to all environments in the project
- **Use Cases**: Project managers, team leads

#### Environment Scope  
- **Level**: Environment within project
- **Applies to**: Resources in the specific environment only
- **Inheritance**: No further inheritance
- **Use Cases**: Developers, environment-specific administrators

### Permission Resolution

When a user attempts an action, permissions are resolved using this hierarchy:

1. **Explicit Permission**: Direct permission for the exact resource and action
2. **Inherited Permission**: Permission from parent scope (project → org)
3. **Wildcard Permission**: Wildcard permission that covers the resource/action
4. **Default Permission**: Built-in role permissions
5. **Deny by Default**: Access denied if no permission matches

### Example Permission Resolution

```yaml
user: john@company.com
roles:
  - role: "project_admin"
    scope: "project:ai-chatbot" 
  - role: "developer"
    scope: "environment:ai-chatbot:production"

# Permission check: Can john delete environment "ai-chatbot:production"?
resolution_process:
  1. Check explicit permission: "environments:delete" on "ai-chatbot:production"
     → No explicit permission
  
  2. Check inherited permissions: "project_admin" on "project:ai-chatbot"  
     → Project admin has "environments:delete" permission
     → ai-chatbot:production is in project ai-chatbot
     → Permission GRANTED
     
result: ALLOW
```

## Configuration

### Basic RBAC Configuration

```yaml
# config.yaml
enterprise:
  rbac:
    enabled: true
    
    # Default role for new users
    default_role: "viewer"
    
    # Built-in role customization
    role_permissions:
      admin:
        - "projects:*"
        - "environments:*"
        - "users:manage"
      developer:
        - "models:read"
        - "models:deploy"
        - "analytics:read"
        - "apikeys:*"
```

### Custom Roles Configuration

```yaml
enterprise:
  rbac:
    enabled: true
    
    # Define custom roles
    custom_roles:
      - name: "platform_engineer"
        description: "Manages platform infrastructure and AI operations"
        permissions:
          - "models:*"
          - "infrastructure:manage"
          - "monitoring:configure"
          - "analytics:advanced"
        scopes: ["organization", "project"]
        
      - name: "data_scientist"
        description: "Develops and deploys AI models"
        permissions:
          - "models:read"
          - "models:deploy"
          - "experiments:*"
          - "analytics:read"
          - "dashboards:create"
        scopes: ["project", "environment"]
        
      - name: "compliance_officer"
        description: "Ensures compliance and data governance"
        permissions:
          - "compliance:*"
          - "audit:read"
          - "data:anonymize"
          - "policies:manage"
        scopes: ["organization"]
```

### Permission Inheritance Rules

```yaml
enterprise:
  rbac:
    inheritance:
      # Organization permissions inherit to projects
      organization_to_project:
        - "analytics:read"
        - "users:view"
        - "projects:read"
        
      # Project permissions inherit to environments
      project_to_environment:
        - "models:read"
        - "analytics:read"
        - "apikeys:read"
        
    # Permissions that never inherit (scope-specific)
    no_inherit:
      - "org:billing"
      - "org:delete"
      - "projects:delete"
      - "environments:delete"
```

### Environment Variables

```bash
# RBAC Configuration
BROKLE_ENTERPRISE_RBAC_ENABLED="true"
BROKLE_ENTERPRISE_RBAC_DEFAULT_ROLE="viewer"
BROKLE_ENTERPRISE_RBAC_STRICT_MODE="true"  # Deny by default

# Custom roles (JSON encoded)
BROKLE_ENTERPRISE_RBAC_CUSTOM_ROLES='[{"name":"platform_engineer","permissions":["models:*","infrastructure:manage"],"scopes":["organization"]}]'
```

## API Integration

### Role Management API

#### Create Custom Role
```bash
curl -X POST /api/v1/rbac/roles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ml_engineer",
    "description": "Machine Learning Engineer with model deployment access",
    "permissions": [
      "models:read",
      "models:deploy", 
      "experiments:manage",
      "analytics:read"
    ],
    "scopes": ["project", "environment"]
  }'

# Response
{
  "id": "role123",
  "name": "ml_engineer",
  "description": "Machine Learning Engineer with model deployment access",
  "permissions": ["models:read", "models:deploy", "experiments:manage", "analytics:read"],
  "scopes": ["project", "environment"],
  "created_at": "2024-09-02T15:30:00Z"
}
```

#### List Roles
```bash
# List all available roles
curl -X GET /api/v1/rbac/roles \
  -H "Authorization: Bearer $TOKEN"

# Response
{
  "builtin_roles": [
    {"name": "owner", "description": "Full organization access"},
    {"name": "admin", "description": "Administrative access"},
    {"name": "developer", "description": "Development access"},
    {"name": "viewer", "description": "Read-only access"}
  ],
  "custom_roles": [
    {
      "id": "role123",
      "name": "ml_engineer",
      "description": "Machine Learning Engineer",
      "permissions": ["models:read", "models:deploy"],
      "scopes": ["project", "environment"]
    }
  ]
}
```

#### Update Role
```bash
curl -X PUT /api/v1/rbac/roles/role123 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Updated ML Engineer role",
    "permissions": [
      "models:read",
      "models:deploy",
      "models:monitor",
      "experiments:manage",
      "analytics:read"
    ]
  }'
```

### User Role Assignment API

#### Assign Role to User
```bash
curl -X POST /api/v1/rbac/users/user123/roles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "ml_engineer",
    "scope": "project:ai-chatbot"
  }'

# Multiple role assignment
curl -X POST /api/v1/rbac/users/user123/roles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "roles": [
      {"role": "developer", "scope": "project:ai-chatbot"},
      {"role": "viewer", "scope": "organization:main"}
    ]
  }'
```

#### Get User Roles
```bash
curl -X GET /api/v1/rbac/users/user123/roles \
  -H "Authorization: Bearer $TOKEN"

# Response
{
  "user_id": "user123",
  "roles": [
    {
      "role": "developer",
      "scope": "project:ai-chatbot",
      "assigned_at": "2024-09-01T10:00:00Z",
      "assigned_by": "admin@company.com"
    },
    {
      "role": "ml_engineer",
      "scope": "project:ai-chatbot",
      "assigned_at": "2024-09-02T15:30:00Z",
      "assigned_by": "manager@company.com"
    }
  ]
}
```

#### Remove Role from User
```bash
curl -X DELETE "/api/v1/rbac/users/user123/roles/ml_engineer?scope=project:ai-chatbot" \
  -H "Authorization: Bearer $TOKEN"
```

### Permission Check API

#### Check Specific Permission
```bash
curl -X GET "/api/v1/rbac/users/user123/permissions/models:deploy?scope=project:ai-chatbot" \
  -H "Authorization: Bearer $TOKEN"

# Response
{
  "user_id": "user123",
  "permission": "models:deploy", 
  "scope": "project:ai-chatbot",
  "allowed": true,
  "reason": "User has role 'ml_engineer' which grants 'models:deploy' permission",
  "effective_role": "ml_engineer"
}
```

#### Bulk Permission Check
```bash
curl -X POST /api/v1/rbac/users/user123/permissions/check \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "permissions": [
      {"permission": "models:deploy", "scope": "project:ai-chatbot"},
      {"permission": "projects:delete", "scope": "organization:main"},
      {"permission": "analytics:export", "scope": "project:ai-chatbot"}
    ]
  }'

# Response
{
  "user_id": "user123",
  "results": [
    {"permission": "models:deploy", "scope": "project:ai-chatbot", "allowed": true},
    {"permission": "projects:delete", "scope": "organization:main", "allowed": false}, 
    {"permission": "analytics:export", "scope": "project:ai-chatbot", "allowed": true}
  ]
}
```

### Audit and Reporting API

#### Get Access Logs
```bash
curl -X GET "/api/v1/rbac/audit/access?user=user123&start=2024-09-01&end=2024-09-02" \
  -H "Authorization: Bearer $TOKEN"

# Response
{
  "logs": [
    {
      "timestamp": "2024-09-02T15:30:00Z",
      "user_id": "user123",
      "action": "models:deploy",
      "resource": "project:ai-chatbot",
      "result": "allowed",
      "role": "ml_engineer",
      "ip_address": "192.168.1.100"
    }
  ]
}
```

#### Generate Role Usage Report
```bash
curl -X GET /api/v1/rbac/reports/role-usage \
  -H "Authorization: Bearer $TOKEN"

# Response
{
  "report_date": "2024-09-02T15:30:00Z",
  "role_usage": [
    {
      "role": "developer",
      "user_count": 25,
      "active_users_7d": 20,
      "most_used_permissions": ["models:read", "analytics:read", "apikeys:create"]
    },
    {
      "role": "ml_engineer",
      "user_count": 8,
      "active_users_7d": 7,
      "most_used_permissions": ["models:deploy", "experiments:manage"]
    }
  ]
}
```

## SSO Integration

### Automatic Role Assignment from SSO

When users authenticate via SSO, roles can be automatically assigned based on their identity provider group membership.

#### Role Mapping Configuration

```yaml
enterprise:
  sso:
    role_mapping:
      enabled: true
      
      # Map IdP groups to Brokle roles
      groups:
        # Engineering teams
        "AI-Platform-Team": 
          role: "platform_engineer"
          scope: "organization:main"
          
        "ML-Engineers":
          role: "ml_engineer" 
          scope: "project:ai-models"
          
        "Data-Scientists":
          role: "data_scientist"
          scope: "project:ai-models"
          
        # Administrative roles
        "Platform-Admins":
          role: "admin"
          scope: "organization:main"
          
        "Security-Team":
          role: "security_engineer"
          scope: "organization:main"
          
        # Project-specific roles
        "Chatbot-Developers":
          role: "developer"
          scope: "project:ai-chatbot"
          
        "Analytics-Team":
          role: "business_analyst"
          scope: "organization:main"
```

#### Dynamic Role Assignment Rules

```yaml
enterprise:
  sso:
    role_mapping:
      # Complex mapping rules
      rules:
        # Senior engineers get platform engineer role
        - condition: "groups contains 'Engineering' AND title contains 'Senior'"
          role: "platform_engineer"
          scope: "organization:main"
          
        # Managers get admin access to their department's projects
        - condition: "title contains 'Manager' AND department == 'AI'"
          role: "admin"
          scope: "project:{{department|lower}}-*"
          
        # Data scientists get ML engineer role on model projects
        - condition: "job_function == 'Data Science'"
          role: "ml_engineer"
          scope: "project:ai-models"
          
        # Contractors get limited developer access
        - condition: "employee_type == 'Contractor'"
          role: "developer"
          scope: "environment:{{assigned_project}}:development"
```

#### API for SSO Role Management

```bash
# Get SSO role mappings
curl -X GET /api/v1/sso/role-mappings \
  -H "Authorization: Bearer $TOKEN"

# Update SSO role mapping
curl -X PUT /api/v1/sso/role-mappings \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "groups": {
      "AI-Platform-Team": {
        "role": "platform_engineer",
        "scope": "organization:main"
      }
    }
  }'

# Test role mapping for user
curl -X POST /api/v1/sso/role-mappings/test \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_attributes": {
      "groups": ["AI-Platform-Team", "Engineering"],
      "title": "Senior ML Engineer",
      "department": "AI"
    }
  }'

# Response
{
  "mapped_roles": [
    {
      "role": "platform_engineer",
      "scope": "organization:main",
      "source": "group:AI-Platform-Team"
    },
    {
      "role": "ml_engineer", 
      "scope": "project:ai-models",
      "source": "rule:senior_engineer"
    }
  ]
}
```

## Best Practices

### Role Design Principles

#### 1. Principle of Least Privilege
```yaml
# Good: Specific permissions for specific needs
data_scientist_role:
  permissions:
    - "models:read"
    - "models:deploy"
    - "experiments:manage" 
    - "analytics:read"
  scopes: ["project"]

# Bad: Overly broad permissions  
data_scientist_role:
  permissions:
    - "*:*"  # Too broad, violates least privilege
  scopes: ["organization"]
```

#### 2. Separation of Duties
```yaml
# Good: Separate deployment from approval
ml_engineer:
  permissions: ["models:deploy"]
  scopes: ["environment:development", "environment:staging"]

ml_approver:
  permissions: ["models:deploy"]
  scopes: ["environment:production"]
  
# Models require approval workflow for production deployment
```

#### 3. Role Granularity Balance
```yaml
# Good: Balanced granularity
roles:
  - frontend_developer    # Specific to frontend needs
  - backend_developer     # Specific to backend needs  
  - fullstack_developer   # Combines both for versatile developers

# Bad: Too granular
roles:
  - react_developer
  - vue_developer
  - angular_developer     # Too specific, hard to manage

# Bad: Not granular enough  
roles:
  - developer             # Too broad for specialized teams
```

### Security Best Practices

#### 1. Regular Role Audits
```bash
# Monthly role audit script
#!/bin/bash

# Get all users and their roles
curl -X GET /api/v1/rbac/users/roles \
  -H "Authorization: Bearer $ADMIN_TOKEN" > current_roles.json

# Check for users with multiple high-privilege roles
curl -X GET "/api/v1/rbac/audit/high-privilege-users" \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Generate role usage report
curl -X GET /api/v1/rbac/reports/role-usage \
  -H "Authorization: Bearer $ADMIN_TOKEN" > role_usage_report.json
```

#### 2. Time-Limited Role Assignments
```bash
# Assign temporary elevated permissions
curl -X POST /api/v1/rbac/users/user123/roles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "admin",
    "scope": "project:ai-chatbot",
    "expires_at": "2024-09-03T17:00:00Z",
    "reason": "Emergency deployment for critical bug fix"
  }'
```

#### 3. Approval Workflows for Sensitive Roles
```yaml
enterprise:
  rbac:
    approval_required:
      roles: ["admin", "security_engineer", "compliance_officer"]
      approvers: ["owner", "security_team"]
      approval_timeout: "24h"
```

### Organizational Patterns

#### 1. Team-Based Role Structure
```yaml
# Development Teams
teams:
  ai_platform_team:
    roles: ["platform_engineer", "developer"]
    projects: ["ai-gateway", "infrastructure"]
    
  ml_research_team:
    roles: ["data_scientist", "ml_engineer"]  
    projects: ["ai-models", "experiments"]
    
  product_team:
    roles: ["business_analyst", "viewer"]
    projects: ["analytics", "reporting"]
```

#### 2. Environment-Based Permissions
```yaml
# Different permissions per environment
environment_roles:
  development:
    developer: ["models:*", "experiments:*", "data:read"]
    
  staging:  
    developer: ["models:deploy", "models:test", "data:read"]
    
  production:
    developer: ["models:read", "analytics:read"]
    ml_approver: ["models:deploy"]
```

#### 3. Project Lifecycle Permissions  
```yaml
# Permissions that change based on project phase
project_phases:
  research:
    data_scientist: ["experiments:*", "data:*", "models:develop"]
    
  development:
    ml_engineer: ["models:deploy", "models:test", "infrastructure:configure"]
    
  production:
    platform_engineer: ["models:deploy", "infrastructure:manage", "monitoring:configure"]
```

## Troubleshooting

### Common RBAC Issues

#### 1. Permission Denied Errors

**Symptoms**: Users get 403 Forbidden responses

**Diagnosis**:
```bash
# Check user's current permissions
curl -X GET /api/v1/rbac/users/user123/permissions \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Check specific permission
curl -X GET "/api/v1/rbac/users/user123/permissions/models:deploy?scope=project:ai-chatbot" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

**Common Solutions**:
- Verify user has correct role assigned
- Check role has necessary permissions
- Ensure permission applies to correct scope
- Verify scope inheritance is working

#### 2. SSO Role Mapping Issues

**Symptoms**: Users login but don't get expected roles

**Diagnosis**:
```bash
# Debug SSO role mapping
curl -X POST /api/v1/debug/sso/role-mapping \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "user": "user@company.com",
    "sso_attributes": {
      "groups": ["Engineering", "AI-Team"],
      "title": "Senior ML Engineer"
    }
  }'
```

**Solutions**:
- Check SSO attribute names match configuration
- Verify group membership in identity provider
- Test role mapping rules with debug endpoint
- Ensure SSO user provisioning is enabled

#### 3. Role Inheritance Problems

**Symptoms**: Users don't inherit expected permissions from parent scopes

**Diagnosis**:
```bash
# Check permission resolution path
curl -X GET "/api/v1/rbac/debug/permission-resolution?user=user123&permission=models:read&scope=environment:ai-chatbot:prod" \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Response shows resolution steps
{
  "user": "user123",
  "permission": "models:read",
  "scope": "environment:ai-chatbot:prod",
  "resolution": [
    {"step": "direct_permission", "found": false},
    {"step": "environment_role", "found": false},
    {"step": "project_inheritance", "found": true, "role": "developer"},
    {"step": "organization_inheritance", "found": false}
  ],
  "result": "allowed"
}
```

**Solutions**:
- Verify inheritance rules are configured correctly
- Check parent scope permissions exist
- Ensure no explicit deny rules block inheritance

#### 4. Custom Role Validation Errors

**Symptoms**: Cannot create or update custom roles

**Common Validation Errors**:
```json
{
  "error": {
    "code": "INVALID_PERMISSION",
    "message": "Permission 'invalid:action' is not recognized",
    "invalid_permissions": ["invalid:action"]
  }
}
```

**Solutions**:
- Use valid permission syntax (`resource:action`)
- Check available permissions: `GET /api/v1/rbac/permissions`
- Ensure custom permissions are enabled for your license tier

### Debug Tools

#### RBAC Debug Endpoints (Development/Admin Only)

```bash
# Get all available permissions
curl -X GET /api/v1/debug/rbac/permissions \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Simulate permission check
curl -X POST /api/v1/debug/rbac/simulate \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "user_id": "user123",
    "action": "models:deploy",
    "resource": "project:ai-chatbot"
  }'

# Get detailed role information
curl -X GET /api/v1/debug/rbac/roles/ml_engineer \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

#### Audit Logging

Enable detailed RBAC audit logging:

```yaml
logging:
  loggers:
    rbac: "debug"
    permissions: "debug"
```

Key log patterns:
```bash
# Permission checks
grep "permission_check" /var/log/brokle.log

# Role assignments  
grep "role_assigned" /var/log/brokle.log

# SSO role mapping
grep "sso_role_mapping" /var/log/brokle.log
```

---

For additional RBAC configuration examples and enterprise use cases, see the [Enterprise Deployment Guide](DEPLOYMENT.md) and [SSO Integration Guide](SSO.md).