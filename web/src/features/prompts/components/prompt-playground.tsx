'use client'

import { useState, useMemo } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Slider } from '@/components/ui/slider'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type {
  PromptVersion,
  ModelConfig,
  ExecutePromptRequest,
  ExecutePromptResponse,
  TextTemplate,
  ChatTemplate,
} from '../types'
import { VariableInputs } from './playground/VariableInputs'
import { ExecutionPanel } from './playground/ExecutionPanel'
import { ResponseViewer } from './playground/ResponseViewer'

// ============================================================================
// Model Config Input
// ============================================================================

interface ModelConfigInputProps {
  config: ModelConfig
  onChange: (config: ModelConfig) => void
}

function ModelConfigInput({ config, onChange }: ModelConfigInputProps) {
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <Label>Model</Label>
        <Select
          value={config.model || 'gpt-4o-mini'}
          onValueChange={(model) => onChange({ ...config, model })}
        >
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="gpt-4o">GPT-4o</SelectItem>
            <SelectItem value="gpt-4o-mini">GPT-4o Mini</SelectItem>
            <SelectItem value="gpt-4-turbo">GPT-4 Turbo</SelectItem>
            <SelectItem value="gpt-3.5-turbo">GPT-3.5 Turbo</SelectItem>
            <SelectItem value="claude-3-5-sonnet-20241022">Claude 3.5 Sonnet</SelectItem>
            <SelectItem value="claude-3-haiku-20240307">Claude 3 Haiku</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <Label>Temperature</Label>
          <span className="text-sm text-muted-foreground">
            {config.temperature ?? 0.7}
          </span>
        </div>
        <Slider
          value={[config.temperature ?? 0.7]}
          onValueChange={([value]) => onChange({ ...config, temperature: value })}
          min={0}
          max={2}
          step={0.1}
          className="w-full"
        />
      </div>

      <div className="space-y-2">
        <Label>Max Tokens</Label>
        <Input
          type="number"
          value={config.max_tokens ?? 1024}
          onChange={(e) =>
            onChange({ ...config, max_tokens: parseInt(e.target.value) || 1024 })
          }
          min={1}
          max={128000}
        />
      </div>

      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <Label>Top P</Label>
          <span className="text-sm text-muted-foreground">
            {config.top_p ?? 1}
          </span>
        </div>
        <Slider
          value={[config.top_p ?? 1]}
          onValueChange={([value]) => onChange({ ...config, top_p: value })}
          min={0}
          max={1}
          step={0.05}
          className="w-full"
        />
      </div>
    </div>
  )
}

// ============================================================================
// Compiled Preview
// ============================================================================

interface CompiledMessage {
  role?: string
  content?: string
}

interface CompiledPreviewProps {
  compiled: { content?: string; messages?: CompiledMessage[] } | null
  type: 'text' | 'chat'
}

function CompiledPreview({ compiled, type }: CompiledPreviewProps) {
  if (!compiled) {
    return (
      <p className="text-sm text-muted-foreground italic">
        Enter variable values to see the compiled prompt
      </p>
    )
  }

  if (type === 'text') {
    return (
      <pre className="whitespace-pre-wrap rounded-md bg-muted p-4 font-mono text-sm">
        {compiled.content || ''}
      </pre>
    )
  }

  const messages = compiled.messages
  if (!messages || !Array.isArray(messages)) {
    return (
      <pre className="whitespace-pre-wrap rounded-md bg-muted p-4 font-mono text-sm">
        {JSON.stringify(compiled, null, 2)}
      </pre>
    )
  }

  return (
    <div className="space-y-2">
      {messages.map((msg: CompiledMessage, i: number) => (
        <div
          key={i}
          className={
            msg.role === 'system'
              ? 'rounded-md p-3 bg-purple-100 dark:bg-purple-900/30'
              : msg.role === 'assistant'
              ? 'rounded-md p-3 bg-blue-100 dark:bg-blue-900/30'
              : 'rounded-md p-3 bg-muted'
          }
        >
          <div className="mb-1 text-xs font-medium uppercase text-muted-foreground">
            {msg.role}
          </div>
          <pre className="whitespace-pre-wrap font-mono text-sm">
            {msg.content}
          </pre>
        </div>
      ))}
    </div>
  )
}

// ============================================================================
// Main Playground
// ============================================================================

interface PromptPlaygroundProps {
  version: PromptVersion
  promptType: 'text' | 'chat'
  onExecute: (request: ExecutePromptRequest) => Promise<ExecutePromptResponse>
  isExecuting?: boolean
}

export function PromptPlayground({
  version,
  promptType,
  onExecute,
  isExecuting,
}: PromptPlaygroundProps) {
  const [variableValues, setVariableValues] = useState<Record<string, string>>({})
  const [configOverrides, setConfigOverrides] = useState<ModelConfig>(
    version.config || { model: 'gpt-4o-mini', temperature: 0.7, max_tokens: 1024 }
  )
  const [result, setResult] = useState<ExecutePromptResponse | null>(null)
  const [error, setError] = useState<string | null>(null)

  // Compile preview (client-side Mustache-like substitution)
  const compiledPreview = useMemo((): { content?: string; messages?: CompiledMessage[] } | null => {
    try {
      if (promptType === 'text') {
        const template = version.template as TextTemplate
        let content = template.content || ''
        for (const [key, value] of Object.entries(variableValues)) {
          content = content.replace(new RegExp(`\\{\\{${key}\\}\\}`, 'g'), value)
        }
        return { content }
      } else {
        const template = version.template as ChatTemplate
        const messages: CompiledMessage[] = (template.messages || [])
          .filter((msg) => msg.type !== 'placeholder')
          .map((msg) => {
            let content = msg.content || ''
            for (const [key, value] of Object.entries(variableValues)) {
              content = content.replace(new RegExp(`\\{\\{${key}\\}\\}`, 'g'), value)
            }
            return { role: msg.role, content }
          })
        return { messages }
      }
    } catch {
      return null
    }
  }, [version.template, promptType, variableValues])

  const handleExecute = async () => {
    setError(null)
    try {
      const response = await onExecute({
        variables: variableValues,
        config_overrides: configOverrides,
      })
      setResult(response)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Execution failed')
    }
  }

  const missingVariables = version.variables.filter(
    (v) => !variableValues[v] || variableValues[v].trim() === ''
  )

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
      {/* Left Panel - Input */}
      <div className="space-y-6">
        <VariableInputs
          variables={version.variables}
          values={variableValues}
          onChange={setVariableValues}
        />

        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Model Configuration</CardTitle>
          </CardHeader>
          <CardContent>
            <ModelConfigInput config={configOverrides} onChange={setConfigOverrides} />
          </CardContent>
        </Card>

        <ExecutionPanel
          onExecute={handleExecute}
          isExecuting={isExecuting || false}
          missingVariables={missingVariables}
        />
      </div>

      {/* Right Panel - Output */}
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Preview & Result</CardTitle>
          </CardHeader>
          <CardContent>
            <Tabs defaultValue="preview">
              <TabsList className="mb-4">
                <TabsTrigger value="preview">Compiled Preview</TabsTrigger>
                <TabsTrigger value="result">Execution Result</TabsTrigger>
              </TabsList>
              <TabsContent value="preview" className="mt-0">
                <CompiledPreview compiled={compiledPreview} type={promptType} />
              </TabsContent>
              <TabsContent value="result" className="mt-0">
                <ResponseViewer
                  result={result}
                  isLoading={isExecuting || false}
                  error={error}
                />
              </TabsContent>
            </Tabs>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
