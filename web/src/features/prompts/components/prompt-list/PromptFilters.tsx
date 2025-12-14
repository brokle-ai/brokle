'use client'

import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Search, X } from 'lucide-react'
import type { PromptType } from '../../types'

interface PromptFiltersProps {
  search: string
  onSearchChange: (search: string) => void
  type?: PromptType
  onTypeChange: (type: PromptType | undefined) => void
  className?: string
}

export function PromptFilters({
  search,
  onSearchChange,
  type,
  onTypeChange,
  className,
}: PromptFiltersProps) {
  return (
    <div className={className}>
      <div className="flex items-center gap-2">
        <div className="relative flex-1">
          <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search prompts..."
            value={search}
            onChange={(e) => onSearchChange(e.target.value)}
            className="pl-8"
          />
          {search && (
            <Button
              variant="ghost"
              size="icon"
              className="absolute right-1 top-1 h-7 w-7"
              onClick={() => onSearchChange('')}
            >
              <X className="h-4 w-4" />
            </Button>
          )}
        </div>

        <Select
          value={type || 'all'}
          onValueChange={(v) => onTypeChange(v === 'all' ? undefined : (v as PromptType))}
        >
          <SelectTrigger className="w-[140px]">
            <SelectValue placeholder="All types" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All types</SelectItem>
            <SelectItem value="text">Text</SelectItem>
            <SelectItem value="chat">Chat</SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>
  )
}
