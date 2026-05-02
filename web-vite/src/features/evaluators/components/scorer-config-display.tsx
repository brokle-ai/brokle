import { Badge } from '@/components/ui/badge'
import { ReadonlyCode } from '@/editors/readonly-code'
import type {
  BuiltinScorerConfig,
  LLMScorerConfig,
  RegexScorerConfig,
  ScorerType,
} from '../api/types'

interface ScorerConfigDisplayProps {
  scorerType: ScorerType
  config: Record<string, unknown>
}

// Narrow `Record<string, unknown>` at the read site against the
// scorer-type discriminator. The backend owns schema validation so we
// trust `scorer_type` and reflectively cast — the alternative (zod
// per scorer_type) is over-strict for a render-only consumer.
function asLLMConfig(c: Record<string, unknown>): LLMScorerConfig {
  return c as unknown as LLMScorerConfig
}
function asBuiltinConfig(c: Record<string, unknown>): BuiltinScorerConfig {
  return c as unknown as BuiltinScorerConfig
}
function asRegexConfig(c: Record<string, unknown>): RegexScorerConfig {
  return c as unknown as RegexScorerConfig
}

/**
 * Read-only structured visualiser for scorer config.
 *
 * - `llm`: model + temperature + messages (with role badges) + output schema fields.
 * - `builtin`: scorer name + JSON config blob.
 * - `regex`: pattern + score names + match/no-match scores.
 *
 * Falls back to a JSON code block if the type is unrecognised so a
 * future scorer kind doesn't render blank.
 */
export function ScorerConfigDisplay({
  scorerType,
  config,
}: ScorerConfigDisplayProps) {
  if (scorerType === 'llm') {
    const c = asLLMConfig(config)
    return (
      <div className="space-y-4">
        <div className="grid grid-cols-2 gap-4">
          <div>
            <p className="text-xs text-muted-foreground">Model</p>
            <p className="font-mono text-sm">{c.model}</p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">Temperature</p>
            <p className="text-sm">{c.temperature}</p>
          </div>
        </div>
        {c.messages && c.messages.length > 0 ? (
          <div>
            <p className="text-xs text-muted-foreground mb-2">Messages</p>
            <div className="space-y-2">
              {c.messages.map((msg, idx) => (
                <div key={idx} className="rounded-md border bg-muted/40 p-3">
                  <div className="mb-2 flex items-center gap-2">
                    <Badge variant="outline" className="capitalize">
                      {msg.role}
                    </Badge>
                  </div>
                  <pre className="whitespace-pre-wrap break-words font-mono text-xs">
                    {msg.content}
                  </pre>
                </div>
              ))}
            </div>
          </div>
        ) : null}
        {c.output_schema && c.output_schema.length > 0 ? (
          <div>
            <p className="text-xs text-muted-foreground mb-2">Output schema</p>
            <div className="space-y-1">
              {c.output_schema.map((field, idx) => (
                <div
                  key={idx}
                  className="flex flex-wrap items-center gap-2 text-sm"
                >
                  <code className="font-mono">{field.name}</code>
                  <span className="text-xs text-muted-foreground">
                    ({field.type})
                  </span>
                  {field.description ? (
                    <span className="text-xs text-muted-foreground">
                      — {field.description}
                    </span>
                  ) : null}
                  {field.type === 'numeric' &&
                  (field.min_value !== undefined ||
                    field.max_value !== undefined) ? (
                    <span className="text-xs text-muted-foreground">
                      [{field.min_value ?? '−∞'}, {field.max_value ?? '+∞'}]
                    </span>
                  ) : null}
                  {field.type === 'categorical' &&
                  field.categories &&
                  field.categories.length > 0 ? (
                    <span className="text-xs text-muted-foreground">
                      {field.categories.join(' | ')}
                    </span>
                  ) : null}
                </div>
              ))}
            </div>
          </div>
        ) : null}
      </div>
    )
  }

  if (scorerType === 'builtin') {
    const c = asBuiltinConfig(config)
    const cfgKeys = Object.keys(c.config ?? {})
    return (
      <div className="space-y-4">
        <div>
          <p className="text-xs text-muted-foreground">Scorer name</p>
          <p className="font-mono text-sm">{c.scorer_name}</p>
        </div>
        {cfgKeys.length > 0 ? (
          <div>
            <p className="text-xs text-muted-foreground mb-2">Configuration</p>
            <ReadonlyCode
              code={JSON.stringify(c.config, null, 2)}
              language="json"
            />
          </div>
        ) : null}
      </div>
    )
  }

  if (scorerType === 'regex') {
    const c = asRegexConfig(config)
    return (
      <div className="space-y-4">
        <div>
          <p className="text-xs text-muted-foreground">Pattern</p>
          <code className="block rounded bg-muted px-2 py-1 font-mono text-sm">
            {c.pattern}
          </code>
        </div>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <p className="text-xs text-muted-foreground">Score name</p>
            <p className="text-sm">{c.score_name}</p>
          </div>
          {c.capture_group !== undefined ? (
            <div>
              <p className="text-xs text-muted-foreground">Capture group</p>
              <p className="text-sm">{c.capture_group}</p>
            </div>
          ) : null}
        </div>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <p className="text-xs text-muted-foreground">Match score</p>
            <p className="text-sm">{c.match_score ?? 1}</p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">No-match score</p>
            <p className="text-sm">{c.no_match_score ?? 0}</p>
          </div>
        </div>
      </div>
    )
  }

  return (
    <ReadonlyCode code={JSON.stringify(config, null, 2)} language="json" />
  )
}
