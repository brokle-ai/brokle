import type { TextTemplate, ChatTemplate, PromptType } from '../types'

/** Extracts variable names ({{variableName}} pattern) from a prompt template. */
export function extractVariables(
  template: TextTemplate | ChatTemplate,
  type: PromptType
): string[] {
  const variablePattern = /\{\{(\w+)\}\}/g
  const variables = new Set<string>()

  if (type === 'text') {
    const content = (template as TextTemplate).content || ''
    let match
    while ((match = variablePattern.exec(content)) !== null) {
      variables.add(match[1])
    }
  } else {
    const messages = (template as ChatTemplate).messages || []
    for (const msg of messages) {
      if (msg.content) {
        let match
        while ((match = variablePattern.exec(msg.content)) !== null) {
          variables.add(match[1])
        }
      }
    }
  }

  return Array.from(variables)
}
