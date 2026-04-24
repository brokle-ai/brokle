import { create } from 'zustand'
import { devtools, persist } from 'zustand/middleware'

interface UIState {
  // Theme
  theme: 'light' | 'dark' | 'system'
  
  // Font
  font: string
  
  // Sidebar
  sidebarOpen: boolean
  sidebarCollapsed: boolean
  
  // Search
  searchOpen: boolean
  searchQuery: string
  searchResults: any[]
  
  // Modals and dialogs
  modals: Record<string, boolean>
  
  // Notifications
  notifications: Notification[]
  
  // Loading states
  loading: Record<string, boolean>
  
  // Actions
  setTheme: (theme: 'light' | 'dark' | 'system') => void
  setFont: (font: string) => void
  toggleSidebar: () => void
  setSidebarOpen: (open: boolean) => void
  setSidebarCollapsed: (collapsed: boolean) => void
  toggleSearch: () => void
  setSearchOpen: (open: boolean) => void
  setSearchQuery: (query: string) => void
  setSearchResults: (results: any[]) => void
  openModal: (modalId: string) => void
  closeModal: (modalId: string) => void
  addNotification: (notification: Omit<Notification, 'id'>) => void
  removeNotification: (id: string) => void
  setLoading: (key: string, loading: boolean) => void
  clearAllNotifications: () => void
}

interface Notification {
  id: string
  type: 'success' | 'error' | 'warning' | 'info'
  title: string
  message?: string
  action?: {
    label: string
    onClick: () => void
  }
  duration?: number
  persistent?: boolean
}

export const useUIStore = create<UIState>()(
  devtools(
    persist(
      (set, get) => ({
        // Initial state
        theme: 'system',
        font: 'inter',
        sidebarOpen: true,
        sidebarCollapsed: false,
        searchOpen: false,
        searchQuery: '',
        searchResults: [],
        modals: {},
        notifications: [],
        loading: {},

        // Actions
        setTheme: (theme) => set({ theme }),
        
        setFont: (font) => set({ font }),
        
        toggleSidebar: () => set((state) => ({ 
          sidebarOpen: !state.sidebarOpen 
        })),
        
        setSidebarOpen: (open) => set({ sidebarOpen: open }),
        
        setSidebarCollapsed: (collapsed) => set({ sidebarCollapsed: collapsed }),

        toggleSearch: () => set((state) => ({
          searchOpen: !state.searchOpen,
          ...(!state.searchOpen ? {} : { searchQuery: '', searchResults: [] })
        })),

        setSearchOpen: (open) => set({
          searchOpen: open,
          ...(open ? {} : { searchQuery: '', searchResults: [] })
        }),
        
        setSearchQuery: (query) => set({ searchQuery: query }),
        
        setSearchResults: (results) => set({ searchResults: results }),
        
        openModal: (modalId) => set((state) => ({ 
          modals: { ...state.modals, [modalId]: true } 
        })),
        
        closeModal: (modalId) => set((state) => ({ 
          modals: { ...state.modals, [modalId]: false } 
        })),
        
        addNotification: (notification) => {
          const id = Date.now().toString()
          const newNotification = { 
            ...notification, 
            id,
            duration: notification.duration ?? 5000 
          }
          
          set((state) => ({ 
            notifications: [...state.notifications, newNotification] 
          }))
          
          // Auto remove after duration (unless persistent)
          if (!notification.persistent && newNotification.duration > 0) {
            setTimeout(() => {
              get().removeNotification(id)
            }, newNotification.duration)
          }
        },
        
        removeNotification: (id) => set((state) => ({
          notifications: state.notifications.filter(n => n.id !== id)
        })),
        
        setLoading: (key, loading) => set((state) => ({
          loading: { ...state.loading, [key]: loading }
        })),
        
        clearAllNotifications: () => set({ notifications: [] }),
      }),
      {
        name: 'brokle-ui-storage',
        partialize: (state) => ({
          theme: state.theme,
          font: state.font,
          sidebarCollapsed: state.sidebarCollapsed,
        }),
      }
    ),
    {
      name: 'ui-store',
    }
  )
)