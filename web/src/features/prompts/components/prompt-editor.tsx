'use client'

import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { cn } from '@/lib/utils'
import type { PromptType, TextTemplate, ChatTemplate, ChatMessage } from '../types'
import { PromptTemplateInput } from './prompt-editor/PromptTemplateInput'
import { ChatMessageEditor } from './prompt-editor/ChatMessageEditor'
import { VariableList } from './prompt-editor/VariableExtractor'

// ============================================================================
// Text Editor (wrapper for PromptTemplateInput)
// ============================================================================

interface TextEditorProps {
  value: TextTemplate
  onChange: (value: TextTemplate) => void
  variables: string[]
}

export function TextEditor({ value, onChange, variables }: TextEditorProps) {
  return <PromptTemplateInput value={value} onChange={onChange} variables={variables} />
}

// ============================================================================
// Chat Editor (uses new ChatMessageEditor with drag-and-drop)
// ============================================================================

interface ChatEditorProps {
  value: ChatTemplate
  onChange: (value: ChatTemplate) => void
  variables: string[]
}

export function ChatEditor({ value, onChange, variables }: ChatEditorProps) {
  const messages = value.messages || []

  const handleMessagesChange = (newMessages: ChatMessage[]) => {
    onChange({ messages: newMessages })
  }

  return (
    <div className="space-y-4">
      <ChatMessageEditor messages={messages} onChange={handleMessagesChange} />
      <div className="space-y-2">
        <Label>Detected Variables</Label>
        <VariableList variables={variables} />
      </div>
    </div>
  )
}

// ============================================================================
// Main Prompt Editor
// ============================================================================

interface PromptEditorProps {
  type: PromptType
  template: TextTemplate | ChatTemplate
  onChange: (template: TextTemplate | ChatTemplate) => void
  onTypeChange?: (type: PromptType) => void
  variables: string[]
  readOnly?: boolean
}

export function PromptEditor({
  type,
  template,
  onChange,
  onTypeChange,
  variables,
  readOnly,
}: PromptEditorProps) {
  if (readOnly) {
    return (
      <div className="space-y-4">
        {type === 'text' ? (
          <pre className="whitespace-pre-wrap rounded-md bg-muted p-4 font-mono text-sm">
            {(template as TextTemplate).content}
          </pre>
        ) : (
          <div className="space-y-2">
            {(template as ChatTemplate).messages?.map((msg, i) => (
              <div
                key={i}
                className={cn(
                  'rounded-md p-3',
                  msg.type === 'placeholder'
                    ? 'bg-amber-100 dark:bg-amber-900/30'
                    : msg.role === 'system'
                    ? 'bg-purple-100 dark:bg-purple-900/30'
                    : msg.role === 'assistant'
                    ? 'bg-blue-100 dark:bg-blue-900/30'
                    : 'bg-muted'
                )}
              >
                <div className="mb-1 text-xs font-medium uppercase text-muted-foreground">
                  {msg.type === 'placeholder' ? `[${msg.name}]` : msg.role}
                </div>
                <pre className="whitespace-pre-wrap font-mono text-sm">
                  {msg.content}
                </pre>
              </div>
            ))}
          </div>
        )}
        <div className="space-y-2">
          <Label>Variables</Label>
          <VariableList variables={variables} />
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {onTypeChange && (
        <div className="space-y-2">
          <Label>Template Type</Label>
          <Select value={type} onValueChange={(v) => onTypeChange(v as PromptType)}>
            <SelectTrigger className="w-[180px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="text">Text</SelectItem>
              <SelectItem value="chat">Chat</SelectItem>
            </SelectContent>
          </Select>
        </div>
      )}

      {type === 'text' ? (
        <TextEditor
          value={template as TextTemplate}
          onChange={onChange}
          variables={variables}
        />
      ) : (
        <ChatEditor
          value={template as ChatTemplate}
          onChange={onChange}
          variables={variables}
        />
      )}
    </div>
  )
}
