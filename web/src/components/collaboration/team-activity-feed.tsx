'use client'

import { useState } from 'react'
import { 
  Users, 
  MessageCircle, 
  FileText, 
  GitBranch, 
  Settings, 
  Key, 
  Plus, 
  Edit, 
  Trash2, 
  Clock,
  Filter,
  RefreshCw,
  Bookmark,
  Star,
  Eye
} from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Separator } from '@/components/ui/separator'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'

interface ActivityItem {
  id: string
  type: 'comment' | 'project_created' | 'settings_changed' | 'member_added' | 'api_key_created' | 'document_shared'
  actor: {
    id: string
    name: string
    email: string
    avatar?: string
  }
  timestamp: string
  title: string
  description: string
  resource?: {
    type: string
    name: string
    id: string
  }
  metadata?: Record<string, any>
  isBookmarked?: boolean
  mentions?: string[]
  reactions?: { emoji: string; count: number; users: string[] }[]
}

interface TeamComment {
  id: string
  author: {
    id: string
    name: string
    email: string
    avatar?: string
  }
  content: string
  timestamp: string
  mentions: string[]
  reactions: { emoji: string; count: number; users: string[] }[]
  replies?: TeamComment[]
}

const MOCK_ACTIVITIES: ActivityItem[] = [
  {
    id: 'activity-001',
    type: 'project_created',
    actor: {
      id: 'user-123',
      name: 'John Doe',
      email: 'john@acmecorp.com',
      avatar: '/avatars/john.jpg'
    },
    timestamp: '2024-03-15T14:30:00Z',
    title: 'Created new project',
    description: 'John created "AI Customer Support" using the Chatbot template',
    resource: {
      type: 'project',
      name: 'AI Customer Support',
      id: 'proj-456'
    },
    metadata: {
      template: 'chatbot'
    }
  },
  {
    id: 'activity-002',
    type: 'comment',
    actor: {
      id: 'user-456',
      name: 'Jane Smith',
      email: 'jane@acmecorp.com',
      avatar: '/avatars/jane.jpg'
    },
    timestamp: '2024-03-15T14:25:00Z',
    title: 'Commented on project',
    description: 'Great work on the new chatbot! The response quality looks impressive. @mike should take a look at the analytics.',
    resource: {
      type: 'project',
      name: 'AI Customer Support',
      id: 'proj-456'
    },
    mentions: ['mike'],
    reactions: [
      { emoji: '👍', count: 3, users: ['user-789', 'user-123', 'user-456'] },
      { emoji: '🚀', count: 1, users: ['user-123'] }
    ]
  },
  {
    id: 'activity-003',
    type: 'settings_changed',
    actor: {
      id: 'user-456',
      name: 'Jane Smith',
      email: 'jane@acmecorp.com',
      avatar: '/avatars/jane.jpg'
    },
    timestamp: '2024-03-15T14:20:00Z',
    title: 'Updated project settings',
    description: 'Changed routing strategy from "balanced" to "quality-optimized"',
    resource: {
      type: 'project',
      name: 'Content Generator',
      id: 'proj-789'
    },
    metadata: {
      changes: {
        routing_strategy: { from: 'balanced', to: 'quality-optimized' }
      }
    }
  },
  {
    id: 'activity-004',
    type: 'member_added',
    actor: {
      id: 'user-123',
      name: 'John Doe',
      email: 'john@acmecorp.com',
      avatar: '/avatars/john.jpg'
    },
    timestamp: '2024-03-15T13:45:00Z',
    title: 'Added team member',
    description: 'Added Sarah Wilson as a Developer',
    resource: {
      type: 'member',
      name: 'Sarah Wilson',
      id: 'member-999'
    },
    metadata: {
      role: 'developer'
    }
  },
  {
    id: 'activity-005',
    type: 'document_shared',
    actor: {
      id: 'user-789',
      name: 'Mike Johnson',
      email: 'mike@acmecorp.com',
      avatar: '/avatars/mike.jpg'
    },
    timestamp: '2024-03-15T13:30:00Z',
    title: 'Shared documentation',
    description: 'Shared "API Integration Guide" with the team',
    resource: {
      type: 'document',
      name: 'API Integration Guide',
      id: 'doc-123'
    },
    isBookmarked: true
  }
]

