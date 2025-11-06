type ContentSectionProps = {
  children: React.ReactNode
  title: string
  description: string
}

export function ContentSection({ children, title, description }: ContentSectionProps) {
  return (
    <div className='space-y-6'>
      <div>
        <h3 className='text-lg font-medium'>{title}</h3>
        <p className='text-sm text-muted-foreground'>
          {description}
        </p>
      </div>
      {children}
    </div>
  )
}