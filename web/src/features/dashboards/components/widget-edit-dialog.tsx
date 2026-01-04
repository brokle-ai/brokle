'use client'

import { useState } from 'react'
import { Plus, Edit } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { ScrollArea } from '@/components/ui/scroll-area'
import { WidgetForm } from './widget-form'
import type { Widget, WidgetType } from '../types'

interface WidgetEditDialogProps {
  widget?: Widget
  onSave: (widget: Omit<Widget, 'id'> | Widget) => void | Promise<void>
  trigger?: React.ReactNode
  isLoading?: boolean
  // Controlled mode props
  open?: boolean
  onOpenChange?: (open: boolean) => void
  // Pre-selected type from palette
  defaultType?: WidgetType
}

export function WidgetEditDialog({
  widget,
  onSave,
  trigger,
  isLoading,
  open: controlledOpen,
  onOpenChange: controlledOnOpenChange,
  defaultType,
}: WidgetEditDialogProps) {
  // Support both controlled and uncontrolled modes
  const [internalOpen, setInternalOpen] = useState(false)
  const [saving, setSaving] = useState(false)

  const isControlled = controlledOpen !== undefined
  const open = isControlled ? controlledOpen : internalOpen
  const setOpen = isControlled
    ? (value: boolean) => controlledOnOpenChange?.(value)
    : setInternalOpen

  const isEditing = Boolean(widget?.id)

  const handleSubmit = async (widgetData: Omit<Widget, 'id'> | Widget) => {
    try {
      setSaving(true)
      await onSave(widgetData)
      setOpen(false)
    } catch (error) {
      console.error('Failed to save widget:', error)
    } finally {
      setSaving(false)
    }
  }

  const defaultTrigger = isEditing ? (
    <Button variant="ghost" size="sm">
      <Edit className="h-4 w-4 mr-1" />
      Edit
    </Button>
  ) : (
    <Button>
      <Plus className="h-4 w-4 mr-2" />
      Add Widget
    </Button>
  )

  // In controlled mode without a trigger, we don't need DialogTrigger
  const dialogContent = (
    <DialogContent className="sm:max-w-[700px] max-h-[90vh]">
      <DialogHeader>
        <DialogTitle>
          {isEditing ? `Edit Widget: ${widget?.title}` : 'Create Widget'}
        </DialogTitle>
        <DialogDescription>
          {isEditing
            ? 'Modify the widget configuration and query.'
            : 'Create a new widget to visualize your observability data.'}
        </DialogDescription>
      </DialogHeader>
      <ScrollArea className="max-h-[calc(90vh-120px)] pr-4">
        <WidgetForm
          widget={widget}
          onSubmit={handleSubmit}
          onCancel={() => setOpen(false)}
          isLoading={isLoading || saving}
          defaultType={defaultType}
        />
      </ScrollArea>
    </DialogContent>
  )

  // If in controlled mode without trigger, render without DialogTrigger
  if (isControlled && !trigger) {
    return (
      <Dialog open={open} onOpenChange={setOpen}>
        {dialogContent}
      </Dialog>
    )
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{trigger ?? defaultTrigger}</DialogTrigger>
      {dialogContent}
    </Dialog>
  )
}

interface AddWidgetButtonProps {
  onAdd: (widget: Omit<Widget, 'id'>) => void | Promise<void>
  disabled?: boolean
}

export function AddWidgetButton({ onAdd, disabled }: AddWidgetButtonProps) {
  return (
    <WidgetEditDialog
      onSave={onAdd as (widget: Omit<Widget, 'id'> | Widget) => void}
      trigger={
        <Button disabled={disabled}>
          <Plus className="h-4 w-4 mr-2" />
          Add Widget
        </Button>
      }
    />
  )
}
