import { Sidebar, SidebarContent, SidebarHeader } from '@/components/ui/sidebar'
import { Skeleton } from '@/components/ui/skeleton'

export function SidebarSkeleton() {
  return (
    <Sidebar collapsible="icon" variant="sidebar">
      <SidebarHeader>
        <Skeleton className="h-8 w-32 m-2" />
      </SidebarHeader>
      <SidebarContent>
        <div className="space-y-2 p-2">
          {[...Array(8)].map((_, i) => (
            <Skeleton key={i} className="h-10 w-full" />
          ))}
        </div>
      </SidebarContent>
    </Sidebar>
  )
}
