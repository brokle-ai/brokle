# Frontend Architecture

## Overview

The Brokle frontend follows a **feature-based architecture** that aligns with Next.js 15 best practices while maintaining clear domain boundaries.

## Directory Structure

```
web/src/
├── app/                          # Next.js App Router (routing only)
│   ├── (auth)/                   # Auth route group
│   ├── (dashboard)/              # Protected dashboard routes
│   └── (errors)/                 # Error pages
│
├── features/                     # Domain features (self-contained)
│   ├── authentication/
│   │   ├── components/           # Auth UI components
│   │   ├── hooks/                # Auth hooks
│   │   ├── stores/               # Auth Zustand store
│   │   ├── api/                  # Auth API functions
│   │   ├── types/                # Auth TypeScript types
│   │   ├── __tests__/            # Feature tests
│   │   └── index.ts              # Public exports
│   ├── organizations/
│   ├── projects/
│   ├── analytics/
│   ├── billing/
│   ├── gateway/
│   └── settings/
│
├── components/                   # Shared components only
│   ├── ui/                      # shadcn/ui primitives
│   ├── layout/                  # App shell (header, sidebar, footer)
│   ├── guards/                  # Auth guards (re-exports from features)
│   └── shared/                  # Generic reusable components
│
├── lib/                         # Core infrastructure
│   ├── api/                     # API client
│   │   └── core/
│   │       ├── client.ts        # BrokleAPIClient class
│   │       └── types.ts         # API types
│   ├── auth/                    # Auth utilities
│   ├── errors/                  # Error handling (planned)
│   └── utils/                   # Pure utilities
│
├── hooks/                       # Global hooks
│   ├── use-mobile.ts
│   └── use-media-query.ts
│
├── stores/                      # Global stores
│   └── ui-store.ts              # Global UI state
│
├── context/                     # Cross-feature context
│   └── workspace-context.tsx    # Organization/project context
│
├── types/                       # Shared types
│   └── api-responses.ts
│
└── __tests__/                   # Test infrastructure
    ├── mocks/                   # MSW request handlers
    ├── utils/                   # Test utilities
    └── integration/             # Integration tests
```

## Architecture Principles

