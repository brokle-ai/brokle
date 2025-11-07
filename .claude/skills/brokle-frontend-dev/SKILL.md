---
name: brokle-frontend-dev
description: Use this skill when developing, implementing, or modifying Next.js/React frontend code for the Brokle web application. This includes creating components, pages, hooks, API clients, stores, or any frontend features. Invoke this skill at the start of frontend development tasks.
---

# Brokle Frontend Development Skill

This skill provides comprehensive guidance for Next.js/React frontend development following Brokle's feature-based architecture.

## Tech Stack

- **Next.js**: 15.5.2 (App Router, Turbopack)
- **React**: 19.2.0
- **TypeScript**: 5.9.3 (strict mode)
- **Styling**: Tailwind CSS 4.1.15
- **Components**: shadcn/ui
- **State Management**: Zustand (client state) + React Query (server state)
- **Forms**: React Hook Form + Zod validation
- **Testing**: Vitest + React Testing Library + MSW
- **Package Manager**: pnpm
- **Port**: :3000

## Feature-Based Architecture

### Directory Structure

```
web/src/
├── app/                      # Next.js App Router (routing only)
│   ├── (auth)/              # Auth route group
│   ├── (dashboard)/         # Dashboard routes
│   ├── (onboarding)/        # Onboarding wizard
│   └── (errors)/            # Error pages
├── features/                # Domain features (self-contained)
│   ├── authentication/      # User auth, sessions, OAuth
│   ├── organizations/       # Org management, members, invitations
│   ├── projects/           # Project dashboard, API keys, settings
│   ├── analytics/          # Usage analytics and metrics
│   ├── billing/            # Billing and subscription management
│   ├── gateway/            # AI gateway configuration
│   ├── settings/           # User settings and preferences
│   └── tasks/              # Task management
├── components/              # Shared components only
│   ├── ui/                 # shadcn/ui primitives
│   ├── layout/             # App shell (header, sidebar, footer)
│   ├── guards/             # Auth guards
│   ├── shared/             # Generic reusable components
│   ├── navigation/         # Navigation components
│   ├── notifications/      # Notification components
│   ├── error-boundary/     # Error boundaries
│   ├── audit/              # Audit components
│   ├── collaboration/      # Collaboration components
│   ├── data/               # Data components
│   ├── templates/          # Template components
│   └── wizard/             # Wizard components
├── lib/                    # Core infrastructure
│   ├── api/core/           # BrokleAPIClient (HTTP client)
│   ├── auth/               # JWT utilities
│   └── utils/              # Pure utilities
├── hooks/                  # Global hooks (use-mobile, etc.)
├── stores/                 # Global stores (ui-store.ts)
├── context/                # Cross-feature context (workspace-context)
├── types/                  # Shared types
├── assets/                 # Static assets (logos, icons, SVGs)
│   ├── brand-icons/        # Provider/brand icons
│   └── custom/             # Custom graphics
├── utils/                  # Small utilities
└── __tests__/              # Test infrastructure (MSW, utilities)
```

### Feature Structure

**Note**: Check `web/src/features/{feature}/index.ts` for current exports and implementation status.

Each feature in `features/[feature]/` has:
- `components/` - Feature-specific UI components
- `hooks/` - Feature-specific React hooks
- `api/` - API functions for this domain
- `stores/` - Zustand stores (optional)
- `types/` - TypeScript definitions
- `__tests__/` - Feature tests
- `index.ts` - Public API exports (ONLY way to import)

## Critical Import Rules (MANDATORY)

### ✅ Allowed Imports

```typescript
// Import from feature public API
import { useAuth, SignInForm } from '@/features/authentication'
import { useOrganization } from '@/features/organizations'

// Shared components
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'

// Utilities
import { cn } from '@/lib/utils'
import { apiClient } from '@/lib/api/core/client'
```

### ❌ Forbidden Imports

```typescript
// DON'T import internal feature files directly
import { useAuth } from '@/features/authentication/hooks/use-auth'  // ❌

// DON'T import from other features' internals
import { AuthStore } from '@/features/authentication/stores/auth-store'  // ❌

// DON'T bypass feature index
import { LoginForm } from '@/features/authentication/components/login-form'  // ❌
```

**Rule**: Always import from `@/features/[feature]` (feature index), never from internal paths.

## Component Patterns

### Server Components (Default)

```typescript
// No 'use client' directive
export default async function Page({ params }: { params: { id: string } }) {
  // Can fetch data server-side
  const data = await fetchData(params.id)

  return (
    <div>
      <ServerComponent data={data} />
      <ClientComponent />
    </div>
  )
}
```

**When to use**:
- Default choice for all components
- Static content
- SEO-friendly pages
- Data fetching at build/request time

### Client Components (When Needed)

