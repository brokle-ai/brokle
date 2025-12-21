import { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Projects',
  description: 'Manage your projects',
}

export default function ProjectsPage() {
  return (
    <div className="container mx-auto py-8">
      <div className="mb-8">
        <h1 className="text-xl font-bold">Projects</h1>
        <p className="text-muted-foreground">
          Manage your projects and their configurations.
        </p>
      </div>

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