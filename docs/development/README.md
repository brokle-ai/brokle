# Development Documentation

This directory contains comprehensive development guides for the Brokle platform, designed to help both human developers and AI assistants understand and maintain the codebase effectively.

## 📚 Core Architecture & Error Handling

### [ERROR_HANDLING_GUIDE.md](./ERROR_HANDLING_GUIDE.md)
**Complete Implementation Guide** - Industrial-grade Go error handling patterns across all layers.

**Contents:**
- Clean architecture error flow (Repository → Service → Handler)
- Professional domain alias patterns
- Comprehensive implementation examples
- Testing strategies and troubleshooting
- Best practices checklist

**Use Cases:**
- Implementing new repositories, services, or handlers
- Understanding error propagation through layers
- Code review guidelines
- Onboarding new developers

### [DOMAIN_ALIAS_PATTERNS.md](./DOMAIN_ALIAS_PATTERNS.md)
**Professional Domain Aliases** - Standardized import patterns for clean, conflict-free code.

**Contents:**
- Standard domain aliases (`authDomain`, `orgDomain`, etc.)
- Multi-domain usage examples
- Migration guide for existing code
- Anti-patterns to avoid
- Code quality validation

**Use Cases:**
- Setting up new repository files
- Refactoring existing domain imports
- Resolving import conflicts
- Maintaining consistent code style

### [ERROR_HANDLING_QUICK_REFERENCE.md](./ERROR_HANDLING_QUICK_REFERENCE.md)
**Developer Cheat Sheet** - Quick reference for daily development tasks.

**Contents:**
- Essential patterns for each layer
- Common mistake prevention
- Error type mappings
- Debugging tips and tricks

**Use Cases:**
- Quick lookup during development
- Code review reference
- New team member onboarding
- AI assistant guidance

## 🔌 API Development Standards

### [API_DEVELOPMENT_GUIDE.md](./API_DEVELOPMENT_GUIDE.md)
**Complete API Standards** - Comprehensive guide for building consistent, professional APIs.

**Contents:**
- RESTful API design principles and URL structure standards
- Handler development patterns and request/response standards
- Industrial error handling integration
- Validation patterns and authentication/authorization
- OpenAPI documentation requirements and testing standards

**Use Cases:**
- Building new API endpoints
- Standardizing existing APIs
- API design reviews
- Team API development training

### [PAGINATION_GUIDE.md](./PAGINATION_GUIDE.md)
**Comprehensive Pagination** - Complete patterns for efficient, user-friendly pagination.

**Contents:**
- Offset-based and cursor-based pagination patterns
- Database implementation with performance optimization
- Advanced filtering, search, and sorting patterns
- Testing strategies and performance guidelines
- Frontend integration examples

**Use Cases:**
- Implementing list endpoints
- Optimizing pagination performance
- Complex filtering and search requirements
- Large dataset handling

## 🎯 Quick Start for New Developers

### Core Patterns (Start Here)
1. **Start with**: [ERROR_HANDLING_QUICK_REFERENCE.md](./ERROR_HANDLING_QUICK_REFERENCE.md) for immediate patterns
2. **Deep dive**: [ERROR_HANDLING_GUIDE.md](./ERROR_HANDLING_GUIDE.md) for complete understanding
3. **Implementation**: [DOMAIN_ALIAS_PATTERNS.md](./DOMAIN_ALIAS_PATTERNS.md) for proper imports

### API Development
4. **API Standards**: [API_DEVELOPMENT_GUIDE.md](./API_DEVELOPMENT_GUIDE.md) for building consistent APIs
5. **Pagination**: [PAGINATION_GUIDE.md](./PAGINATION_GUIDE.md) for list endpoints and data handling

## 🤖 AI Assistant Integration

These guides are specifically structured to help AI assistants:

- **Pattern Recognition**: Clear examples for each layer and use case
- **Context Understanding**: Professional domain separation and error flow
- **Code Generation**: Complete templates and implementation patterns
- **Quality Assurance**: Validation rules and common mistake prevention

## 📋 Development Workflow

### For Repository Development
1. Use professional domain aliases from [DOMAIN_ALIAS_PATTERNS.md](./DOMAIN_ALIAS_PATTERNS.md)
2. Follow GORM error handling patterns from [ERROR_HANDLING_GUIDE.md](./ERROR_HANDLING_GUIDE.md)
3. Test error propagation as shown in examples

### For Service Development
1. Use AppError constructors for business logic errors
2. Convert domain errors to business errors appropriately
3. Maintain clean separation from cross-cutting concerns

### For Handler Development
1. Use structured response handling via `response.Error()`
2. Validate inputs before service calls
3. Let automatic HTTP status mapping handle responses

## 🔧 Code Quality Standards

All code in the Brokle platform follows these industrial patterns:

- **Clean Architecture**: Proper layer separation and error flow
- **Domain Aliases**: Consistent `{domain}Domain` import patterns  
- **Error Context**: Meaningful error messages with relevant context
- **Professional Standards**: Following Go best practices for enterprise applications

## 📖 Related Documentation

- [CLAUDE.md](../CLAUDE.md) - Platform overview and development commands
- [INDUSTRIAL_ERROR_HANDLING_GUIDE.md](../INDUSTRIAL_ERROR_HANDLING_GUIDE.md) - Detailed implementation patterns
- Architecture documentation in `/docs/03-technical-architecture/`

## 🚀 Implementation Status

**✅ Completed**: Full industrial error handling transformation across all layers
- 18+ repository files with professional domain aliases
- 14+ services with structured AppError handling  
- HTTP handlers with automatic response mapping
- Complete dependency injection updates

This documentation ensures consistent, maintainable, and professional code across the entire Brokle platform.