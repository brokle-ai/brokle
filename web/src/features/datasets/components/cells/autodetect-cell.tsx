'use client'

import { JsonCell } from './json-cell'
import { TextCell } from './text-cell'
import { MediaCell, detectMediaType, isUrl } from './media-cell'
import type { RowHeight } from './types'

interface AutodetectCellProps {
  value: unknown
  rowHeight?: RowHeight
  className?: string
}

/**
 * AutodetectCell automatically determines the best cell renderer based on content type.
 * Detection cascade: Media URL → JSON Object/Array → Text
 */
export function AutodetectCell({ value, rowHeight = 'medium', className }: AutodetectCellProps) {
  // Handle null/undefined
  if (value === null || value === undefined) {
    return <span className="text-muted-foreground">-</span>
  }

  // Check for media URLs (string that's a URL with media extension)
  if (typeof value === 'string' && isUrl(value)) {
    const mediaType = detectMediaType(value)
    if (mediaType) {
      return (
        <MediaCell
          url={value}
          type={mediaType}
          rowHeight={rowHeight}
          className={className}
        />
      )
    }
    // It's a URL but not media - show as text with link behavior
    return (
      <a
        href={value}
        target="_blank"
        rel="noopener noreferrer"
        className="text-sm text-blue-600 hover:underline truncate block"
        title={value}
      >
        {value}
      </a>
    )
  }

  // Check for JSON objects or arrays
  if (typeof value === 'object') {
    return (
      <JsonCell
        value={value as Record<string, unknown> | unknown[]}
        rowHeight={rowHeight}
        className={className}
      />
    )
  }

  // Check for JSON strings
  if (typeof value === 'string') {
    // Try to parse as JSON
    try {
      const trimmed = value.trim()
      if (trimmed.startsWith('{') || trimmed.startsWith('[')) {
        const parsed = JSON.parse(trimmed)
        if (typeof parsed === 'object' && parsed !== null) {
          return (
            <JsonCell
              value={parsed}
              rowHeight={rowHeight}
              className={className}
            />
          )
        }
      }
    } catch {
      // Not valid JSON, continue to text
    }
  }

  // Default to text for primitives
  return (
    <TextCell
      value={String(value)}
      rowHeight={rowHeight}
      className={className}
    />
  )
}
