/**
 * Auto-Save Hook Tests
 *
 * Tests for the debounced auto-save functionality in dashboard editing.
 * Focuses on synchronous state management tests to avoid timer/promise complexity.
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useAutoSave, getAutoSaveStatusLabel } from '../hooks/use-auto-save'
import type { Dashboard } from '../types'
import type { ReactNode } from 'react'

// Mock the API module
vi.mock('../api/dashboards-api', () => ({
  updateDashboard: vi.fn(),
}))

// Import after mocking
import { updateDashboard } from '../api/dashboards-api'

const mockUpdateDashboard = vi.mocked(updateDashboard)

// Test wrapper with QueryClient
function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })

  return function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  }
}

// Mock dashboard data
const mockDashboard: Dashboard = {
  id: 'dashboard-1',
  project_id: 'project-1',
  name: 'Test Dashboard',
  description: 'Test description',
  is_locked: false,
  layout: [],
  config: {
    widgets: [],
    time_range: { relative: '24h' },
  },
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

describe('useAutoSave', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.clearAllMocks()
    mockUpdateDashboard.mockResolvedValue(mockDashboard)
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  describe('initial state', () => {
    it('starts with idle status', () => {
      const { result } = renderHook(
        () => useAutoSave('project-1', 'dashboard-1', mockDashboard),
        { wrapper: createWrapper() }
      )

      expect(result.current.status).toBe('idle')
      expect(result.current.hasUnsavedChanges).toBe(false)
      expect(result.current.error).toBeNull()
    })

    it('exposes save functions', () => {
      const { result } = renderHook(
        () => useAutoSave('project-1', 'dashboard-1', mockDashboard),
        { wrapper: createWrapper() }
      )

      expect(typeof result.current.scheduleSave).toBe('function')
      expect(typeof result.current.saveNow).toBe('function')
      expect(typeof result.current.cancelPendingSave).toBe('function')
    })
  })

  describe('scheduleSave', () => {
    it('sets status to pending when called', () => {
      const { result } = renderHook(
        () => useAutoSave('project-1', 'dashboard-1', mockDashboard),
        { wrapper: createWrapper() }
      )

      act(() => {
        result.current.scheduleSave({ config: { widgets: [] } })
      })

      expect(result.current.status).toBe('pending')
      expect(result.current.hasUnsavedChanges).toBe(true)
    })

    it('does not save immediately - waits for debounce', () => {
      const { result } = renderHook(
        () => useAutoSave('project-1', 'dashboard-1', mockDashboard, { debounceMs: 1000 }),
        { wrapper: createWrapper() }
      )

      act(() => {
        result.current.scheduleSave({ config: { widgets: [] } })
      })

      // Should not save immediately
      expect(mockUpdateDashboard).not.toHaveBeenCalled()
      expect(result.current.status).toBe('pending')
    })

    it('does not save before debounce time', () => {
      const { result } = renderHook(
        () => useAutoSave('project-1', 'dashboard-1', mockDashboard, { debounceMs: 1000 }),
        { wrapper: createWrapper() }
      )

      act(() => {
        result.current.scheduleSave({ config: { widgets: [] } })
      })

      // Advance timer but not fully
      act(() => {
        vi.advanceTimersByTime(500)
      })

      expect(mockUpdateDashboard).not.toHaveBeenCalled()
    })

    it('does nothing when disabled', () => {
      const { result } = renderHook(
        () => useAutoSave('project-1', 'dashboard-1', mockDashboard, { enabled: false }),
        { wrapper: createWrapper() }
      )

      act(() => {
        result.current.scheduleSave({ config: { widgets: [] } })
      })

      // Status should remain idle
      expect(result.current.status).toBe('idle')
      expect(result.current.hasUnsavedChanges).toBe(false)

      act(() => {
        vi.advanceTimersByTime(2000)
      })

      expect(mockUpdateDashboard).not.toHaveBeenCalled()
    })

    it('resets debounce timer on subsequent calls', () => {
      const { result } = renderHook(
        () => useAutoSave('project-1', 'dashboard-1', mockDashboard, { debounceMs: 1000 }),
        { wrapper: createWrapper() }
      )

      act(() => {
        result.current.scheduleSave({ config: { widgets: [] } })
      })

      // Advance part way
      act(() => {
        vi.advanceTimersByTime(500)
      })

      // Schedule again - this resets the timer
      act(() => {
        result.current.scheduleSave({ layout: [] })
      })

      // Advance another 500ms (would have triggered if timer wasn't reset)
      act(() => {
        vi.advanceTimersByTime(500)
      })

      // Still should not have called (timer reset, need another 500ms)
      expect(mockUpdateDashboard).not.toHaveBeenCalled()
    })
  })

  describe('cancelPendingSave', () => {
    it('resets to idle state', () => {
      const { result } = renderHook(
        () => useAutoSave('project-1', 'dashboard-1', mockDashboard, { debounceMs: 1000 }),
        { wrapper: createWrapper() }
      )

      act(() => {
        result.current.scheduleSave({ config: { widgets: [] } })
      })

      expect(result.current.status).toBe('pending')
      expect(result.current.hasUnsavedChanges).toBe(true)

      act(() => {
        result.current.cancelPendingSave()
      })

      expect(result.current.status).toBe('idle')
      expect(result.current.hasUnsavedChanges).toBe(false)
      expect(result.current.error).toBeNull()
    })

    it('prevents scheduled save from executing', () => {
      const { result } = renderHook(
        () => useAutoSave('project-1', 'dashboard-1', mockDashboard, { debounceMs: 100 }),
        { wrapper: createWrapper() }
      )

      act(() => {
        result.current.scheduleSave({ config: { widgets: [] } })
      })

      act(() => {
        result.current.cancelPendingSave()
      })

      // Advance past debounce time
      act(() => {
        vi.advanceTimersByTime(200)
      })

      expect(mockUpdateDashboard).not.toHaveBeenCalled()
    })
  })

  describe('saveNow', () => {
    it('does nothing without pending changes', () => {
      const { result } = renderHook(
        () => useAutoSave('project-1', 'dashboard-1', mockDashboard),
        { wrapper: createWrapper() }
      )

      act(() => {
        result.current.saveNow()
      })

      expect(mockUpdateDashboard).not.toHaveBeenCalled()
    })

    it('does nothing when disabled', () => {
      const { result } = renderHook(
        () => useAutoSave('project-1', 'dashboard-1', mockDashboard, { enabled: false }),
        { wrapper: createWrapper() }
      )

      // Try to trigger a save
      act(() => {
        result.current.scheduleSave({ config: { widgets: [] } })
        result.current.saveNow()
      })

      expect(mockUpdateDashboard).not.toHaveBeenCalled()
    })
  })

  describe('without dashboard', () => {
    it('handles null dashboard gracefully', () => {
      const { result } = renderHook(
        () => useAutoSave('project-1', 'dashboard-1', null),
        { wrapper: createWrapper() }
      )

      act(() => {
        result.current.scheduleSave({ config: { widgets: [] } })
      })

      expect(result.current.status).toBe('pending')
    })

    it('handles undefined dashboard gracefully', () => {
      const { result } = renderHook(
        () => useAutoSave('project-1', 'dashboard-1', undefined),
        { wrapper: createWrapper() }
      )

      act(() => {
        result.current.scheduleSave({ config: { widgets: [] } })
      })

      expect(result.current.status).toBe('pending')
    })
  })
})

describe('getAutoSaveStatusLabel', () => {
  it('returns empty string for idle', () => {
    expect(getAutoSaveStatusLabel('idle')).toBe('')
  })

  it('returns appropriate label for pending', () => {
    expect(getAutoSaveStatusLabel('pending')).toBe('Unsaved changes')
  })

  it('returns appropriate label for saving', () => {
    expect(getAutoSaveStatusLabel('saving')).toBe('Saving...')
  })

  it('returns appropriate label for saved', () => {
    expect(getAutoSaveStatusLabel('saved')).toBe('Saved')
  })

  it('returns appropriate label for error', () => {
    expect(getAutoSaveStatusLabel('error')).toBe('Save failed')
  })
})
