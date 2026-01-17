'use client'

import * as React from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Plus, X, Tags } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useUpdateTraceTags } from '../hooks/use-update-trace-tags'

interface TraceTagsEditorProps {
  projectId: string
  traceId: string
  tags: string[]
  className?: string
}

/**
 * TraceTagsEditor - Inline tag editor for traces
 *
 * Features:
 * - Display existing tags as badges
 * - Add new tags via popover input
 * - Remove tags by clicking X on badge
 * - Tags are normalized on backend (lowercase, unique, sorted)
 *
 * Constraints (enforced by backend):
 * - Max 50 tags per trace
 * - Max 100 characters per tag
 */
export function TraceTagsEditor({
  projectId,
  traceId,
  tags = [],
  className,
}: TraceTagsEditorProps) {
  const [isOpen, setIsOpen] = React.useState(false)
  const [newTag, setNewTag] = React.useState('')
  const inputRef = React.useRef<HTMLInputElement>(null)
  const mutation = useUpdateTraceTags(projectId)

  // Focus input when popover opens
  React.useEffect(() => {
    if (isOpen) {
      // Small delay to ensure popover is rendered
      const timer = setTimeout(() => {
        inputRef.current?.focus()
      }, 0)
      return () => clearTimeout(timer)
    }
  }, [isOpen])

  const handleAddTag = () => {
    const tag = newTag.trim().toLowerCase()
    if (tag && !tags.includes(tag)) {
      mutation.mutate({ traceId, tags: [...tags, tag] })
    }
    setNewTag('')
    setIsOpen(false)
  }

  const handleRemoveTag = (tagToRemove: string) => {
    mutation.mutate({ traceId, tags: tags.filter((t) => t !== tagToRemove) })
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      handleAddTag()
    } else if (e.key === 'Escape') {
      setNewTag('')
      setIsOpen(false)
    }
  }

  return (
    <div className={cn('flex flex-wrap items-center gap-1.5', className)}>
      <Tags className='h-3.5 w-3.5 text-muted-foreground shrink-0' />

      {tags.length === 0 && (
        <span className='text-xs text-muted-foreground'>No tags</span>
      )}

      {tags.map((tag) => (
        <Badge
          key={tag}
          variant='secondary'
          size='sm'
          className='gap-0.5 pr-0.5 h-5'
        >
          <span className='max-w-[100px] truncate'>{tag}</span>
          <Button
            variant='ghost'
            size='sm'
            className='h-4 w-4 p-0 hover:bg-transparent hover:text-destructive'
            onClick={() => handleRemoveTag(tag)}
            disabled={mutation.isPending}
          >
            <X className='h-3 w-3' />
            <span className='sr-only'>Remove {tag}</span>
          </Button>
        </Badge>
      ))}

      <Popover open={isOpen} onOpenChange={setIsOpen}>
        <PopoverTrigger asChild>
          <Button
            variant='ghost'
            size='sm'
            className='h-5 px-1.5 text-xs text-muted-foreground hover:text-foreground'
            disabled={mutation.isPending || tags.length >= 50}
          >
            <Plus className='h-3 w-3 mr-0.5' />
            Add
          </Button>
        </PopoverTrigger>
        <PopoverContent className='w-52 p-2' align='start'>
          <div className='flex gap-1.5'>
            <Input
              ref={inputRef}
              value={newTag}
              onChange={(e) => setNewTag(e.target.value)}
              placeholder='Tag name'
              className='h-7 text-xs'
              onKeyDown={handleKeyDown}
              maxLength={100}
            />
            <Button
              size='sm'
              className='h-7 px-2 text-xs'
              onClick={handleAddTag}
              disabled={!newTag.trim() || mutation.isPending}
            >
              Add
            </Button>
          </div>
          <p className='text-[10px] text-muted-foreground mt-1.5'>
            Press Enter to add, Escape to cancel
          </p>
        </PopoverContent>
      </Popover>
    </div>
  )
}
