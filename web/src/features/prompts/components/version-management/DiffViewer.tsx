'use client'

import React, { useEffect, useState } from 'react'
import { cn } from '@/lib/utils'
import { diffLines as calculateDiffLines, diffWords } from 'diff'

// ============================================================================
// Types
// ============================================================================

type DiffSegmentPart = {
  value: string
  type?: 'unchanged' | 'removed' | 'added' | 'empty'
}

type DiffSegment = {
  text: string
  type: 'unchanged' | 'removed' | 'added' | 'empty'
  parts?: DiffSegmentPart[]
}

interface DiffViewerProps {
  oldString: string
  newString: string
  oldLabel?: string
  newLabel?: string
  oldSubLabel?: string
  newSubLabel?: string
  className?: string
}

// ============================================================================
// Constants
// ============================================================================

const DIFF_COLORS = {
  added: {
    text: 'bg-green-200 dark:bg-green-800/50',
    line: 'bg-green-100 dark:bg-green-900/30',
  },
  removed: {
    text: 'bg-red-200 dark:bg-red-800/50',
    line: 'bg-red-100 dark:bg-red-900/30',
  },
  unchanged: {
    text: '',
    line: '',
  },
  empty: {
    text: 'bg-muted/50',
    line: 'bg-muted/50',
  },
} as const

// ============================================================================
// Helpers
// ============================================================================

/**
 * Calculates the diff between two segments, word by word.
 */
function calculateSegmentDiff(oldString: string, newString: string) {
  const segmentChanges = diffWords(oldString, newString, {})
  const leftWords: DiffSegmentPart[] = []
  const rightWords: DiffSegmentPart[] = []

  for (let charIndex = 0; charIndex < segmentChanges.length; charIndex++) {
    const change = segmentChanges[charIndex]

    if (!change.added && !change.removed) {
      leftWords.push({ value: change.value, type: 'unchanged' })
      rightWords.push({ value: change.value, type: 'unchanged' })
    } else if (change.removed) {
      const nextChange = segmentChanges[charIndex + 1]
      const addsCharacterNext = nextChange !== undefined && nextChange.added
      if (addsCharacterNext) {
        leftWords.push({ value: change.value, type: 'removed' })
        rightWords.push({ value: nextChange.value, type: 'added' })
        charIndex++
      } else {
        leftWords.push({ value: change.value, type: 'removed' })
        rightWords.push({ value: '', type: 'empty' })
      }
    } else {
      leftWords.push({ value: '', type: 'empty' })
      rightWords.push({ value: change.value, type: 'added' })
    }
  }

  return { leftWords, rightWords }
}

// ============================================================================
// DiffRow Component
// ============================================================================

interface DiffRowProps {
  leftLine: DiffSegment
  rightLine: DiffSegment
}

function DiffRow({ leftLine, rightLine }: DiffRowProps) {
  const typeClasses = {
    unchanged: '',
    removed: DIFF_COLORS.removed.line,
    added: DIFF_COLORS.added.line,
    empty: DIFF_COLORS.empty.line,
  }

  const renderContent = (line: DiffSegment) =>
    line.parts
      ? line.parts.map((part, idx) => (
          <span
            key={idx}
            className={cn(
              part.type ? DIFF_COLORS[part.type].text : undefined,
              part.type === 'added' || part.type === 'removed' ? 'px-0.5 rounded' : ''
            )}
          >
            {part.value}
          </span>
        ))
      : line.text || '\u00A0'

  return (
    <div className="grid grid-cols-2 border-b last:border-b-0">
      <div
        className={cn(
          'whitespace-pre-wrap break-words border-r px-4 py-2.5 font-mono text-sm',
          typeClasses[leftLine.type]
        )}
      >
        {renderContent(leftLine)}
      </div>
      <div
        className={cn(
          'whitespace-pre-wrap break-words px-4 py-2.5 font-mono text-sm',
          typeClasses[rightLine.type]
        )}
      >
        {renderContent(rightLine)}
      </div>
    </div>
  )
}

// ============================================================================
// Main DiffViewer Component
// ============================================================================

export function DiffViewer({
  oldString,
  newString,
  oldLabel = 'Original Version',
  newLabel = 'New Version',
  oldSubLabel,
  newSubLabel,
  className,
}: DiffViewerProps) {
  const [diffLines, setDiffLines] = useState<{
    left: DiffSegment[]
    right: DiffSegment[]
  }>({ left: [], right: [] })

  useEffect(() => {
    const left: DiffSegment[] = []
    const right: DiffSegment[] = []

    const lineChanges = calculateDiffLines(oldString, newString, {})

    for (let diffIndex = 0; diffIndex < lineChanges.length; diffIndex++) {
      const part = lineChanges[diffIndex]

      if (!part.added && !part.removed) {
        left.push({ text: part.value, type: 'unchanged' })
        right.push({ text: part.value, type: 'unchanged' })
      } else if (part.removed) {
        const areThereMoreChanges = diffIndex < lineChanges.length - 1
        const isThereAnAdditionNext =
          areThereMoreChanges && lineChanges[diffIndex + 1].added
        if (isThereAnAdditionNext) {
          const { leftWords, rightWords } = calculateSegmentDiff(
            part.value,
            lineChanges[diffIndex + 1].value
          )
          left.push({ parts: leftWords, text: '', type: 'removed' })
          right.push({ parts: rightWords, text: '', type: 'added' })
          diffIndex++
        } else {
          left.push({ text: part.value, type: 'removed' })
          right.push({ text: '', type: 'empty' })
        }
      } else {
        left.push({ text: '', type: 'empty' })
        right.push({ text: part.value, type: 'added' })
      }
    }

    setDiffLines({ left, right })
  }, [oldString, newString])

  if (oldString === newString) {
    return <div className="text-sm text-muted-foreground py-2">No changes</div>
  }

  return (
    <div className={cn('w-full', className)}>
      <div className="rounded-lg border overflow-hidden">
        {/* Header row */}
        <div className="grid grid-cols-2 bg-muted/60">
          <div className="border-r border-b px-4 py-3">
            <span className="text-sm font-medium">{oldLabel}</span>
            {oldSubLabel && (
              <span className="ml-2 text-sm text-muted-foreground">
                {oldSubLabel}
              </span>
            )}
          </div>
          <div className="border-b px-4 py-3">
            <span className="text-sm font-medium">{newLabel}</span>
            {newSubLabel && (
              <span className="ml-2 text-sm text-muted-foreground">
                {newSubLabel}
              </span>
            )}
          </div>
        </div>
        {/* Diff content */}
        <div className="max-h-[500px] overflow-y-auto">
          {diffLines.left.map((leftLine, idx) => (
            <DiffRow
              key={idx}
              leftLine={leftLine}
              rightLine={diffLines.right[idx]}
            />
          ))}
        </div>
      </div>
    </div>
  )
}
