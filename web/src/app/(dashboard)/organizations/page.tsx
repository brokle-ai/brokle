import { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Organizations',
  description: 'Manage your organizations',
}

export default function OrganizationsPage() {
  return (
    <div className="container mx-auto py-8">
      <div className="mb-8">
        <h1 className="text-2xl font-bold">Organizations</h1>
        <p className="text-muted-foreground">
          Manage your organizations and their settings.
        </p>
      </div>

      <div className="space-y-4">
        <div className="rounded-lg border p-6">
          <h3 className="text-lg font-semibold mb-2">Your Organizations</h3>
          <p className="text-muted-foreground">
            Organization list will be loaded here.
          </p>
        </div>
      </div>
    </div>
  )
}