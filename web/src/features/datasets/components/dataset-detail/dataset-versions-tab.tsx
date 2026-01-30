'use client'

import { useState } from 'react'
import { formatDistanceToNow, isToday, isYesterday, subDays, isAfter } from 'date-fns'
import { Check, Pin, PinOff, Plus, FileText, Clock } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion'
import { useDatasetDetail } from '../../context/dataset-detail-context'
import {
  useDatasetWithVersionInfoQuery,
  useDatasetVersionsQuery,
  useCreateDatasetVersionMutation,
  usePinDatasetVersionMutation,
  useUnpinDatasetVersionMutation,
} from '../../hooks/use-datasets'
import type { DatasetVersion } from '../../types'

interface GroupedVersions {
  today: DatasetVersion[]
  yesterday: DatasetVersion[]
  last7Days: DatasetVersion[]
  last30Days: DatasetVersion[]
  older: DatasetVersion[]
}

function groupVersionsByDate(versions: DatasetVersion[]): GroupedVersions {
  const now = new Date()
  const sevenDaysAgo = subDays(now, 7)
  const thirtyDaysAgo = subDays(now, 30)

  return versions.reduce<GroupedVersions>(
    (groups, version) => {
      const createdAt = new Date(version.created_at)

      if (isToday(createdAt)) {
        groups.today.push(version)
      } else if (isYesterday(createdAt)) {
        groups.yesterday.push(version)
      } else if (isAfter(createdAt, sevenDaysAgo)) {
        groups.last7Days.push(version)
      } else if (isAfter(createdAt, thirtyDaysAgo)) {
        groups.last30Days.push(version)
      } else {
        groups.older.push(version)
      }

      return groups
    },
    { today: [], yesterday: [], last7Days: [], last30Days: [], older: [] }
  )
}

