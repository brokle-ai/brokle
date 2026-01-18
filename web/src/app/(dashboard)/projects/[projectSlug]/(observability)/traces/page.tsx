'use client'

import { useParams } from 'next/navigation'
import { Traces } from '@/features/traces'
import type { ProjectParams } from '@/features/organizations'

export default function ProjectTracesPage() {
  const params = useParams() as ProjectParams

  return <Traces projectSlug={params.projectSlug} />
}
