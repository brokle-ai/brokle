import { createFileRoute } from '@tanstack/react-router'
import { DatasetSettingsTab } from '@/features/datasets/components/dataset-detail/dataset-settings-tab'

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/datasets/$datasetId/settings',
)({
  component: DatasetSettingsPage,
})

function DatasetSettingsPage() {
  return <DatasetSettingsTab />
}
