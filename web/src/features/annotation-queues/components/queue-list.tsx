'use client'

import { QueueCard } from './queue-card'
import type { QueueWithStats } from '../types'

interface QueueListProps {
  data: QueueWithStats[]
}

export function QueueList({ data }: QueueListProps) {
  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      {data.map((item) => (
        <QueueCard key={item.queue.id} data={item} />
      ))}
    </div>
  )
}
