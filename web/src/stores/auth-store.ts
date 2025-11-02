import { create } from 'zustand'
import { devtools, persist } from 'zustand/middleware'
import type { User, Organization, Project, ApiKey } from '@/types/auth'

interface AuthState {
  // Auth state
  user: User | null
  organization: Organization | null
  currentProject: Project | null
  isAuthenticated: boolean
  isLoading: boolean

  // API keys
  apiKeys: ApiKey[]

  // Actions
  setUser: (user: User | null) => void
  setOrganization: (organization: Organization | null) => void
  setCurrentProject: (project: Project | null) => void
  setApiKeys: (apiKeys: ApiKey[]) => void
  login: (user: User, organization: Organization) => void
  logout: () => void
  setLoading: (loading: boolean) => void

  // NOTE: Permission checks removed - use useHasAccess hook instead
  // Old methods (hasPermission, isOwner, isAdmin) deprecated
  // Use: const canInvite = useHasAccess({ scope: "members:invite" })
}

export const useAuthStore = create<AuthState>()(
  devtools(
    persist(
      (set, get) => ({
        // Initial state
        user: null,
        organization: null,
        currentProject: null,
        isAuthenticated: false,
        isLoading: false,
        apiKeys: [],

        // Actions
        setUser: (user) => set({ user }),
        
        setOrganization: (organization) => set({ organization }),
        
        setCurrentProject: (project) => set({ currentProject: project }),
        
        setApiKeys: (apiKeys) => set({ apiKeys }),
        
        login: (user, organization) => set({ 
          user, 
          organization, 
          isAuthenticated: true,
          isLoading: false 
        }),
        
        logout: () => set({
          user: null,
          organization: null,
          currentProject: null,
          isAuthenticated: false,
          apiKeys: [],
          isLoading: false
        }),

        setLoading: (loading) => set({ isLoading: loading }),

        // Deprecated: Permission checks removed
        // Use useHasAccess hook instead:
        //   const canInvite = useHasAccess({ scope: "members:invite" })
        //   const canDelete = useHasAccess({ scope: "traces:delete", projectId })
      }),
      {
        name: 'brokle-auth-storage',
        partialize: (state) => ({
          user: state.user,
          organization: state.organization,
          currentProject: state.currentProject,
          isAuthenticated: state.isAuthenticated,
        }),
      }
    ),
    {
      name: 'auth-store',
    }
  )
)