const MOCK_COMMENTS: TeamComment[] = [
  {
    id: 'comment-001',
    author: {
      id: 'user-456',
      name: 'Jane Smith',
      email: 'jane@acmecorp.com',
      avatar: '/avatars/jane.jpg'
    },
    content: 'The new routing configuration is working great! We\'re seeing 23% better response times. @john might want to apply this to other projects too.',
    timestamp: '2024-03-15T15:30:00Z',
    mentions: ['john'],
    reactions: [
      { emoji: '🎉', count: 2, users: ['user-123', 'user-789'] },
      { emoji: '👍', count: 4, users: ['user-123', 'user-789', 'user-999', 'user-456'] }
    ],
    replies: [
      {
        id: 'reply-001',
        author: {
          id: 'user-123',
          name: 'John Doe',
          email: 'john@acmecorp.com',
          avatar: '/avatars/john.jpg'
        },
        content: 'Absolutely! I\'ll roll this out to the other production projects this week.',
        timestamp: '2024-03-15T15:45:00Z',
        mentions: [],
        reactions: [
          { emoji: '✅', count: 1, users: ['user-456'] }
        ]
      }
    ]
  },
  {
    id: 'comment-002',
    author: {
      id: 'user-789',
      name: 'Mike Johnson',
      email: 'mike@acmecorp.com',
      avatar: '/avatars/mike.jpg'
    },
    content: 'I\'ve updated the API documentation with the latest changes. Everyone should review the new authentication flow.',
    timestamp: '2024-03-15T14:15:00Z',
    mentions: [],
    reactions: [
      { emoji: '📚', count: 1, users: ['user-456'] },
      { emoji: '👍', count: 2, users: ['user-123', 'user-456'] }
    ]
  }
]

