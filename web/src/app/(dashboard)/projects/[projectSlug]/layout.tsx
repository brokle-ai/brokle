interface ProjectLayoutProps {
  children: React.ReactNode
  params: Promise<{ projectSlug: string }>
}

export default function ProjectLayout({
  children,
}: ProjectLayoutProps) {
  // Sidebar is now handled by parent (dashboard) layout
  return <>{children}</>
}
