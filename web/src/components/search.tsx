import { IconSearch } from '@tabler/icons-react'
import { cn } from '@/lib/utils'
import { useSearch } from '@/context/search-context'
import { Button } from './ui/button'

interface Props {
  className?: string
}

export function Search({ className = '' }: Props) {
  const { setOpen } = useSearch()

  return (
    <Button
      variant='ghost'
      size='icon'
      className={cn('scale-95', className)}
      onClick={() => setOpen(true)}
    >
      <IconSearch className='size-[1.2rem]' />
      <span className='sr-only'>Search</span>
    </Button>
  )
}