### 1. Feature Self-Containment
Each feature in `features/` is self-contained with:
- **components/**: Feature-specific UI
- **hooks/**: Feature-specific hooks
- **api/**: API functions for this domain
- **stores/**: Zustand stores (if needed)
- **types/**: TypeScript definitions
- **__tests__/**: Feature tests
- **index.ts**: Public API (only way to import from feature)

### 2. Import Rules

**✅ Allowed:**
```typescript
// Import from feature public API
import { useAuth, SignInForm } from '@/features/authentication'

// Shared components
import { Button } from '@/components/ui/button'

// Utilities
import { cn } from '@/lib/utils'
```

**❌ Forbidden:**
```typescript
// DON'T import internal feature files
import { useAuth } from '@/features/authentication/hooks/use-auth'

// DON'T import from other features' internals
import { AuthStore } from '@/features/authentication/stores/auth-store'
```

### 3. State Management Strategy

**Server State** (React Query):
- Managed in `context/` providers
- Wraps React Query for convenience
- Example: `workspace-context.tsx`

**Client State** (Zustand):
- Feature-specific stores in `features/[feature]/stores/`
- Global UI state in `stores/ui-store.ts`
- Example: `authentication/stores/auth-store.ts`

**URL State**:
- Use `useSearchParams()` for filters, pagination
- Managed at route level

**Form State**:
- React Hook Form + Zod validation
- Scoped to components

### 4. Routes Are Thin

Pages in `app/` should be minimal - just import and compose from features:

```tsx
// app/(dashboard)/projects/[projectSlug]/page.tsx
import { ProjectDashboard } from '@/features/projects'

export default function ProjectPage({ params }) {
  return <ProjectDashboard projectSlug={params.projectSlug} />
}
```

### 5. Server Components by Default

Keep pages as Server Components when possible:
```tsx
// Server Component (no 'use client')
export default async function Page({ params }) {
  // Can fetch data server-side
  return <ClientComponent data={data} />
}
```

Only use `'use client'` when you need:
- Event handlers
- React hooks (useState, useEffect)
- Browser APIs

## Tech Stack

- **Framework**: Next.js 15.5.2 (App Router, Turbopack)
- **React**: 19.2.0
- **TypeScript**: 5.9.3 (strict mode enabled)
- **Styling**: Tailwind CSS 4.1.15
- **UI Components**: shadcn/ui
- **State Management**: Zustand (client) + React Query (server)
- **Forms**: React Hook Form + Zod
- **Testing**: Vitest + React Testing Library + MSW
- **Package Manager**: pnpm

## Adding New Features

### 1. Create Feature Structure
```bash
cd web/src
mkdir -p features/my-feature/{components,hooks,api,types,__tests__/{api,hooks,components}}
touch features/my-feature/index.ts
```

### 2. Implement Feature
- Add components in `components/`
- Add hooks in `hooks/`
- Add API functions in `api/`
- Add types in `types/`

### 3. Create Public API
Edit `index.ts`:
```typescript
// Public exports for my-feature

// Hooks
export { useMyFeature } from './hooks/use-my-feature'

// Components
export { MyComponent } from './components/my-component'

// Types
export type { MyType } from './types'
```

### 4. Use in Routes
```typescript
// app/some-route/page.tsx
import { MyComponent } from '@/features/my-feature'
```

### 5. Add Tests
Create tests in `__tests__/{api,hooks,components}/`

## Testing Strategy

**Coverage Target**: 30-40% (critical paths)

**Test Structure**:
```
features/authentication/__tests__/
├── api/
│   └── auth-api.test.ts          # API function tests
├── hooks/
│   └── use-auth.test.ts          # Hook tests
└── components/
    └── login-form.test.tsx       # Component tests
```

**Test Utilities**:
- `__tests__/utils/test-utils.tsx` - Custom render with providers
- `__tests__/utils/factories.ts` - Test data factories
- `__tests__/mocks/handlers.ts` - MSW request handlers

**Run Tests**:
```bash
pnpm test              # Run all tests
pnpm test:watch        # Watch mode
pnpm test:ui           # Vitest UI
pnpm test:coverage     # With coverage report
```

## Build & Development

```bash
# Development
pnpm dev               # Start dev server (Turbopack)

# Type checking
pnpm tsc --noEmit      # Check types

# Linting
pnpm lint              # ESLint

# Testing
pnpm test              # Run tests

# Build
pnpm build             # Production build
```

## Migration Status

### ✅ Migrated Features
- Authentication (10 components, 4 hooks, 1 store, 1 API)
- Organizations (7 components, 2 hooks, 1 API)
- Projects (4 components, 1 hook, 1 store, 1 API)
- Analytics (1 component, 1 API)
- Settings (7 components)

### 📋 Remaining Work
- Fix TypeScript strict mode errors
- Add comprehensive tests (current: 2 tests)
- Add error boundaries per feature
- Add loading states per route
- Performance optimization (code splitting)
- Accessibility audit

## Key Files

- `tsconfig.json` - TypeScript config with strict mode + path aliases
- `vitest.config.ts` - Test configuration
- `next.config.ts` - Next.js configuration
- `tailwind.config.ts` - Tailwind CSS configuration
- `proxy.ts` - Next.js middleware (auth, CSRF)

## Feature Index

| Feature | Components | Hooks | API | Store |
|---------|-----------|-------|-----|-------|
| authentication | 12 | 4 | ✅ | ✅ |
| organizations | 7 | 2 | ✅ | - |
| projects | 4 | 1 | ✅ | ✅ |
| analytics | 1 | - | ✅ | - |
| billing | - | - | - | - |
| gateway | - | - | - | - |
| settings | 7 | - | - | - |

## Best Practices

1. **Always import from feature index**: `@/features/[feature]`
2. **Keep routes thin**: Delegate to feature components
3. **Use Server Components**: Default to server, add `'use client'` only when needed
4. **Test critical paths**: Focus on auth, navigation, API calls
5. **Type everything**: No `any` types, enable strict mode
6. **Error boundaries**: Add per feature and route
7. **Loading states**: Add `loading.tsx` for async routes

## Resources

- [Next.js Documentation](https://nextjs.org/docs)
- [React Documentation](https://react.dev)
- [Tailwind CSS](https://tailwindcss.com)
- [shadcn/ui](https://ui.shadcn.com)
- [Zustand](https://zustand.docs.pmnd.rs)
- [React Query](https://tanstack.com/query/latest)
- [Vitest](https://vitest.dev)
