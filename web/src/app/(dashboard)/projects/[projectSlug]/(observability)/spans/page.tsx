'use client'

import { useParams } from 'next/navigation'
import { Spans } from '@/features/spans'
import type { ProjectParams } from '@/features/organizations'

export default function ProjectSpansPage() {
  const params = useParams() as ProjectParams

  return <Spans projectSlug={params.projectSlug} />
}
