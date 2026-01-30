'use client'

import { History, RotateCcw, Trash2, AlertTriangle, Clock, Coins, Zap } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import type { RunHistoryEntry } from '../types'

interface RunHistoryPanelProps {
  history: RunHistoryEntry[]
  onRestore: (id: string) => void
  onClear: () => void
  disabled?: boolean
}

function formatTimestamp(timestamp: string): string {
  const date = new Date(timestamp)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)

  if (diffMins < 1) return 'Just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  return date.toLocaleDateString()
}

function truncateContent(content: string, maxLength: number = 150): string {
  if (content.length <= maxLength) return content
  return content.slice(0, maxLength).trim() + '...'
}

function formatCost(cost: number | undefined): string {
  if (cost === undefined) return '-'
  if (cost < 0.01) return `$${cost.toFixed(4)}`
  return `$${cost.toFixed(2)}`
}

function formatTokens(tokens: number | undefined): string {
  if (tokens === undefined) return '-'
  if (tokens >= 1000) return `${(tokens / 1000).toFixed(1)}k`
  return tokens.toString()
}

function HistoryEntryCard({
  entry,
  onRestore,
}: {
  entry: RunHistoryEntry
  onRestore: () => void
}) {
  return (
    <div
      className={`border rounded-lg p-3 space-y-2 ${
        entry.isStale ? 'opacity-70 border-dashed border-muted-foreground/40' : ''
      }`}
    >
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <Clock className="h-3 w-3" />
          <span>{formatTimestamp(entry.timestamp)}</span>
          {entry.isStale && (
            <Tooltip>
              <TooltipTrigger asChild>
                <Badge variant="outline" className="text-[10px] px-1 py-0 gap-1 text-yellow-600 border-yellow-600/50">
                  <AlertTriangle className="h-2.5 w-2.5" />
                  Stale
                </Badge>
              </TooltipTrigger>
              <TooltipContent>
                <p>Prompt has changed since this run</p>
              </TooltipContent>
            </Tooltip>
          )}
        </div>
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="ghost"
              size="icon"
              className="h-6 w-6"
              onClick={onRestore}
            >
              <RotateCcw className="h-3.5 w-3.5" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p>Restore this run</p>
          </TooltipContent>
        </Tooltip>
      </div>

      {entry.metrics && (
        <div className="flex items-center gap-3 text-xs text-muted-foreground">
          {entry.metrics.model && (
            <span className="font-medium text-foreground/80">
              {entry.metrics.model.split('/').pop()}
            </span>
          )}
          {entry.metrics.total_tokens !== undefined && (
            <span className="flex items-center gap-1">
              <Zap className="h-3 w-3" />
              {formatTokens(entry.metrics.total_tokens)}
            </span>
          )}
          {entry.metrics.cost !== undefined && (
            <span className="flex items-center gap-1">
              <Coins className="h-3 w-3" />
              {formatCost(entry.metrics.cost)}
            </span>
          )}
          {entry.metrics.latency_ms !== undefined && (
            <span>{(entry.metrics.latency_ms / 1000).toFixed(1)}s</span>
          )}
        </div>
      )}

      <div className="text-sm text-muted-foreground bg-muted/50 rounded p-2 font-mono text-xs leading-relaxed">
        {truncateContent(entry.content || '(empty response)')}
      </div>

      {entry.error && (
        <div className="text-xs text-destructive bg-destructive/10 rounded p-2">
          {entry.error}
        </div>
      )}
    </div>
  )
}

export function RunHistoryPanel({
  history,
  onRestore,
  onClear,
  disabled,
}: RunHistoryPanelProps) {
  const historyCount = history.length

  return (
    <Sheet>
      <Tooltip>
        <TooltipTrigger asChild>
          <SheetTrigger asChild>
            <Button
              variant="outline"
              size="icon"
              className="h-8 w-8 relative"
              disabled={disabled}
              aria-label="Run history"
            >
              <History className="h-4 w-4" />
              {historyCount > 0 && (
                <span className="absolute -top-1 -right-1 h-4 w-4 rounded-full bg-primary text-primary-foreground text-[10px] flex items-center justify-center">
                  {historyCount}
                </span>
              )}
            </Button>
          </SheetTrigger>
        </TooltipTrigger>
        <TooltipContent>
          <p>Run history ({historyCount})</p>
        </TooltipContent>
      </Tooltip>
      <SheetContent className="w-[400px] sm:w-[540px]">
        <SheetHeader>
          <SheetTitle className="flex items-center gap-2">
            <History className="h-5 w-5" />
            Run History
          </SheetTitle>
          <SheetDescription>
            Last {historyCount} execution{historyCount !== 1 ? 's' : ''} for this window.
            Restore any run to reload its messages, variables, and config.
          </SheetDescription>
        </SheetHeader>

        <div className="mt-4 space-y-4">
          {historyCount > 0 && (
            <div className="flex justify-end">
              <Button
                variant="ghost"
                size="sm"
                onClick={onClear}
                className="text-destructive hover:text-destructive hover:bg-destructive/10"
              >
                <Trash2 className="h-4 w-4 mr-1" />
                Clear history
              </Button>
            </div>
          )}

          <ScrollArea className="h-[calc(100vh-200px)]">
            {historyCount === 0 ? (
              <div className="flex flex-col items-center justify-center py-12 text-center text-muted-foreground">
                <History className="h-12 w-12 mb-4 opacity-30" />
                <p className="text-sm">No runs yet</p>
                <p className="text-xs mt-1">Execute a prompt to see history here</p>
              </div>
            ) : (
              <div className="space-y-3 pr-4">
                {history.map((entry) => (
                  <HistoryEntryCard
                    key={entry.id}
                    entry={entry}
                    onRestore={() => onRestore(entry.id)}
                  />
                ))}
              </div>
            )}
          </ScrollArea>
        </div>
      </SheetContent>
    </Sheet>
  )
}
