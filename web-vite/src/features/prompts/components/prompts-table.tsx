import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import type { PromptListItem } from '../api/types'

interface PromptsTableProps {
  rows: PromptListItem[]
}

function formatTimestamp(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

function TagList({ tags }: { tags: string[] }) {
  if (tags.length === 0) return <span className="text-muted-foreground">—</span>
  const shown = tags.slice(0, 3)
  const extra = tags.length - shown.length
  return (
    <div className="flex flex-wrap gap-1">
      {shown.map((t) => (
        <Badge key={t} variant="outline" className="font-normal">
          {t}
        </Badge>
      ))}
      {extra > 0 ? (
        <Badge variant="outline" className="font-normal">
          +{extra}
        </Badge>
      ) : null}
    </div>
  )
}

export function PromptsTable({ rows }: PromptsTableProps) {
  if (rows.length === 0) {
    return (
      <div className="rounded-lg border p-12 text-center">
        <p className="text-sm text-muted-foreground">No prompts yet</p>
        <p className="mt-1 text-xs text-muted-foreground">
          Create a prompt to manage versioned templates for this project.
        </p>
      </div>
    )
  }

  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Type</TableHead>
            <TableHead>Version</TableHead>
            <TableHead>Tags</TableHead>
            <TableHead>Updated</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((prompt) => (
            <TableRow key={prompt.id}>
              <TableCell className="font-medium">
                <div>{prompt.name}</div>
                {prompt.description ? (
                  <div className="text-xs text-muted-foreground">
                    {prompt.description}
                  </div>
                ) : null}
              </TableCell>
              <TableCell>
                <Badge variant="secondary" className="uppercase">
                  {prompt.type}
                </Badge>
              </TableCell>
              <TableCell className="text-muted-foreground">
                v{prompt.latest_version}
              </TableCell>
              <TableCell>
                <TagList tags={prompt.tags} />
              </TableCell>
              <TableCell className="text-muted-foreground">
                {formatTimestamp(prompt.updated_at)}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
