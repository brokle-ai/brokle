import { useMemo, useState, type FormEvent } from 'react'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { PromptEditor } from '@/editors/prompt-editor'
import type { PromptFormState, PromptType } from '../api/types'

// Mustache/Jinja-style `{{ name }}` extractor. The backend runs its own
// variable detection at save time (sqlc-hydrated via the template
// parser); this is for the preview list only — source of truth lives
// server-side.
const VARIABLE_REGEX = /\{\{\s*([a-zA-Z_][a-zA-Z0-9_.-]*)\s*\}\}/g

function deriveVariables(body: string): string[] {
  const found = new Set<string>()
  for (const match of body.matchAll(VARIABLE_REGEX)) {
    const name = match[1]
    if (name) found.add(name)
  }
  return Array.from(found).sort()
}

interface PromptFormProps {
  mode: 'create' | 'edit'
  initial: PromptFormState
  // Parent wires useMutation + navigation.
  onSubmit: (state: PromptFormState) => void
  isSubmitting?: boolean
  error?: string | null
  onCancel?: () => void
}

// Shared form for create + edit. In edit mode, `name` and `type` are
// locked — the backend treats those as immutable identity fields; new
// versions only carry template + commit message + labels.
export function PromptForm({
  mode,
  initial,
  onSubmit,
  isSubmitting,
  error,
  onCancel,
}: PromptFormProps) {
  const [state, setState] = useState<PromptFormState>(initial)

  const derivedVariables = useMemo(
    () => deriveVariables(state.body),
    [state.body],
  )

  const locked = mode === 'edit'

  function update<K extends keyof PromptFormState>(
    key: K,
    value: PromptFormState[K],
  ) {
    setState((prev) => ({ ...prev, [key]: value }))
  }

  function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    onSubmit(state)
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Basics</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="prompt-name">Name</Label>
            <Input
              id="prompt-name"
              value={state.name}
              onChange={(e) => update('name', e.target.value)}
              disabled={locked || isSubmitting}
              required
              placeholder="welcome-email"
            />
            {locked ? (
              <p className="text-xs text-muted-foreground">
                Name is immutable — edits create a new version.
              </p>
            ) : null}
          </div>

          <div className="space-y-2">
            <Label htmlFor="prompt-type">Type</Label>
            <Select
              value={state.type}
              onValueChange={(v) => update('type', v as PromptType)}
              disabled={locked || isSubmitting}
            >
              <SelectTrigger id="prompt-type">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="text">Text</SelectItem>
                <SelectItem value="chat">Chat</SelectItem>
              </SelectContent>
            </Select>
            {locked ? (
              <p className="text-xs text-muted-foreground">
                Type is immutable — edits create a new version.
              </p>
            ) : null}
          </div>

          <div className="space-y-2">
            <Label htmlFor="prompt-tags">Tags</Label>
            <Input
              id="prompt-tags"
              value={state.tagsCsv}
              onChange={(e) => update('tagsCsv', e.target.value)}
              disabled={isSubmitting}
              placeholder="onboarding, email"
            />
            <p className="text-xs text-muted-foreground">
              Comma-separated.
            </p>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Template</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {state.type === 'chat' ? (
            <p className="rounded-md border border-amber-300/50 bg-amber-50 p-3 text-xs text-amber-900 dark:border-amber-700/40 dark:bg-amber-950/40 dark:text-amber-200">
              Chat-message editing is not yet implemented in the Vite
              dashboard. Use the text editor below to enter raw JSON
              with a <code>messages</code> array, or edit the chat
              template in the legacy dashboard.
            </p>
          ) : null}
          <PromptEditor
            value={state.body}
            onChange={(next) => update('body', next)}
            language={state.type === 'chat' ? 'json' : 'jinja'}
            readOnly={isSubmitting}
            placeholder="Hello {{ name }}!"
            minHeight="240px"
            ariaLabel="Prompt template body"
          />
          <div className="space-y-1">
            <Label>Detected variables</Label>
            {derivedVariables.length === 0 ? (
              <p className="text-xs text-muted-foreground">
                None yet — write <code className="font-mono">{'{{ name }}'}</code>{' '}
                in the template above.
              </p>
            ) : (
              <div className="flex flex-wrap gap-1">
                {derivedVariables.map((v) => (
                  <span
                    key={v}
                    className="rounded-md border bg-muted px-2 py-0.5 font-mono text-xs"
                  >
                    {v}
                  </span>
                ))}
              </div>
            )}
          </div>
          <div className="space-y-2">
            <Label htmlFor="prompt-vars">Declared variables (override)</Label>
            <Input
              id="prompt-vars"
              value={state.variablesCsv}
              onChange={(e) => update('variablesCsv', e.target.value)}
              disabled={isSubmitting}
              placeholder="name, email"
            />
            <p className="text-xs text-muted-foreground">
              Comma-separated. Leave empty to use detected variables from
              the template.
            </p>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Commit</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            <Label htmlFor="prompt-commit">Commit message</Label>
            <Input
              id="prompt-commit"
              value={state.commitMessage}
              onChange={(e) => update('commitMessage', e.target.value)}
              disabled={isSubmitting}
              placeholder={
                mode === 'edit'
                  ? 'Describe the changes in this version'
                  : 'Initial version'
              }
            />
          </div>
        </CardContent>
      </Card>

      {error ? (
        <div className="rounded-md border border-destructive/40 bg-destructive/5 p-3 text-sm text-destructive">
          {error}
        </div>
      ) : null}

      <div className="flex items-center justify-end gap-2">
        {onCancel ? (
          <Button
            type="button"
            variant="outline"
            onClick={onCancel}
            disabled={isSubmitting}
          >
            Cancel
          </Button>
        ) : null}
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting
            ? 'Saving…'
            : mode === 'edit'
              ? 'Save new version'
              : 'Create prompt'}
        </Button>
      </div>
    </form>
  )
}
