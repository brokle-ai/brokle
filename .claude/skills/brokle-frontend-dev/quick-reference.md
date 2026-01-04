# Quick Reference

Decision trees and cheat sheets for Brokle frontend development.

## Decision Trees

### Creating New Functionality

```
New functionality needed?
├─ Feature-specific? → Create in features/[feature]/
├─ Shared across features? → Create in components/shared/
├─ UI primitive? → Use shadcn/ui or components/ui/
└─ Utility logic? → Create in lib/ or hooks/
```

### Adding State

```
What kind of state?
├─ Server data? → React Query in feature hook
├─ Client-only? → Zustand store in feature
├─ Form state? → React Hook Form + Zod
└─ URL state? → useSearchParams()
```

### Component Type

```
Component needs?
├─ Interactivity (onClick, onChange)? → 'use client'
├─ React hooks (useState, useEffect)? → 'use client'
├─ Browser APIs (localStorage, window)? → 'use client'
└─ Static content/SEO? → Server Component (default)
```

## Import Rules Cheat Sheet

```typescript
// ALLOWED
import { useAuth } from '@/features/authentication'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { apiClient } from '@/lib/api/core/client'

// FORBIDDEN
import { useAuth } from '@/features/authentication/hooks/use-auth'
import { AuthStore } from '@/features/authentication/stores/auth-store'
```

**Rule**: Import from `@/features/[feature]`, never internal paths.

## Feature Index Template

```typescript
// features/my-feature/index.ts

// Hooks
export { useMyFeature } from './hooks/use-my-feature'

// Components
export { MyComponent } from './components/my-component'

// Types (selective)
export type { MyFeatureData } from './types'

// DO NOT export stores directly (use hooks)
```

## Development Commands

```bash
cd web

pnpm dev          # Dev server with Turbopack
pnpm build        # Production build
pnpm lint         # Next.js linter
pnpm format       # Format with Prettier
pnpm test         # Run tests (Vitest)
pnpm test:watch   # Watch mode
```

## Best Practices Summary

1. **Import from feature index** - Never from internal paths
2. **Keep routes thin** - Delegate to feature components
3. **Server Components default** - Only add 'use client' when needed
4. **Type everything** - No `any` types
5. **Error boundaries** - Add per feature and route
6. **Loading states** - Add `loading.tsx` for async routes

## File Organization

```
features/[feature]/
├── components/     # Feature UI
├── hooks/          # React hooks
├── api/            # API functions
├── stores/         # Zustand (optional)
├── types/          # TypeScript types
├── __tests__/      # Tests
└── index.ts        # Public API
```

## Common Patterns

### API Function

```typescript
export async function getData(id: string): Promise<Response> {
  const response = await apiClient.get<Response>(`/endpoint/${id}`)
  return response.data
}
```

### Query Hook

```typescript
export function useData(id: string) {
  return useQuery({
    queryKey: ['data', id],
    queryFn: () => getData(id),
  })
}
```

### Mutation Hook

```typescript
const queryClient = useQueryClient()
const mutation = useMutation({
  mutationFn: updateData,
  onSuccess: () => queryClient.invalidateQueries({ queryKey: ['data'] }),
})
```

## shadcn/ui Integration

Use MCP tools to add components:

```bash
# Search for components
mcp__shadcn__search_items_in_registries

# Get add command
mcp__shadcn__get_add_command_for_items
```
