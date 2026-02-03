import { redirect } from 'next/navigation'

interface DisplayPageProps {
  params: Promise<{ projectSlug: string }>
}

export default async function DisplayPage({ params }: DisplayPageProps) {
  const { projectSlug } = await params
  redirect(`/projects/${projectSlug}/settings/appearance`)
}
