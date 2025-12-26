'use client'

import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { formatDistanceToNow } from 'date-fns'
import { FlaskConical, MoreVertical, Trash2, GitCompare } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { ExperimentStatusBadge } from './experiment-status-badge'
import { useExperiments } from '../context/experiments-context'
import type { Experiment } from '../types'

interface ExperimentCardProps {
  experiment: Experiment
}

export function ExperimentCard({ experiment }: ExperimentCardProps) {
  const router = useRouter()
  const { setOpen, setCurrentRow, projectSlug } = useExperiments()

  const handleDelete = () => {
    setCurrentRow(experiment)
    setOpen('delete')
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-lg font-medium">
          <Link
            href={`/projects/${projectSlug}/experiments/${experiment.id}`}
            className="hover:underline"
          >
            {experiment.name}
          </Link>
        </CardTitle>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon">
              <MoreVertical className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem
              onClick={() =>
                router.push(
                  `/projects/${projectSlug}/experiments/compare?ids=${experiment.id}`
                )
              }
            >
              <GitCompare className="mr-2 h-4 w-4" />
              Compare
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              className="text-destructive focus:text-destructive"
              onClick={handleDelete}
            >
              <Trash2 className="mr-2 h-4 w-4" />
              Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-muted-foreground mb-4 line-clamp-2">
          {experiment.description || 'No description'}
        </p>
        <div className="flex items-center justify-between text-sm">
          <div className="flex items-center gap-2">
            <FlaskConical className="h-4 w-4 text-muted-foreground" />
            <ExperimentStatusBadge status={experiment.status} />
          </div>
          <span className="text-muted-foreground">
            {formatDistanceToNow(new Date(experiment.created_at), {
              addSuffix: true,
            })}
          </span>
        </div>
      </CardContent>
    </Card>
  )
}
