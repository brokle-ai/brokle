import { Link } from '@tanstack/react-router'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { ReadonlyCode } from '@/editors/readonly-code'
import type { Prompt, PromptVersion } from '../api/types'
import { isTextTemplate, templateToBody } from '../api/types'

interface PromptDetailProps {
  orgId: string
  projectId: string
  prompt: Prompt
  versions: PromptVersion[]
}

function formatTimestamp(iso: string | undefined): string {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

export function PromptDetail({
  orgId,
  projectId,
  prompt,
  versions,
}: PromptDetailProps) {
  const body = templateToBody(prompt.template)
  const isText = isTextTemplate(prompt.template)

  return (
    <div className="space-y-6">
      <header className="flex flex-wrap items-start justify-between gap-4">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-semibold">{prompt.name}</h1>
            <Badge variant="secondary" className="uppercase">
              {prompt.type}
            </Badge>
            <Badge variant="outline">v{prompt.version}</Badge>
          </div>
          {prompt.description ? (
            <p className="text-sm text-muted-foreground">{prompt.description}</p>
          ) : null}
          <div className="flex flex-wrap gap-3 text-xs text-muted-foreground">
            <span>Updated {formatTimestamp(prompt.created_at)}</span>
            {prompt.created_by ? (
              <span>by {prompt.created_by}</span>
            ) : null}
          </div>
        </div>
        <Button asChild variant="default" size="sm">
          <Link
            to="/o/$orgId/p/$projectId/prompts/$promptId/edit"
            params={{ orgId, projectId, promptId: prompt.id }}
          >
            Edit
          </Link>
        </Button>
      </header>

      {prompt.tags.length > 0 ? (
        <div className="flex flex-wrap gap-1">
          {prompt.tags.map((t) => (
            <Badge key={t} variant="outline" className="font-normal">
              {t}
            </Badge>
          ))}
        </div>
      ) : null}

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Template</CardTitle>
        </CardHeader>
        <CardContent>
          {isText ? (
            <ReadonlyCode code={body} language="jinja" />
          ) : (
            <div className="space-y-2">
              <p className="text-xs text-muted-foreground">
                Chat template — raw JSON view (editor coming soon).
              </p>
              <ReadonlyCode code={body} language="json" />
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Variables</CardTitle>
        </CardHeader>
        <CardContent>
          {prompt.variables.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              No variables declared.
            </p>
          ) : (
            <div className="flex flex-wrap gap-1">
              {prompt.variables.map((v) => (
                <Badge key={v} variant="secondary" className="font-mono">
                  {v}
                </Badge>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">
            Versions ({versions.length})
          </CardTitle>
        </CardHeader>
        <CardContent>
          {versions.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              No version history yet.
            </p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Version</TableHead>
                  <TableHead>Commit message</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead>By</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {versions.map((v) => (
                  <TableRow key={v.id}>
                    <TableCell>
                      <Badge variant="outline">v{v.version}</Badge>
                    </TableCell>
                    <TableCell>
                      {v.commit_message ? (
                        v.commit_message
                      ) : (
                        <span className="text-muted-foreground">—</span>
                      )}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {formatTimestamp(v.created_at)}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {v.created_by ?? '—'}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
