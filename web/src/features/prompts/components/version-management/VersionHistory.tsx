'use client'

import { useState } from 'react'
import { formatDistanceToNow, format } from 'date-fns'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { cn } from '@/lib/utils'
import {
  GitBranch,
  Tag,
  Clock,
  User,
  MessageSquare,
  ChevronRight,
  ArrowLeftRight,
} from 'lucide-react'
import type { PromptVersion } from '../../types'
import { LabelBadge } from '../label-badge'

interface VersionHistoryProps {
  versions: PromptVersion[]
  protectedLabels?: string[]
  selectedVersionId?: string
  onVersionSelect?: (version: PromptVersion) => void
  onCompare?: (fromVersion: number, toVersion: number) => void
}

export function VersionHistory({
  versions,
  protectedLabels = [],
  selectedVersionId,
  onVersionSelect,
  onCompare,
}: VersionHistoryProps) {
  const [compareFrom, setCompareFrom] = useState<number | null>(null)

  const handleCompareClick = (version: PromptVersion) => {
    if (compareFrom === null) {
      setCompareFrom(version.version)
    } else if (compareFrom !== version.version) {
      const from = Math.min(compareFrom, version.version)
      const to = Math.max(compareFrom, version.version)
      onCompare?.(from, to)
      setCompareFrom(null)
    } else {
      setCompareFrom(null)
    }
  }

  if (versions.length === 0) {
    return (
      <div className="text-center py-12 text-muted-foreground">
        No versions found
      </div>
    )
  }

  return (
    <div className="relative">
      {/* Timeline line */}
      <div className="absolute left-6 top-0 bottom-0 w-px bg-border" />

      <div className="space-y-4">
        {versions.map((version, index) => {
          const isSelected = version.id === selectedVersionId
          const isCompareFrom = compareFrom === version.version
          const isLatest = index === 0

          return (
            <div
              key={version.id}
              className={cn(
                'relative pl-12',
                onVersionSelect && 'cursor-pointer'
              )}
              onClick={() => onVersionSelect?.(version)}
            >
              {/* Timeline dot */}
              <div
                className={cn(
                  'absolute left-4 top-4 h-4 w-4 rounded-full border-2 bg-background',
                  isSelected
                    ? 'border-primary bg-primary'
                    : isLatest
                    ? 'border-green-500'
                    : 'border-muted-foreground'
                )}
              />

              <Card
                className={cn(
                  'transition-colors',
                  isSelected && 'border-primary',
                  onVersionSelect && 'hover:border-primary/50'
                )}
              >
                <CardHeader className="pb-2">
                  <div className="flex items-start justify-between">
                    <div className="flex items-center gap-2">
                      <Badge
                        variant={isLatest ? 'default' : 'secondary'}
                        className="font-mono"
                      >
                        v{version.version}
                      </Badge>
                      {version.labels.map((label) => (
                        <LabelBadge
                          key={label}
                          label={label}
                          isProtected={protectedLabels.includes(label)}
                        />
                      ))}
                    </div>
                    {onCompare && (
                      <Button
                        variant={isCompareFrom ? 'default' : 'ghost'}
                        size="sm"
                        onClick={(e) => {
                          e.stopPropagation()
                          handleCompareClick(version)
                        }}
                      >
                        <ArrowLeftRight className="mr-2 h-4 w-4" />
                        {isCompareFrom ? 'Select to compare' : 'Compare'}
                      </Button>
                    )}
                  </div>
                </CardHeader>
                <CardContent className="pt-2">
                  <div className="space-y-2">
                    {/* Commit message */}
                    {version.commit_message && (
                      <div className="flex items-start gap-2 text-sm">
                        <MessageSquare className="h-4 w-4 text-muted-foreground mt-0.5" />
                        <span>{version.commit_message}</span>
                      </div>
                    )}

                    {/* Metadata */}
                    <div className="flex flex-wrap items-center gap-4 text-xs text-muted-foreground">
                      <div className="flex items-center gap-1">
                        <Clock className="h-3 w-3" />
                        <span title={format(new Date(version.created_at), 'PPpp')}>
                          {formatDistanceToNow(new Date(version.created_at), {
                            addSuffix: true,
                          })}
                        </span>
                      </div>
                      {version.created_by && (
                        <div className="flex items-center gap-1">
                          <User className="h-3 w-3" />
                          <span>{version.created_by}</span>
                        </div>
                      )}
                      <div className="flex items-center gap-1">
                        <GitBranch className="h-3 w-3" />
                        <span>{version.variables.length} variables</span>
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          )
        })}
      </div>

      {compareFrom !== null && (
        <div className="fixed bottom-4 left-1/2 -translate-x-1/2 z-50">
          <Card className="border-primary shadow-lg">
            <CardContent className="py-2 px-4 flex items-center gap-2">
              <span className="text-sm">
                Comparing from <Badge variant="secondary">v{compareFrom}</Badge>
              </span>
              <span className="text-sm text-muted-foreground">
                — Select another version to compare
              </span>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setCompareFrom(null)}
              >
                Cancel
              </Button>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  )
}

// ============================================================================
// Version Diff View
// ============================================================================

interface VersionDiffProps {
  fromVersion: number
  toVersion: number
  templateFrom: any
  templateTo: any
  variablesAdded: string[]
  variablesRemoved: string[]
}

export function VersionDiff({
  fromVersion,
  toVersion,
  templateFrom,
  templateTo,
  variablesAdded,
  variablesRemoved,
}: VersionDiffProps) {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-center gap-4">
        <Badge variant="outline" className="font-mono text-lg px-4 py-2">
          v{fromVersion}
        </Badge>
        <ChevronRight className="h-6 w-6 text-muted-foreground" />
        <Badge variant="default" className="font-mono text-lg px-4 py-2">
          v{toVersion}
        </Badge>
      </div>

      {/* Variables changes */}
      {(variablesAdded.length > 0 || variablesRemoved.length > 0) && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Variable Changes</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            {variablesAdded.length > 0 && (
              <div className="flex items-center gap-2">
                <span className="text-sm text-green-600 dark:text-green-400">
                  Added:
                </span>
                {variablesAdded.map((v) => (
                  <Badge
                    key={v}
                    variant="outline"
                    className="bg-green-100 dark:bg-green-900/30 font-mono"
                  >
                    +{`{{${v}}}`}
                  </Badge>
                ))}
              </div>
            )}
            {variablesRemoved.length > 0 && (
              <div className="flex items-center gap-2">
                <span className="text-sm text-red-600 dark:text-red-400">
                  Removed:
                </span>
                {variablesRemoved.map((v) => (
                  <Badge
                    key={v}
                    variant="outline"
                    className="bg-red-100 dark:bg-red-900/30 font-mono"
                  >
                    -{`{{${v}}}`}
                  </Badge>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Template comparison */}
      <div className="grid grid-cols-2 gap-4">
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">
              Version {fromVersion}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <pre className="whitespace-pre-wrap rounded-md bg-muted p-4 font-mono text-sm">
              {JSON.stringify(templateFrom, null, 2)}
            </pre>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">
              Version {toVersion}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <pre className="whitespace-pre-wrap rounded-md bg-muted p-4 font-mono text-sm">
              {JSON.stringify(templateTo, null, 2)}
            </pre>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
