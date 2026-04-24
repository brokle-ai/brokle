import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/')({
  component: Landing,
})

function Landing() {
  return (
    <main className="flex min-h-screen items-center justify-center">
      <div className="text-center">
        <h1 className="text-2xl font-semibold">Brokle Dashboard</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          web-vite scaffold · Phase 1.1 complete
        </p>
      </div>
    </main>
  )
}
