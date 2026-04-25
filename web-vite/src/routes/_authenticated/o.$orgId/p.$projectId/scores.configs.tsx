import { useState } from 'react'
import { createFileRoute, Link } from '@tanstack/react-router'
import { z } from 'zod'
import { Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { ScoreConfigsSection } from '@/features/scores/components'

// Empty + tolerant search schema so this route doesn't inherit the
// /scores list paginator/filters via TanStack's parent-search merging.
// `passthrough` keeps the inferred type open — without it the schema
// resolves to `[x: string]: never` and rejects Links from siblings
// that pass any field at all.
const searchSchema = z.object({}).passthrough().catch({})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/scores/configs',
)({
  validateSearch: searchSchema,
  component: ScoreConfigsPage,
})

function ScoreConfigsPage() {
  const { orgId, projectId } = Route.useParams()
  const [createDialogOpen, setCreateDialogOpen] = useState(false)

  return (
    <main className="mx-auto max-w-7xl space-y-6 p-6">
      <header className="flex items-baseline justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Score Configs</h1>
          <p className="text-sm text-muted-foreground">
            Define data types, ranges, and categories that score recordings
            must conform to.
          </p>
        </div>
        <div className="flex gap-2">
          <Button asChild variant="outline" size="sm">
            <Link
              to="/o/$orgId/p/$projectId/scores"
              params={{ orgId, projectId }}
              search={{ page: 1, limit: 20 }}
            >
              Back to scores
            </Link>
          </Button>
          <Button size="sm" onClick={() => setCreateDialogOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />
            New config
          </Button>
        </div>
      </header>

      <ScoreConfigsSection
        projectId={projectId}
        createDialogOpen={createDialogOpen}
        onCreateDialogOpenChange={setCreateDialogOpen}
      />
    </main>
  )
}
