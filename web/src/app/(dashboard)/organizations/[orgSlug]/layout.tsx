interface OrganizationLayoutProps {
  children: React.ReactNode
  params: Promise<{ orgSlug: string }>
}

export default function OrganizationLayout({
  children,
}: OrganizationLayoutProps) {
  // Sidebar is now handled by parent (dashboard) layout
  return <>{children}</>
}
