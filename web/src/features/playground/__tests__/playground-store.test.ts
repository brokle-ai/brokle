import { describe, it, expect, beforeEach } from 'vitest'
import { usePlaygroundStore, createContentSnapshot, isWindowDirty } from '../stores/playground-store'
import { createMessage, type ChatMessage, type ModelConfig } from '../types'

/**
 * Type for captured inputs passed to setWindowOutput
 */
interface CapturedInputs {
  messages: ChatMessage[]
  variables: Record<string, string>
  config: ModelConfig | null
}

/**
 * Helper to create mock CapturedInputs for history tests
 */
const createMockCapturedInputs = (overrides?: Partial<CapturedInputs>): CapturedInputs => ({
  messages: [{ id: crypto.randomUUID(), role: 'user', content: 'Test message' }],
  variables: {},
  config: {
    model: 'gpt-4',
    provider: 'openai',
    temperature: 0.7,
    temperature_enabled: true,
  },
  ...overrides,
})

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

  // ============================================================================
  // HIGH-VALUE TESTS: Business Logic - Run History
  // ============================================================================

  describe('run history', () => {
    describe('setWindowOutput - creates history entries', () => {
      it('should store full config with _enabled flags in history entry', () => {
        const store = usePlaygroundStore.getState()

        // Full config with enabled/disabled params
        const fullConfig: ModelConfig = {
          model: 'gpt-4',
          provider: 'openai',
          temperature: 0.7,
          temperature_enabled: true,
          max_tokens: 1000,
          max_tokens_enabled: false, // Disabled but has value
          top_p: 0.9,
          top_p_enabled: true,
          frequency_penalty: 0.5,
          frequency_penalty_enabled: false,
          presence_penalty: 0.3,
          presence_penalty_enabled: true,
        }

        // Captured inputs (what streaming hook passes)
        const capturedInputs: CapturedInputs = {
          messages: [{ id: '1', role: 'user', content: 'Hello' }],
          variables: { name: 'test' },
          config: fullConfig,
        }

        store.setWindowOutput(0, 'Response', { model: 'gpt-4' }, capturedInputs)

        const historyEntry = usePlaygroundStore.getState().windows[0].runHistory[0]

        // Verify ALL config fields preserved including _enabled flags
        expect(historyEntry.config).toEqual(fullConfig)
        expect(historyEntry.config?.temperature_enabled).toBe(true)
        expect(historyEntry.config?.max_tokens_enabled).toBe(false)
        expect(historyEntry.config?.max_tokens).toBe(1000) // Value preserved even when disabled
        expect(historyEntry.config?.top_p_enabled).toBe(true)
        expect(historyEntry.config?.frequency_penalty_enabled).toBe(false)
        expect(historyEntry.config?.presence_penalty_enabled).toBe(true)
      })

      it('should create history entry with correct structure', () => {
        const store = usePlaygroundStore.getState()

        const capturedInputs = createMockCapturedInputs({
          messages: [{ id: 'm1', role: 'user', content: 'Test prompt' }],
          variables: { key: 'value' },
        })

        const metrics = {
          model: 'gpt-4',
          prompt_tokens: 10,
          completion_tokens: 20,
          total_tokens: 30,
          cost: 0.001,
          ttft_ms: 100,
          total_duration_ms: 500,
        }

        store.setWindowOutput(0, 'Generated response', metrics, capturedInputs)

        const historyEntry = usePlaygroundStore.getState().windows[0].runHistory[0]

        // Verify complete structure
        expect(historyEntry.id).toBeDefined()
        expect(historyEntry.id.length).toBeGreaterThan(0)
        expect(historyEntry.content).toBe('Generated response')
        expect(historyEntry.timestamp).toBeDefined()
        expect(new Date(historyEntry.timestamp).getTime()).not.toBeNaN()
        expect(historyEntry.isStale).toBe(false)
        expect(historyEntry.messages).toHaveLength(1)
        expect(historyEntry.messages[0].content).toBe('Test prompt')
        expect(historyEntry.variables).toEqual({ key: 'value' })
        expect(historyEntry.config).toEqual(capturedInputs.config)

        // Verify metrics mapping
        expect(historyEntry.metrics?.model).toBe('gpt-4')
        expect(historyEntry.metrics?.prompt_tokens).toBe(10)
        expect(historyEntry.metrics?.completion_tokens).toBe(20)
        expect(historyEntry.metrics?.total_tokens).toBe(30)
        expect(historyEntry.metrics?.cost).toBe(0.001)
        expect(historyEntry.metrics?.ttft_ms).toBe(100)
        expect(historyEntry.metrics?.latency_ms).toBe(500)
      })

      it('should enforce max 10 history entries (newest first)', () => {
        const store = usePlaygroundStore.getState()

        // Add 12 entries
        for (let i = 1; i <= 12; i++) {
          store.setWindowOutput(
            0,
            `Response ${i}`,
            { model: 'gpt-4' },
            createMockCapturedInputs({
              messages: [{ id: `m${i}`, role: 'user', content: `Message ${i}` }],
            })
          )
        }

        const history = usePlaygroundStore.getState().windows[0].runHistory

        // Verify only 10 entries remain
        expect(history).toHaveLength(10)

        // Verify newest first (entry 12 should be first)
        expect(history[0].content).toBe('Response 12')
        expect(history[9].content).toBe('Response 3') // Entry 1 and 2 were evicted
      })

      it('should capture messages from inputSnapshot not current window state', () => {
        const store = usePlaygroundStore.getState()

        // Create captured inputs at execution start
        const capturedInputs = createMockCapturedInputs({
          messages: [{ id: 'm1', role: 'user', content: 'Original message at execution start' }],
        })

        // Modify window messages AFTER creating capturedInputs (simulating user typing during execution)
        store.updateWindow(0, {
          messages: [createMessage('user', 'Modified message during execution')],
        })

        // Set output with original capturedInputs
        store.setWindowOutput(0, 'Response', { model: 'gpt-4' }, capturedInputs)

        const historyEntry = usePlaygroundStore.getState().windows[0].runHistory[0]

        // History should use capturedInputs.messages, not current window state
        expect(historyEntry.messages[0].content).toBe('Original message at execution start')
      })

      it('should fall back to current window state when inputSnapshot not provided', () => {
        const store = usePlaygroundStore.getState()

        // Set up window with specific messages
        store.updateWindow(0, {
          messages: [createMessage('user', 'Current window message')],
          variables: { key: 'window-value' },
          config: { model: 'gpt-3.5' },
        })

        // Set output without inputSnapshot
        store.setWindowOutput(0, 'Response', { model: 'gpt-3.5' })

        const historyEntry = usePlaygroundStore.getState().windows[0].runHistory[0]

        // Should use current window state as fallback
        expect(historyEntry.messages[0].content).toBe('Current window message')
        expect(historyEntry.variables).toEqual({ key: 'window-value' })
        expect(historyEntry.config?.model).toBe('gpt-3.5')
      })

      it('should handle null metrics gracefully', () => {
        const store = usePlaygroundStore.getState()

        store.setWindowOutput(0, 'Response', null, createMockCapturedInputs())

        const historyEntry = usePlaygroundStore.getState().windows[0].runHistory[0]
        expect(historyEntry.metrics).toBeNull()
      })
    })

    describe('restoreFromHistory - restores full state', () => {
      it('should restore full config with _enabled flags from history', () => {
        const store = usePlaygroundStore.getState()

        const originalConfig: ModelConfig = {
          model: 'gpt-4',
          provider: 'openai',
          temperature: 0.7,
          temperature_enabled: true,
          max_tokens: 1000,
          max_tokens_enabled: false,
          top_p: 0.9,
          top_p_enabled: false,
          frequency_penalty: 0.5,
          frequency_penalty_enabled: true,
          presence_penalty: 0.3,
          presence_penalty_enabled: false,
        }

        // Create history entry with full config
        store.setWindowOutput(
          0,
          'Original response',
          { model: 'gpt-4' },
          {
            messages: [{ id: '1', role: 'user', content: 'Test' }],
            variables: { key: 'value' },
            config: originalConfig,
          }
        )

        const historyId = usePlaygroundStore.getState().windows[0].runHistory[0].id

        // Modify current window config completely
        store.updateWindow(0, {
          config: {
            model: 'gpt-3.5',
            temperature: 0.5,
            temperature_enabled: false,
            max_tokens: 500,
            max_tokens_enabled: true,
          },
        })

        // Restore from history
        store.restoreFromHistory(0, historyId)

        // Verify FULL config restored including _enabled flags
        const restoredConfig = usePlaygroundStore.getState().windows[0].config
        expect(restoredConfig?.model).toBe('gpt-4')
        expect(restoredConfig?.provider).toBe('openai')
        expect(restoredConfig?.temperature).toBe(0.7)
        expect(restoredConfig?.temperature_enabled).toBe(true)
        expect(restoredConfig?.max_tokens).toBe(1000)
        expect(restoredConfig?.max_tokens_enabled).toBe(false) // Critical: disabled state preserved
        expect(restoredConfig?.top_p).toBe(0.9)
        expect(restoredConfig?.top_p_enabled).toBe(false)
        expect(restoredConfig?.frequency_penalty).toBe(0.5)
        expect(restoredConfig?.frequency_penalty_enabled).toBe(true)
        expect(restoredConfig?.presence_penalty).toBe(0.3)
        expect(restoredConfig?.presence_penalty_enabled).toBe(false)
      })

      it('should generate new message IDs on restore (for drag-drop)', () => {
        const store = usePlaygroundStore.getState()

        const originalMessageId = 'original-msg-id'
        store.setWindowOutput(
          0,
          'Response',
          { model: 'gpt-4' },
          {
            messages: [{ id: originalMessageId, role: 'user', content: 'Test' }],
            variables: {},
            config: null,
          }
        )

        const historyId = usePlaygroundStore.getState().windows[0].runHistory[0].id

        // Clear window and restore
        store.updateWindow(0, { messages: [] })
        store.restoreFromHistory(0, historyId)

        const restoredMessages = usePlaygroundStore.getState().windows[0].messages

        // Messages should have NEW IDs (for drag-drop), not original IDs
        expect(restoredMessages).toHaveLength(1)
        expect(restoredMessages[0].id).not.toBe(originalMessageId)
        expect(restoredMessages[0].content).toBe('Test')
        expect(restoredMessages[0].role).toBe('user')
      })

      it('should restore messages, variables, output, and metrics', () => {
        const store = usePlaygroundStore.getState()

        const capturedInputs: CapturedInputs = {
          messages: [
            { id: 'm1', role: 'system', content: 'System prompt' },
            { id: 'm2', role: 'user', content: 'User message' },
          ],
          variables: { var1: 'val1', var2: 'val2' },
          config: { model: 'gpt-4' },
        }

        const metrics = {
          model: 'gpt-4',
          prompt_tokens: 50,
          completion_tokens: 100,
          total_tokens: 150,
          cost: 0.01,
          ttft_ms: 200,
          total_duration_ms: 1000,
        }

        store.setWindowOutput(0, 'Historical response', metrics, capturedInputs)
        const historyId = usePlaygroundStore.getState().windows[0].runHistory[0].id

        // Clear current state
        store.updateWindow(0, {
          messages: [],
          variables: {},
          config: null,
          lastOutput: null,
          lastMetrics: null,
        })

        // Restore from history
        store.restoreFromHistory(0, historyId)

        const window = usePlaygroundStore.getState().windows[0]

        // Verify comprehensive restoration
        expect(window.messages).toHaveLength(2)
        expect(window.messages[0].role).toBe('system')
        expect(window.messages[0].content).toBe('System prompt')
        expect(window.messages[1].role).toBe('user')
        expect(window.messages[1].content).toBe('User message')
        expect(window.variables).toEqual({ var1: 'val1', var2: 'val2' })
        expect(window.config?.model).toBe('gpt-4')
        expect(window.lastOutput).toBe('Historical response')
        expect(window.lastMetrics?.model).toBe('gpt-4')
        expect(window.lastMetrics?.prompt_tokens).toBe(50)
        expect(window.lastMetrics?.total_duration_ms).toBe(1000)
      })

      it('should be no-op for non-existent history entry', () => {
        const store = usePlaygroundStore.getState()

        // Set up window with specific state
        store.updateWindow(0, {
          messages: [createMessage('user', 'Current message')],
          variables: { key: 'current-value' },
        })

        const originalState = usePlaygroundStore.getState().windows[0]

        // Attempt to restore from non-existent history ID
        store.restoreFromHistory(0, 'non-existent-id')

        const afterState = usePlaygroundStore.getState().windows[0]

        // State should be unchanged
        expect(afterState.messages[0].content).toBe(originalState.messages[0].content)
        expect(afterState.variables).toEqual(originalState.variables)
      })

      it('should be no-op for non-existent window index', () => {
        const store = usePlaygroundStore.getState()

        // Add history to window 0
        store.setWindowOutput(0, 'Response', null, createMockCapturedInputs())
        const historyId = usePlaygroundStore.getState().windows[0].runHistory[0].id

        // Attempt to restore to non-existent window
        store.restoreFromHistory(99, historyId)

        // No error should occur, state unchanged
        expect(usePlaygroundStore.getState().windows).toHaveLength(1)
      })

      it('should handle null config in history entry', () => {
        const store = usePlaygroundStore.getState()

        store.setWindowOutput(
          0,
          'Response',
          null,
          {
            messages: [{ id: 'm1', role: 'user', content: 'Test' }],
            variables: {},
            config: null,
          }
        )

        const historyId = usePlaygroundStore.getState().windows[0].runHistory[0].id

        // Set current config to something
        store.updateWindow(0, { config: { model: 'gpt-4' } })

        // Restore from history with null config
        store.restoreFromHistory(0, historyId)

        expect(usePlaygroundStore.getState().windows[0].config).toBeNull()
      })
    })

    describe('markHistoryAsStale', () => {
      it('should mark all history entries as stale', () => {
        const store = usePlaygroundStore.getState()

        // Add multiple history entries
        store.setWindowOutput(0, 'R1', null, createMockCapturedInputs())
        store.setWindowOutput(0, 'R2', null, createMockCapturedInputs())
        store.setWindowOutput(0, 'R3', null, createMockCapturedInputs())

        // Verify initially not stale
        let history = usePlaygroundStore.getState().windows[0].runHistory
        expect(history).toHaveLength(3)
        expect(history[0].isStale).toBe(false)
        expect(history[1].isStale).toBe(false)
        expect(history[2].isStale).toBe(false)

        // Mark as stale
        store.markHistoryAsStale(0)

        // Verify all entries marked stale
        history = usePlaygroundStore.getState().windows[0].runHistory
        expect(history[0].isStale).toBe(true)
        expect(history[1].isStale).toBe(true)
        expect(history[2].isStale).toBe(true)
      })

      it('should be no-op when window has no history', () => {
        const store = usePlaygroundStore.getState()

        // Window starts with empty history
        expect(usePlaygroundStore.getState().windows[0].runHistory).toHaveLength(0)

        // Should not throw error
        store.markHistoryAsStale(0)

        // Still empty
        expect(usePlaygroundStore.getState().windows[0].runHistory).toHaveLength(0)
      })

      it('should be no-op when window does not exist', () => {
        const store = usePlaygroundStore.getState()

        // Should not throw error for invalid index
        store.markHistoryAsStale(99)

        // Original state unchanged
        expect(usePlaygroundStore.getState().windows).toHaveLength(1)
      })

      it('should not affect new runs added after marking stale', () => {
        const store = usePlaygroundStore.getState()

        // Add initial history and mark stale
        store.setWindowOutput(0, 'Old run', null, createMockCapturedInputs())
        store.markHistoryAsStale(0)

        // Add new run after marking stale
        store.setWindowOutput(0, 'New run', null, createMockCapturedInputs())

        const history = usePlaygroundStore.getState().windows[0].runHistory

        // New run should NOT be stale
        expect(history[0].content).toBe('New run')
        expect(history[0].isStale).toBe(false)

        // Old run should still be stale
        expect(history[1].content).toBe('Old run')
        expect(history[1].isStale).toBe(true)
      })
    })

    describe('clearWindowHistory', () => {
      it('should clear all history entries for window', () => {
        const store = usePlaygroundStore.getState()

        // Add entries
        store.setWindowOutput(0, 'R1', null, createMockCapturedInputs())
        store.setWindowOutput(0, 'R2', null, createMockCapturedInputs())
        expect(usePlaygroundStore.getState().windows[0].runHistory).toHaveLength(2)

        // Clear history
        store.clearWindowHistory(0)

        // Verify empty
        expect(usePlaygroundStore.getState().windows[0].runHistory).toHaveLength(0)
      })

      it('should not affect other window state', () => {
        const store = usePlaygroundStore.getState()

        // Set up window with state and history
        store.updateWindow(0, {
          messages: [createMessage('user', 'Keep this message')],
          variables: { key: 'keep-this-value' },
          config: { model: 'gpt-4' },
        })
        store.setWindowOutput(0, 'Response', null, createMockCapturedInputs())

        // Clear history
        store.clearWindowHistory(0)

        // Verify other state unchanged
        const window = usePlaygroundStore.getState().windows[0]
        expect(window.runHistory).toHaveLength(0)
        expect(window.messages[0].content).toBe('Keep this message')
        expect(window.variables).toEqual({ key: 'keep-this-value' })
        expect(window.config?.model).toBe('gpt-4')
      })

      it('should be no-op for non-existent window', () => {
        const store = usePlaygroundStore.getState()

        // Should not throw error
        store.clearWindowHistory(99)

        expect(usePlaygroundStore.getState().windows).toHaveLength(1)
      })
    })
  })
})