```typescript
'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'

export function InteractiveComponent() {
  const [count, setCount] = useState(0)

  return (
    <Button onClick={() => setCount(count + 1)}>
      Count: {count}
    </Button>
  )
}
```

**When to use**:
- Event handlers (onClick, onChange, etc.)
- React hooks (useState, useEffect, useContext)
- Browser APIs (localStorage, window, etc.)
- Third-party libraries requiring browser

## State Management Strategy

### 1. Server State (React Query)
Managed in `context/` providers:

```typescript
// context/workspace-context.tsx
'use client'

import { createContext, useContext } from 'react'
import { useQuery } from '@tanstack/react-query'

export function WorkspaceProvider({ children }: { children: React.Node }) {
  const { data: workspace } = useQuery({
    queryKey: ['workspace'],
    queryFn: fetchWorkspace,
  })

  return (
    <WorkspaceContext.Provider value={{ workspace }}>
      {children}
    </WorkspaceContext.Provider>
  )
}
```

### 2. Client State (Zustand)
Feature-specific stores in `features/[feature]/stores/`:

```typescript
// features/authentication/stores/auth-store.ts
import { create } from 'zustand'

interface AuthState {
  isAuthenticated: boolean
  user: User | null
  setUser: (user: User | null) => void
}

export const useAuthStore = create<AuthState>((set) => ({
  isAuthenticated: false,
  user: null,
  setUser: (user) => set({ user, isAuthenticated: !!user }),
}))
```

### 3. URL State
Use `useSearchParams()` for filters, pagination:

```typescript
'use client'

import { useSearchParams, useRouter } from 'next/navigation'

export function FilterComponent() {
  const searchParams = useSearchParams()
  const router = useRouter()

  const filter = searchParams.get('filter') ?? 'all'

  const setFilter = (newFilter: string) => {
    const params = new URLSearchParams(searchParams)
    params.set('filter', newFilter)
    router.push(`?${params.toString()}`)
  }

  return <div>Filter: {filter}</div>
}
```

### 4. Form State
React Hook Form + Zod validation:

```typescript
'use client'

import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'

const schema = z.object({
  email: z.string().email('Invalid email'),
  password: z.string().min(8, 'At least 8 characters'),
})

type FormData = z.infer<typeof schema>

export function LoginForm() {
  const { register, handleSubmit, formState: { errors } } = useForm<FormData>({
    resolver: zodResolver(schema),
  })

  const onSubmit = (data: FormData) => {
    console.log(data)
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <input {...register('email')} />
      {errors.email && <span>{errors.email.message}</span>}
    </form>
  )
}
```

## Feature Templates

### Feature Index Pattern

```typescript
// features/my-feature/index.ts

// Hooks
export { useMyFeature } from './hooks/use-my-feature'
export { useMyFeatureData } from './hooks/use-my-feature-data'

// Components
export { MyComponent } from './components/my-component'
export { MyForm } from './components/my-form'

// Types (selective export)
export type { MyFeatureData, MyFeatureConfig } from './types'

// DO NOT export internal implementation details
// DO NOT export stores directly (use hooks instead)
```

### Component Template

```typescript
// features/my-feature/components/my-component.tsx
'use client'

import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { useMyFeature } from '../hooks/use-my-feature'

interface MyComponentProps {
  id: string
  onComplete?: () => void
}

export function MyComponent({ id, onComplete }: MyComponentProps) {
  const { data, isLoading, error } = useMyFeature(id)

  if (isLoading) return <div>Loading...</div>
  if (error) return <div>Error: {error.message}</div>

  return (
    <Card>
      <h2 className="text-2xl font-bold">{data?.title}</h2>
      <Button onClick={onComplete}>Complete</Button>
    </Card>
  )
}
```

### API Client Pattern

```typescript
// features/my-feature/api/my-feature-api.ts
import { apiClient } from '@/lib/api/core/client'
import type { MyFeatureResponse, CreateMyFeatureRequest } from '../types'

export async function getMyFeature(id: string): Promise<MyFeatureResponse> {
  const response = await apiClient.get<MyFeatureResponse>(`/my-feature/${id}`)
  return response.data
}

export async function createMyFeature(
  data: CreateMyFeatureRequest
): Promise<MyFeatureResponse> {
  const response = await apiClient.post<MyFeatureResponse>('/my-feature', data)
  return response.data
}

export async function updateMyFeature(
  id: string,
  data: Partial<CreateMyFeatureRequest>
): Promise<MyFeatureResponse> {
  const response = await apiClient.put<MyFeatureResponse>(`/my-feature/${id}`, data)
  return response.data
}

export async function deleteMyFeature(id: string): Promise<void> {
  await apiClient.delete(`/my-feature/${id}`)
}
```

