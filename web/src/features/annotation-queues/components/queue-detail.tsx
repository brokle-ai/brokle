'use client'

import { useState } from 'react'
import Link from 'next/link'
import { formatDistanceToNow } from 'date-fns'
import {
  ArrowLeft,
  Loader2,
  Play,
  Pause,
  Archive,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Skeleton } from '@/components/ui/skeleton'
import {
  useAnnotationQueueQuery,
  useQueueStatsQuery,
  useQueueItemsQuery,
} from '../hooks/use-annotation-queues'
import { useProjectOnly } from '@/features/projects'
import { QueueItemTable } from './queue-item-table'
import { AnnotationPanel } from './annotation-panel'
import { AssignmentDialog } from './assignment-dialog'
import { SettingsDialog } from './settings-dialog'
import { AddItemsDialogStandalone } from './add-items-dialog-standalone'
import type { QueueItem, QueueStatus } from '../types'

interface QueueDetailProps {
  projectSlug: string
  queueId: string
}

function getStatusIcon(status: QueueStatus) {
  switch (status) {
    case 'active':
      return Play
    case 'paused':
      return Pause
    case 'archived':
      return Archive
  }
}

export function QueueDetail({ projectSlug, queueId }: QueueDetailProps) {
  const [currentItem, setCurrentItem] = useState<QueueItem | null>(null)
  const [activeTab, setActiveTab] = useState<string>('annotate')
  const { currentProject } = useProjectOnly()
  const projectId = currentProject?.id

  const { data: queue, isLoading: queueLoading } = useAnnotationQueueQuery(projectId, queueId)
  const { data: stats, isLoading: statsLoading } = useQueueStatsQuery(projectId, queueId)
  const { data: itemsData, isLoading: itemsLoading } = useQueueItemsQuery(projectId, queueId)

  if (queueLoading) {
    return <QueueDetailSkeleton />
  }

  if (!projectId) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-muted-foreground">No project selected</p>
      </div>
    )
  }

  if (!queue) {
    return (
      <div className="flex flex-col items-center justify-center py-12">
        <p className="text-muted-foreground">Queue not found</p>
        <Button asChild variant="link" className="mt-2">
          <Link href={`/projects/${projectSlug}/annotation-queues`}>
            Back to queues
          </Link>
        </Button>
      </div>
    )
  }

  const completionPercentage = stats
    ? stats.total_items > 0
      ? Math.round(
          ((stats.completed_items + stats.skipped_items) / stats.total_items) * 100
        )
      : 0
    : 0

  const StatusIcon = getStatusIcon(queue.status)

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Button asChild variant="ghost" size="icon">
          <Link href={`/projects/${projectSlug}/annotation-queues`}>
            <ArrowLeft className="h-4 w-4" />
          </Link>
        </Button>
        <div className="flex-1">
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-bold">{queue.name}</h1>
            <Badge variant="outline" className="capitalize">
              <StatusIcon className="mr-1 h-3 w-3" />
              {queue.status}
            </Badge>
          </div>
          {queue.description && (
            <p className="text-muted-foreground mt-1">{queue.description}</p>
          )}
        </div>
        <div className="flex items-center gap-2">
          <AssignmentDialog
            projectId={projectId}
            queueId={queueId}
            queueName={queue.name}
          />
          <SettingsDialog
            projectId={projectId}
            queue={queue}
          />
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Total Items</CardDescription>
            <CardTitle className="text-2xl">
              {statsLoading ? <Skeleton className="h-8 w-16" /> : stats?.total_items ?? 0}
            </CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Pending</CardDescription>
            <CardTitle className="text-2xl text-yellow-600">
              {statsLoading ? <Skeleton className="h-8 w-16" /> : stats?.pending_items ?? 0}
            </CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Completed</CardDescription>
            <CardTitle className="text-2xl text-green-600">
              {statsLoading ? <Skeleton className="h-8 w-16" /> : stats?.completed_items ?? 0}
            </CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Progress</CardDescription>
            <CardTitle className="text-2xl">{completionPercentage}%</CardTitle>
          </CardHeader>
          <CardContent className="pt-0">
            <Progress value={completionPercentage} className="h-2" />
          </CardContent>
        </Card>
      </div>

      {/* Main Content Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-4">
        <TabsList>
          <TabsTrigger value="annotate">Annotate</TabsTrigger>
          <TabsTrigger value="items">Items ({stats?.total_items ?? 0})</TabsTrigger>
        </TabsList>

        <TabsContent value="annotate" className="space-y-4">
          <div className="grid gap-6 lg:grid-cols-2">
            {/* Left: Annotation Panel */}
            <AnnotationPanel
              projectId={projectId}
              queue={queue}
              currentItem={currentItem}
              onItemClaimed={setCurrentItem}
              onItemCompleted={() => setCurrentItem(null)}
              onItemSkipped={() => setCurrentItem(null)}
            />

            {/* Right: Instructions / Help */}
            <Card>
              <CardHeader>
                <CardTitle>Quick Stats</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-2 gap-4 text-center">
                  <div className="rounded-lg border p-4">
                    <div className="text-3xl font-bold text-green-600">
                      {stats?.completed_items ?? 0}
                    </div>
                    <div className="text-sm text-muted-foreground">Completed</div>
                  </div>
                  <div className="rounded-lg border p-4">
                    <div className="text-3xl font-bold text-yellow-600">
                      {stats?.pending_items ?? 0}
                    </div>
                    <div className="text-sm text-muted-foreground">Remaining</div>
                  </div>
                </div>
                <p className="text-sm text-muted-foreground">
                  Created {formatDistanceToNow(new Date(queue.created_at), { addSuffix: true })}
                </p>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="items" className="space-y-4">
          <div className="flex items-center justify-between">
            <p className="text-sm text-muted-foreground">
              All items in this annotation queue
            </p>
            <AddItemsDialogStandalone
              projectId={projectId}
              queueId={queueId}
              queueName={queue.name}
            />
          </div>
          {itemsLoading ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : (
            <QueueItemTable items={itemsData?.data ?? []} />
          )}
        </TabsContent>
      </Tabs>
    </div>
  )
}

function QueueDetailSkeleton() {
  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Skeleton className="h-10 w-10" />
        <div className="flex-1 space-y-2">
          <Skeleton className="h-8 w-64" />
          <Skeleton className="h-4 w-96" />
        </div>
      </div>
      <div className="grid gap-4 md:grid-cols-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <Card key={i}>
            <CardHeader className="pb-2">
              <Skeleton className="h-4 w-20" />
              <Skeleton className="h-8 w-16 mt-2" />
            </CardHeader>
          </Card>
        ))}
      </div>
      <Skeleton className="h-64 w-full" />
    </div>
  )
}
