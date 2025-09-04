import { AUTH_CONSTANTS, AUTH_EVENTS } from './constants'
import type { User } from '@/types/auth'

interface AuthMessage {
  type: keyof typeof AUTH_EVENTS
  payload?: {
    user?: User
    timestamp?: number
  }
}

type AuthEventCallback = (event: AuthMessage) => void

export class SessionSync {
  private channel: BroadcastChannel | null = null
  private callbacks = new Map<string, AuthEventCallback>()
  private isDestroyed = false

  constructor() {
    if (typeof window !== 'undefined' && 'BroadcastChannel' in window) {
      this.channel = new BroadcastChannel(AUTH_CONSTANTS.BROADCAST_CHANNEL)
      this.setupListeners()
    }
  }

  // Public methods
  on(event: keyof typeof AUTH_EVENTS, callback: AuthEventCallback): () => void {
    const id = `${event}_${Date.now()}_${Math.random()}`
    this.callbacks.set(id, callback)
    
    return () => {
      this.callbacks.delete(id)
    }
  }

  broadcastLogin(user: User): void {
    this.broadcast({
      type: AUTH_EVENTS.LOGIN,
      payload: {
        user,
        timestamp: Date.now(),
      },
    })
  }

  broadcastLogout(): void {
    this.broadcast({
      type: AUTH_EVENTS.LOGOUT,
      payload: {
        timestamp: Date.now(),
      },
    })
  }

  broadcastTokenRefresh(): void {
    this.broadcast({
      type: AUTH_EVENTS.TOKEN_REFRESH,
      payload: {
        timestamp: Date.now(),
      },
    })
  }

  broadcastSessionExpired(): void {
    this.broadcast({
      type: AUTH_EVENTS.SESSION_EXPIRED,
      payload: {
        timestamp: Date.now(),
      },
    })
  }

  broadcastUserUpdate(user: User): void {
    this.broadcast({
      type: AUTH_EVENTS.USER_UPDATED,
      payload: {
        user,
        timestamp: Date.now(),
      },
    })
  }

  // Private methods
  private setupListeners(): void {
    if (!this.channel) return

    this.channel.addEventListener('message', (event: MessageEvent<AuthMessage>) => {
      if (this.isDestroyed) return

      const { type, payload } = event.data
      
      // Debug logging in development
      if (process.env.NODE_ENV === 'development') {
        console.debug(`[SessionSync] Received: ${type}`, payload)
      }

      // Notify all registered callbacks
      this.callbacks.forEach((callback, id) => {
        try {
          callback(event.data)
        } catch (error) {
          console.error(`[SessionSync] Callback error for ${id}:`, error)
        }
      })
    })

    // Handle channel errors
    this.channel.addEventListener('messageerror', (error) => {
      console.error('[SessionSync] Message error:', error)
    })
  }

  private broadcast(message: AuthMessage): void {
    if (!this.channel || this.isDestroyed) return

    try {
      this.channel.postMessage(message)
      
      if (process.env.NODE_ENV === 'development') {
        console.debug(`[SessionSync] Broadcasted: ${message.type}`, message.payload)
      }
    } catch (error) {
      console.error('[SessionSync] Broadcast error:', error)
    }
  }

  // Cleanup
  destroy(): void {
    this.isDestroyed = true
    this.callbacks.clear()
    
    if (this.channel) {
      this.channel.close()
      this.channel = null
    }
  }

  // Utility methods
  isSupported(): boolean {
    return typeof window !== 'undefined' && 'BroadcastChannel' in window
  }

  getActiveTabCount(): Promise<number> {
    if (!this.channel || !this.isSupported()) {
      return Promise.resolve(1)
    }

    return new Promise((resolve) => {
      let responseCount = 0
      const timeout = setTimeout(() => resolve(responseCount + 1), 100)

      const cleanup = this.on('TAB_PING_RESPONSE', () => {
        responseCount++
      })

      this.broadcast({ type: 'TAB_PING' as any })
      
      setTimeout(() => {
        cleanup()
        clearTimeout(timeout)
        resolve(responseCount + 1)
      }, 100)
    })
  }
}

// Global singleton instance
let sessionSync: SessionSync | null = null

export function getSessionSync(): SessionSync {
  if (!sessionSync) {
    sessionSync = new SessionSync()
  }
  return sessionSync
}

// Cleanup on page unload
if (typeof window !== 'undefined') {
  window.addEventListener('beforeunload', () => {
    if (sessionSync) {
      sessionSync.destroy()
      sessionSync = null
    }
  })
}