import { create } from 'zustand'
import { devtools } from 'zustand/middleware'
import type {
  DashboardData,
  DashboardWidget,
  DashboardLayout,
  DashboardPreferences
} from '@/types/dashboard'
import type { TimeRange } from '@/types/api'

interface DashboardState {
  // Dashboard data
  data: DashboardData | null

  // Time range and filters
  timeRange: TimeRange

  // Real-time data (placeholder for future implementation)
  realtimeRequests: unknown[]
  realtimeEnabled: boolean
  
  // Customization
  currentLayout: DashboardLayout | null
  availableLayouts: DashboardLayout[]
  preferences: DashboardPreferences | null
  
  // Loading states
  isLoading: boolean
  isRefreshing: boolean
  lastUpdated: string | null
  
  // Error states
  error: string | null
  
  // Actions
  setData: (data: DashboardData) => void
  setTimeRange: (range: TimeRange) => void
  setRealtimeEnabled: (enabled: boolean) => void
  addRealtimeRequest: (request: unknown) => void
  setCurrentLayout: (layout: DashboardLayout) => void
  updateWidget: (widgetId: string, updates: Partial<DashboardWidget>) => void
  addWidget: (widget: DashboardWidget) => void
  removeWidget: (widgetId: string) => void
  setPreferences: (preferences: DashboardPreferences) => void
  setLoading: (loading: boolean) => void
  setRefreshing: (refreshing: boolean) => void
  setError: (error: string | null) => void
  refreshData: () => Promise<void>
  clearRealtimeData: () => void
}

export const useDashboardStore = create<DashboardState>()(
  devtools(
    (set, get) => ({
      // Initial state
      data: null,
      timeRange: '24h',
      realtimeRequests: [],
      realtimeEnabled: false,
      currentLayout: null,
      availableLayouts: [],
      preferences: null,
      isLoading: false,
      isRefreshing: false,
      lastUpdated: null,
      error: null,

      // Actions
      setData: (data) => set({ 
        data, 
        lastUpdated: new Date().toISOString(),
        error: null 
      }),
      
      setTimeRange: (range) => set({ timeRange: range }),

      setRealtimeEnabled: (enabled) => set({ 
        realtimeEnabled: enabled,
        ...(enabled ? {} : { realtimeRequests: [] })
      }),
      
      addRealtimeRequest: (request) => set((state) => ({
        realtimeRequests: [request, ...state.realtimeRequests].slice(0, 100) // Keep last 100
      })),
      
      setCurrentLayout: (layout) => set({ currentLayout: layout }),
      
      updateWidget: (widgetId, updates) => set((state) => {
        if (!state.currentLayout) return state
        
        const updatedWidgets = state.currentLayout.widgets.map(widget =>
          widget.id === widgetId ? { ...widget, ...updates } : widget
        )
        
        return {
          currentLayout: {
            ...state.currentLayout,
            widgets: updatedWidgets,
            updatedAt: new Date().toISOString()
          }
        }
      }),
      
      addWidget: (widget) => set((state) => {
        if (!state.currentLayout) return state
        
        return {
          currentLayout: {
            ...state.currentLayout,
            widgets: [...state.currentLayout.widgets, widget],
            updatedAt: new Date().toISOString()
          }
        }
      }),
      
      removeWidget: (widgetId) => set((state) => {
        if (!state.currentLayout) return state
        
        return {
          currentLayout: {
            ...state.currentLayout,
            widgets: state.currentLayout.widgets.filter(w => w.id !== widgetId),
            updatedAt: new Date().toISOString()
          }
        }
      }),
      
      setPreferences: (preferences) => set({ preferences }),
      
      setLoading: (loading) => set({ isLoading: loading }),
      
      setRefreshing: (refreshing) => set({ isRefreshing: refreshing }),
      
      setError: (error) => set({ error }),
      
      refreshData: async () => {
        const state = get()
        set({ isRefreshing: true, error: null })
        
        try {
          // This would be replaced with actual API call
          // const data = await dashboardAPI.getData({
          //   timeRange: state.timeRange,
          //   providers: state.selectedProviders,
          // })
          // state.setData(data)

          console.log('Refreshing dashboard data...')
        } catch (error) {
          set({ error: error instanceof Error ? error.message : 'Failed to refresh data' })
        } finally {
          set({ isRefreshing: false })
        }
      },
      
      clearRealtimeData: () => set({ realtimeRequests: [] }),
    }),
    {
      name: 'dashboard-store',
    }
  )
)