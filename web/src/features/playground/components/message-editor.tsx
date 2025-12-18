'use client'

import { useMemo } from 'react'
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
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Trash2, GripVertical } from 'lucide-react'
import type { ChatMessage } from '../types'

interface MessageEditorProps {
  messages: ChatMessage[]
  onChange: (messages: ChatMessage[]) => void
  disabled?: boolean
}

interface SortableMessageCardProps {
  message: ChatMessage
  totalCount: number
  onRoleChange: (value: string) => void
  onContentChange: (value: string) => void
  onRemove: () => void
  disabled?: boolean
}

function SortableMessageCard({
  message,
  totalCount,
  onRoleChange,
  onContentChange,
  onRemove,
  disabled,
}: SortableMessageCardProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: message.id })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
    zIndex: isDragging ? 10 : 'auto',
  }

  // Role-based styling
  const roleStyles: Record<string, string> = {
    system: 'border-l-4 border-l-blue-500/50',
    user: 'border-l-4 border-l-green-500/50',
    assistant: 'border-l-4 border-l-purple-500/50',
  }

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={`space-y-2 p-4 border rounded-lg bg-card ${roleStyles[message.role] || ''}`}
    >
      <div className="flex items-center gap-2">
        <button
          type="button"
          className="cursor-grab touch-none text-muted-foreground hover:text-foreground"
          {...attributes}
          {...listeners}
        >
          <GripVertical className="h-4 w-4" />
        </button>

        <Select
          value={message.role}
          onValueChange={onRoleChange}
          disabled={disabled}
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

        <div className="ml-auto" />

        {totalCount > 1 && (
          <Button
            variant="ghost"
            size="icon"
            onClick={onRemove}
            disabled={disabled}
            className="h-8 w-8"
          >
            <Trash2 className="h-4 w-4 text-muted-foreground hover:text-destructive" />
          </Button>
        )}
      </div>
      <Textarea
        value={message.content}
        onChange={(e) => onContentChange(e.target.value)}
        placeholder={`Enter ${message.role} message... Use {{variable}} for dynamic content.`}
        className="min-h-[80px] font-mono text-sm"
        rows={3}
        disabled={disabled}
      />
    </div>
  )
}

export function MessageEditor({ messages, onChange, disabled }: MessageEditorProps) {
  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8, // Require 8px movement before starting drag
      },
    }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  )

  const messageIds = useMemo(() => messages.map((m) => m.id), [messages])

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event

    if (over && active.id !== over.id) {
      const oldIndex = messages.findIndex((m) => m.id === active.id)
      const newIndex = messages.findIndex((m) => m.id === over.id)

      if (oldIndex !== -1 && newIndex !== -1) {
        onChange(arrayMove(messages, oldIndex, newIndex))
      }
    }
  }

  const handleMessageChange = (index: number, field: 'role' | 'content', value: string) => {
    const newMessages = [...messages]
    newMessages[index] = { ...newMessages[index], [field]: value as ChatMessage['role'] }
    onChange(newMessages)
  }

  const handleRemoveMessage = (index: number) => {
    if (messages.length <= 1) return
    onChange(messages.filter((_, i) => i !== index))
  }

  return (
    <div className="space-y-3">
      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragEnd={handleDragEnd}
      >
        <SortableContext
          items={messageIds}
          strategy={verticalListSortingStrategy}
        >
          {messages.map((message, index) => (
            <SortableMessageCard
              key={message.id}
              message={message}
              totalCount={messages.length}
              onRoleChange={(value) => handleMessageChange(index, 'role', value)}
              onContentChange={(value) => handleMessageChange(index, 'content', value)}
              onRemove={() => handleRemoveMessage(index)}
              disabled={disabled}
            />
          ))}
        </SortableContext>
      </DndContext>
    </div>
  )
}
