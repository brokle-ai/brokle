import { Zap } from 'lucide-react'
import { cn } from '@/lib/utils'

interface BrokleLogoProps {
  className?: string
  showText?: boolean
  showTagline?: boolean
}

export const BrokleLogo = ({ className, showText = false, showTagline = false }: BrokleLogoProps) => {
  if (showText || showTagline) {
    return (
      <div className="flex items-center gap-2">
        <div className={cn("flex items-center justify-center rounded-lg bg-gradient-to-br from-blue-600 to-purple-600 text-white", className || "h-8 w-8")}>
          <Zap className="h-5 w-5 fill-white" />
        </div>
        {showText && (
          <div className="flex flex-col">
            <span className="font-bold text-lg leading-none">Brokle</span>
            {showTagline && (
              <span className="text-[10px] text-muted-foreground leading-none">AI Control Plane</span>
            )}
          </div>
        )}
      </div>
    )
  }

  return (
    <div className={cn("flex items-center justify-center rounded-lg bg-gradient-to-br from-blue-600 to-purple-600 text-white", className || "h-6 w-6")}>
      <Zap className={cn("fill-white", className ? "h-4 w-4" : "h-3 w-3")} />
    </div>
  )
}