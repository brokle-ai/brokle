/**
 * IOPreview Component - MIME-type-driven rendering for trace/span input/output
 *
 * Renders input/output based on MIME type:
 * - application/json + ChatML → Chat messages UI
 * - application/json + generic → JSON viewer
 * - text/plain → Text viewer
 */

'use client';

import { useState } from 'react';
import { isChatMLFormat, type ChatMessage } from '@/utils/chatml';

interface IOPreviewProps {
  value: string | null | undefined;
  mimeType?: string | null;
  label: string;
  truncated?: boolean;
}

/**
 * IOPreview component for displaying trace/span input or output
 *
 * @param value - Raw value string (JSON or text)
 * @param mimeType - MIME type from backend ("application/json" or "text/plain")
 * @param label - Display label ("Input" or "Output")
 * @param truncated - Whether the value was truncated by backend
 */
export function IOPreview({ value, mimeType, label, truncated }: IOPreviewProps) {
  const [renderError, setRenderError] = useState<string | null>(null);

  if (!value) {
    return (
      <div className="text-sm text-muted-foreground italic">
        No {label.toLowerCase()} data
      </div>
    );
  }

  // Show truncation warning
  if (truncated) {
    return (
      <div className="space-y-2">
        <div className="flex items-center gap-2 text-sm text-amber-600 dark:text-amber-400">
          <svg
            className="h-4 w-4"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
            />
          </svg>
          <span>Content truncated (exceeded 1MB limit)</span>
        </div>
        <TextViewer value={value} label={label} />
      </div>
    );
  }

  // Render based on MIME type
  try {
    if (mimeType === 'application/json') {
      // Try to parse as JSON
      const parsed = JSON.parse(value);

      // Check if ChatML format
      if (isChatMLFormat(parsed)) {
        return <ChatMessagesView messages={parsed} label={label} />;
      }

      // Generic JSON
      return <JSONViewer data={parsed} label={label} />;
    }

    // Default to text viewer
    return <TextViewer value={value} label={label} />;
  } catch (error) {
    // Fallback to text if JSON parsing fails
    const err = error as Error;
    setRenderError(`Invalid JSON format: ${err.message}`);
    return (
      <div className="space-y-2">
        {renderError && (
          <div className="text-sm text-amber-600 dark:text-amber-400">
            ⚠️ {renderError} - displaying as text
          </div>
        )}
        <TextViewer value={value} label={label} />
      </div>
    );
  }
}

/**
 * Chat messages view for ChatML format
 */
function ChatMessagesView({ messages, label }: { messages: ChatMessage[]; label: string }) {
  return (
    <div className="space-y-2">
      <h4 className="text-sm font-medium text-foreground">{label}</h4>
      <div className="space-y-3 rounded-md border border-border bg-muted/50 p-4">
        {messages.map((msg, idx) => (
          <div key={idx} className="space-y-1">
            <div className="flex items-center gap-2">
              <span
                className={`text-xs font-medium px-2 py-0.5 rounded ${getRoleBadgeClass(msg.role)}`}
              >
                {msg.role}
              </span>
              {msg.name && (
                <span className="text-xs text-muted-foreground">({msg.name})</span>
              )}
            </div>

            {msg.content && (
              <div className="text-sm text-foreground whitespace-pre-wrap pl-2 border-l-2 border-border">
                {msg.content}
              </div>
            )}

            {msg.tool_calls && msg.tool_calls.length > 0 && (
              <div className="pl-2 space-y-1">
                {msg.tool_calls.map((tool, toolIdx) => (
                  <div
                    key={toolIdx}
                    className="text-xs bg-blue-50 dark:bg-blue-950 border border-blue-200 dark:border-blue-800 rounded p-2"
                  >
                    <div className="font-medium text-blue-900 dark:text-blue-100">
                      Tool: {tool.function.name}
                    </div>
                    <div className="text-blue-700 dark:text-blue-300 mt-1 font-mono">
                      {tool.function.arguments}
                    </div>
                  </div>
                ))}
              </div>
            )}

            {msg.tool_call_id && (
              <div className="text-xs text-muted-foreground pl-2">
                Tool call ID: {msg.tool_call_id}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}

/**
 * Get badge class for message role
 */
function getRoleBadgeClass(role: string): string {
  switch (role) {
    case 'user':
      return 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300';
    case 'assistant':
      return 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300';
    case 'system':
      return 'bg-purple-100 text-purple-700 dark:bg-purple-900 dark:text-purple-300';
    case 'tool':
      return 'bg-orange-100 text-orange-700 dark:bg-orange-900 dark:text-orange-300';
    default:
      return 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300';
  }
}

/**
 * JSON viewer for generic structured data
 */
function JSONViewer({ data, label }: { data: unknown; label: string }) {
  return (
    <div className="space-y-2">
      <h4 className="text-sm font-medium text-foreground">{label}</h4>
      <pre className="text-xs bg-muted rounded-md p-4 overflow-x-auto border border-border">
        <code>{JSON.stringify(data, null, 2)}</code>
      </pre>
    </div>
  );
}

/**
 * Text viewer for plain text data
 */
function TextViewer({ value, label }: { value: string; label: string }) {
  return (
    <div className="space-y-2">
      <h4 className="text-sm font-medium text-foreground">{label}</h4>
      <div className="text-sm bg-muted rounded-md p-4 whitespace-pre-wrap border border-border font-mono">
        {value}
      </div>
    </div>
  );
}
