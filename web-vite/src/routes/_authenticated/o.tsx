import { createFileRoute, Outlet } from '@tanstack/react-router'

// Layout-only segment for `/o/*`. The exact `/o` (org picker) lives
// in `o.index.tsx`; child routes (`o.$orgId.tsx`, transitively
// `o.$orgId/p.$projectId.tsx`, etc.) render through this Outlet.
//
// Without this split, `o.tsx` would render the picker directly and
// override every nested URL under `/o/*` — the picker would have no
// <Outlet />, so children couldn't mount and the URL would update
// while the page still showed the org list.
//
// Mirrors `o.$orgId.tsx` and Opik's
// `competitors/opik/apps/opik-frontend/src/v1/router.tsx:217-235`.
export const Route = createFileRoute('/_authenticated/o')({
  component: () => <Outlet />,
})
