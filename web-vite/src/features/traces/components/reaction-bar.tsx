import { useState } from 'react'
import { SmilePlus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { cn } from '@/lib/utils'
import type { TraceReaction } from '../api/types'

// v1 emoji palette. The backend accepts any short emoji string; this
// list is the deliberately-small set we surface in the picker so the
// palette stays curated. Adding more is a one-line change.
const QUICK_REACTIONS = ['👍', '👎', '🚀', '❤️', '🎉', '👀', '🤔']

interface ReactionBarProps {
  reactions: TraceReaction[]
  onToggle: (emoji: string) => void
  disabled?: boolean
}

export function ReactionBar({
  reactions,
  onToggle,
  disabled,
}: ReactionBarProps) {
  const [pickerOpen, setPickerOpen] = useState(false)

  const handlePick = (emoji: string) => {
    onToggle(emoji)
    setPickerOpen(false)
  }

  return (
    <div className="flex flex-wrap items-center gap-1">
      {reactions.map((r) => (
        <ReactionChip
          key={r.emoji}
          reaction={r}
          onClick={() => onToggle(r.emoji)}
          disabled={disabled}
        />
      ))}
      <Popover open={pickerOpen} onOpenChange={setPickerOpen}>
        <PopoverTrigger asChild>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            className="h-7 px-2 text-muted-foreground hover:text-foreground"
            disabled={disabled}
          >
            <SmilePlus className="h-3.5 w-3.5" />
            <span className="sr-only">Add reaction</span>
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-2" side="top" align="start">
          <div className="flex gap-1">
            {QUICK_REACTIONS.map((emoji) => {
              const active = reactions.find((r) => r.emoji === emoji)?.has_user
              return (
                <button
                  key={emoji}
                  type="button"
                  onClick={() => handlePick(emoji)}
                  className={cn(
                    'rounded p-1.5 text-lg transition-colors hover:bg-muted',
                    active && 'bg-primary/10',
                  )}
                >
                  {emoji}
                </button>
              )
            })}
          </div>
        </PopoverContent>
      </Popover>
    </div>
  )
}

interface ReactionChipProps {
  reaction: TraceReaction
  onClick: () => void
  disabled?: boolean
}

function ReactionChip({ reaction, onClick, disabled }: ReactionChipProps) {
  const tooltip =
    reaction.users.length > 0
      ? reaction.users.slice(0, 10).join(', ') +
        (reaction.users.length > 10
          ? ` and ${reaction.users.length - 10} more`
          : '')
      : reaction.has_user
        ? 'You'
        : `${reaction.count} ${reaction.count === 1 ? 'person' : 'people'}`

  return (
    <TooltipProvider delayDuration={300}>
      <Tooltip>
        <TooltipTrigger asChild>
          <button
            type="button"
            onClick={onClick}
            disabled={disabled}
            className={cn(
              'inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-sm transition-colors',
              reaction.has_user
                ? 'border-primary/30 bg-primary/10 text-foreground'
                : 'border-border bg-muted/50 hover:bg-muted',
              disabled && 'cursor-not-allowed opacity-50',
            )}
          >
            <span>{reaction.emoji}</span>
            <span className="text-xs font-medium">{reaction.count}</span>
          </button>
        </TooltipTrigger>
        <TooltipContent side="top">
          <p className="text-xs">{tooltip}</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}
