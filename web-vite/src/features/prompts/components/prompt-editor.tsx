import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type {
  PromptType,
  TextTemplate,
  ChatTemplate,
  ChatMessage,
  PromptTemplate,
} from '../types'
import { PromptTemplateInput } from './prompt-editor/PromptTemplateInput'
import { ChatMessageEditor } from './prompt-editor/ChatMessageEditor'
import { ChatMessageList } from './prompt-editor/ChatMessageList'
import { VariableList } from './prompt-editor/VariableExtractor'

interface TextEditorProps {
  value: TextTemplate
  onChange: (value: TextTemplate) => void
  variables: string[]
}

export function TextEditor({ value, onChange, variables }: TextEditorProps) {
  return <PromptTemplateInput value={value} onChange={onChange} variables={variables} />
}

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

interface PromptEditorProps {
  type: PromptType
  template: PromptTemplate
  onChange: (template: PromptTemplate) => void
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
          <ChatMessageList
            messages={(template as ChatTemplate).messages ?? []}
          />
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
