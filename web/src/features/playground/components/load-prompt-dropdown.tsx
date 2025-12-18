'use client'

import { useState } from 'react'
import { FileText, ChevronDown, Loader2, Database, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Input } from '@/components/ui/input'
import { useQuery } from '@tanstack/react-query'
import { getPrompts, getPromptById } from '@/features/prompts/api/prompts-api'
import type { PromptListItem } from '@/features/prompts/types'
import type { ChatMessage } from '../types'
import { createMessage } from '../types'

interface LoadPromptDropdownProps {
  projectId: string
  // Currently linked prompt info (Opik-style)
  selectedPromptName?: string | null
  selectedPromptVersionNumber?: number | null
  hasUnsavedChanges?: boolean
  onLoad: (data: {
    messages: ChatMessage[]
    promptId: string
    promptName: string
    promptVersionId: string
    promptVersionNumber: number
    originalTemplate: string // JSON string for change detection
  }) => void
  // Unlink callback (clears link but keeps content)
  onUnlink?: () => void
  disabled?: boolean
}

export function LoadPromptDropdown({
  projectId,
  selectedPromptName,
  selectedPromptVersionNumber,
  hasUnsavedChanges,
  onLoad,
  onUnlink,
  disabled,
}: LoadPromptDropdownProps) {
  const [search, setSearch] = useState('')
  const [open, setOpen] = useState(false)
  const isLinked = Boolean(selectedPromptName)

  const { data: promptsResponse, isLoading } = useQuery({
    queryKey: ['prompts', projectId, 'list'],
    queryFn: () => getPrompts({ projectId, limit: 50, page: 1 }),
    enabled: !!projectId && open,
    staleTime: 30000, // Cache for 30s
  })

  const prompts = promptsResponse?.prompts || []
  const filteredPrompts = prompts.filter((p: PromptListItem) =>
    p.name.toLowerCase().includes(search.toLowerCase())
  )

  const handleSelectPrompt = async (promptId: string, promptName: string) => {
    try {
      // Fetch full prompt (returns latest version by default)
      const prompt = await getPromptById(projectId, promptId)
      if (!prompt) return

      let messages: ChatMessage[] = []

      if (prompt.type === 'chat') {
        const template = prompt.template as { messages?: Array<{ role: string; content: string }> }
        if (!template?.messages) return
        messages = template.messages.map((msg) =>
          createMessage(msg.role as ChatMessage['role'], msg.content)
        )
      } else {
        const template = prompt.template as { content?: string }
        messages = [
          createMessage('user', template.content || ''),
        ]
      }

      // Store role+content (without IDs) for consistent change detection
      const originalTemplate = JSON.stringify(
        messages.map(({ role, content }) => ({ role, content }))
      )

      onLoad({
        messages,
        promptId,
        promptName,
        promptVersionId: prompt.version_id,
        promptVersionNumber: prompt.version,
        originalTemplate,
      })
      setOpen(false)
      setSearch('')
    } catch (error) {
      console.error('Failed to load prompt:', error)
    }
  }

  const handleUnlink = (e: React.MouseEvent) => {
    e.stopPropagation()
    e.preventDefault()
    onUnlink?.()
  }

  return (
    <div className="flex">
      <DropdownMenu open={open} onOpenChange={setOpen}>
        <DropdownMenuTrigger asChild>
          {isLinked ? (
            <Button
              variant="outline"
              size="sm"
              disabled={disabled}
              className="rounded-r-none border-r-0"
            >
              <Database className="mr-2 h-4 w-4" />
              <span className="truncate max-w-32">{selectedPromptName}</span>
              {selectedPromptVersionNumber && (
                <Badge variant="outline" className="ml-2 h-4 text-[10px] px-1">
                  v{selectedPromptVersionNumber}
                </Badge>
              )}
              {hasUnsavedChanges && (
                <span className="ml-1 h-2 w-2 rounded-full bg-amber-500" />
              )}
              <ChevronDown className="ml-2 h-4 w-4" />
            </Button>
          ) : (
            <Button variant="outline" size="sm" disabled={disabled}>
              <FileText className="mr-2 h-4 w-4" />
              Load Prompt
              <ChevronDown className="ml-2 h-4 w-4" />
            </Button>
          )}
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" className="w-64">
          <DropdownMenuLabel>Load from saved prompts</DropdownMenuLabel>
          <DropdownMenuSeparator />
          <div className="p-2">
            <Input
              placeholder="Search prompts..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="h-8"
            />
          </div>
          <DropdownMenuSeparator />
          <div className="max-h-[300px] overflow-y-auto">
            {isLoading ? (
              <div className="flex items-center justify-center py-4">
                <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
              </div>
            ) : filteredPrompts.length === 0 ? (
              <div className="px-2 py-4 text-center text-sm text-muted-foreground">
                {prompts.length === 0
                  ? 'No prompts in this project'
                  : 'No prompts match your search'}
              </div>
            ) : (
              filteredPrompts.map((prompt: PromptListItem) => (
                <DropdownMenuItem
                  key={prompt.id}
                  onClick={() => handleSelectPrompt(prompt.id, prompt.name)}
                  className="cursor-pointer"
                >
                  <FileText className="mr-2 h-4 w-4 text-muted-foreground" />
                  <div className="flex flex-col">
                    <span className="truncate">{prompt.name}</span>
                    {prompt.description && (
                      <span className="text-xs text-muted-foreground truncate">
                        {prompt.description}
                      </span>
                    )}
                  </div>
                </DropdownMenuItem>
              ))
            )}
          </div>
        </DropdownMenuContent>
      </DropdownMenu>
      {isLinked && (
        <Button
          variant="outline"
          size="sm"
          className="rounded-l-none border-l-0 px-2"
          onClick={handleUnlink}
          disabled={disabled}
          title="Unlink prompt"
        >
          <X className="h-4 w-4" />
        </Button>
      )}
    </div>
  )
}
