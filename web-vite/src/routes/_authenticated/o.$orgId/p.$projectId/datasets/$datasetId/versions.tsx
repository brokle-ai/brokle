import { createFileRoute } from '@tanstack/react-router'
import { DatasetVersionsTab } from '@/features/datasets/components/dataset-detail/dataset-versions-tab'

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/datasets/$datasetId/versions',
)({
  component: DatasetVersionsPage,
})

function DatasetVersionsPage() {
  return <DatasetVersionsTab />
}
