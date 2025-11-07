# RBAC Integration TODO

## Current State

Permissions are currently **stubbed** in `web/src/hooks/rbac/route-rbac-utils.ts`.

All scopes return `true` for UI development, allowing all routes to be visible regardless of user permissions.

## Implementation Checklist

### Backend

- [ ] Create `/api/v1/rbac/check` endpoint
  - [ ] Accept POST with `{ scope: string, projectId?: string }`
  - [ ] Return `{ hasAccess: boolean }`
  - [ ] Implement permission checking logic
  - [ ] Add caching strategy (Redis?)

### Frontend

- [ ] Replace stub in `useRoutePermissions` hook
  - [ ] Implement with `useQueries` from TanStack Query
  - [ ] Add proper error handling
  - [ ] Add retry logic
  - [ ] Configure stale time and cache duration

### Testing

- [ ] Test permission filtering with real data
- [ ] Test loading states
- [ ] Test error states
- [ ] Test permission updates (real-time?)

### Documentation

- [ ] Document RBAC patterns in ARCHITECTURE.md
- [ ] Add examples of adding new protected routes
- [ ] Document permission caching strategy

## Example Implementation

### Option 1: API-based with useQueries (Recommended)

```typescript
export function useRoutePermissions(
  scopes: Scope[],
  projectId?: string | null
): {
  permissions: Record<string, boolean>
  isLoading: boolean
} {
  const queries = useQueries({
    queries: scopes.map(scope => ({
      queryKey: ['rbac', 'permission', scope, projectId],
      queryFn: async () => {
        const response = await fetch('/api/v1/rbac/check', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ scope, projectId }),
        })
        if (!response.ok) throw new Error('RBAC check failed')
        const data = await response.json()
        return data.hasAccess === true
      },
      staleTime: 60000, // 1 min
      gcTime: 5 * 60 * 1000, // 5 min
    })),
  })

  const isLoading = queries.some(q =>
    q.isLoading || q.isPending || q.isFetching
  )

  const permissions = useMemo(() => {
    const map: Record<string, boolean> = {}
    scopes.forEach((scope, index) => {
      map[scope] = queries[index]?.data === true
    })
    return map
  }, [scopes, queries])

  return { permissions, isLoading }
}
```

### Option 2: Session-based (If permissions are preloaded)

```typescript
export function useRoutePermissions(
  scopes: Scope[],
  projectId?: string | null
): {
  permissions: Record<string, boolean>
  isLoading: boolean
} {
  const session = useSession()
  const workspace = useWorkspace()

  const isLoading = session.status === 'loading' || workspace.isLoading

  const permissions = useMemo(() => {
    if (isLoading) return {}

    const map: Record<string, boolean> = {}
    scopes.forEach(scope => {
      // Check if user has this scope
      map[scope] = session?.user?.permissions?.includes(scope) ?? false
    })
    return map
  }, [scopes, session, isLoading])

  return { permissions, isLoading }
}
```

## Files to Update

1. **`web/src/hooks/rbac/route-rbac-utils.ts`** - Replace stub with real implementation
2. **Backend API** - Implement `/api/v1/rbac/check` endpoint
3. **Tests** - Add tests for RBAC filtering
4. **Documentation** - Update ARCHITECTURE.md with RBAC patterns

## Migration Strategy

1. **Phase 1**: Keep stub, implement backend endpoint
2. **Phase 2**: Test endpoint with Postman/curl
3. **Phase 3**: Replace stub with real implementation
4. **Phase 4**: Test frontend permission filtering
5. **Phase 5**: Add caching and optimization

## Security Considerations

- ⚠️ Current stub grants ALL permissions - not suitable for production
- ✅ Once implemented, use pessimistic defaults (deny by default)
- ✅ Show sidebar skeleton during permission loading
- ✅ Cache permissions to reduce API calls
- ✅ Handle permission updates gracefully

## Performance Considerations

- Use React Query for automatic caching
- Batch permission checks where possible
- Consider WebSocket for real-time permission updates
- Monitor API call frequency and optimize if needed
