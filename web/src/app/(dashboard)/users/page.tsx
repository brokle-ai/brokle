import { Metadata } from 'next'
import { Suspense } from 'react'
import UsersView from '@/views/users-view'

export const metadata: Metadata = {
  title: 'Users | Brokle Dashboard',
  description: 'Manage your users and their roles. View user information, invite new users, and control access permissions.',
  keywords: ['users', 'management', 'roles', 'permissions', 'team'],
}

function UsersLoading() {
  return (
    <div className="flex items-center justify-center min-h-[400px]">
      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
    </div>
  )
}

export default function UsersPage() {
  return (
    <Suspense fallback={<UsersLoading />}>
      <UsersView />
    </Suspense>
  )
}