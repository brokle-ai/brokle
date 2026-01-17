'use client'

import * as React from 'react'
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
import { SmilePlus } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { ReactionSummary } from '../api/comments-api'

// Quick reaction emojis for the picker
const QUICK_REACTIONS = ['👍', '🚀', '❤️', '🎉', '👀', '🤔']

interface ReactionBarProps {
  reactions: ReactionSummary[]
  onToggleReaction: (emoji: string) => void
  disabled?: boolean
}

/**
 * ReactionBar - Displays reactions and allows toggling emoji reactions
 *
 * Features:
 * - Quick reaction picker with common emojis
 * - Existing reactions with counts
 * - Hover tooltip showing users who reacted
 * - Click to toggle (visual feedback for user's reactions)
 */
export function ReactionBar({
  reactions,
  onToggleReaction,
  disabled = false,
}: ReactionBarProps) {
  const [isPickerOpen, setIsPickerOpen] = React.useState(false)

  const handleReaction = (emoji: string) => {
    onToggleReaction(emoji)
    setIsPickerOpen(false)
  }

  return (
    <div className="flex items-center gap-1 flex-wrap">
      {/* Existing reactions */}
      {reactions.map((reaction) => (
        <ReactionButton
          key={reaction.emoji}
          reaction={reaction}
          onClick={() => onToggleReaction(reaction.emoji)}
          disabled={disabled}
        />
      ))}

      {/* Add reaction button with picker */}
      <Popover open={isPickerOpen} onOpenChange={setIsPickerOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="ghost"
            size="sm"
            className="h-7 px-2 text-muted-foreground hover:text-foreground"
            disabled={disabled}
          >
            <SmilePlus className="h-4 w-4" />
            <span className="sr-only">Add reaction</span>
          </Button>
        </PopoverTrigger>
        <PopoverContent
          className="w-auto p-2"
          side="top"
          align="start"
        >
          <div className="flex gap-1">
            {QUICK_REACTIONS.map((emoji) => (
              <button
                key={emoji}
                type="button"
                onClick={() => handleReaction(emoji)}
                className={cn(
                  'text-lg p-1.5 rounded hover:bg-muted transition-colors',
                  reactions.find((r) => r.emoji === emoji)?.has_user &&
                    'bg-primary/10'
                )}
              >
                {emoji}
              </button>
            ))}
          </div>
        </PopoverContent>
      </Popover>
    </div>
  )
}

interface ReactionButtonProps {
  reaction: ReactionSummary
  onClick: () => void
  disabled?: boolean
}

/**
 * Individual reaction button with count and tooltip
 */
function ReactionButton({ reaction, onClick, disabled }: ReactionButtonProps) {
  const userList = reaction.users.length > 0
    ? reaction.users.slice(0, 10).join(', ') +
      (reaction.users.length > 10 ? ` and ${reaction.users.length - 10} more` : '')
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
              'inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-sm',
              'border transition-colors',
              reaction.has_user
                ? 'bg-primary/10 border-primary/30 text-foreground'
                : 'bg-muted/50 border-border hover:bg-muted',
              disabled && 'opacity-50 cursor-not-allowed'
            )}
          >
            <span>{reaction.emoji}</span>
            <span className="text-xs font-medium">{reaction.count}</span>
          </button>
        </TooltipTrigger>
        <TooltipContent side="top">
          <p className="text-xs">{userList}</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}
