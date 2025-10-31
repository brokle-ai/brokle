import { redirect } from 'next/navigation'
import { cookies } from 'next/headers'
import { AuthenticatedLayout } from "@/components/layout/authenticated-layout"

async function checkOnboardingStatus() {
  try {
    const cookieStore = await cookies()
    const token = cookieStore.get('access_token')?.value

    if (!token) {
      return true
    }

    const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
    const response = await fetch(`${apiUrl}/api/v1/users/me`, {
      headers: {
        'Authorization': `Bearer ${token}`,
      },
      cache: 'no-store',
      next: { revalidate: 0 },
    })

    if (!response.ok) {
      return true
    }

    const data = await response.json()
    return data.data.onboarding_completed_at != null
  } catch (error) {
    console.error('Error checking onboarding:', error)
    return true
  }
}

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const onboardingCompleted = await checkOnboardingStatus()

  if (!onboardingCompleted) {
    redirect('/onboarding')
  }

  return (
    <AuthenticatedLayout>
      {children}
    </AuthenticatedLayout>
  )
}
