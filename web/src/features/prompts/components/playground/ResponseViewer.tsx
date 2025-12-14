'use client'

import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { AlertCircle, Loader2 } from 'lucide-react'
import type { ExecutePromptResponse } from '../../types'

interface ResponseViewerProps {
  result: ExecutePromptResponse | null
  isLoading: boolean
  error: string | null
}

// Detect code blocks in markdown-style text
const parseResponseContent = (content: string) => {
  const codeBlockRegex = /```(\w+)?\n([\s\S]*?)```/g
  const parts: Array<{ type: 'text' | 'code'; content: string; language?: string }> = []
  let lastIndex = 0
  let match

  while ((match = codeBlockRegex.exec(content)) !== null) {
    // Add text before code block
    if (match.index > lastIndex) {
      parts.push({
        type: 'text',
        content: content.substring(lastIndex, match.index),
      })
    }

    // Add code block
    parts.push({
      type: 'code',
      content: match[2],
      language: match[1] || 'text',
    })

    lastIndex = match.index + match[0].length
  }

  // Add remaining text
  if (lastIndex < content.length) {
    parts.push({
      type: 'text',
      content: content.substring(lastIndex),
    })
  }

  // If no code blocks found, return entire content as text
  if (parts.length === 0) {
    parts.push({ type: 'text', content })
  }

  return parts
}

export function ResponseViewer({ result, isLoading, error }: ResponseViewerProps) {
  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        <span className="ml-2 text-sm text-muted-foreground">Executing prompt...</span>
      </div>
    )
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    )
  }

  if (!result) {
    return (
      <p className="text-sm text-muted-foreground italic py-12 text-center">
        Click "Execute" to run the prompt with an LLM
      </p>
    )
  }

  if (result.error) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>{result.error}</AlertDescription>
      </Alert>
    )
  }

  const contentParts = result.response?.content ? parseResponseContent(result.response.content) : []

  return (
    <div className="space-y-4">
      {/* Response Content with Syntax Highlighting */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Response</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {contentParts.map((part, index) => (
              <div key={index}>
                {part.type === 'text' ? (
                  <div className="whitespace-pre-wrap text-sm">{part.content}</div>
                ) : (
                  <div className="rounded-md overflow-hidden">
                    <SyntaxHighlighter
                      language={part.language}
                      style={vscDarkPlus}
                      customStyle={{
                        margin: 0,
                        borderRadius: '0.375rem',
                        fontSize: '0.875rem',
                      }}
                    >
                      {part.content}
                    </SyntaxHighlighter>
                  </div>
                )}
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Metadata */}
      <div className="grid grid-cols-2 gap-4">
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium">Performance</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">Model:</span>
              <span className="font-medium">{result.response?.model}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">Latency:</span>
              <span className="font-medium">{result.latency_ms}ms</span>
            </div>
          </CardContent>
        </Card>

        {result.response?.usage && (
          <Card>
            <CardHeader>
              <CardTitle className="text-sm font-medium">Token Usage</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Prompt:</span>
                <span className="font-medium">{result.response.usage.prompt_tokens}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Completion:</span>
                <span className="font-medium">{result.response.usage.completion_tokens}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Total:</span>
                <span className="font-medium font-mono">
                  {result.response.usage.total_tokens}
                </span>
              </div>
              {result.response.cost !== undefined && (
                <div className="flex justify-between text-sm pt-2 border-t">
                  <span className="text-muted-foreground">Estimated Cost:</span>
                  <span className="font-medium">${result.response.cost.toFixed(6)}</span>
                </div>
              )}
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  )
}
