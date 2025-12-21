import { Separator } from '@/components/ui/separator'

type ContentSectionProps = {
  children: React.ReactNode
  title: string
  description: string
  action?: React.ReactNode
}

export function ContentSection({ children, title, description, action }: ContentSectionProps) {
  return (
    <div className='flex flex-1 flex-col'>
      <div className='flex-none'>
        <div className='flex items-center justify-between'>
          <h3 className='text-lg font-medium'>{title}</h3>
          {action}
        </div>
        <p className='text-muted-foreground text-sm'>{description}</p>
      </div>
      <Separator className='my-4 flex-none' />
      <div className='relative h-full w-full overflow-y-auto scroll-smooth pe-4 pb-12'>
        <div className='-mx-1 px-1.5'>{children}</div>
      </div>
    </div>
  )
}