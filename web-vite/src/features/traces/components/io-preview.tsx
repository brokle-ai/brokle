import type { ChatMlMessage } from '../api/types'

interface IoPreviewProps {
  value: string | null | undefined
  label: string
}

// Fallback order: ChatML -> generic JSON -> plain text. No MIME
// auto-detect, no truncation warning, no error boundary — we lean on
// the component tree's ancestor boundary for anything that throws here
// (in practice only the JSON.parse branch can, and it's caught).
export function IoPreview({ value, label }: IoPreviewProps) {
  if (!value) {
    return (
      <div className="text-sm italic text-muted-foreground">
        No {label.toLowerCase()} data
      </div>
    )
  }

  const trimmed = value.trim()
  const looksJsonish = trimmed.startsWith('{') || trimmed.startsWith('[')

  if (looksJsonish) {
    try {
      const parsed: unknown = JSON.parse(trimmed)
      if (isChatMl(parsed)) {
        return <ChatView messages={parsed} label={label} />
      }
      return <JsonView data={parsed} label={label} />
    } catch {
      // fall through to plain-text render
    }
  }

  return <TextView value={value} label={label} />
}

function isChatMl(data: unknown): data is ChatMlMessage[] {
  if (!Array.isArray(data) || data.length === 0) return false
  return data.every(
    (item) =>
      typeof item === 'object' &&
      item !== null &&
      'role' in item &&
      typeof (item as { role: unknown }).role === 'string',
  )
}

function roleBadgeClass(role: string): string {
  switch (role) {
    case 'user':
      return 'bg-blue-100 text-blue-700 dark:bg-blue-900/50 dark:text-blue-300'
    case 'assistant':
      return 'bg-green-100 text-green-700 dark:bg-green-900/50 dark:text-green-300'
    case 'system':
      return 'bg-purple-100 text-purple-700 dark:bg-purple-900/50 dark:text-purple-300'
    case 'tool':
      return 'bg-orange-100 text-orange-700 dark:bg-orange-900/50 dark:text-orange-300'
    default:
      return 'bg-muted text-muted-foreground'
  }
}

function ChatView({
  messages,
  label,
}: {
  messages: ChatMlMessage[]
  label: string
}) {
  return (
    <div className="space-y-2">
      <h4 className="text-sm font-medium">{label}</h4>
      <div className="space-y-3 rounded-md border bg-muted/30 p-4">
        {messages.map((msg, idx) => (
          <div key={idx} className="space-y-1">
            <div className="flex items-center gap-2">
              <span
                className={`rounded px-2 py-0.5 text-xs font-medium ${roleBadgeClass(msg.role)}`}
              >
                {msg.role}
              </span>
              {msg.name && (
                <span className="text-xs text-muted-foreground">
                  ({msg.name})
                </span>
              )}
            </div>
            {msg.content != null && msg.content.length > 0 && (
              <div className="whitespace-pre-wrap border-l-2 pl-2 text-sm">
                {msg.content}
              </div>
            )}
            {msg.tool_calls && msg.tool_calls.length > 0 && (
              <div className="space-y-1 pl-2">
                {msg.tool_calls.map((tool, toolIdx) => (
                  <div
                    key={toolIdx}
                    className="rounded border border-blue-200 bg-blue-50 p-2 text-xs dark:border-blue-800 dark:bg-blue-950/50"
                  >
                    <div className="font-medium text-blue-900 dark:text-blue-100">
                      Tool: {tool.function?.name ?? 'unknown'}
                    </div>
                    {tool.function?.arguments && (
                      <div className="mt-1 font-mono text-blue-700 dark:text-blue-300">
                        {tool.function.arguments}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
            {msg.tool_call_id && (
              <div className="pl-2 text-xs text-muted-foreground">
                Tool call ID: {msg.tool_call_id}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}

function JsonView({ data, label }: { data: unknown; label: string }) {
  return (
    <div className="space-y-2">
      <h4 className="text-sm font-medium">{label}</h4>
      <pre className="overflow-x-auto rounded-md border bg-muted p-4 text-xs">
        <code>{JSON.stringify(data, null, 2)}</code>
      </pre>
    </div>
  )
}

function TextView({ value, label }: { value: string; label: string }) {
  return (
    <div className="space-y-2">
      <h4 className="text-sm font-medium">{label}</h4>
      <div className="whitespace-pre-wrap rounded-md border bg-muted p-4 font-mono text-sm">
        {value}
      </div>
    </div>
  )
}
