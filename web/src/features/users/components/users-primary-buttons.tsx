'use client'

import { IconMailPlus, IconUserPlus } from '@tabler/icons-react'
import { Button } from '@/components/ui/button'
import { useUsers } from '../context/users-context'

export function UsersPrimaryButtons() {
  const { setOpen } = useUsers()
  return (
    <div className='flex space-x-2'>
      <Button
        variant='outline'
        className='space-x-2'
        onClick={() => setOpen('invite')}
        aria-label="Invite a new user via email"
      >
        <IconMailPlus size={16} />
        <span>Invite User</span>
      </Button>
      <Button 
        className='space-x-2' 
        onClick={() => setOpen('add')}
        aria-label="Add a new user directly"
      >
        <IconUserPlus size={16} />
        <span>Add User</span>
      </Button>
    </div>
  )
}