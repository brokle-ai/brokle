'use client'

import { useState } from 'react'
import { Image as ImageIcon, Video, FileText, ExternalLink } from 'lucide-react'
import { cn } from '@/lib/utils'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import type { RowHeight } from './types'

interface MediaCellProps {
  url: string
  type: 'image' | 'video' | 'document'
  rowHeight?: RowHeight
  className?: string
}

const ROW_HEIGHT_CONFIG: Record<RowHeight, { showPreview: boolean; previewSize: number }> = {
  small: { showPreview: false, previewSize: 0 },
  medium: { showPreview: true, previewSize: 40 },
  large: { showPreview: true, previewSize: 120 },
}

export function MediaCell({ url, type, rowHeight = 'medium', className }: MediaCellProps) {
  const [isPreviewOpen, setIsPreviewOpen] = useState(false)
  const [imageError, setImageError] = useState(false)

  const config = ROW_HEIGHT_CONFIG[rowHeight]

  const getIcon = () => {
    switch (type) {
      case 'image':
        return <ImageIcon className="h-4 w-4" />
      case 'video':
        return <Video className="h-4 w-4" />
      case 'document':
        return <FileText className="h-4 w-4" />
    }
  }

  const getFileName = () => {
    try {
      const urlObj = new URL(url)
      const pathname = urlObj.pathname
      const fileName = pathname.split('/').pop() || url
      return fileName.length > 30 ? fileName.slice(0, 27) + '...' : fileName
    } catch {
      return url.length > 30 ? url.slice(0, 27) + '...' : url
    }
  }

  if (type === 'image' && config.showPreview && !imageError) {
    return (
      <>
        <button
          onClick={() => setIsPreviewOpen(true)}
          className={cn(
            'flex items-center gap-2 hover:bg-muted/50 rounded p-1 -m-1 transition-colors',
            className
          )}
        >
          <div
            className="rounded overflow-hidden border bg-muted"
            style={{
              width: config.previewSize,
              height: config.previewSize,
            }}
          >
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={url}
              alt=""
              className="w-full h-full object-cover"
              onError={() => setImageError(true)}
            />
          </div>
          {rowHeight === 'large' && (
            <span className="text-xs text-muted-foreground truncate max-w-[150px]">
              {getFileName()}
            </span>
          )}
        </button>

        <Dialog open={isPreviewOpen} onOpenChange={setIsPreviewOpen}>
          <DialogContent className="max-w-3xl">
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2">
                <ImageIcon className="h-4 w-4" />
                Image Preview
              </DialogTitle>
            </DialogHeader>
            <div className="flex justify-center">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={url}
                alt=""
                className="max-h-[70vh] object-contain rounded"
              />
            </div>
            <div className="flex justify-between items-center text-sm text-muted-foreground">
              <span className="truncate max-w-[400px]">{url}</span>
              <a
                href={url}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-1 hover:text-foreground"
              >
                Open <ExternalLink className="h-3 w-3" />
              </a>
            </div>
          </DialogContent>
        </Dialog>
      </>
    )
  }

  // Fallback to link for other types or when preview is disabled
  return (
    <a
      href={url}
      target="_blank"
      rel="noopener noreferrer"
      className={cn(
        'flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground transition-colors',
        className
      )}
    >
      {getIcon()}
      <span className="truncate">{getFileName()}</span>
      <ExternalLink className="h-3 w-3 shrink-0" />
    </a>
  )
}

// Helper to detect media type from URL
export function detectMediaType(url: string): 'image' | 'video' | 'document' | null {
  const imageExtensions = /\.(jpg|jpeg|png|gif|webp|svg|bmp|ico)$/i
  const videoExtensions = /\.(mp4|webm|ogg|mov|avi)$/i
  const documentExtensions = /\.(pdf|doc|docx|xls|xlsx|ppt|pptx|txt)$/i

  try {
    const urlObj = new URL(url)
    const pathname = urlObj.pathname.toLowerCase()

    if (imageExtensions.test(pathname)) return 'image'
    if (videoExtensions.test(pathname)) return 'video'
    if (documentExtensions.test(pathname)) return 'document'

    // Check for common image hosting patterns
    if (pathname.includes('/image') || pathname.includes('/img')) return 'image'
  } catch {
    // Not a valid URL
    return null
  }

  return null
}

// Check if a string is a URL
export function isUrl(value: string): boolean {
  try {
    new URL(value)
    return true
  } catch {
    return false
  }
}
