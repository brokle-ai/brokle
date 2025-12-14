'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { formatDistanceToNow } from 'date-fns'
import { MoreVertical, Play, History, Pencil, Trash2 } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { PromptTypeIcon } from '../common/PromptTypeIcon'
import { LabelList } from '../label-badge'
import type { PromptListItem } from '../../types'

interface PromptCardProps {
  prompt: PromptListItem
  onEdit?: (prompt: PromptListItem) => void
  onDelete?: (prompt: PromptListItem) => void
  onPlayground?: (prompt: PromptListItem) => void
  onViewHistory?: (prompt: PromptListItem) => void
}

export function PromptCard({
  prompt,
  onEdit,
  onDelete,
  onPlayground,
  onViewHistory,
}: PromptCardProps) {
  return (
    <Card className="hover:border-primary/50 transition-colors">
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-2">
            <PromptTypeIcon type={prompt.type} className="h-4 w-4 text-muted-foreground" />
            <CardTitle className="text-lg">{prompt.name}</CardTitle>
            <Badge variant="outline" className="font-mono">
              v{prompt.latest_version}
            </Badge>
          </div>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="h-8 w-8">
                <MoreVertical className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              {onEdit && (
                <DropdownMenuItem onClick={() => onEdit(prompt)}>
                  <Pencil className="mr-2 h-4 w-4" />
                  Edit
                </DropdownMenuItem>
              )}
              {onPlayground && (
                <DropdownMenuItem onClick={() => onPlayground(prompt)}>
                  <Play className="mr-2 h-4 w-4" />
                  Playground
                </DropdownMenuItem>
              )}
              {onViewHistory && (
                <DropdownMenuItem onClick={() => onViewHistory(prompt)}>
                  <History className="mr-2 h-4 w-4" />
                  Version History
                </DropdownMenuItem>
              )}
              {onDelete && (
                <>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem
                    onClick={() => onDelete(prompt)}
                    className="text-destructive focus:text-destructive"
                  >
                    <Trash2 className="mr-2 h-4 w-4" />
                    Delete
                  </DropdownMenuItem>
                </>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        {prompt.description && (
          <p className="text-sm text-muted-foreground line-clamp-2">
            {prompt.description}
          </p>
        )}

        <div className="flex items-center justify-between">
          <div className="flex flex-wrap gap-1">
            <LabelList labels={prompt.labels} />
          </div>
          <span className="text-xs text-muted-foreground">
            {formatDistanceToNow(new Date(prompt.updated_at), { addSuffix: true })}
          </span>
        </div>

        {prompt.tags.length > 0 && (
          <div className="flex flex-wrap gap-1">
            {prompt.tags.map((tag) => (
              <Badge key={tag} variant="secondary" className="text-xs">
                {tag}
              </Badge>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
