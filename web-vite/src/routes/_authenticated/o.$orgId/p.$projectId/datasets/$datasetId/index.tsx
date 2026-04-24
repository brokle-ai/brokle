import { createFileRoute } from '@tanstack/react-router'
import { DatasetItemsTab } from '@/features/datasets/components/dataset-detail/dataset-items-tab'

// Items tab — mirrors web/'s /datasets/[datasetId]/page.tsx which
// mounts DatasetItemsTab directly. The tab reads projectId + datasetId
// via useDatasetDetail() → DatasetDetailProvider (installed by the
// parent route's DatasetDetailLayout).
export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/datasets/$datasetId/',
)({
  component: DatasetItemsPage,
})

function DatasetItemsPage() {
  return <DatasetItemsTab />
}
