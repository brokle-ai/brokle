import { describe, it, expect, beforeEach } from 'vitest'
import { usePlaygroundStore, createContentSnapshot, isWindowDirty } from '../stores/playground-store'
import { createMessage } from '../types'

/**
 * Playground Store Tests
 *
 * Following testing philosophy from docs/TESTING.md:
 * - Test business logic (dirty tracking via snapshots, window management, limit enforcement)
 * - Don't test trivial operations (simple state updates)
 */

describe('PlaygroundStore', () => {
  beforeEach(() => {
    // Reset store state before each test
    usePlaygroundStore.getState().clearAll()
  })

  // ============================================================================
  // HIGH-VALUE TESTS: Business Logic - Window Management
  // ============================================================================

  describe('window management', () => {
    it('should enforce maximum of 20 windows', () => {
      const store = usePlaygroundStore.getState()

      // Start with 1 window (default)
      expect(store.windows.length).toBe(1)

      // Add windows up to practical limit
      for (let i = 0; i < 19; i++) {
        store.addWindow()
      }
      expect(usePlaygroundStore.getState().windows.length).toBe(20)

      // Attempt to add 21st window should be blocked
      store.addWindow()
      expect(usePlaygroundStore.getState().windows.length).toBe(20)
    })

    it('should enforce minimum of 1 window', () => {
      const store = usePlaygroundStore.getState()

      // Start with 1 window (default)
      expect(store.windows.length).toBe(1)

      // Attempt to remove last window should be blocked
      store.removeWindow(0)
      expect(usePlaygroundStore.getState().windows.length).toBe(1)
    })

    it('should duplicate window correctly', () => {
      const store = usePlaygroundStore.getState()

      // Update first window with custom state
      store.updateWindow(0, {
        messages: [
          createMessage('system', 'Custom system prompt'),
          createMessage('user', 'Hello'),
        ],
        variables: { name: 'test' },
        config: { model: 'gpt-4' },
      })

      // Mark it as saved (simulate save)
      const snapshot = createContentSnapshot(usePlaygroundStore.getState().windows[0])
      store.setLastSavedSnapshot(0, snapshot)

      // Duplicate window
      store.duplicateWindow(0)

      const windows = usePlaygroundStore.getState().windows
      expect(windows.length).toBe(2)

      // Duplicated window should have same content but different ID
      expect(windows[1].messages.length).toBe(windows[0].messages.length)
      expect(windows[1].messages[0].content).toBe(windows[0].messages[0].content)
      expect(windows[1].variables).toEqual(windows[0].variables)
      expect(windows[1].config).toEqual(windows[0].config)
      expect(windows[1].id).not.toBe(windows[0].id)

      // Duplicated window should have clean execution state
      expect(windows[1].lastOutput).toBeNull()
      expect(windows[1].lastMetrics).toBeNull()
      expect(windows[1].isExecuting).toBe(false)

      // Duplicated window should have snapshot set (clean after duplication)
      expect(windows[1].lastSavedSnapshot).not.toBeNull()
    })
  })

  // ============================================================================
  // HIGH-VALUE TESTS: Business Logic - Dirty Tracking via Snapshots
  // ============================================================================

  describe('dirty tracking via snapshots', () => {
    it('should not be dirty before any save (lastSavedSnapshot is null)', () => {
      const store = usePlaygroundStore.getState()
      const window = store.windows[0]

      // Window starts with null snapshot - isWindowDirty returns false
      expect(window.lastSavedSnapshot).toBeNull()
      expect(isWindowDirty(window)).toBe(false)
    })

    it('should detect dirty state when content changes after save', () => {
      const store = usePlaygroundStore.getState()

      // Set initial content
      store.updateWindow(0, {
        messages: [createMessage('user', 'original')],
      })

      // Simulate save by capturing snapshot
      const snapshot = createContentSnapshot(usePlaygroundStore.getState().windows[0])
      store.setLastSavedSnapshot(0, snapshot)

      // After save, should not be dirty (content matches snapshot)
      let window = usePlaygroundStore.getState().windows[0]
      expect(isWindowDirty(window)).toBe(false)

      // Change content - should now be dirty
      store.updateWindow(0, {
        messages: [createMessage('user', 'modified')],
      })
      window = usePlaygroundStore.getState().windows[0]
      expect(isWindowDirty(window)).toBe(true)
    })

    it('should detect dirty state when variables change after save', () => {
      const store = usePlaygroundStore.getState()

      // Set initial variables
      store.updateWindow(0, { variables: { foo: 'bar' } })

      // Simulate save
      const snapshot = createContentSnapshot(usePlaygroundStore.getState().windows[0])
      store.setLastSavedSnapshot(0, snapshot)

      // After save, should not be dirty
      let window = usePlaygroundStore.getState().windows[0]
      expect(isWindowDirty(window)).toBe(false)

      // Change variables - should now be dirty
      store.updateWindow(0, { variables: { foo: 'baz' } })
      window = usePlaygroundStore.getState().windows[0]
      expect(isWindowDirty(window)).toBe(true)
    })

    it('should detect dirty state when config changes after save', () => {
      const store = usePlaygroundStore.getState()

      // Set initial config
      store.updateWindow(0, { config: { model: 'gpt-4' } })

      // Simulate save
      const snapshot = createContentSnapshot(usePlaygroundStore.getState().windows[0])
      store.setLastSavedSnapshot(0, snapshot)

      // After save, should not be dirty
      let window = usePlaygroundStore.getState().windows[0]
      expect(isWindowDirty(window)).toBe(false)

      // Change config - should now be dirty
      store.updateWindow(0, { config: { model: 'gpt-4o' } })
      window = usePlaygroundStore.getState().windows[0]
      expect(isWindowDirty(window)).toBe(true)
    })

    it('should NOT be dirty when execution state changes', () => {
      const store = usePlaygroundStore.getState()

      // Set initial content and save
      store.updateWindow(0, {
        messages: [createMessage('user', 'test')],
      })
      const snapshot = createContentSnapshot(usePlaygroundStore.getState().windows[0])
      store.setLastSavedSnapshot(0, snapshot)

      // Should not be dirty after save
      let window = usePlaygroundStore.getState().windows[0]
      expect(isWindowDirty(window)).toBe(false)

      // Update execution state - should NOT affect dirty state
      store.setWindowExecuting(0, true)
      window = usePlaygroundStore.getState().windows[0]
      expect(isWindowDirty(window)).toBe(false)

      store.setWindowOutput(0, 'some output', { model: 'gpt-4' })
      window = usePlaygroundStore.getState().windows[0]
      expect(isWindowDirty(window)).toBe(false)
    })

    it('should become clean when content reverts to saved state', () => {
      const store = usePlaygroundStore.getState()

      // Set initial content and save
      const originalMessages = [createMessage('user', 'original')]
      store.updateWindow(0, { messages: originalMessages })
      const snapshot = createContentSnapshot(usePlaygroundStore.getState().windows[0])
      store.setLastSavedSnapshot(0, snapshot)

      // Change content - should be dirty
      store.updateWindow(0, {
        messages: [createMessage('user', 'modified')],
      })
      let window = usePlaygroundStore.getState().windows[0]
      expect(isWindowDirty(window)).toBe(true)

      // Revert to original content - should be clean again
      // Note: Need to recreate with same role/content but new IDs won't match
      // The snapshot comparison strips IDs, so content match is what matters
      store.updateWindow(0, {
        messages: [createMessage('user', 'original')],
      })
      window = usePlaygroundStore.getState().windows[0]
      expect(isWindowDirty(window)).toBe(false)
    })

    it('should set all saved snapshots at once', () => {
      const store = usePlaygroundStore.getState()

      // Add another window (addWindow initializes snapshot)
      store.addWindow()

      // Set content for both windows
      store.updateWindow(0, {
        messages: [createMessage('user', 'window 1')],
      })
      store.updateWindow(1, {
        messages: [createMessage('user', 'window 2')],
      })

      // Window 0: null snapshot (never saved) → isWindowDirty = false
      // Window 1: has snapshot from addWindow, content changed → isWindowDirty = true
      let windows = usePlaygroundStore.getState().windows
      expect(windows[0].lastSavedSnapshot).toBeNull()
      expect(isWindowDirty(windows[0])).toBe(false) // null snapshot = not dirty
      expect(isWindowDirty(windows[1])).toBe(true)  // has snapshot, content changed

      // Save all - sets snapshots for all windows
      store.setAllSavedSnapshots()

      // Both windows should now have snapshots matching their current content
      windows = usePlaygroundStore.getState().windows
      expect(windows[0].lastSavedSnapshot).toBe(createContentSnapshot(windows[0]))
      expect(windows[1].lastSavedSnapshot).toBe(createContentSnapshot(windows[1]))
      expect(isWindowDirty(windows[0])).toBe(false)
      expect(isWindowDirty(windows[1])).toBe(false)
    })
  })

  // ============================================================================
  // HIGH-VALUE TESTS: Business Logic - Session ID Management
  // ============================================================================

  describe('session ID management', () => {
    it('should track current session ID', () => {
      const store = usePlaygroundStore.getState()

      // Starts with no session ID
      expect(store.currentSessionId).toBeNull()

      // Set session ID
      store.setCurrentSessionId('session-123')
      expect(usePlaygroundStore.getState().currentSessionId).toBe('session-123')

      // Clear session ID
      store.setCurrentSessionId(null)
      expect(usePlaygroundStore.getState().currentSessionId).toBeNull()
    })

    it('should clear session ID on clearAll', () => {
      const store = usePlaygroundStore.getState()

      // Set session ID
      store.setCurrentSessionId('session-123')
      expect(usePlaygroundStore.getState().currentSessionId).toBe('session-123')

      // Clear all
      store.clearAll()
      expect(usePlaygroundStore.getState().currentSessionId).toBeNull()
    })
  })

  // ============================================================================
  // HIGH-VALUE TESTS: Business Logic - Clear All Reset
  // ============================================================================

  describe('clearAll', () => {
    it('should reset to initial state', () => {
      const store = usePlaygroundStore.getState()

      // Modify state
      store.setCurrentSessionId('session-123')
      store.addWindow()
      store.updateWindow(0, {
        messages: [createMessage('user', 'test')],
      })
      store.setSharedVariables({ foo: 'bar' })
      store.toggleSharedVariables()

      // Verify state is modified
      let state = usePlaygroundStore.getState()
      expect(state.currentSessionId).toBe('session-123')
      expect(state.windows.length).toBe(2)
      expect(state.sharedVariables).toEqual({ foo: 'bar' })
      expect(state.useSharedVariables).toBe(true)

      // Clear all
      store.clearAll()

      // Verify reset to initial state
      state = usePlaygroundStore.getState()
      expect(state.currentSessionId).toBeNull()
      expect(state.windows.length).toBe(1)
      expect(state.windows[0].messages.length).toBeGreaterThan(0) // Has default messages
      expect(state.sharedVariables).toEqual({})
      expect(state.useSharedVariables).toBe(false)
      expect(state.isExecutingAll).toBe(false)
    })
  })
})
