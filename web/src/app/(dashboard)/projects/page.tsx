import { Metadata } from 'next'
import { PageHeader } from '@/components/layout/page-header'

export const metadata: Metadata = {
  title: 'Projects',
  description: 'Manage your projects',
}

export default function ProjectsPage() {
  return (
    <div className="container mx-auto py-8">
      <PageHeader title="Projects" />

      <div className="space-y-4">
        <div className="rounded-lg border p-6">
          <h3 className="text-lg font-semibold mb-2">Your Projects</h3>
          <p className="text-muted-foreground">
            Project list will be loaded here.
          </p>
        </div>
      </div>
    </div>
  )
}