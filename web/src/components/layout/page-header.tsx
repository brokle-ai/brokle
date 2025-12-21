import { Separator } from '@/components/ui/separator'

interface PageHeaderProps {
  title: string
  children?: React.ReactNode
}

export function PageHeader({ title, children }: PageHeaderProps) {
  return (
    <>
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold">{title}</h1>
        {children && <div className="flex items-center gap-2">{children}</div>}
      </div>
      <Separator className="mt-2 mb-4" />
    </>
  )
}
