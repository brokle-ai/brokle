import { createFileRoute, Outlet } from '@tanstack/react-router'

// Prompts shell. The index route (`prompts/index.tsx`) renders the
// list; `$promptId`, `$promptId/edit`, and `new` render detail/edit/
// create views. Keeping the shell thin keeps search-param state on
// the leaf route where it belongs (see CRITICAL cascade note in
// the migration handoff — each child MUST install its own
// `validateSearch: z.object({}).catch({})` to decouple).
export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/prompts',
)({
  component: PromptsShell,
})

function PromptsShell() {
  return <Outlet />
}
