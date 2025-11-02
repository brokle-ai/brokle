import { AuthenticatedLayout } from "@/components/layout/authenticated-layout"

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  // No onboarding redirect needed - users complete onboarding during signup
  return (
    <AuthenticatedLayout>
      {children}
    </AuthenticatedLayout>
  )
}
