/**
 * Tool/Function Calling types for the Playground.
 * Ported verbatim from web/; OpenAI format is canonical.
 */

export interface Tool {
  /** Unique identifier for the tool in the UI */
  id: string
  /** Tool type - currently only "function" is supported */
  type: 'function'
  /** Function definition */
  function: ToolFunction
}

export interface ToolFunction {
  /** Function name (must be a-z, A-Z, 0-9, underscores, max 64 chars) */
  name: string
  /** Description of what the function does */
  description?: string
  /** JSON Schema defining the function parameters */
  parameters?: Record<string, unknown>
  /** Whether to enable strict schema validation (OpenAI feature) */
  strict?: boolean
}

export interface ToolCall {
  id: string
  type: 'function'
  function: {
    name: string
    arguments: string
  }
}

export type ToolChoice =
  | 'auto'
  | 'none'
  | 'required'
  | SpecificToolChoice

export interface SpecificToolChoice {
  type: 'function'
  function: {
    name: string
  }
}

export type ResponseFormat =
  | { type: 'text' }
  | { type: 'json_object' }
  | JsonSchemaResponseFormat

export interface JsonSchemaResponseFormat {
  type: 'json_schema'
  json_schema: {
    name: string
    schema: Record<string, unknown>
    strict?: boolean
    description?: string
  }
}

export function createEmptyTool(): Tool {
  return {
    id: crypto.randomUUID(),
    type: 'function',
    function: {
      name: '',
      description: '',
      parameters: {
        type: 'object',
        properties: {},
        required: [],
      },
    },
  }
}

export function createToolFromJSON(json: Record<string, unknown>): Tool {
  return {
    id: crypto.randomUUID(),
    type: 'function',
    function: {
      name: (json.name as string) || '',
      description: json.description as string | undefined,
      parameters: json.parameters as Record<string, unknown> | undefined,
    },
  }
}

export function validateTool(tool: Tool): string[] {
  const errors: string[] = []

  if (!tool.function.name) {
    errors.push('Function name is required')
  } else if (!/^[a-zA-Z0-9_]+$/.test(tool.function.name)) {
    errors.push('Function name can only contain letters, numbers, and underscores')
  } else if (tool.function.name.length > 64) {
    errors.push('Function name must be 64 characters or less')
  }

  if (tool.function.parameters) {
    if (typeof tool.function.parameters !== 'object') {
      errors.push('Parameters must be an object')
    }
  }

  return errors
}

export function toolsToAPIFormat(tools: Tool[]): object[] {
  return tools.map((tool) => ({
    type: tool.type,
    function: tool.function,
  }))
}

export function parseToolCalls(rawToolCalls: unknown[]): ToolCall[] {
  if (!Array.isArray(rawToolCalls)) return []

  return rawToolCalls.map((tc) => {
    const toolCall = tc as Record<string, unknown>
    const func = toolCall.function as Record<string, unknown> | undefined

    return {
      id: (toolCall.id as string) || crypto.randomUUID(),
      type: 'function',
      function: {
        name: (func?.name as string) || '',
        arguments: (func?.arguments as string) || '{}',
      },
    }
  })
}

export function formatToolCallArguments(args: string): string {
  try {
    const parsed = JSON.parse(args)
    return JSON.stringify(parsed, null, 2)
  } catch {
    return args
  }
}
