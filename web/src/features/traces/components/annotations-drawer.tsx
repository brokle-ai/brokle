'use client'

import * as React from 'react'
import { useSearchParams, usePathname, useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { ClipboardCheck, Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useCurrentUser } from '@/features/authentication'
import { useScoreConfigsQuery } from '@/features/scores/hooks/use-score-configs'
import {
  useAnnotations,
  useCreateAnnotation,
  useDeleteAnnotation,
} from '../hooks/use-annotations'
import { AnnotationList } from './annotation-list'
import { AnnotationForm } from './annotation-form'

interface AnnotationsDrawerProps {
  projectId: string
  traceId: string
  className?: string
}

/**
 * AnnotationsDrawer - Drawer with badge count for trace annotations
 *
 * Features:
 * - Sheet component (side="right", width 400-540px)
 * - Trigger button with ClipboardCheck icon + badge count
 * - Header with title and annotation count
 * - Score type selector based on ScoreConfigs
 * - Dynamic value input based on score type (numeric, categorical, boolean)
 * - Reason/explanation field
 * - Scrollable annotation list
 * - Deep linking support: ?annotations=open query param
 */
export function AnnotationsDrawer({
  projectId,
  traceId,
  className,
}: AnnotationsDrawerProps) {
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()

  // Deep linking: use URL as source of truth for open state
  const isOpenFromUrl = searchParams.get('annotations') === 'open'

  // Handle open/close by updating URL (URL is the source of truth)
  const handleOpenChange = React.useCallback(
    (newOpen: boolean) => {
      const params = new URLSearchParams(searchParams.toString())
      if (newOpen) {
        params.set('annotations', 'open')
      } else {
        params.delete('annotations')
      }

      const newUrl = params.toString()
        ? `${pathname}?${params.toString()}`
        : pathname
      router.replace(newUrl, { scroll: false })
    },
    [pathname, router, searchParams]
  )

  // Get current user for ownership checks
  const { data: currentUser } = useCurrentUser()

  // Queries
  const { data: annotations, isLoading: isLoadingAnnotations } = useAnnotations(
    projectId,
    traceId
  )
  const { data: scoreConfigs, isLoading: isLoadingConfigs } = useScoreConfigsQuery(projectId)

  // Mutations
  const createMutation = useCreateAnnotation(projectId, traceId)
  const deleteMutation = useDeleteAnnotation(projectId, traceId)

  const annotationCount = annotations?.length ?? 0
  const humanAnnotations = annotations?.filter(a => a.source === 'annotation') ?? []
  const automatedScores = annotations?.filter(a => a.source !== 'annotation') ?? []

  const handleDelete = (scoreId: string) => {
    deleteMutation.mutate(scoreId)
  }

  // Format badge count
  const badgeText = annotationCount > 99 ? '99+' : annotationCount.toString()

  return (
    <Sheet open={isOpenFromUrl} onOpenChange={handleOpenChange}>
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <SheetTrigger asChild>
              <Button
                variant='ghost'
                size='icon'
                className={cn('h-8 w-8 relative', className)}
              >
                <ClipboardCheck className='h-4 w-4' />
                {annotationCount > 0 && (
                  <Badge
                    variant='secondary'
                    className='absolute -top-1 -right-1 h-4 min-w-4 px-1 text-[10px] flex items-center justify-center'
                  >
                    {badgeText}
                  </Badge>
                )}
                <span className='sr-only'>Annotations</span>
              </Button>
            </SheetTrigger>
          </TooltipTrigger>
          <TooltipContent>
            {annotationCount === 0
              ? 'Add annotation'
              : `${annotationCount} score${annotationCount === 1 ? '' : 's'}`}
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>

      <SheetContent
        side='right'
        className='w-full sm:max-w-md lg:max-w-lg flex flex-col'
        hideCloseButton={false}
      >
        <SheetHeader className='border-b pb-4'>
          <SheetTitle className='flex items-center gap-2'>
            <ClipboardCheck className='h-5 w-5' />
            Annotations & Scores
            {annotationCount > 0 && (
              <Badge variant='secondary' className='ml-1'>
                {annotationCount}
              </Badge>
            )}
          </SheetTitle>
        </SheetHeader>

        {/* Annotation Form */}
        <div className='border-b py-4'>
          {isLoadingConfigs ? (
            <div className='flex items-center justify-center py-4'>
              <Loader2 className='h-5 w-5 animate-spin text-muted-foreground' />
            </div>
          ) : (
            <AnnotationForm
              scoreConfigs={scoreConfigs?.data ?? []}
              onSubmit={(data) => createMutation.mutate(data)}
              isSubmitting={createMutation.isPending}
            />
          )}
        </div>

        {/* Scrollable annotation list */}
        <div className='flex-1 overflow-y-auto py-4'>
          <AnnotationList
            humanAnnotations={humanAnnotations}
            automatedScores={automatedScores}
            currentUserId={currentUser?.id}
            onDelete={handleDelete}
            deletingId={deleteMutation.isPending ? deleteMutation.variables : undefined}
            isLoading={isLoadingAnnotations}
          />
        </div>
      </SheetContent>
    </Sheet>
  )
}
