'use client'

import { useState } from 'react'
import { formatDistanceToNow } from 'date-fns'
import { X, Copy, Check, ChevronDown, ChevronRight, ExternalLink } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { cn } from '@/lib/utils'
import type { DatasetItem } from '../../types'

interface ItemPreviewSidebarProps {
  item: DatasetItem | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function ItemPreviewSidebar({ item, open, onOpenChange }: ItemPreviewSidebarProps) {
  const [copiedField, setCopiedField] = useState<string | null>(null)

  const handleCopy = async (field: string, value: unknown) => {
    const text = typeof value === 'string' ? value : JSON.stringify(value, null, 2)
    await navigator.clipboard.writeText(text)
    setCopiedField(field)
    setTimeout(() => setCopiedField(null), 2000)
  }

  if (!item) return null

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="w-[500px] sm:max-w-[500px] p-0">
        <SheetHeader className="px-6 py-4 border-b">
          <div className="flex items-center justify-between">
            <SheetTitle className="text-lg">Item Details</SheetTitle>
            <Button variant="ghost" size="sm" onClick={() => onOpenChange(false)}>
              <X className="h-4 w-4" />
            </Button>
          </div>
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <span className="font-mono text-xs">{item.id.slice(0, 8)}...</span>
            <span>•</span>
            <span>
              {formatDistanceToNow(new Date(item.created_at), { addSuffix: true })}
            </span>
          </div>
        </SheetHeader>

        <ScrollArea className="h-[calc(100vh-100px)]">
          <div className="px-6 py-4 space-y-6">
            {/* Input Section */}
            <FieldSection
              title="Input"
              value={item.input}
              onCopy={() => handleCopy('input', item.input)}
              isCopied={copiedField === 'input'}
            />

            {/* Expected Output Section */}
            {item.expected_output && (
              <FieldSection
                title="Expected Output"
                value={item.expected_output}
                onCopy={() => handleCopy('expected', item.expected_output)}
                isCopied={copiedField === 'expected'}
              />
            )}

            {/* Metadata Section */}
            {item.metadata && Object.keys(item.metadata).length > 0 && (
              <FieldSection
                title="Metadata"
                value={item.metadata}
                onCopy={() => handleCopy('metadata', item.metadata)}
                isCopied={copiedField === 'metadata'}
              />
            )}

            <Separator />

            {/* Item Info */}
            <div className="space-y-3">
              <h4 className="text-sm font-medium text-muted-foreground">Item Information</h4>
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <span className="text-muted-foreground">ID</span>
                  <div className="font-mono text-xs mt-1">{item.id}</div>
                </div>
                <div>
                  <span className="text-muted-foreground">Dataset ID</span>
                  <div className="font-mono text-xs mt-1">{item.dataset_id}</div>
                </div>
                <div>
                  <span className="text-muted-foreground">Created</span>
                  <div className="mt-1">
                    {new Date(item.created_at).toLocaleDateString()}
                  </div>
                </div>
                <div>
                  <span className="text-muted-foreground">Updated</span>
                  <div className="mt-1">
                    {new Date(item.updated_at).toLocaleDateString()}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </ScrollArea>
      </SheetContent>
    </Sheet>
  )
}

interface FieldSectionProps {
  title: string
  value: unknown
  onCopy: () => void
  isCopied: boolean
}

function FieldSection({ title, value, onCopy, isCopied }: FieldSectionProps) {
  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-medium">{title}</h4>
        <Button
          variant="ghost"
          size="sm"
          className="h-7 px-2"
          onClick={onCopy}
        >
          {isCopied ? (
            <Check className="h-3 w-3 text-green-500" />
          ) : (
            <Copy className="h-3 w-3" />
          )}
        </Button>
      </div>
      <div className="rounded-md border bg-muted/30 p-3 overflow-hidden">
        <JsonViewer data={value} />
      </div>
    </div>
  )
}

interface JsonViewerProps {
  data: unknown
  level?: number
}

function JsonViewer({ data, level = 0 }: JsonViewerProps) {
  const [expanded, setExpanded] = useState(level < 3)

  if (data === null) {
    return <span className="text-orange-500 font-mono text-sm">null</span>
  }

  if (typeof data === 'boolean') {
    return <span className="text-purple-500 font-mono text-sm">{String(data)}</span>
  }

  if (typeof data === 'number') {
    return <span className="text-blue-500 font-mono text-sm">{data}</span>
  }

  if (typeof data === 'string') {
    // Check if it's a URL
    if (isValidUrl(data)) {
      return (
        <a
          href={data}
          target="_blank"
          rel="noopener noreferrer"
          className="text-blue-600 hover:underline font-mono text-sm flex items-center gap-1"
        >
          {data.length > 60 ? data.slice(0, 60) + '...' : data}
          <ExternalLink className="h-3 w-3 inline" />
        </a>
      )
    }
    // Wrap long strings
    return (
      <span className="text-green-600 font-mono text-sm whitespace-pre-wrap break-words">
        &quot;{data}&quot;
      </span>
    )
  }

  if (Array.isArray(data)) {
    if (data.length === 0) {
      return <span className="text-muted-foreground font-mono text-sm">[]</span>
    }

    return (
      <div className="font-mono text-sm">
        <button
          onClick={() => setExpanded(!expanded)}
          className="flex items-center gap-1 hover:bg-muted/50 rounded -ml-1 px-1"
        >
          {expanded ? (
            <ChevronDown className="h-3 w-3" />
          ) : (
            <ChevronRight className="h-3 w-3" />
          )}
          <span className="text-muted-foreground">
            Array [{data.length} items]
          </span>
        </button>
        {expanded && (
          <div className="ml-4 border-l-2 border-muted pl-3 mt-1 space-y-1">
            {data.map((item, index) => (
              <div key={index} className="flex gap-2">
                <span className="text-muted-foreground shrink-0">{index}:</span>
                <div className="min-w-0 flex-1">
                  <JsonViewer data={item} level={level + 1} />
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    )
  }

  if (typeof data === 'object') {
    const entries = Object.entries(data)
    if (entries.length === 0) {
      return <span className="text-muted-foreground font-mono text-sm">{'{}'}</span>
    }

    return (
      <div className="font-mono text-sm">
        <button
          onClick={() => setExpanded(!expanded)}
          className="flex items-center gap-1 hover:bg-muted/50 rounded -ml-1 px-1"
        >
          {expanded ? (
            <ChevronDown className="h-3 w-3" />
          ) : (
            <ChevronRight className="h-3 w-3" />
          )}
          <span className="text-muted-foreground">
            Object {'{'}{entries.length} keys{'}'}
          </span>
        </button>
        {expanded && (
          <div className="ml-4 border-l-2 border-muted pl-3 mt-1 space-y-1">
            {entries.map(([key, value]) => (
              <div key={key} className="flex gap-2">
                <span className="text-cyan-600 shrink-0">{key}:</span>
                <div className="min-w-0 flex-1">
                  <JsonViewer data={value} level={level + 1} />
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    )
  }

  return <span className="text-muted-foreground font-mono text-sm">{String(data)}</span>
}

function isValidUrl(str: string): boolean {
  try {
    new URL(str)
    return str.startsWith('http://') || str.startsWith('https://')
  } catch {
    return false
  }
}
