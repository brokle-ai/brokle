'use client'

import { Badge } from '@/components/ui/badge'
import { GitBranch, Pin } from 'lucide-react'
import type { DatasetWithVersionInfo } from '../types'

interface DatasetVersionBadgeProps {
  versionInfo?: DatasetWithVersionInfo
  compact?: boolean
}

export function DatasetVersionBadge({ versionInfo, compact = false }: DatasetVersionBadgeProps) {
  if (!versionInfo) {
    return null
  }

  const isPinned = !!versionInfo.current_version_id
  const currentVersion = versionInfo.current_version
  const latestVersion = versionInfo.latest_version

  if (!latestVersion) {
    return (
      <Badge variant="outline" className="text-muted-foreground">
        <GitBranch className="mr-1 h-3 w-3" />
        No versions
      </Badge>
    )
  }

  if (compact) {
    return isPinned ? (
      <Badge variant="secondary" className="text-xs">
        <Pin className="mr-1 h-3 w-3" />
        v{currentVersion}
      </Badge>
    ) : (
      <Badge variant="outline" className="text-xs">
        v{latestVersion}
      </Badge>
    )
  }

  return (
    <div className="flex items-center gap-2">
      {isPinned ? (
        <Badge variant="secondary">
          <Pin className="mr-1 h-3 w-3" />
          Pinned to v{currentVersion}
        </Badge>
      ) : (
        <Badge variant="outline">
          <GitBranch className="mr-1 h-3 w-3" />
          Latest (v{latestVersion})
        </Badge>
      )}
      {isPinned && currentVersion !== latestVersion && (
        <Badge variant="outline" className="text-muted-foreground">
          Latest: v{latestVersion}
        </Badge>
      )}
    </div>
  )
}
