'use client'

import { useParams } from 'next/navigation'
import { Sessions } from '@/features/sessions'
import type { ProjectParams } from '@/features/organizations'

export default function ProjectSessionsPage() {
  const params = useParams() as ProjectParams

  return <Sessions projectSlug={params.projectSlug} />
}
