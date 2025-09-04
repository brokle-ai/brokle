# Single Sign-On (SSO) Integration

## Table of Contents
- [Overview](#overview)
- [Supported Providers](#supported-providers)
- [Configuration](#configuration)
- [SAML 2.0 Setup](#saml-20-setup)
- [OIDC/OAuth2 Setup](#oidcoauth2-setup)
- [User Provisioning](#user-provisioning)
- [Role Mapping](#role-mapping)
- [API Integration](#api-integration)
- [Troubleshooting](#troubleshooting)

## Overview

Brokle Enterprise provides comprehensive Single Sign-On (SSO) integration, allowing organizations to authenticate users through their existing identity providers. SSO integration is available in **Business tier and above**.

### Benefits

- **Centralized Authentication**: Users authenticate through existing corporate identity systems
- **Enhanced Security**: Leverage enterprise-grade security policies and MFA
- **Simplified User Management**: Automatic user provisioning and role assignment
- **Audit Compliance**: Complete authentication audit trails
- **Reduced Password Fatigue**: Single login for all enterprise applications

### Architecture

```
[Identity Provider] --> [SAML/OIDC] --> [Brokle SSO Service] --> [User Session]
                                                    |
                                                    v
                                            [Role Mapping Engine]
                                                    |
                                                    v
                                            [User Provisioning]
```

## Supported Providers

### SAML 2.0 Providers
- **Active Directory Federation Services (ADFS)**
- **Azure Active Directory**
- **Okta**
- **OneLogin**
- **PingFederate**
- **Google Workspace**
- **Auth0**
- **Custom SAML 2.0 providers**

### OIDC/OAuth2 Providers
- **Azure Active Directory**
- **Google Workspace**
- **Okta**
- **Auth0**
- **Keycloak**
- **Custom OIDC providers**

## Configuration

### License Requirement

SSO integration requires a Business tier license or higher:

```yaml
enterprise:
  license:
    type: "business"  # or "enterprise"
    features:
      - "sso_integration"
```

### Basic SSO Configuration

```yaml
# config.yaml
enterprise:
  sso:
    enabled: true
    provider: "saml"  # or "oidc"
    metadata_url: "https://your-idp.com/FederationMetadata/2007-06/FederationMetadata.xml"
    entity_id: "brokle-ai-platform"
    certificate: "/etc/brokle/saml-cert.pem"
    
    # Role mapping from IdP attributes to Brokle roles
    attributes:
      role: "http://schemas.microsoft.com/ws/2008/06/identity/claims/role"
      email: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
      name: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name"
      groups: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/groups"
```

### Environment Variables

```bash
# SSO Configuration
BROKLE_ENTERPRISE_SSO_ENABLED="true"
BROKLE_ENTERPRISE_SSO_PROVIDER="saml"
BROKLE_ENTERPRISE_SSO_METADATA_URL="https://idp.example.com/metadata"
BROKLE_ENTERPRISE_SSO_ENTITY_ID="brokle-platform"
BROKLE_ENTERPRISE_SSO_CERTIFICATE_PATH="/etc/brokle/saml.pem"

# Role mapping attributes
BROKLE_ENTERPRISE_SSO_ATTR_ROLE="http://schemas.microsoft.com/ws/2008/06/identity/claims/role"
BROKLE_ENTERPRISE_SSO_ATTR_EMAIL="http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
BROKLE_ENTERPRISE_SSO_ATTR_NAME="http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name"
```

## SAML 2.0 Setup

### Step 1: Configure Brokle as Service Provider

1. **Generate SAML Configuration**:
```bash
# Get Brokle SAML metadata
curl https://your-brokle.com/api/v1/sso/saml/metadata
```

2. **Service Provider Details**:
- **Entity ID**: `brokle-ai-platform` (configurable)
- **ACS URL**: `https://your-brokle.com/api/v1/sso/saml/callback`
- **Logout URL**: `https://your-brokle.com/api/v1/sso/saml/logout`
- **Name ID Format**: `urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress`

### Step 2: Configure Identity Provider

#### Azure Active Directory

1. **Create Enterprise Application**:
   - Go to Azure Portal → Enterprise Applications
   - Create new application → Non-gallery application
   - Name: "Brokle AI Platform"

2. **Configure SAML**:
   - Identifier: `brokle-ai-platform`
   - Reply URL: `https://your-brokle.com/api/v1/sso/saml/callback`
   - Logout URL: `https://your-brokle.com/api/v1/sso/saml/logout`

3. **Attribute Mapping**:
```xml
<!-- User.email -->
<Attribute Name="http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress">
    <AttributeValue>user.mail</AttributeValue>
</Attribute>

<!-- User.displayname -->
<Attribute Name="http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name">
    <AttributeValue>user.displayname</AttributeValue>
</Attribute>

<!-- Group membership for role mapping -->
<Attribute Name="http://schemas.microsoft.com/ws/2008/06/identity/claims/role">
    <AttributeValue>user.assignedroles</AttributeValue>
</Attribute>
```

4. **Download Certificate**:
   - Download Federation Metadata XML
   - Or download Certificate (Base64) for manual configuration

#### Okta Configuration

1. **Create SAML Application**:
   - Applications → Create App Integration
   - Sign-in method: SAML 2.0

2. **General Settings**:
   - App name: "Brokle AI Platform"

3. **SAML Settings**:
   - Single sign on URL: `https://your-brokle.com/api/v1/sso/saml/callback`
   - Audience URI: `brokle-ai-platform`
   - Attribute Statements:
     - `email`: `user.email`
     - `name`: `user.firstName + " " + user.lastName`
     - `role`: `user.role`
     - `groups`: `getFilteredGroups({"brokle"}, "group.name", 50)`

### Step 3: Configure Brokle

```yaml
enterprise:
  sso:
    enabled: true
    provider: "saml"
    metadata_url: "https://login.microsoftonline.com/tenant-id/federationmetadata/2007-06/federationmetadata.xml"
    entity_id: "brokle-ai-platform"
    certificate: "/etc/brokle/azure-saml.pem"
    
    # Attribute mapping
    attributes:
      email: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
      name: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name"
      role: "http://schemas.microsoft.com/ws/2008/06/identity/claims/role"
      groups: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/groups"
```

## OIDC/OAuth2 Setup

### Step 1: Register Application

#### Azure Active Directory (OIDC)

1. **App Registration**:
   - Go to Azure Portal → App registrations
   - New registration → Name: "Brokle AI Platform"
   - Redirect URI: `https://your-brokle.com/api/v1/sso/oidc/callback`

2. **Configure Authentication**:
   - Platform configurations → Web
   - Redirect URIs: `https://your-brokle.com/api/v1/sso/oidc/callback`
   - Logout URL: `https://your-brokle.com/api/v1/sso/logout`

3. **API Permissions**:
   - Microsoft Graph → Delegated permissions
   - `User.Read`, `Group.Read.All`

4. **Token Configuration**:
   - Optional claims → Add groups claim

#### Google Workspace

1. **Google Cloud Console**:
   - APIs & Services → Credentials
   - Create OAuth 2.0 Client ID

2. **Configuration**:
   - Application type: Web application
   - Authorized redirect URIs: `https://your-brokle.com/api/v1/sso/oidc/callback`

### Step 2: Configure Brokle

```yaml
enterprise:
  sso:
    enabled: true
    provider: "oidc"
    
    # OIDC Configuration
    oidc:
      issuer: "https://login.microsoftonline.com/tenant-id/v2.0"
      client_id: "your-client-id"
      client_secret: "your-client-secret"
      redirect_uri: "https://your-brokle.com/api/v1/sso/oidc/callback"
      scopes: ["openid", "profile", "email", "groups"]
      
    # Attribute mapping from OIDC claims
    attributes:
      email: "email"
      name: "name"
      role: "roles"
      groups: "groups"
```

### Environment Variables for OIDC

```bash
BROKLE_ENTERPRISE_SSO_PROVIDER="oidc"
BROKLE_ENTERPRISE_SSO_OIDC_ISSUER="https://login.microsoftonline.com/tenant-id/v2.0"
BROKLE_ENTERPRISE_SSO_OIDC_CLIENT_ID="your-client-id"
BROKLE_ENTERPRISE_SSO_OIDC_CLIENT_SECRET="your-client-secret"
BROKLE_ENTERPRISE_SSO_OIDC_REDIRECT_URI="https://your-brokle.com/api/v1/sso/oidc/callback"
BROKLE_ENTERPRISE_SSO_OIDC_SCOPES="openid,profile,email,groups"
```

## User Provisioning

### Automatic User Creation

When users successfully authenticate via SSO, Brokle automatically creates user accounts with the following process:

1. **Extract User Information**: Email, name, and other attributes from SSO assertion
2. **Create User Account**: Generate user record in Brokle database
3. **Apply Role Mapping**: Assign roles based on IdP group membership
4. **Set Organization**: Assign to appropriate organization (configurable)

### User Provisioning Configuration

```yaml
enterprise:
  sso:
    provisioning:
      enabled: true
      default_organization: "main"  # Default organization for new users
      update_on_login: true         # Update user info on each login
      create_organization: false    # Create org from IdP attribute
      
      # User attribute mapping
      user_mapping:
        email: "email"
        name: "name" 
        first_name: "given_name"
        last_name: "family_name"
        organization: "organization"  # IdP attribute for org assignment
```

### Manual User Linking

For existing users, you can link SSO identities:

```bash
# Link existing user to SSO identity
curl -X POST /api/v1/users/user123/sso-link \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "provider": "saml",
    "external_id": "user@company.com",
    "attributes": {
      "name": "John Doe",
      "email": "user@company.com"
    }
  }'
```

## Role Mapping

### Role Mapping Configuration

Map IdP groups/roles to Brokle roles:

```yaml
enterprise:
  sso:
    role_mapping:
      enabled: true
      default_role: "viewer"  # Default role for new users
      
      # Map IdP groups to Brokle roles
      groups:
        "Brokle-Admins": "admin"
        "Brokle-Developers": "developer" 
        "Brokle-Viewers": "viewer"
        "Brokle-Owners": "owner"
        
      # Map IdP roles to Brokle roles (for role-based IdPs)
      roles:
        "BrokleAdmin": "admin"
        "BrokleDeveloper": "developer"
        
      # Advanced mapping rules
      rules:
        - condition: "groups contains 'Engineering' and department == 'AI'"
          role: "developer"
        - condition: "title contains 'Manager'"
          role: "admin"
```

### Dynamic Role Assignment

```yaml
enterprise:
  sso:
    role_mapping:
      # Organization-level roles
      organization_roles:
        "Global-Admins": "admin"
        
      # Project-level roles  
      project_roles:
        "Project-{{project}}-Admins": "admin"
        "Project-{{project}}-Developers": "developer"
        
      # Environment-level roles
      environment_roles:
        "{{env}}-Admins": "admin"
```

### Custom Role Mapping Function

For complex role mapping logic, implement custom functions:

```go
// Custom role mapping hook
func CustomRoleMapper(user *SSOUser, claims map[string]interface{}) []UserRole {
    var roles []UserRole
    
    // Extract groups from claims
    groups := extractGroups(claims)
    department := extractAttribute(claims, "department")
    title := extractAttribute(claims, "title")
    
    // Complex role logic
    if contains(groups, "Engineering") && department == "AI" {
        roles = append(roles, UserRole{
            Role:  "ai_engineer",
            Scope: "project:ai-chatbot",
        })
    }
    
    if strings.Contains(title, "Manager") {
        roles = append(roles, UserRole{
            Role:  "manager",
            Scope: "organization:main",
        })
    }
    
    return roles
}
```

## API Integration

### SSO Login Flow

1. **Initiate Login**:
```bash
# Get SSO login URL
curl -X GET /api/v1/sso/login \
  -H "Accept: application/json"

# Response
{
  "login_url": "https://your-brokle.com/api/v1/sso/redirect",
  "provider": "saml",
  "state": "random-state-token"
}
```

2. **Handle Callback**:
   - User authenticates with IdP
   - IdP redirects to callback URL with SAML assertion or OIDC code
   - Brokle validates assertion/code and creates session

3. **Session Information**:
```bash
# Get current session
curl -X GET /api/v1/auth/session \
  -H "Authorization: Bearer $JWT_TOKEN"

# Response  
{
  "user": {
    "id": "user123",
    "email": "user@company.com",
    "name": "John Doe",
    "roles": ["developer"],
    "organization": "main",
    "sso_provider": "saml"
  },
  "session": {
    "expires_at": "2024-09-03T15:30:00Z",
    "sso_session_id": "session456"
  }
}
```

### SSO Management API

```bash
# List SSO providers
curl -X GET /api/v1/admin/sso/providers \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Get provider configuration
curl -X GET /api/v1/admin/sso/providers/saml \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Update provider configuration
curl -X PUT /api/v1/admin/sso/providers/saml \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "enabled": true,
    "metadata_url": "https://new-idp.com/metadata"
  }'

# Test SSO configuration
curl -X POST /api/v1/admin/sso/test \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "provider": "saml",
    "test_user": "test@company.com"
  }'
```

### User SSO Information API

```bash
# Get user SSO information
curl -X GET /api/v1/users/user123/sso \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Response
{
  "sso_provider": "saml",
  "external_id": "user@company.com",
  "last_login": "2024-09-02T14:30:00Z",
  "attributes": {
    "name": "John Doe",
    "email": "user@company.com",
    "groups": ["Engineering", "AI-Team"]
  },
  "role_mappings": [
    {
      "source": "group:Engineering",
      "role": "developer",
      "scope": "project:ai-chatbot"
    }
  ]
}
```

## Troubleshooting

### Common Issues

#### 1. SAML Assertion Validation Failed

**Symptoms**: Users get "Invalid SAML assertion" error

**Diagnosis**:
```bash
# Check SAML configuration
curl -X GET /api/v1/debug/sso/saml/config \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Validate SAML assertion manually
curl -X POST /api/v1/debug/sso/saml/validate \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "assertion": "base64-encoded-assertion"
  }'
```

**Solutions**:
- Verify certificate is valid and matches IdP
- Check clock synchronization between Brokle and IdP
- Ensure Entity ID matches exactly
- Verify ACS URL is configured correctly in IdP

#### 2. OIDC Token Validation Failed

**Symptoms**: "Invalid ID token" or "Token signature verification failed"

**Diagnosis**:
```bash
# Check OIDC configuration
curl -X GET /api/v1/debug/sso/oidc/config \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Test OIDC flow
curl -X POST /api/v1/debug/sso/oidc/test \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "client_id": "your-client-id",
    "issuer": "https://login.microsoftonline.com/tenant-id/v2.0"
  }'
```

**Solutions**:
- Verify client ID and secret are correct
- Check issuer URL matches exactly
- Ensure redirect URI is registered in IdP
- Verify scopes are granted in IdP

#### 3. Role Mapping Not Working

**Symptoms**: Users login but have incorrect roles

**Diagnosis**:
```bash
# Debug role mapping for user
curl -X GET "/api/v1/debug/sso/role-mapping?user=user@company.com" \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Response shows mapping process
{
  "user_claims": {
    "groups": ["Engineering", "AI-Team"],
    "role": "Developer"
  },
  "mapping_rules": [
    {
      "rule": "groups contains 'Engineering'",
      "matched": true,
      "result": "developer"
    }
  ],
  "final_roles": ["developer"]
}
```

**Solutions**:
- Check attribute names match between IdP and Brokle config
- Verify group/role claims are included in assertions/tokens
- Test role mapping rules with debug endpoint
- Ensure default role is configured

#### 4. User Provisioning Issues

**Symptoms**: Users can't be created or updated

**Diagnosis**:
```bash
# Check user provisioning logs
curl -X GET "/api/v1/debug/sso/provisioning?user=user@company.com" \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Check organization assignment
curl -X GET /api/v1/users/user@company.com \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

**Solutions**:
- Verify required user attributes are provided by IdP
- Check default organization exists
- Ensure user has permission to be created in target organization
- Verify email uniqueness constraints

### Debugging Tools

#### Enable Debug Logging

```yaml
logging:
  level: "debug"
  
  # Enable SSO-specific logging
  loggers:
    sso: "debug"
    saml: "debug"
    oidc: "debug"
```

#### SSO Debug Endpoints (Development Only)

```bash
# Get SAML metadata
curl -X GET /api/v1/debug/sso/saml/metadata

# Get OIDC discovery document
curl -X GET /api/v1/debug/sso/oidc/discovery

# Validate SSO configuration
curl -X POST /api/v1/debug/sso/validate \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "provider": "saml",
    "config": {
      "metadata_url": "https://idp.com/metadata"
    }
  }'
```

### Log Analysis

Key log patterns to look for:

```bash
# SAML assertion validation
grep "SAML assertion" /var/log/brokle.log

# OIDC token validation
grep "OIDC token" /var/log/brokle.log

# Role mapping events
grep "role mapping" /var/log/brokle.log

# User provisioning events  
grep "user provisioning" /var/log/brokle.log
```

### Support Resources

- **Documentation**: https://docs.brokle.com/enterprise/sso
- **Community**: https://community.brokle.com/sso
- **Enterprise Support**: support@brokle.com
- **Professional Services**: Contact your account manager for SSO implementation assistance

---

For additional SSO configuration examples and advanced use cases, see the [Enterprise Deployment Guide](DEPLOYMENT.md).