### Hook Pattern

```typescript
// features/my-feature/hooks/use-my-feature.ts
'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getMyFeature, updateMyFeature } from '../api/my-feature-api'

export function useMyFeature(id: string) {
  const queryClient = useQueryClient()

  const query = useQuery({
    queryKey: ['my-feature', id],
    queryFn: () => getMyFeature(id),
  })

  const updateMutation = useMutation({
    mutationFn: (data: UpdateData) => updateMyFeature(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['my-feature', id] })
    },
  })

  return {
    data: query.data,
    isLoading: query.isLoading,
    error: query.error,
    update: updateMutation.mutate,
    isUpdating: updateMutation.isPending,
  }
}
```

### Store Pattern (Optional)

```typescript
// features/my-feature/stores/my-feature-store.ts
import { create } from 'zustand'
import type { MyFeatureState } from '../types'

interface MyFeatureStore extends MyFeatureState {
  setData: (data: MyFeatureState) => void
  reset: () => void
}

const initialState: MyFeatureState = {
  items: [],
  selectedId: null,
}

export const useMyFeatureStore = create<MyFeatureStore>((set) => ({
  ...initialState,
  setData: (data) => set(data),
  reset: () => set(initialState),
}))
```

## Testing Pattern

```typescript
// features/my-feature/__tests__/my-component.test.tsx
import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MyComponent } from '../components/my-component'

describe('MyComponent', () => {
  it('renders successfully', () => {
    render(<MyComponent id="123" />)
    expect(screen.getByText('Expected Text')).toBeInTheDocument()
  })

  it('handles loading state', () => {
    render(<MyComponent id="123" />)
    expect(screen.getByText('Loading...')).toBeInTheDocument()
  })
})
```

## Development Commands

From `web/package.json:5-16`:

```bash
cd web

# Install dependencies
pnpm install

# Development
pnpm dev                   # Start dev server (Next.js with Turbopack)

# Build
pnpm build                 # Build for production
pnpm start                 # Start production server

# Code quality
pnpm lint                  # Next.js linter
pnpm format                # Format with Prettier
pnpm format:check          # Check formatting

# Testing
pnpm test                  # Run tests (Vitest)
pnpm test:watch            # Watch mode
pnpm test:ui               # Vitest UI
pnpm test:coverage         # Coverage report
```

## Best Practices

### 1. Import from Feature Index Only
```typescript
// ✅ Correct
import { useAuth } from '@/features/authentication'

// ❌ Wrong
import { useAuth } from '@/features/authentication/hooks/use-auth'
```

### 2. Keep Routes Thin
```typescript
// app/(dashboard)/projects/page.tsx
import { ProjectsList } from '@/features/projects'

export default function ProjectsPage() {
  return <ProjectsList />  // Delegate to feature component
}
```

### 3. Use Server Components by Default
Only add `'use client'` when absolutely necessary.

### 4. Type Everything
No `any` types. Use TypeScript strict mode.

### 5. Error Boundaries
Add error boundaries per feature and route:

```typescript
// app/(dashboard)/error.tsx
'use client'

export default function Error({ error, reset }: {
  error: Error
  reset: () => void
}) {
  return (
    <div>
      <h2>Something went wrong!</h2>
      <button onClick={reset}>Try again</button>
    </div>
  )
}
```

### 6. Loading States
Add `loading.tsx` for async routes:

```typescript
// app/(dashboard)/projects/loading.tsx
export default function Loading() {
  return <div>Loading projects...</div>
}
```

### 7. Test Critical Paths
Focus on:
- Authentication flows
- Navigation
- API calls
- User interactions
- Error states

### 8. shadcn/ui Integration
Use MCP tools to add components:

```bash
# Use shadcn MCP tools
mcp__shadcn__search_items_in_registries
mcp__shadcn__get_add_command_for_items
```

## Quick Decision Tree

**Creating new functionality?**
1. Is it feature-specific? → Create in `features/[feature]/`
2. Is it shared across features? → Create in `components/shared/`
3. Is it a UI primitive? → Use shadcn/ui or create in `components/ui/`
4. Is it utility logic? → Create in `lib/` or `hooks/`

**Adding state?**
1. Server data? → Use React Query in feature hook
2. Client-only state? → Use Zustand store in feature
3. Form state? → Use React Hook Form + Zod
4. URL state? → Use useSearchParams()

**Component decisions?**
1. Needs interactivity? → `'use client'`
2. Static content? → Server Component (default)
3. SEO important? → Server Component
4. Browser APIs needed? → `'use client'`

## Documentation

- **Architecture**: `web/ARCHITECTURE.md`
- **Backend API**: `CLAUDE.md` (API routes section)
- **Existing Features**: `web/src/features/` for reference patterns