export function TeamActivityFeed() {
  const [activities, setActivities] = useState<ActivityItem[]>(MOCK_ACTIVITIES)
  const [comments, setComments] = useState<TeamComment[]>(MOCK_COMMENTS)
  const [filterType, setFilterType] = useState<string>('all')
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [newComment, setNewComment] = useState('')
  const [isCommentDialogOpen, setIsCommentDialogOpen] = useState(false)

  const filteredActivities = activities.filter(activity => {
    if (filterType === 'all') return true
    return activity.type === filterType
  })

  const handleRefresh = async () => {
    setIsRefreshing(true)
    // TODO: Implement actual refresh
    await new Promise(resolve => setTimeout(resolve, 1000))
    setIsRefreshing(false)
    toast.success('Activity feed refreshed')
  }

  const toggleBookmark = (activityId: string) => {
    setActivities(activities.map(activity =>
      activity.id === activityId
        ? { ...activity, isBookmarked: !activity.isBookmarked }
        : activity
    ))
  }

  const addReaction = (activityId: string, emoji: string) => {
    setActivities(activities.map(activity => {
      if (activity.id === activityId) {
        const reactions = activity.reactions || []
        const existingReaction = reactions.find(r => r.emoji === emoji)
        
        if (existingReaction) {
          // Remove reaction if user already reacted
          if (existingReaction.users.includes('current-user')) {
            return {
              ...activity,
              reactions: reactions.map(r =>
                r.emoji === emoji
                  ? { ...r, count: r.count - 1, users: r.users.filter(u => u !== 'current-user') }
                  : r
              ).filter(r => r.count > 0)
            }
          } else {
            // Add reaction
            return {
              ...activity,
              reactions: reactions.map(r =>
                r.emoji === emoji
                  ? { ...r, count: r.count + 1, users: [...r.users, 'current-user'] }
                  : r
              )
            }
          }
        } else {
          // New reaction
          return {
            ...activity,
            reactions: [...reactions, { emoji, count: 1, users: ['current-user'] }]
          }
        }
      }
      return activity
    }))
  }

  const postComment = () => {
    if (!newComment.trim()) return

    const comment: TeamComment = {
      id: `comment-${Date.now()}`,
      author: {
        id: 'current-user',
        name: 'Current User',
        email: 'user@example.com'
      },
      content: newComment,
      timestamp: new Date().toISOString(),
      mentions: [],
      reactions: []
    }

    setComments([comment, ...comments])
    setNewComment('')
    setIsCommentDialogOpen(false)
    toast.success('Comment posted successfully')
  }

  const getActivityIcon = (type: string) => {
    switch (type) {
      case 'comment':
        return <MessageCircle className="h-4 w-4" />
      case 'project_created':
        return <Plus className="h-4 w-4" />
      case 'settings_changed':
        return <Settings className="h-4 w-4" />
      case 'member_added':
        return <Users className="h-4 w-4" />
      case 'api_key_created':
        return <Key className="h-4 w-4" />
      case 'document_shared':
        return <FileText className="h-4 w-4" />
      default:
        return <Eye className="h-4 w-4" />
    }
  }

  const getActivityColor = (type: string) => {
    switch (type) {
      case 'comment':
        return 'text-blue-500'
      case 'project_created':
        return 'text-green-500'
      case 'settings_changed':
        return 'text-orange-500'
      case 'member_added':
        return 'text-purple-500'
      case 'api_key_created':
        return 'text-red-500'
      case 'document_shared':
        return 'text-indigo-500'
      default:
        return 'text-gray-500'
    }
  }

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp)
    const now = new Date()
    const diff = now.getTime() - date.getTime()
    const hours = Math.floor(diff / (1000 * 60 * 60))
    const days = Math.floor(hours / 24)

    if (days > 0) return `${days} day${days > 1 ? 's' : ''} ago`
    if (hours > 0) return `${hours} hour${hours > 1 ? 's' : ''} ago`
    return 'Just now'
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-foreground">Team Activity</h2>
          <p className="text-muted-foreground">
            Stay updated with your team's latest activities and collaborations
          </p>
        </div>
        
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={handleRefresh} disabled={isRefreshing}>
            <RefreshCw className={cn("mr-2 h-4 w-4", isRefreshing && "animate-spin")} />
            Refresh
          </Button>
          
          <Dialog open={isCommentDialogOpen} onOpenChange={setIsCommentDialogOpen}>
            <DialogTrigger asChild>
              <Button size="sm">
                <MessageCircle className="mr-2 h-4 w-4" />
                New Comment
              </Button>
            </DialogTrigger>
            
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Post Team Comment</DialogTitle>
                <DialogDescription>
                  Share an update or announcement with your team
                </DialogDescription>
              </DialogHeader>

              <div className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="comment">Comment</Label>
                  <Textarea
                    id="comment"
                    value={newComment}
                    onChange={(e) => setNewComment(e.target.value)}
                    placeholder="What's happening with your projects? Use @username to mention team members..."
                    rows={4}
                  />
                </div>
              </div>

              <div className="flex justify-end gap-2">
                <Button variant="outline" onClick={() => setIsCommentDialogOpen(false)}>
                  Cancel
                </Button>
                <Button onClick={postComment} disabled={!newComment.trim()}>
                  Post Comment
                </Button>
              </div>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      <Tabs defaultValue="feed" className="space-y-6">
        <TabsList>
          <TabsTrigger value="feed">Activity Feed</TabsTrigger>
          <TabsTrigger value="comments">Team Comments</TabsTrigger>
          <TabsTrigger value="bookmarks">Bookmarks</TabsTrigger>
        </TabsList>

        <TabsContent value="feed" className="space-y-6">
          {/* Filters */}
          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center gap-4">
                <Label>Filter by type:</Label>
                <Select value={filterType} onValueChange={setFilterType}>
                  <SelectTrigger className="w-48">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Activities</SelectItem>
                    <SelectItem value="comment">Comments</SelectItem>
                    <SelectItem value="project_created">Projects</SelectItem>
                    <SelectItem value="settings_changed">Settings</SelectItem>
                    <SelectItem value="member_added">Team Changes</SelectItem>
                    <SelectItem value="document_shared">Documents</SelectItem>
                  </SelectContent>
                </Select>
                
                <div className="text-sm text-muted-foreground">
                  Showing {filteredActivities.length} activities
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Activity Feed */}
          <div className="space-y-2">
            {filteredActivities.map((activity) => (
              <Card key={activity.id} className="transition-all hover:shadow-md">
                <CardContent className="pt-6">
                  <div className="flex gap-2.5">
                    <Avatar className="h-7 w-7">
                      <AvatarImage src={activity.actor.avatar} alt={activity.actor.name} />
                      <AvatarFallback>
                        {activity.actor.name.split(' ').map(n => n[0]).join('')}
                      </AvatarFallback>
                    </Avatar>
                    
                    <div className="flex-1 space-y-3">
                      {/* Header */}
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <div className={cn("p-1 rounded", getActivityColor(activity.type))}>
                            {getActivityIcon(activity.type)}
                          </div>
                          <span className="font-medium">{activity.actor.name}</span>
                          <span className="text-muted-foreground">{activity.title.toLowerCase()}</span>
                          {activity.resource && (
                            <>
                              <span className="text-muted-foreground">·</span>
                              <Badge variant="outline" className="text-xs">
                                {activity.resource.name}
                              </Badge>
                            </>
                          )}
                        </div>
                        
                        <div className="flex items-center gap-2">
                          <div className="text-sm text-muted-foreground flex items-center gap-1">
                            <Clock className="h-3 w-3" />
                            {formatTimestamp(activity.timestamp)}
                          </div>
                          
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => toggleBookmark(activity.id)}
                            className="h-8 w-8 p-0"
                          >
                            <Bookmark 
                              className={cn(
                                "h-4 w-4",
                                activity.isBookmarked ? "fill-current text-yellow-500" : "text-muted-foreground"
                              )} 
                            />
                          </Button>
                        </div>
                      </div>

                      {/* Content */}
                      <div className="text-sm">
                        {activity.description}
                      </div>

                      {/* Metadata */}
                      {activity.metadata && (
                        <div className="p-2 bg-muted rounded-lg text-xs">
                          {activity.metadata.template && (
                            <div>Template: <span className="font-medium">{activity.metadata.template}</span></div>
                          )}
                          {activity.metadata.changes && (
                            <div>
                              Changes: {Object.entries(activity.metadata.changes).map(([key, change]: [string, any]) => (
                                <span key={key} className="font-medium">
                                  {key}: {change.from} → {change.to}
                                </span>
                              ))}
                            </div>
                          )}
                          {activity.metadata.role && (
                            <div>Role: <span className="font-medium capitalize">{activity.metadata.role}</span></div>
                          )}
                        </div>
                      )}

                      {/* Reactions */}
                      <div className="flex items-center gap-4">
                        <div className="flex items-center gap-2">
                          {['👍', '🚀', '❤️', '🎉'].map(emoji => (
                            <Button
                              key={emoji}
                              variant="ghost"
                              size="sm"
                              className="h-8 px-2 text-sm"
                              onClick={() => addReaction(activity.id, emoji)}
                            >
                              {emoji}
                            </Button>
                          ))}
                        </div>
                        
                        {activity.reactions && activity.reactions.length > 0 && (
                          <div className="flex items-center gap-2">
                            <Separator orientation="vertical" className="h-4" />
                            {activity.reactions.map(reaction => (
                              <Badge key={reaction.emoji} variant="secondary" className="text-xs">
                                {reaction.emoji} {reaction.count}
                              </Badge>
                            ))}
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </TabsContent>

        <TabsContent value="comments" className="space-y-6">
          <div className="space-y-4">
            {comments.map((comment) => (
              <Card key={comment.id}>
                <CardContent className="pt-6">
                  <div className="flex gap-2.5">
                    <Avatar className="h-7 w-7">
                      <AvatarImage src={comment.author.avatar} alt={comment.author.name} />
                      <AvatarFallback>
                        {comment.author.name.split(' ').map(n => n[0]).join('')}
                      </AvatarFallback>
                    </Avatar>
                    
                    <div className="flex-1 space-y-3">
                      <div className="flex items-center justify-between">
                        <div>
                          <span className="font-medium">{comment.author.name}</span>
                          <span className="text-muted-foreground ml-2 text-sm">
                            {formatTimestamp(comment.timestamp)}
                          </span>
                        </div>
                      </div>

                      <div className="text-sm whitespace-pre-wrap">
                        {comment.content}
                      </div>

                      {comment.reactions && comment.reactions.length > 0 && (
                        <div className="flex items-center gap-2">
                          {comment.reactions.map(reaction => (
                            <Badge key={reaction.emoji} variant="secondary" className="text-xs">
                              {reaction.emoji} {reaction.count}
                            </Badge>
                          ))}
                        </div>
                      )}

                      {/* Replies */}
                      {comment.replies && comment.replies.length > 0 && (
                        <div className="ml-8 space-y-3 border-l pl-4">
                          {comment.replies.map(reply => (
                            <div key={reply.id} className="flex gap-2">
                              <Avatar className="h-6 w-6">
                                <AvatarImage src={reply.author.avatar} alt={reply.author.name} />
                                <AvatarFallback className="text-xs">
                                  {reply.author.name.split(' ').map(n => n[0]).join('')}
                                </AvatarFallback>
                              </Avatar>
                              
                              <div className="flex-1">
                                <div className="flex items-center gap-2">
                                  <span className="font-medium text-sm">{reply.author.name}</span>
                                  <span className="text-muted-foreground text-xs">
                                    {formatTimestamp(reply.timestamp)}
                                  </span>
                                </div>
                                <div className="text-sm mt-1">{reply.content}</div>
                                
                                {reply.reactions && reply.reactions.length > 0 && (
                                  <div className="flex items-center gap-2 mt-2">
                                    {reply.reactions.map(reaction => (
                                      <Badge key={reaction.emoji} variant="secondary" className="text-xs">
                                        {reaction.emoji} {reaction.count}
                                      </Badge>
                                    ))}
                                  </div>
                                )}
                              </div>
                            </div>
                          ))}
                        </div>
                      )}
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </TabsContent>

        <TabsContent value="bookmarks" className="space-y-6">
          <div className="space-y-4">
            {activities.filter(activity => activity.isBookmarked).map((activity) => (
              <Card key={activity.id} className="border-yellow-200">
                <CardContent className="pt-6">
                  <div className="flex gap-2">
                    <div className={cn("p-2 rounded", getActivityColor(activity.type))}>
                      {getActivityIcon(activity.type)}
                    </div>
                    
                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        <span className="font-medium">{activity.title}</span>
                        <span className="text-muted-foreground">by {activity.actor.name}</span>
                        {activity.resource && (
                          <Badge variant="outline" className="text-xs">
                            {activity.resource.name}
                          </Badge>
                        )}
                        <Star className="h-4 w-4 text-yellow-500 fill-current ml-auto" />
                      </div>
                      
                      <div className="text-sm text-muted-foreground mt-1">
                        {activity.description}
                      </div>
                      
                      <div className="text-xs text-muted-foreground mt-2">
                        {formatTimestamp(activity.timestamp)}
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
            
            {activities.filter(activity => activity.isBookmarked).length === 0 && (
              <Card>
                <CardContent className="text-center py-12">
                  <Bookmark className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
                  <h3 className="text-lg font-medium mb-2">No bookmarked activities</h3>
                  <p className="text-muted-foreground">
                    Bookmark important activities to easily find them later
                  </p>
                </CardContent>
              </Card>
            )}
          </div>
        </TabsContent>
      </Tabs>
    </div>
  )
}