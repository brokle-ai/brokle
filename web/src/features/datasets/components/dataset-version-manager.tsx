'use client'

import { useState } from 'react'
import { formatDistanceToNow } from 'date-fns'
import { Check, ChevronDown, GitBranch, Pin, PinOff, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
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
import {
  useDatasetWithVersionInfoQuery,
  useDatasetVersionsQuery,
  useCreateDatasetVersionMutation,
  usePinDatasetVersionMutation,
  useUnpinDatasetVersionMutation,
} from '../hooks/use-datasets'
import type { DatasetVersion } from '../types'

interface DatasetVersionManagerProps {
  projectId: string
  datasetId: string
}

export function DatasetVersionManager({ projectId, datasetId }: DatasetVersionManagerProps) {
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [description, setDescription] = useState('')

  const { data: versionInfo, isLoading: isLoadingInfo } = useDatasetWithVersionInfoQuery(projectId, datasetId)
  const { data: versions, isLoading: isLoadingVersions } = useDatasetVersionsQuery(projectId, datasetId)

  const createVersionMutation = useCreateDatasetVersionMutation(projectId, datasetId)
  const pinVersionMutation = usePinDatasetVersionMutation(projectId, datasetId)
  const unpinVersionMutation = useUnpinDatasetVersionMutation(projectId, datasetId)

  const isLoading = isLoadingInfo || isLoadingVersions
  const isPinned = !!versionInfo?.current_version_id
  const currentVersion = versionInfo?.current_version
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
      <Button variant="outline" size="sm" disabled>
        <GitBranch className="mr-2 h-4 w-4" />
        Loading...
      </Button>
    )
  }

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="outline" size="sm">
            {isPinned ? (
              <>
                <Pin className="mr-2 h-4 w-4" />
                v{currentVersion}
              </>
            ) : (
              <>
                <GitBranch className="mr-2 h-4 w-4" />
                {latestVersion ? `v${latestVersion}` : 'No versions'}
              </>
            )}
            <ChevronDown className="ml-2 h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="w-64">
          <DropdownMenuLabel className="flex items-center justify-between">
            <span>Versions</span>
            {isPinned && (
              <Badge variant="secondary" className="text-xs">
                Pinned
              </Badge>
            )}
          </DropdownMenuLabel>
          <DropdownMenuSeparator />

          {/* Create new version */}
          <DropdownMenuItem onSelect={() => setCreateDialogOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Create new version
          </DropdownMenuItem>

          {/* Unpin option if currently pinned */}
          {isPinned && (
            <DropdownMenuItem onSelect={handleUnpinVersion}>
              <PinOff className="mr-2 h-4 w-4" />
              Use latest items
            </DropdownMenuItem>
          )}

          {versions && versions.length > 0 && (
            <>
              <DropdownMenuSeparator />
              <DropdownMenuLabel className="text-xs text-muted-foreground">
                Version History
              </DropdownMenuLabel>

              {versions.map((version) => (
                <VersionMenuItem
                  key={version.id}
                  version={version}
                  isSelected={version.id === versionInfo?.current_version_id}
                  isLatest={version.version === latestVersion}
                  onPin={() => handlePinVersion(version.id)}
                />
              ))}
            </>
          )}

          {(!versions || versions.length === 0) && (
            <div className="px-2 py-4 text-center text-sm text-muted-foreground">
              No versions yet. Create one to snapshot current items.
            </div>
          )}
        </DropdownMenuContent>
      </DropdownMenu>

      {/* Create Version Dialog */}
      <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Version</DialogTitle>
            <DialogDescription>
              Create a snapshot of the current dataset items. This preserves the exact set of items
              at this point in time.
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
    </>
  )
}

interface VersionMenuItemProps {
  version: DatasetVersion
  isSelected: boolean
  isLatest: boolean
  onPin: () => void
}

function VersionMenuItem({ version, isSelected, isLatest, onPin }: VersionMenuItemProps) {
  return (
    <DropdownMenuItem
      onSelect={onPin}
      className="flex items-center justify-between"
    >
      <div className="flex items-center gap-2 min-w-0">
        {isSelected ? (
          <Check className="h-4 w-4 text-primary shrink-0" />
        ) : (
          <Pin className="h-4 w-4 text-muted-foreground shrink-0" />
        )}
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-medium">v{version.version}</span>
            {isLatest && (
              <Badge variant="outline" className="text-xs">
                Latest
              </Badge>
            )}
          </div>
          <p className="text-xs text-muted-foreground truncate">
            {version.item_count} items • {formatDistanceToNow(new Date(version.created_at), { addSuffix: true })}
          </p>
        </div>
      </div>
    </DropdownMenuItem>
  )
}
