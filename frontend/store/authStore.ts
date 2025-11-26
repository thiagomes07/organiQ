import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'
import type { User } from '@/types'

// ============================================
// TYPES
// ============================================

interface AuthState {
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
}

interface AuthActions {
  setUser: (user: User | null) => void
  updateUser: (updates: Partial<User>) => void
  clearUser: () => void
  setLoading: (loading: boolean) => void
}

type AuthStore = AuthState & AuthActions

// ============================================
// INITIAL STATE
// ============================================

const initialState: AuthState = {
  user: null,
  isAuthenticated: false,
  isLoading: true
}

// ============================================
// STORE
// ============================================

export const useAuthStore = create<AuthStore>()(
  persist(
    (set) => ({
      ...initialState,
      
      /**
       * Define o usuário atual
       */
      setUser: (user) =>
        set({
          user,
          isAuthenticated: !!user,
          isLoading: false
        }),
      
      /**
       * Atualiza dados parciais do usuário
       */
      updateUser: (updates) =>
        set((state) => ({
          user: state.user ? { ...state.user, ...updates } : null
        })),
      
      /**
       * Limpa o estado de autenticação
       */
      clearUser: () =>
        set({
          user: null,
          isAuthenticated: false,
          isLoading: false
        }),
      
      /**
       * Define estado de loading
       */
      setLoading: (loading) =>
        set({ isLoading: loading })
    }),
    {
      name: 'organiq-auth', // Nome da chave no localStorage
      storage: createJSONStorage(() => localStorage),
      
      // Particionar o que será persistido
      partialize: (state) => ({
        user: state.user,
        isAuthenticated: state.isAuthenticated
        // isLoading NÃO é persistido
      }),
      
      // Callback após hidratar do localStorage
      onRehydrateStorage: () => (state) => {
        // Finaliza loading após hidratar
        if (state) {
          state.isLoading = false
        }
      }
    }
  )
)

// ============================================
// SELECTORS (Para uso otimizado)
// ============================================

export const selectUser = (state: AuthStore) => state.user
export const selectIsAuthenticated = (state: AuthStore) => state.isAuthenticated
export const selectIsLoading = (state: AuthStore) => state.isLoading
export const selectHasCompletedOnboarding = (state: AuthStore) => 
  state.user?.hasCompletedOnboarding ?? false
export const selectArticlesRemaining = (state: AuthStore) => 
  state.user ? state.user.maxArticles - state.user.articlesUsed : 0
export const selectCanCreateArticles = (state: AuthStore) => 
  state.user ? state.user.articlesUsed < state.user.maxArticles : false

// ============================================
// HELPER HOOKS
// ============================================

/**
 * Hook otimizado que só re-renderiza quando o usuário muda
 */
export const useUser = () => useAuthStore(selectUser)

/**
 * Hook otimizado que só re-renderiza quando isAuthenticated muda
 */
export const useIsAuthenticated = () => useAuthStore(selectIsAuthenticated)

/**
 * Hook otimizado para loading
 */
export const useAuthLoading = () => useAuthStore(selectIsLoading)

/**
 * Hook para verificar se completou onboarding
 */
export const useHasCompletedOnboarding = () => 
  useAuthStore(selectHasCompletedOnboarding)

/**
 * Hook para verificar limite de artigos
 */
export const useArticlesRemaining = () => 
  useAuthStore(selectArticlesRemaining)

/**
 * Hook para verificar se pode criar artigos
 */
export const useCanCreateArticles = () => 
  useAuthStore(selectCanCreateArticles)