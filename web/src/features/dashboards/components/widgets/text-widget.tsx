'use client'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import type { Widget } from '../../types'

interface TextWidgetProps {
  widget: Widget
  data: TextData | null
  isLoading: boolean
  error?: string
}

interface TextData {
  content: string
}

// Simple markdown-like renderer for basic formatting
function renderContent(content: string): React.ReactNode {
  // Split into lines for processing
  const lines = content.split('\n')
  const elements: React.ReactNode[] = []
  let listItems: string[] = []
  let listType: 'ul' | 'ol' | null = null

  const flushList = () => {
    if (listItems.length > 0 && listType) {
      const ListTag = listType
      elements.push(
        <ListTag key={elements.length} className="text-sm pl-4 my-2">
          {listItems.map((item, i) => (
            <li key={i}>{renderInline(item)}</li>
          ))}
        </ListTag>
      )
      listItems = []
      listType = null
    }
  }

  lines.forEach((line, index) => {
    // Headers
    if (line.startsWith('### ')) {
      flushList()
      elements.push(
        <h3 key={index} className="text-sm font-semibold mt-3 mb-1">
          {renderInline(line.slice(4))}
        </h3>
      )
      return
    }
    if (line.startsWith('## ')) {
      flushList()
      elements.push(
        <h2 key={index} className="text-base font-semibold mt-3 mb-1">
          {renderInline(line.slice(3))}
        </h2>
      )
      return
    }
    if (line.startsWith('# ')) {
      flushList()
      elements.push(
        <h1 key={index} className="text-lg font-bold mt-3 mb-1">
          {renderInline(line.slice(2))}
        </h1>
      )
      return
    }

    // Unordered list
    if (line.match(/^[-*]\s/)) {
      if (listType !== 'ul') {
        flushList()
        listType = 'ul'
      }
      listItems.push(line.slice(2))
      return
    }

    // Ordered list
    if (line.match(/^\d+\.\s/)) {
      if (listType !== 'ol') {
        flushList()
        listType = 'ol'
      }
      listItems.push(line.replace(/^\d+\.\s/, ''))
      return
    }

    // Horizontal rule
    if (line.match(/^[-*_]{3,}$/)) {
      flushList()
      elements.push(<hr key={index} className="my-3 border-border" />)
      return
    }

    // Empty line
    if (line.trim() === '') {
      flushList()
      return
    }

    // Regular paragraph
    flushList()
    elements.push(
      <p key={index} className="text-sm my-1">
        {renderInline(line)}
      </p>
    )
  })

  flushList()
  return elements
}

// Render inline formatting (bold, italic, code, links)
function renderInline(text: string): React.ReactNode {
  const parts: React.ReactNode[] = []
  let remaining = text
  let key = 0

  while (remaining) {
    // Bold: **text** or __text__
    let match = remaining.match(/^(.*?)\*\*(.+?)\*\*(.*)$/) ||
                remaining.match(/^(.*?)__(.+?)__(.*)$/)
    if (match) {
      if (match[1]) parts.push(<span key={key++}>{match[1]}</span>)
      parts.push(<strong key={key++} className="font-semibold">{match[2]}</strong>)
      remaining = match[3]
      continue
    }

    // Italic: *text* or _text_
    match = remaining.match(/^(.*?)\*(.+?)\*(.*)$/) ||
            remaining.match(/^(.*?)_(.+?)_(.*)$/)
    if (match) {
      if (match[1]) parts.push(<span key={key++}>{match[1]}</span>)
      parts.push(<em key={key++}>{match[2]}</em>)
      remaining = match[3]
      continue
    }

    // Code: `text`
    match = remaining.match(/^(.*?)`(.+?)`(.*)$/)
    if (match) {
      if (match[1]) parts.push(<span key={key++}>{match[1]}</span>)
      parts.push(
        <code key={key++} className="px-1 py-0.5 bg-muted rounded text-xs font-mono">
          {match[2]}
        </code>
      )
      remaining = match[3]
      continue
    }

    // Link: [text](url)
    match = remaining.match(/^(.*?)\[(.+?)\]\((.+?)\)(.*)$/)
    if (match) {
      if (match[1]) parts.push(<span key={key++}>{match[1]}</span>)
      parts.push(
        <a
          key={key++}
          href={match[3]}
          target="_blank"
          rel="noopener noreferrer"
          className="text-primary hover:underline"
        >
          {match[2]}
        </a>
      )
      remaining = match[4]
      continue
    }

    // No more special formatting, add remaining text
    parts.push(<span key={key++}>{remaining}</span>)
    break
  }

  return parts.length === 1 ? parts[0] : parts
}

export function TextWidget({ widget, data, isLoading, error }: TextWidgetProps) {
  const variant = (widget.config?.variant as 'default' | 'card' | 'minimal') || 'default'
  const alignment = (widget.config?.alignment as 'left' | 'center' | 'right') || 'left'

  // Text widget can also use static content from config
  const content = data?.content || (widget.config?.content as string) || ''

  if (isLoading) {
    return (
      <Card className="h-full">
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">
            <Skeleton className="h-4 w-32" />
          </CardTitle>
        </CardHeader>
        <CardContent>
          <Skeleton className="h-4 w-full mb-2" />
          <Skeleton className="h-4 w-3/4 mb-2" />
          <Skeleton className="h-4 w-1/2" />
        </CardContent>
      </Card>
    )
  }

  if (error) {
    return (
      <Card className="h-full">
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">{widget.title}</CardTitle>
        </CardHeader>
        <CardContent className="flex items-center justify-center h-[200px]">
          <p className="text-sm text-destructive">{error}</p>
        </CardContent>
      </Card>
    )
  }

  if (!content) {
    return (
      <Card className="h-full">
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">{widget.title}</CardTitle>
        </CardHeader>
        <CardContent className="flex items-center justify-center h-[200px]">
          <p className="text-sm text-muted-foreground">No content</p>
        </CardContent>
      </Card>
    )
  }

  // Minimal variant - just text, no card
  if (variant === 'minimal') {
    return (
      <div
        className={cn(
          'text-muted-foreground',
          alignment === 'center' && 'text-center',
          alignment === 'right' && 'text-right'
        )}
      >
        {renderContent(content)}
      </div>
    )
  }

  return (
    <Card className="h-full">
      {(widget.title || widget.description) && (
        <CardHeader className="pb-2">
          {widget.title && (
            <CardTitle className="text-sm font-medium">{widget.title}</CardTitle>
          )}
          {widget.description && (
            <CardDescription className="text-xs">{widget.description}</CardDescription>
          )}
        </CardHeader>
      )}
      <CardContent
        className={cn(
          'text-muted-foreground',
          alignment === 'center' && 'text-center',
          alignment === 'right' && 'text-right'
        )}
      >
        {renderContent(content)}
      </CardContent>
    </Card>
  )
}

export type { TextData }
