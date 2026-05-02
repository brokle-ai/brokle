import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import type { Invoice } from '../api/types'

interface InvoicesTableProps {
  rows: Invoice[]
}

function formatDate(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

function formatMoney(decimalStr: string, currency: string): string {
  const n = Number(decimalStr)
  if (!Number.isFinite(n)) return `${decimalStr} ${currency}`
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency,
  }).format(n)
}

function StatusBadge({ status }: { status: Invoice['status'] }) {
  switch (status) {
    case 'paid':
      return <Badge variant="secondary">Paid</Badge>
    case 'open':
      return <Badge variant="outline">Open</Badge>
    case 'void':
      return <Badge variant="outline">Void</Badge>
    case 'uncollectible':
      return <Badge variant="destructive">Uncollectible</Badge>
  }
}

export function InvoicesTable({ rows }: InvoicesTableProps) {
  if (rows.length === 0) {
    return (
      <div className="rounded-lg border p-12 text-center">
        <p className="text-sm text-muted-foreground">No invoices yet</p>
        <p className="mt-1 text-xs text-muted-foreground">
          Invoices will appear here once your first billing period closes.
        </p>
      </div>
    )
  }

  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Number</TableHead>
            <TableHead>Issued</TableHead>
            <TableHead>Period</TableHead>
            <TableHead>Status</TableHead>
            <TableHead className="text-right">Amount</TableHead>
            <TableHead className="w-[100px] text-right">Download</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((inv) => (
            <TableRow key={inv.id}>
              <TableCell className="font-mono text-sm">{inv.number}</TableCell>
              <TableCell className="text-muted-foreground">
                {formatDate(inv.issued_at)}
              </TableCell>
              <TableCell className="text-muted-foreground">
                {formatDate(inv.period_start)} – {formatDate(inv.period_end)}
              </TableCell>
              <TableCell>
                <StatusBadge status={inv.status} />
              </TableCell>
              <TableCell className="text-right font-medium">
                {formatMoney(inv.amount, inv.currency)}
              </TableCell>
              <TableCell className="text-right">
                {inv.download_url ? (
                  <a
                    href={inv.download_url}
                    className="text-sm text-primary hover:underline"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    PDF
                  </a>
                ) : (
                  <span className="text-xs text-muted-foreground">—</span>
                )}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
