/**
 * Tool/Function Calling types for the Playground
 *
 * Follows OpenAI format as the canonical representation.
 * Other providers (Anthropic, etc.) will have their formats converted
 * at the backend level.
 */

/**
 * Tool definition (OpenAI format)
 */
export interface Tool {
  /** Unique identifier for the tool in the UI */
  id: string
  /** Tool type - currently only "function" is supported */
  type: 'function'
  /** Function definition */
  function: ToolFunction
}

/**
 * Function definition within a tool
 */
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

/**
 * Tool call returned by the model
 */
export interface ToolCall {
  /** Unique identifier for this tool call */
  id: string
  /** Tool type */
  type: 'function'
  /** Function call details */
  function: {
    /** Name of the function to call */
    name: string
    /** Arguments as a JSON string */
    arguments: string
  }
}

/**
 * Tool choice configuration
 * Controls how the model selects which tools to use
 */
export type ToolChoice =
  | 'auto' // Model decides whether to call tools (default)
  | 'none' // Model will not call any tools
  | 'required' // Model must call at least one tool
  | SpecificToolChoice // Model must call a specific tool

/**
 * Force a specific tool to be called
 */
export interface SpecificToolChoice {
  type: 'function'
  function: {
    name: string
  }
}

/**
 * Response format configuration for structured outputs
 */
export type ResponseFormat =
  | { type: 'text' }
  | { type: 'json_object' }
  | JsonSchemaResponseFormat

/**
 * JSON Schema response format for structured outputs
 */
export interface JsonSchemaResponseFormat {
  type: 'json_schema'
  json_schema: {
    /** Name for the schema (required by OpenAI) */
    name: string
    /** The JSON Schema definition */
    schema: Record<string, unknown>
    /** Enable strict schema validation (default: true) */
    strict?: boolean
    /** Optional description */
    description?: string
  }
}

/**
 * Helper to create a new empty tool
 */
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

/**
 * Helper to create a tool from a JSON definition
 */
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

/**
 * Validate a tool definition
 */
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
    try {
      // Basic validation that parameters is a valid JSON Schema-like object
      if (typeof tool.function.parameters !== 'object') {
        errors.push('Parameters must be an object')
      }
    } catch {
      errors.push('Invalid parameters schema')
    }
  }

  return errors
}

/**
 * Convert tools array to the format expected by the API
 * (removes the UI-only 'id' field)
 */
export function toolsToAPIFormat(tools: Tool[]): object[] {
  return tools.map((tool) => ({
    type: tool.type,
    function: tool.function,
  }))
}

/**
 * Parse tool calls from API response
 */
export function parseToolCalls(rawToolCalls: unknown[]): ToolCall[] {
  if (!Array.isArray(rawToolCalls)) return []

  return rawToolCalls.map((tc) => {
    const toolCall = tc as Record<string, unknown>
    const func = toolCall.function as Record<string, unknown>

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

/**
 * Pretty print tool call arguments
 */
export function formatToolCallArguments(args: string): string {
  try {
    const parsed = JSON.parse(args)
    return JSON.stringify(parsed, null, 2)
  } catch {
    return args
  }
}
