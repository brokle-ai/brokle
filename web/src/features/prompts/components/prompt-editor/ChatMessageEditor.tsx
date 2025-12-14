'use client'

import { useCallback } from 'react'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Plus, Trash2, GripVertical } from 'lucide-react'
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
} from '@dnd-kit/core'
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import type { ChatMessage } from '../../types'

interface SortableMessageProps {
  message: ChatMessage
  index: number
  onUpdate: (index: number, message: ChatMessage) => void
  onDelete: (index: number) => void
  canDelete: boolean
}

function SortableMessage({
  message,
  index,
  onUpdate,
  onDelete,
  canDelete,
}: SortableMessageProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: `message-${index}`,
  })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  }

  const isPlaceholder = message.type === 'placeholder'

  return (
    <Card ref={setNodeRef} style={style} className="relative">
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <button
              {...attributes}
              {...listeners}
              className="cursor-grab hover:bg-muted rounded p-1 active:cursor-grabbing"
            >
              <GripVertical className="h-4 w-4 text-muted-foreground" />
            </button>
            <Select
              value={message.type}
              onValueChange={(type) =>
                onUpdate(index, {
                  ...message,
                  type,
                  role: type === 'message' ? message.role || 'user' : undefined,
                  content: type === 'message' ? message.content || '' : undefined,
                  name: type === 'placeholder' ? message.name || '' : undefined,
                })
              }
            >
              <SelectTrigger className="w-[140px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="message">Message</SelectItem>
                <SelectItem value="placeholder">Placeholder</SelectItem>
              </SelectContent>
            </Select>
            {!isPlaceholder && (
              <Select
                value={message.role || 'user'}
                onValueChange={(role) => onUpdate(index, { ...message, role })}
              >
                <SelectTrigger className="w-[120px]">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="system">System</SelectItem>
                  <SelectItem value="user">User</SelectItem>
                  <SelectItem value="assistant">Assistant</SelectItem>
                </SelectContent>
              </Select>
            )}
          </div>
          {canDelete && (
            <Button
              variant="ghost"
              size="icon"
              onClick={() => onDelete(index)}
              className="h-8 w-8 text-destructive hover:text-destructive"
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          )}
        </div>
      </CardHeader>
      <CardContent className="pt-2">
        {isPlaceholder ? (
          <div className="space-y-2">
            <Input
              value={message.name || ''}
              onChange={(e) => onUpdate(index, { ...message, name: e.target.value })}
              placeholder="Placeholder name (e.g., history)"
              className="font-mono text-sm"
            />
            <p className="text-xs text-muted-foreground">
              This placeholder will be replaced with content from the SDK
            </p>
          </div>
        ) : (
          <Textarea
            value={message.content || ''}
            onChange={(e) => onUpdate(index, { ...message, content: e.target.value })}
            placeholder={`Enter ${message.role || 'user'} message...`}
            className="min-h-[100px] font-mono text-sm"
          />
        )}
      </CardContent>
    </Card>
  )
}

interface ChatMessageEditorProps {
  messages: ChatMessage[]
  onChange: (messages: ChatMessage[]) => void
}

export function ChatMessageEditor({ messages, onChange }: ChatMessageEditorProps) {
  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  )

  const handleDragEnd = useCallback(
    (event: DragEndEvent) => {
      const { active, over } = event
      if (over && active.id !== over.id) {
        const oldIndex = messages.findIndex((_, i) => `message-${i}` === active.id)
        const newIndex = messages.findIndex((_, i) => `message-${i}` === over.id)
        onChange(arrayMove(messages, oldIndex, newIndex))
      }
    },
    [messages, onChange]
  )

  const handleMessageUpdate = useCallback(
    (index: number, message: ChatMessage) => {
      const newMessages = [...messages]
      newMessages[index] = message
      onChange(newMessages)
    },
    [messages, onChange]
  )

  const handleDeleteMessage = useCallback(
    (index: number) => {
      onChange(messages.filter((_, i) => i !== index))
    },
    [messages, onChange]
  )

  const handleAddMessage = useCallback(
    (type: 'message' | 'placeholder') => {
      const newMessage: ChatMessage =
        type === 'message'
          ? { type: 'message', role: 'user', content: '' }
          : { type: 'placeholder', name: '' }
      onChange([...messages, newMessage])
    },
    [messages, onChange]
  )

  return (
    <div className="space-y-4">
      <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
        <SortableContext items={messages.map((_, i) => `message-${i}`)} strategy={verticalListSortingStrategy}>
          <div className="space-y-3">
            {messages.map((message, index) => (
              <SortableMessage
                key={`message-${index}`}
                message={message}
                index={index}
                onUpdate={handleMessageUpdate}
                onDelete={handleDeleteMessage}
                canDelete={messages.length > 1}
              />
            ))}
          </div>
        </SortableContext>
      </DndContext>

      <div className="flex gap-2">
        <Button variant="outline" size="sm" onClick={() => handleAddMessage('message')}>
          <Plus className="mr-2 h-4 w-4" />
          Add Message
        </Button>
        <Button variant="outline" size="sm" onClick={() => handleAddMessage('placeholder')}>
          <Plus className="mr-2 h-4 w-4" />
          Add Placeholder
        </Button>
      </div>
    </div>
  )
}
