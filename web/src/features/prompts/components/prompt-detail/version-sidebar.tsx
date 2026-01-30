'use client'

import * as React from 'react'
import { Search, Plus, PanelLeftClose } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import type { PromptVersion, PromptType } from '../../types'
import { VersionSidebarItem } from './version-sidebar-item'
import { VersionDiffDialog } from '../version-management/VersionDiffDialog'

// ============================================================================
// Types
// ============================================================================

interface VersionSidebarProps {
  versions: PromptVersion[]
  selectedVersionId: string | null
  protectedLabels: string[]
  promptType: PromptType
  promptName?: string
  onVersionSelect: (version: PromptVersion) => void
  onCompare?: (fromVersion: number, toVersion: number) => void
  onCreateVersion?: () => void
  onCollapse: () => void
  onRestore?: (version: PromptVersion) => void
}

// ============================================================================
// Main VersionSidebar Component
// ============================================================================

export function VersionSidebar({
  versions,
  selectedVersionId,
  protectedLabels,
  promptType,
  promptName,
  onVersionSelect,
  onCompare,
  onCreateVersion,
  onCollapse,
  onRestore,
}: VersionSidebarProps) {
  const [searchQuery, setSearchQuery] = React.useState('')
  const [compareFrom, setCompareFrom] = React.useState<PromptVersion | null>(null)
  const [compareTo, setCompareTo] = React.useState<PromptVersion | null>(null)
  const [dialogOpen, setDialogOpen] = React.useState(false)

  // Filter versions by search query (version number or commit message)
  const filteredVersions = React.useMemo(() => {
    if (!searchQuery.trim()) return versions

    const query = searchQuery.toLowerCase()
    return versions.filter((version) => {
      // Match version number
      if (version.version.toString().includes(query)) return true
      // Match commit message
      if (version.commit_message?.toLowerCase().includes(query)) return true
      // Match labels
      if (version.labels.some((label) => label.toLowerCase().includes(query)))
        return true
      return false
    })
  }, [versions, searchQuery])

  const handleCompareClick = (e: React.MouseEvent, version: PromptVersion) => {
    e.stopPropagation()
    if (compareFrom === null) {
      // First click - set the "from" version
      setCompareFrom(version)
    } else if (compareFrom.id !== version.id) {
      // Second click on a different version - open dialog
      // Ensure from is the older version (lower version number)
      const [from, to] =
        compareFrom.version < version.version
          ? [compareFrom, version]
          : [version, compareFrom]
      setCompareTo(to)
      setCompareFrom(from)
      setDialogOpen(true)
    } else {
      // Clicked on the same version - cancel comparison
      setCompareFrom(null)
    }
  }

  const handleDialogClose = (open: boolean) => {
    setDialogOpen(open)
    if (!open) {
      // Reset comparison state when dialog closes
      setCompareFrom(null)
      setCompareTo(null)
    }
  }

  const handleRestoreClick = (e: React.MouseEvent, version: PromptVersion) => {
    e.stopPropagation()
    onRestore?.(version)
  }

  const handleCopyIdClick = (e: React.MouseEvent) => {
    e.stopPropagation()
    // Copy is handled in the CopyIdButton component
  }

  return (
    <div className="flex h-full flex-col border-r bg-background">
      {/* Header */}
      <div className="flex items-center justify-between border-b px-3 py-2">
        <h3 className="text-sm font-medium">Versions</h3>
        <div className="flex items-center gap-1">
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className="h-7 w-7"
                onClick={onCollapse}
              >
                <PanelLeftClose className="h-4 w-4" />
                <span className="sr-only">Collapse sidebar</span>
              </Button>
            </TooltipTrigger>
            <TooltipContent side="right">
              <p className="text-xs">Collapse sidebar</p>
            </TooltipContent>
          </Tooltip>
        </div>
      </div>

      {/* Search + New Version */}
      <div className="p-2 space-y-2 border-b">
        {/* Search input */}
        <div className="relative">
          <Search className="absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground" />
          <Input
            type="search"
            placeholder="Search versions..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="h-8 pl-8 text-sm"
          />
        </div>

        {/* New Version button - only show on detail page, not on edit page */}
        {onCreateVersion && (
          <Button
            size="sm"
            className="w-full"
            onClick={onCreateVersion}
          >
            <Plus className="mr-2 h-3.5 w-3.5" />
            New Version
          </Button>
        )}
      </div>

      {/* Version list */}
      <ScrollArea className="flex-1">
        <div className="relative p-2">
          {/* Timeline line */}
          <div className="absolute left-[18px] top-0 bottom-0 w-px bg-border" />

          {/* Version items */}
          <div className="space-y-0.5">
            {filteredVersions.length === 0 ? (
              <div className="py-8 text-center text-sm text-muted-foreground">
                {searchQuery ? 'No versions match your search' : 'No versions'}
              </div>
            ) : (
              filteredVersions.map((version, index) => (
                <VersionSidebarItem
                  key={version.id}
                  version={version}
                  isSelected={version.id === selectedVersionId}
                  isLatest={index === 0}
                  isCompareFrom={compareFrom?.id === version.id}
                  protectedLabels={protectedLabels}
                  onClick={() => onVersionSelect(version)}
                  onCompareClick={(e) => handleCompareClick(e, version)}
                  onRestoreClick={
                    onRestore ? (e) => handleRestoreClick(e, version) : undefined
                  }
                  onCopyIdClick={handleCopyIdClick}
                />
              ))
            )}
          </div>
        </div>
      </ScrollArea>

      {/* Compare floating tooltip */}
      {compareFrom !== null && !dialogOpen && (
        <div className="border-t bg-muted/50 p-2">
          <Card className="border-primary">
            <CardContent className="py-2 px-3 flex items-center justify-between">
              <div className="flex items-center gap-2 text-sm">
                <span>Comparing from</span>
                <Badge variant="secondary" className="font-mono">
                  #{compareFrom.version}
                </Badge>
              </div>
              <Button
                variant="ghost"
                size="sm"
                className="h-7 text-xs"
                onClick={() => setCompareFrom(null)}
              >
                Cancel
              </Button>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Version Diff Dialog */}
      <VersionDiffDialog
        isOpen={dialogOpen}
        onOpenChange={handleDialogClose}
        fromVersion={compareFrom}
        toVersion={compareTo}
        promptType={promptType}
        promptName={promptName}
      />
    </div>
  )
}
