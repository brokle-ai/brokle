'use client'

import type { LLMScorerConfig, BuiltinScorerConfig, RegexScorerConfig } from '../types'

interface ScorerConfigDisplayProps {
  scorerType: 'llm' | 'builtin' | 'regex'
  config: LLMScorerConfig | BuiltinScorerConfig | RegexScorerConfig
}

export function ScorerConfigDisplay({ scorerType, config }: ScorerConfigDisplayProps) {
  if (scorerType === 'llm') {
    const llmConfig = config as LLMScorerConfig
    return (
      <div className="space-y-4">
        <div className="grid grid-cols-2 gap-4">
          <div>
            <p className="text-sm font-medium text-muted-foreground">Model</p>
            <p className="text-sm font-mono">{llmConfig.model}</p>
          </div>
          <div>
            <p className="text-sm font-medium text-muted-foreground">Temperature</p>
            <p className="text-sm">{llmConfig.temperature}</p>
          </div>
        </div>
        {llmConfig.messages && llmConfig.messages.length > 0 && (
          <div>
            <p className="text-sm font-medium text-muted-foreground mb-2">Messages</p>
            <div className="space-y-2">
              {llmConfig.messages.map((msg, index) => (
                <div key={index} className="p-2 bg-muted rounded text-sm">
                  <span className="font-medium capitalize">{msg.role}:</span>
                  <p className="text-muted-foreground mt-1 whitespace-pre-wrap">{msg.content}</p>
                </div>
              ))}
            </div>
          </div>
        )}
        {llmConfig.output_schema && llmConfig.output_schema.length > 0 && (
          <div>
            <p className="text-sm font-medium text-muted-foreground mb-2">Output Schema</p>
            <div className="space-y-1">
              {llmConfig.output_schema.map((field, index) => (
                <div key={index} className="flex items-center gap-2 text-sm">
                  <span className="font-mono">{field.name}</span>
                  <span className="text-muted-foreground">({field.type})</span>
                  {field.description && (
                    <span className="text-muted-foreground">- {field.description}</span>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    )
  }

  if (scorerType === 'builtin') {
    const builtinConfig = config as BuiltinScorerConfig
    return (
      <div className="space-y-4">
        <div>
          <p className="text-sm font-medium text-muted-foreground">Scorer Name</p>
          <p className="text-sm font-mono">{builtinConfig.scorer_name}</p>
        </div>
        {Object.keys(builtinConfig.config || {}).length > 0 && (
          <div>
            <p className="text-sm font-medium text-muted-foreground mb-2">Configuration</p>
            <pre className="p-2 bg-muted rounded text-xs overflow-auto">
              {JSON.stringify(builtinConfig.config, null, 2)}
            </pre>
          </div>
        )}
      </div>
    )
  }

  if (scorerType === 'regex') {
    const regexConfig = config as RegexScorerConfig
    return (
      <div className="space-y-4">
        <div>
          <p className="text-sm font-medium text-muted-foreground">Pattern</p>
          <p className="text-sm font-mono bg-muted px-2 py-1 rounded">{regexConfig.pattern}</p>
        </div>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <p className="text-sm font-medium text-muted-foreground">Score Name</p>
            <p className="text-sm">{regexConfig.score_name}</p>
          </div>
          {regexConfig.capture_group !== undefined && (
            <div>
              <p className="text-sm font-medium text-muted-foreground">Capture Group</p>
              <p className="text-sm">{regexConfig.capture_group}</p>
            </div>
          )}
        </div>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <p className="text-sm font-medium text-muted-foreground">Match Score</p>
            <p className="text-sm">{regexConfig.match_score ?? 1}</p>
          </div>
          <div>
            <p className="text-sm font-medium text-muted-foreground">No Match Score</p>
            <p className="text-sm">{regexConfig.no_match_score ?? 0}</p>
          </div>
        </div>
      </div>
    )
  }

  return null
}
