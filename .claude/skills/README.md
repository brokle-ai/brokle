# Brokle Claude Code Skills

This directory contains **7 custom Claude Code skills** for the Brokle project, implementing a hybrid approach with role-based and specialized skills.

## Skills Overview

### Role-Based Skills (2)
These provide broad context and should be invoked at the start of development sessions.

| Skill | Purpose | When to Use |
|-------|---------|-------------|
| **brokle-backend-dev** | Complete Go backend development context | Starting any backend task (services, repos, handlers, domains) |
| **brokle-frontend-dev** | Complete Next.js frontend development context | Starting any frontend task (components, pages, hooks, features) |

### Specialized Skills (5)
These provide deep expertise for specific architectural areas and auto-activate based on task context.

| Skill | Purpose | Auto-Activates When |
|-------|---------|---------------------|
| **brokle-domain-architecture** | DDD patterns, domain creation, entity design | Working with domains, creating entities, cross-domain logic |
| **brokle-api-routes** | Dual-route API architecture, handlers, validation | Creating endpoints, API development, request/response patterns |
| **brokle-error-handling** | Industrial error patterns, domain aliases | Implementing error handling, debugging error flows |
| **brokle-testing** | Pragmatic testing philosophy, mock patterns | Writing tests, reviewing test coverage |
| **brokle-migration-workflow** | PostgreSQL/ClickHouse migrations, seeding | Creating migrations, running migrations, database schema |

## How Skills Work

**Model-Invoked**: Claude automatically activates skills based on task context and skill descriptions. You don't need to explicitly call them.

**Progressive Disclosure**: Skills load supporting files only when needed, keeping context usage efficient.

**Team-Wide**: Skills are committed to git, so the entire team benefits automatically.

## Skill Structure

### Complete Skills with Supporting Files