export function DatasetVersionsTab() {
  const { projectId, datasetId } = useDatasetDetail()
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [description, setDescription] = useState('')

  const { data: versionInfo, isLoading: isLoadingInfo } = useDatasetWithVersionInfoQuery(projectId, datasetId)
  const { data: versions = [], isLoading: isLoadingVersions } = useDatasetVersionsQuery(projectId, datasetId)

  const createVersionMutation = useCreateDatasetVersionMutation(projectId, datasetId)
  const pinVersionMutation = usePinDatasetVersionMutation(projectId, datasetId)
  const unpinVersionMutation = useUnpinDatasetVersionMutation(projectId, datasetId)

  const isLoading = isLoadingInfo || isLoadingVersions
  const isPinned = !!versionInfo?.current_version_id
  const latestVersion = versionInfo?.latest_version

  const handleCreateVersion = () => {
    createVersionMutation.mutate(
      { description: description || undefined },
      {
        onSuccess: () => {
          setCreateDialogOpen(false)
          setDescription('')
        },
      }
    )
  }

  const handlePinVersion = (versionId: string) => {
    pinVersionMutation.mutate({ version_id: versionId })
  }

  const handleUnpinVersion = () => {
    unpinVersionMutation.mutate()
  }

  if (isLoading) {
    return (
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <Skeleton className="h-6 w-32" />
          <Skeleton className="h-9 w-32" />
        </div>
        <Skeleton className="h-[300px]" />
      </div>
    )
  }

  const groupedVersions = groupVersionsByDate(versions)
  const hasVersions = versions.length > 0

  // Determine which groups should be open by default (groups with content)
  const defaultOpenGroups = Object.entries(groupedVersions)
    .filter(([_, items]) => items.length > 0)
    .map(([key]) => key)
    .slice(0, 2) // Open first two non-empty groups

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-medium">Version History</h2>
          <p className="text-sm text-muted-foreground">
            {isPinned
              ? `Pinned to v${versionInfo?.current_version}`
              : 'Using latest items'}
          </p>
        </div>
        <div className="flex items-center gap-2">
          {isPinned && (
            <Button
              variant="outline"
              onClick={handleUnpinVersion}
              disabled={unpinVersionMutation.isPending}
            >
              <PinOff className="mr-2 h-4 w-4" />
              Use Latest
            </Button>
          )}
          <Button onClick={() => setCreateDialogOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Create Version
          </Button>
        </div>
      </div>

      {!hasVersions ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Clock className="h-12 w-12 text-muted-foreground mb-4" />
            <CardTitle className="text-lg mb-2">No versions yet</CardTitle>
            <CardDescription className="text-center max-w-md mb-4">
              Create a version to snapshot the current dataset items.
              Versions allow you to pin experiments to a specific set of items.
            </CardDescription>
            <Button onClick={() => setCreateDialogOpen(true)}>
              <Plus className="mr-2 h-4 w-4" />
              Create First Version
            </Button>
          </CardContent>
        </Card>
      ) : (
        <Accordion type="multiple" defaultValue={defaultOpenGroups} className="space-y-2">
          <VersionGroup
            title="Today"
            versions={groupedVersions.today}
            pinnedVersionId={versionInfo?.current_version_id}
            latestVersion={latestVersion}
            onPin={handlePinVersion}
            groupKey="today"
          />
          <VersionGroup
            title="Yesterday"
            versions={groupedVersions.yesterday}
            pinnedVersionId={versionInfo?.current_version_id}
            latestVersion={latestVersion}
            onPin={handlePinVersion}
            groupKey="yesterday"
          />
          <VersionGroup
            title="Last 7 Days"
            versions={groupedVersions.last7Days}
            pinnedVersionId={versionInfo?.current_version_id}
            latestVersion={latestVersion}
            onPin={handlePinVersion}
            groupKey="last7Days"
          />
          <VersionGroup
            title="Last 30 Days"
            versions={groupedVersions.last30Days}
            pinnedVersionId={versionInfo?.current_version_id}
            latestVersion={latestVersion}
            onPin={handlePinVersion}
            groupKey="last30Days"
          />
          <VersionGroup
            title="Older"
            versions={groupedVersions.older}
            pinnedVersionId={versionInfo?.current_version_id}
            latestVersion={latestVersion}
            onPin={handlePinVersion}
            groupKey="older"
          />
        </Accordion>
      )}

      {/* Create Version Dialog */}
      <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Version</DialogTitle>
            <DialogDescription>
              Create a snapshot of the current dataset items. This preserves the exact set of items
              at this point in time for reproducible experiments.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="description">Description (optional)</Label>
              <Textarea
                id="description"
                placeholder="What's notable about this version?"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateDialogOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleCreateVersion}
              disabled={createVersionMutation.isPending}
            >
              {createVersionMutation.isPending ? 'Creating...' : 'Create Version'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

interface VersionGroupProps {
  title: string
  versions: DatasetVersion[]
  pinnedVersionId?: string
  latestVersion?: number
  onPin: (versionId: string) => void
  groupKey: string
}

function VersionGroup({
  title,
  versions,
  pinnedVersionId,
  latestVersion,
  onPin,
  groupKey,
}: VersionGroupProps) {
  if (versions.length === 0) {
    return null
  }

  return (
    <AccordionItem value={groupKey} className="border rounded-lg px-4">
      <AccordionTrigger className="hover:no-underline">
        <div className="flex items-center gap-2">
          <span className="font-medium">{title}</span>
          <Badge variant="secondary" className="text-xs">
            {versions.length}
          </Badge>
        </div>
      </AccordionTrigger>
      <AccordionContent>
        <div className="space-y-2 pt-2">
          {versions.map((version) => (
            <VersionCard
              key={version.id}
              version={version}
              isPinned={version.id === pinnedVersionId}
              isLatest={version.version === latestVersion}
              onPin={() => onPin(version.id)}
            />
          ))}
        </div>
      </AccordionContent>
    </AccordionItem>
  )
}

interface VersionCardProps {
  version: DatasetVersion
  isPinned: boolean
  isLatest: boolean
  onPin: () => void
}

function VersionCard({ version, isPinned, isLatest, onPin }: VersionCardProps) {
  return (
    <Card className={isPinned ? 'border-primary' : ''}>
      <CardContent className="flex items-center justify-between p-4">
        <div className="flex items-center gap-3 min-w-0">
          {isPinned ? (
            <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary/10">
              <Check className="h-4 w-4 text-primary" />
            </div>
          ) : (
            <div className="flex h-8 w-8 items-center justify-center rounded-full bg-muted">
              <FileText className="h-4 w-4 text-muted-foreground" />
            </div>
          )}
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <span className="font-medium">Version {version.version}</span>
              {isLatest && (
                <Badge variant="outline" className="text-xs">
                  Latest
                </Badge>
              )}
              {isPinned && (
                <Badge variant="default" className="text-xs">
                  Pinned
                </Badge>
              )}
            </div>
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <span>{version.item_count} items</span>
              <span>•</span>
              <span>{formatDistanceToNow(new Date(version.created_at), { addSuffix: true })}</span>
            </div>
            {version.description && (
              <p className="text-sm text-muted-foreground mt-1 truncate">
                {version.description}
              </p>
            )}
          </div>
        </div>
        {!isPinned && (
          <Button variant="ghost" size="sm" onClick={onPin}>
            <Pin className="h-4 w-4" />
          </Button>
        )}
      </CardContent>
    </Card>
  )
}
