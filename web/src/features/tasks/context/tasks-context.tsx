'use client'

import React, { useState } from 'react'
import useDialogState from '@/hooks/use-dialog-state'
import { type Task } from '../data/schema'

type TasksDialogType = 'create' | 'update' | 'delete' | 'import'

type TasksContextType = {
  open: TasksDialogType | null
  setOpen: (str: TasksDialogType | null) => void
  currentRow: Task | null
  setCurrentRow: React.Dispatch<React.SetStateAction<Task | null>>
  projectSlug?: string
}

const TasksContext = React.createContext<TasksContextType | null>(null)

interface TasksProviderProps {
  children: React.ReactNode
  projectSlug?: string
}

export function TasksProvider({ children, projectSlug }: TasksProviderProps) {
  const [open, setOpen] = useDialogState<TasksDialogType>(null)
  const [currentRow, setCurrentRow] = useState<Task | null>(null)

  return (
    <TasksContext value={{ open, setOpen, currentRow, setCurrentRow, projectSlug }}>
      {children}
    </TasksContext>
  )
}

export const useTasks = () => {
  const tasksContext = React.useContext(TasksContext)

  if (!tasksContext) {
    throw new Error('useTasks has to be used within <TasksProvider>')
  }

  return tasksContext
}