import { redirect } from 'next/navigation'

interface NotificationsPageProps {
  params: Promise<{ projectSlug: string }>
}

export default async function NotificationsPage({ params }: NotificationsPageProps) {
  const { projectSlug } = await params
  redirect(`/projects/${projectSlug}/settings/profile`)
}
