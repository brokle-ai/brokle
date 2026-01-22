'use client'

import * as React from 'react'
import { Badge } from '@/components/ui/badge'
import { Loader2 } from 'lucide-react'
import type { Annotation } from '../api/scores-api'
import { AnnotationItem } from './annotation-item'

interface AnnotationListProps {
  humanAnnotations: Annotation[]
  automatedScores: Annotation[]
  currentUserId?: string
  onDelete: (scoreId: string) => void
  deletingId?: string
  isLoading?: boolean
}

/**
 * AnnotationList - Display list of annotations organized by source
 *
 * Features:
 * - Two sections: Human Annotations and Automated Scores
 * - Visual distinction between sections
 * - Empty state for no annotations
 * - Loading state
 */
export function AnnotationList({
  humanAnnotations,
  automatedScores,
  currentUserId,
  onDelete,
  deletingId,
  isLoading = false,
}: AnnotationListProps) {
  if (isLoading) {
    return (
      <div className='flex items-center justify-center py-8'>
        <Loader2 className='h-6 w-6 animate-spin text-muted-foreground' />
      </div>
    )
  }

  const totalCount = humanAnnotations.length + automatedScores.length

  if (totalCount === 0) {
    return (
      <div className='flex flex-col items-center justify-center py-12 text-center'>
        <p className='text-sm text-muted-foreground'>
          No annotations yet.
        </p>
        <p className='text-xs text-muted-foreground mt-1'>
          Add one above to get started.
        </p>
      </div>
    )
  }

  return (
    <div className='space-y-6'>
      {/* Human Annotations Section */}
      {humanAnnotations.length > 0 && (
        <div className='space-y-3'>
          <div className='flex items-center gap-2'>
            <h4 className='text-sm font-medium'>Human Annotations</h4>
            <Badge variant='outline' className='text-xs'>
              {humanAnnotations.length}
            </Badge>
          </div>
          <div className='space-y-2'>
            {humanAnnotations.map((annotation) => (
              <AnnotationItem
                key={annotation.id}
                annotation={annotation}
                currentUserId={currentUserId}
                onDelete={onDelete}
                isDeleting={deletingId === annotation.id}
                canDelete
              />
            ))}
          </div>
        </div>
      )}

      {/* Automated Scores Section */}
      {automatedScores.length > 0 && (
        <div className='space-y-3'>
          <div className='flex items-center gap-2'>
            <h4 className='text-sm font-medium text-muted-foreground'>
              Automated Scores
            </h4>
            <Badge variant='secondary' className='text-xs'>
              {automatedScores.length}
            </Badge>
          </div>
          <div className='space-y-2'>
            {automatedScores.map((score) => (
              <AnnotationItem
                key={score.id}
                annotation={score}
                currentUserId={currentUserId}
                onDelete={onDelete}
                isDeleting={deletingId === score.id}
                canDelete={false}
              />
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