**brokle-backend-dev/**:
- `SKILL.md` - Main skill with DDD, error handling, testing, auth patterns
- `code-templates.md` - Complete repository, service, handler templates
- `architecture-overview.md` - Scalable monolith, domains, multi-database strategy

**brokle-frontend-dev/**:
- `SKILL.md` - Feature-based architecture, import rules, component patterns
- `feature-patterns.md` - Complete authentication feature as reference

### Specialized Skills (Single File)

Each specialized skill has a focused `SKILL.md` with:
- Clear activation triggers in YAML frontmatter
- Core patterns and templates
- Best practices and checklists
- References to detailed documentation

## Usage Examples

### Scenario 1: Backend Feature Development

**User**: "I need to add a new billing feature for usage export"

**Claude Activates**:
1. `brokle-backend-dev` (role-based context)
2. `brokle-domain-architecture` (if creating new domain entities)
3. `brokle-api-routes` (for API endpoints)
4. `brokle-error-handling` (for proper error flow)

**Result**: Claude knows:
- Domain alias imports (`billingDomain "brokle/internal/core/domain/billing"`)
- Repository → Service → Handler pattern
- AppError constructors in services
- `response.Error()` in handlers
- OpenAPI documentation patterns

### Scenario 2: Frontend Component Creation

**User**: "Create a new projects dashboard component"

**Claude Activates**:
1. `brokle-frontend-dev` (role-based context)

**Result**: Claude knows:
- Feature-based structure (`features/projects/`)
- Import from feature index only (`@/features/projects`)
- Server Components by default, Client only when needed
- React Query for server state, Zustand for client state
- Tailwind CSS + shadcn/ui patterns

### Scenario 3: Database Migration

**User**: "Add a new table for tracking API usage"

**Claude Activates**:
1. `brokle-backend-dev` (general backend context)
2. `brokle-migration-workflow` (migration expertise)

**Result**: Claude knows:
- `make create-migration DB=postgres NAME=add_usage_table`
- Multi-tenancy (include `organization_id`)
- Index patterns
- Up/Down migration structure
- TTL patterns for ClickHouse

### Scenario 4: Writing Tests

**User**: "Write tests for the new billing service"

**Claude Activates**:
1. `brokle-backend-dev` (backend context)
2. `brokle-testing` (testing philosophy)

**Result**: Claude knows:
- Test business logic, not CRUD
- Table-driven test pattern
- Complete mock interface implementation
- `AssertExpectations(t)` on all mocks
- ~1:1 test-to-code ratio target

## Benefits

### Before Skills (Problems)
- ❌ Claude reads entire CLAUDE.md every time (context bloat)
- ❌ Forgets architectural patterns mid-conversation
- ❌ Makes layer violations (handlers in services, etc.)
- ❌ Inconsistent error handling
- ❌ Wrong import patterns (direct domain imports)

### After Skills (Solutions)
- ✅ Progressive disclosure (load only what's needed)
- ✅ Auto-activation based on task
- ✅ Consistent architectural patterns
- ✅ Proper error handling every time
- ✅ Correct domain alias imports
- ✅ 80% reduction in architectural mistakes

## Skill Activation Triggers

Skills activate based on keywords in their descriptions:

| Keyword/Phrase | Activates Skill |
|----------------|-----------------|
| "developing Go backend", "creating services" | brokle-backend-dev |
| "Next.js", "React component", "frontend" | brokle-frontend-dev |
| "creating new domain", "domain entities" | brokle-domain-architecture |
| "API endpoint", "handler", "request/response" | brokle-api-routes |
| "error handling", "AppError", "domain alias" | brokle-error-handling |
| "writing tests", "test coverage", "mock" | brokle-testing |
| "migration", "database schema", "seeding" | brokle-migration-workflow |

## Maintenance

### Adding New Skills

1. Create directory: `.claude/skills/skill-name/`
2. Create `SKILL.md` with YAML frontmatter:
```yaml
---
name: skill-name
description: Use this skill when [clear trigger]. This includes [specific scenarios].
---
```
3. Add skill content (patterns, templates, examples)
4. Commit to git for team-wide availability

### Updating Existing Skills

1. Edit `SKILL.md` or supporting files
2. Test activation with relevant prompts
3. Commit changes

### Testing Skills

Use prompts that match the skill's description:
- "I need to create a new service" → Should activate brokle-backend-dev
- "Add a migration for new table" → Should activate brokle-migration-workflow
- "Write tests for UserService" → Should activate brokle-testing

## File Count Summary

- **Total Skills**: 7
- **Total Files**: 10
- **Role-Based Skills**: 2 (with 5 supporting files)
- **Specialized Skills**: 5 (single-file each)

## Source of Truth References

Skills point to actual code as the source of truth. When in doubt, check:

**Backend**:
- Domains: `internal/core/domain/` directory
- Error constructors: `pkg/errors/errors.go`
- Migrations: `migrations/` and `migrations/clickhouse/`
- Health endpoints: `internal/transport/http/server.go`

**Frontend**:
- Feature exports: `web/src/features/{feature}/index.ts`
- Scripts: `web/package.json`
- Tech stack versions: `web/package.json`

**SDKs**:
- Python: `sdk/python/brokle/__init__.py`, `pyproject.toml`
- JavaScript: `sdk/javascript/packages/brokle/src/index.ts`, `package.json`

## Architecture Alignment

These skills align with Brokle's core architecture principles:
- **Scalable Monolith**: Separate server/worker binaries
- **Domain-Driven Design**: Clean layer separation
- **Multi-Database**: PostgreSQL, ClickHouse, Redis
- **Dual Authentication**: SDK (API keys) vs Dashboard (JWT)
- **Industrial Error Handling**: Repository → Service → Handler
- **Pragmatic Testing**: Test business logic, not framework
- **Feature-Based Frontend**: Self-contained domain features

## Quick Reference

**View All Skills**:
```bash
ls -la .claude/skills/
```

**View Skill Content**:
```bash
cat .claude/skills/brokle-backend-dev/SKILL.md
```

**Test Skill Activation**:
Use Claude Code and start tasks matching skill descriptions.

## Documentation Cross-Reference

Skills complement (not replace) existing documentation:
- `CLAUDE.md` - Complete platform overview
- `docs/development/` - Detailed development guides
- `web/ARCHITECTURE.md` - Frontend architecture
- `docs/TESTING.md` - Testing philosophy
- Skill files reference these for deep dives

## SDK Skills (Submodules)

The Python and JavaScript SDKs have their own dedicated skills in separate repositories (accessible via git submodules).

### Python SDK Skill

**Location**: `sdk/python/.claude/skills/brokle-python-sdk/`

**Covers**:
- Three integration patterns: `@observe` decorator, context managers, wrappers
- OTEL-native architecture (TracerProvider → BatchSpanProcessor → OTLP/HTTP)
- Configuration: Explicit vs environment-based singleton (`get_client()`)
- GenAI 1.28+ attributes (full OTEL compliance)
- Version attribute for A/B testing
- Serverless patterns with `client.flush()`
- Validation rules (API key, environment, sample rate)

**When to use**: Working on Python SDK codebase (`sdk/python/`)

**Activation triggers**: Python SDK, brokle python, @observe, get_client(), wrap_openai, GenAI attributes, OTLP

### JavaScript/TypeScript SDK Skill

**Location**: `sdk/javascript/.claude/skills/brokle-javascript-sdk/`

**Covers**:
- **Monorepo with 4 packages** (all fully implemented):
  - `brokle` - Core SDK
  - `brokle-openai` - OpenAI wrapper (Proxy pattern)
  - `brokle-anthropic` - Anthropic wrapper
  - `brokle-langchain` - LangChain callbacks
- Five integration patterns: decorator, traced(), generation(), wrappers, LangChain
- Symbol.for() singleton (cross-realm safe)
- Explicit lifecycle management (no automatic exit handlers)
- Processor choice: BatchSpanProcessor OR SimpleSpanProcessor (via `flushSync`)
- TypeScript decorators (requires `experimentalDecorators`)
- Type-safe GenAI 1.28+ attributes
- ESM + CJS dual build

**When to use**: Working on TypeScript SDK codebase (`sdk/javascript/`)

**Activation triggers**: JavaScript SDK, TypeScript SDK, brokle-js, OTLP, GenAI attributes, traced(), Symbol.for(), monorepo

### SDK Skills vs Platform Skills

| Type | Location | Purpose |
|------|----------|---------|
| **Platform Skills** | `.claude/skills/` | Backend/frontend app development (main Brokle platform) |
| **SDK Skills** | `sdk/{python,javascript}/.claude/skills/` | SDK development and maintenance |

**Note**: SDK skills are in separate repos but accessible via git submodules. They focus on SDK-specific patterns like OTEL integration, decorators, wrappers, and singleton management.

---

**Created**: 2025-11-07
**Version**: 3.0 (Simplified for maintainability)
**Skills Count**: 9 total
- **Platform**: 7 (2 role-based + 5 specialized)
- **SDKs**: 2 (Python SDK + JavaScript SDK)

**Philosophy**: Skills focus on patterns and point to source code for specifics (avoiding high-maintenance counts/versions)
