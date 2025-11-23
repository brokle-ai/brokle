/**
 * ChatML (Chat Markup Language) format detection and utilities
 *
 * ChatML is the standard format for LLM conversations:
 * [{ role: "user" | "assistant" | "system" | "tool", content: string, ... }]
 */

export interface ChatMessage {
  role: 'user' | 'assistant' | 'system' | 'tool' | string;
  content?: string | null;
  name?: string;
  tool_calls?: ToolCall[];
  tool_call_id?: string;
}

export interface ToolCall {
  id: string;
  type: 'function';
  function: {
    name: string;
    arguments: string;
  };
}

/**
 * Check if data is in ChatML messages format
 *
 * @param data - Data to check
 * @returns True if ChatML format, false otherwise
 *
 * @example
 * ```typescript
 * isChatMLFormat([{ role: 'user', content: 'Hello' }]) // true
 * isChatMLFormat({ key: 'value' }) // false
 * ```
 */
export function isChatMLFormat(data: unknown): data is ChatMessage[] {
  if (!Array.isArray(data)) {
    return false;
  }

  if (data.length === 0) {
    return false;
  }

  // Check if all items are objects with 'role' field
  return data.every(
    (item) =>
      typeof item === 'object' &&
      item !== null &&
      'role' in item &&
      typeof item.role === 'string'
  );
}

/**
 * Normalize input data to ChatML format if possible
 *
 * Handles various input formats and converts to ChatMessage[]
 *
 * @param input - Raw input data
 * @returns Normalized ChatMessage array or null
 */
export function normalizeToChatML(input: unknown): ChatMessage[] | null {
  if (isChatMLFormat(input)) {
    return input;
  }

  // Try parsing nested array [[messages]]
  if (Array.isArray(input) && input.length === 1 && isChatMLFormat(input[0])) {
    return input[0];
  }

  // Try parsing { messages: [...] }
  if (
    typeof input === 'object' &&
    input !== null &&
    'messages' in input &&
    isChatMLFormat((input as { messages: unknown }).messages)
  ) {
    return (input as { messages: ChatMessage[] }).messages;
  }

  return null;
}

/**
 * Extract tool calls from ChatML messages
 *
 * @param messages - ChatML messages array
 * @returns Array of tool calls found in messages
 */
export function extractToolCalls(messages: ChatMessage[]): ToolCall[] {
  const toolCalls: ToolCall[] = [];

  for (const message of messages) {
    if (message.tool_calls && Array.isArray(message.tool_calls)) {
      toolCalls.push(...message.tool_calls);
    }
  }

  return toolCalls;
}

/**
 * Count messages by role
 *
 * @param messages - ChatML messages array
 * @returns Object with counts per role
 */
export function countMessagesByRole(messages: ChatMessage[]): Record<string, number> {
  const counts: Record<string, number> = {};

  for (const message of messages) {
    const role = message.role;
    counts[role] = (counts[role] || 0) + 1;
  }

  return counts;
}

/**
 * Check if messages contain tool calls
 *
 * @param messages - ChatML messages array
 * @returns True if any message has tool_calls
 */
export function hasToolCalls(messages: ChatMessage[]): boolean {
  return messages.some(
    (msg) => msg.tool_calls && Array.isArray(msg.tool_calls) && msg.tool_calls.length > 0
  );
}

/**
 * Get first and last message roles
 *
 * @param messages - ChatML messages array
 * @returns Object with first and last roles
 */
export function getMessageRoles(messages: ChatMessage[]): {
  firstRole: string | null;
  lastRole: string | null;
} {
  if (messages.length === 0) {
    return { firstRole: null, lastRole: null };
  }

  return {
    firstRole: messages[0].role,
    lastRole: messages[messages.length - 1].role,
  };
}
