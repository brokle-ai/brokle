import { FileText, MessageSquare } from 'lucide-react'
import type { PromptType } from '../../types'

interface PromptTypeIconProps {
  type: PromptType
  className?: string
}

export function PromptTypeIcon({ type, className }: PromptTypeIconProps) {
  if (type === 'chat') {
    return <MessageSquare className={className} />
  }
  return <FileText className={className} />
}
