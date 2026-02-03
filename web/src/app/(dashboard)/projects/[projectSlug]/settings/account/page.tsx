import { redirect } from 'next/navigation'

interface AccountPageProps {
  params: Promise<{ projectSlug: string }>
}

export default async function AccountPage({ params }: AccountPageProps) {
  const { projectSlug } = await params
  redirect(`/projects/${projectSlug}/settings/profile`)
}
