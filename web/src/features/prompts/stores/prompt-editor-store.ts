import { create } from 'zustand'
import type { PromptType, ChatMessage, Prompt } from '../types'

interface Variable {
  name: string
  type: 'string'
  required: boolean
}

interface PromptEditorState {
  // Form state
  name: string
  description: string
  type: PromptType
  template: string | ChatMessage[]
  variables: Variable[]
  isDirty: boolean

  // Actions
  setName: (name: string) => void
  setDescription: (description: string) => void
  setType: (type: PromptType) => void
  setTemplate: (template: string | ChatMessage[]) => void
  reset: () => void
  loadPrompt: (prompt: Prompt) => void
}

// Extract variables from template using Mustache pattern
const extractVariables = (template: string | ChatMessage[]): Variable[] => {
  let text = ''

  if (typeof template === 'string') {
    text = template
  } else if (Array.isArray(template)) {
    // For chat templates, extract from message content
    text = template
      .map((m) => m.content || '')
      .join('\n')
  }

  const regex = /\{\{(\w+)\}\}/g
  const variables = new Map<string, Variable>()

  let match
  while ((match = regex.exec(text)) !== null) {
    const name = match[1]
    if (!variables.has(name)) {
      variables.set(name, {
        name,
        type: 'string',
        required: true,
      })
    }
  }

  return Array.from(variables.values())
}

const initialState = {
  name: '',
  description: '',
  type: 'text' as PromptType,
  template: '',
  variables: [],
  isDirty: false,
}

export const usePromptEditorStore = create<PromptEditorState>((set, get) => ({
  ...initialState,

  setName: (name) => set({ name, isDirty: true }),

  setDescription: (description) => set({ description, isDirty: true }),

  setType: (type) => {
    const currentType = get().type
    if (type === currentType) return

    // Reset template when switching types
    const template = type === 'text' ? '' : []
    set({ type, template, variables: [], isDirty: true })
  },

  setTemplate: (template) => {
    const variables = extractVariables(template)
    set({ template, variables, isDirty: true })
  },

  reset: () => set(initialState),

  loadPrompt: (prompt) => {
    // Extract raw template value from structured type
    let template: string | ChatMessage[]
    if (prompt.type === 'text') {
      template = (prompt.template as { content: string }).content || ''
    } else {
      template = (prompt.template as { messages: ChatMessage[] }).messages || []
    }

    set({
      name: prompt.name,
      description: prompt.description || '',
      type: prompt.type,
      template,
      variables: extractVariables(template),
      isDirty: false,
    })
  },
}))
