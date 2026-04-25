import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { AlertTriangle, Bell } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import {
  acknowledgeAlert,
  alertsQueryOptions,
  billingKeys,
} from '../api/queries'
import type { AlertDimension, AlertSeverity, UsageAlert } from '../api/types'

interface AlertsListProps {
  orgId: string
  /** Cap on the alerts to fetch. Default 25 — matches the budget page footprint. */
  limit?: number
  /** When true, only show alerts in `triggered` state (the actionable subset). */
  triggeredOnly?: boolean
}

const SEVERITY_TONE: Record<AlertSeverity, string> = {
  info: 'border-blue-200 bg-blue-50 dark:border-blue-900 dark:bg-blue-950/20',
  warning:
    'border-yellow-200 bg-yellow-50 dark:border-yellow-900 dark:bg-yellow-950/20',
  critical:
    'border-red-200 bg-red-50 dark:border-red-900 dark:bg-red-950/20',
}

const SEVERITY_BADGE: Record<AlertSeverity, string> = {
  info: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300',
  warning:
    'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300',
  critical: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300',
}

const DIMENSION_LABEL: Record<AlertDimension, string> = {
  spans: 'Spans',
  bytes: 'Data',
  scores: 'Scores',
  cost: 'Cost',
}

export function AlertsList({
  orgId,
  limit = 25,
  triggeredOnly = false,
}: AlertsListProps) {
  const queryClient = useQueryClient()
  const { data, isLoading, isError, error } = useQuery(
    alertsQueryOptions(orgId, limit),
  )

  const ackMut = useMutation({
    mutationFn: async (alertId: string) => acknowledgeAlert(orgId, alertId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: billingKeys.alerts() })
      toast.success('Alert acknowledged')
    },
    onError: (err) => {
      toast.error(
        err instanceof Error ? err.message : 'Failed to acknowledge alert',
      )
    },
  })

  if (isLoading) {
    return (
      <div className="rounded-lg border p-4">
        <p className="text-sm text-muted-foreground">Loading alerts…</p>
      </div>
    )
  }
  if (isError) {
    return (
      <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-4">
        <p className="text-sm font-medium text-destructive">
          Unable to load alerts
        </p>
        <p className="mt-1 text-xs text-muted-foreground">
          {error instanceof Error ? error.message : 'Unknown error'}
        </p>
      </div>
    )
  }

  const alerts = (data ?? []).filter((a) =>
    triggeredOnly ? a.status === 'triggered' : true,
  )

  if (alerts.length === 0) {
    return (
      <div className="rounded-lg border border-dashed p-6 text-center">
        <Bell className="mx-auto h-6 w-6 text-muted-foreground" />
        <p className="mt-2 text-sm text-muted-foreground">
          {triggeredOnly
            ? 'No active alerts.'
            : 'No alerts yet — set thresholds on a budget to get notified.'}
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-2">
      {alerts.map((alert) => (
        <AlertRow
          key={alert.id}
          alert={alert}
          onAcknowledge={() => ackMut.mutate(alert.id)}
          ackPending={ackMut.isPending && ackMut.variables === alert.id}
        />
      ))}
    </div>
  )
}

function AlertRow({
  alert,
  onAcknowledge,
  ackPending,
}: {
  alert: UsageAlert
  onAcknowledge: () => void
  ackPending: boolean
}) {
  const pct = Number(alert.percent_used)
  const pctLabel = Number.isFinite(pct) ? `${pct.toFixed(0)}%` : alert.percent_used
  return (
    <div
      className={cn(
        'flex items-center justify-between gap-3 rounded-lg border p-3',
        SEVERITY_TONE[alert.severity],
      )}
    >
      <div className="flex items-center gap-3">
        <AlertTriangle className="h-4 w-4 text-muted-foreground" />
        <div>
          <p className="text-sm font-medium">
            {DIMENSION_LABEL[alert.dimension]} at {pctLabel}
          </p>
          <p className="text-xs text-muted-foreground">
            {new Date(alert.triggered_at).toLocaleString()}
          </p>
        </div>
      </div>
      <div className="flex items-center gap-2">
        <Badge className={SEVERITY_BADGE[alert.severity]}>
          {alert.severity}
        </Badge>
        {alert.status === 'triggered' && (
          <Button
            variant="ghost"
            size="sm"
            onClick={onAcknowledge}
            disabled={ackPending}
          >
            {ackPending ? 'Acknowledging…' : 'Acknowledge'}
          </Button>
        )}
        {alert.status === 'acknowledged' && (
          <Badge variant="secondary">Acknowledged</Badge>
        )}
      </div>
    </div>
  )
}
