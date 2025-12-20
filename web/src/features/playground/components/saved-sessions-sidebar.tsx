'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import {
  Bookmark,
  Clock,
  Loader2,
  MoreHorizontal,
  Plus,
  Search,
  Trash2,
} from 'lucide-react'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Skeleton } from '@/components/ui/skeleton'
import {
  useSessionsQuery,
  useDeleteSessionMutation,
} from '../hooks/use-playground-queries'
import type { PlaygroundSessionSummary } from '../types'

interface SavedSessionsSidebarProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  projectId: string
  projectSlug: string
  currentSessionId?: string
  onNewSession: () => void
}

export function SavedSessionsSidebar({
  open,
  onOpenChange,
  projectId,
  projectSlug,
  currentSessionId,
  onNewSession,
}: SavedSessionsSidebarProps) {
  const router = useRouter()
  const [searchQuery, setSearchQuery] = useState('')
  const [deleteSession, setDeleteSession] = useState<PlaygroundSessionSummary | null>(null)

  const { data: sessions, isLoading } = useSessionsQuery(projectId, undefined, {
    enabled: open,
  })

  const deleteSessionMutation = useDeleteSessionMutation(projectId)

  const filteredSessions = sessions?.filter((session) => {
    if (!searchQuery) return true
    const query = searchQuery.toLowerCase()
    return (
      session.name?.toLowerCase().includes(query) ||
      session.description?.toLowerCase().includes(query) ||
      session.tags?.some((tag) => tag.toLowerCase().includes(query))
    )
  })

  const handleSessionClick = (sessionId: string) => {
    router.push(`/projects/${projectSlug}/playground/${sessionId}`)
    onOpenChange(false)
  }

  const handleDeleteConfirm = async () => {
    if (!deleteSession) return

    try {
      await deleteSessionMutation.mutateAsync({
        sessionId: deleteSession.id,
        sessionName: deleteSession.name,
      })

      // If we deleted the currently-viewed session, navigate to playground home
      if (deleteSession.id === currentSessionId) {
        onOpenChange(false)
        router.push(`/projects/${projectSlug}/playground`)
      }

      setDeleteSession(null)
    } catch (error) {
      // Error handled by mutation
    }
  }

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffMins = Math.floor(diffMs / (1000 * 60))
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60))
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))

    if (diffMins < 60) return `${diffMins}m ago`
    if (diffHours < 24) return `${diffHours}h ago`
    if (diffDays < 7) return `${diffDays}d ago`
    return date.toLocaleDateString()
  }

  return (
    <>
      <Sheet open={open} onOpenChange={onOpenChange}>
        <SheetContent side="right" className="sm:max-w-[540px] p-0">
          <SheetHeader className="p-4 pb-2">
            <SheetTitle className="flex items-center gap-2">
              <Bookmark className="h-5 w-5" />
              Saved Sessions
            </SheetTitle>
            <SheetDescription>
              Your saved playground sessions
            </SheetDescription>
          </SheetHeader>

          <div className="px-4 pb-2">
            <div className="flex gap-2">
              <div className="relative flex-1">
                <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search sessions..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-8"
                />
              </div>
              <Button size="sm" variant="outline" onClick={onNewSession}>
                <Plus className="h-4 w-4" />
              </Button>
            </div>
          </div>

          <ScrollArea className="flex-1 h-[calc(100vh-180px)]">
            <div className="px-4 space-y-2">
              {isLoading ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <SessionSkeleton key={i} />
                ))
              ) : filteredSessions?.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">
                  {searchQuery ? (
                    <p>No sessions match your search</p>
                  ) : (
                    <>
                      <Bookmark className="h-8 w-8 mx-auto mb-2 opacity-50" />
                      <p>No saved sessions yet</p>
                      <p className="text-xs mt-1">
                        Save a session to see it here
                      </p>
                    </>
                  )}
                </div>
              ) : (
                filteredSessions?.map((session) => (
                  <SessionItem
                    key={session.id}
                    session={session}
                    isActive={session.id === currentSessionId}
                    onClick={() => handleSessionClick(session.id)}
                    onDelete={() => setDeleteSession(session)}
                    formatDate={formatDate}
                  />
                ))
              )}
            </div>
          </ScrollArea>
        </SheetContent>
      </Sheet>

      <AlertDialog open={!!deleteSession} onOpenChange={() => setDeleteSession(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Session</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete &quot;{deleteSession?.name}&quot;? This
              action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDeleteConfirm}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
              disabled={deleteSessionMutation.isPending}
            >
              {deleteSessionMutation.isPending ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                'Delete'
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}

interface SessionItemProps {
  session: PlaygroundSessionSummary
  isActive: boolean
  onClick: () => void
  onDelete: () => void
  formatDate: (date: string) => string
}

function SessionItem({
  session,
  isActive,
  onClick,
  onDelete,
  formatDate,
}: SessionItemProps) {
  return (
    <div
      className={`group relative p-3 rounded-lg border cursor-pointer transition-colors ${
        isActive
          ? 'bg-accent border-accent-foreground/20'
          : 'hover:bg-muted/50 border-transparent hover:border-border'
      }`}
      onClick={onClick}
    >
      <div className="flex items-start justify-between gap-2">
        <div className="flex-1 min-w-0">
          <p className="font-medium truncate">{session.name || 'Untitled'}</p>
          {session.description && (
            <p className="text-xs text-muted-foreground truncate mt-0.5">
              {session.description}
            </p>
          )}
        </div>

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              variant="ghost"
              size="sm"
              className="h-6 w-6 p-0 opacity-0 group-hover:opacity-100"
              onClick={(e) => e.stopPropagation()}
            >
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem
              className="text-destructive"
              onClick={(e) => {
                e.stopPropagation()
                onDelete()
              }}
            >
              <Trash2 className="mr-2 h-4 w-4" />
              Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      <div className="flex items-center gap-1 mt-2 text-xs text-muted-foreground">
        <Clock className="h-3 w-3" />
        {formatDate(session.last_used_at)}
      </div>

      {session.tags && session.tags.length > 0 && (
        <div className="flex flex-wrap gap-1 mt-2">
          {session.tags.slice(0, 3).map((tag) => (
            <Badge key={tag} variant="secondary" className="text-[10px] h-4">
              {tag}
            </Badge>
          ))}
          {session.tags.length > 3 && (
            <Badge variant="secondary" className="text-[10px] h-4">
              +{session.tags.length - 3}
            </Badge>
          )}
        </div>
      )}
    </div>
  )
}

function SessionSkeleton() {
  return (
    <div className="p-3 rounded-lg border">
      <Skeleton className="h-4 w-3/4" />
      <Skeleton className="h-3 w-1/2 mt-2" />
      <div className="flex gap-2 mt-2">
        <Skeleton className="h-4 w-16" />
        <Skeleton className="h-4 w-12" />
      </div>
    </div>
  )
}
