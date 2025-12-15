import { create } from "zustand";
import type { User } from "@/types";
import api from "@/lib/axios";

// ============================================
// TYPES
// ============================================

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  isHydrated: boolean;
}

interface AuthActions {
  setUser: (user: User | null) => void;
  updateUser: (updates: Partial<User>) => void;
  clearUser: () => void;
  setLoading: (loading: boolean) => void;
  logout: () => Promise<void>;
  hydrate: () => Promise<void>;
}

type AuthStore = AuthState & AuthActions;

// ============================================
// INITIAL STATE
// ============================================

const initialState: AuthState = {
  user: null,
  isAuthenticated: false,
  isLoading: true,
  isHydrated: false,
};

// ============================================
// STORE
// ============================================

/**
 * SECURITY: This store does NOT persist to localStorage.
 * Auth state is hydrated from the server on each page load.
 * Tokens are stored in httpOnly cookies managed by the backend.
 */
export const useAuthStore = create<AuthStore>()((set, get) => ({
  ...initialState,

  /**
   * Define o usuário atual
   */
  setUser: (user) =>
    set({
      user,
      isAuthenticated: !!user,
      isLoading: false,
      isHydrated: true,
    }),

  /**
   * Atualiza dados parciais do usuário
   */
  updateUser: (updates) =>
    set((state) => ({
      user: state.user ? { ...state.user, ...updates } : null,
    })),

  /**
   * Limpa o estado de autenticação
   */
  clearUser: () =>
    set({
      user: null,
      isAuthenticated: false,
      isLoading: false,
      isHydrated: true,
    }),

  /**
   * Define estado de loading
   */
  setLoading: (loading) => set({ isLoading: loading }),

  /**
   * Faz logout (API) e limpa estado local
   */
  logout: async () => {
    try {
      await api.post("/auth/logout");
    } finally {
      set({
        user: null,
        isAuthenticated: false,
        isLoading: false,
        isHydrated: true,
      });
    }
  },

  /**
   * Hydrate auth state from server
   * Called on app initialization to validate session
   */
  hydrate: async () => {
    // Prevent multiple hydrations
    if (get().isHydrated) return;

    try {
      const { data } = await api.get("/auth/me");
      set({
        user: data.user,
        isAuthenticated: true,
        isLoading: false,
        isHydrated: true,
      });
    } catch {
      // No valid session - user is not authenticated
      set({
        user: null,
        isAuthenticated: false,
        isLoading: false,
        isHydrated: true,
      });
    }
  },
}));

// ============================================
// SELECTORS (Para uso otimizado)
// ============================================

export const selectUser = (state: AuthStore) => state.user;
export const selectIsAuthenticated = (state: AuthStore) =>
  state.isAuthenticated;
export const selectIsLoading = (state: AuthStore) => state.isLoading;
export const selectIsHydrated = (state: AuthStore) => state.isHydrated;
export const selectHasCompletedOnboarding = (state: AuthStore) =>
  state.user?.hasCompletedOnboarding ?? false;
export const selectArticlesRemaining = (state: AuthStore) =>
  state.user ? state.user.maxArticles - state.user.articlesUsed : 0;
export const selectCanCreateArticles = (state: AuthStore) =>
  state.user ? state.user.articlesUsed < state.user.maxArticles : false;

// ============================================
// HELPER HOOKS
// ============================================

/**
 * Hook otimizado que só re-renderiza quando o usuário muda
 */
export const useUser = () => useAuthStore(selectUser);

/**
 * Hook otimizado que só re-renderiza quando isAuthenticated muda
 */
export const useIsAuthenticated = () => useAuthStore(selectIsAuthenticated);

/**
 * Hook otimizado para loading
 */
export const useAuthLoading = () => useAuthStore(selectIsLoading);

/**
 * Hook para verificar se o store foi hidratado
 */
export const useIsHydrated = () => useAuthStore(selectIsHydrated);

/**
 * Hook para verificar se completou onboarding
 */
export const useHasCompletedOnboarding = () =>
  useAuthStore(selectHasCompletedOnboarding);

/**
 * Hook para verificar limite de artigos
 */
export const useArticlesRemaining = () => useAuthStore(selectArticlesRemaining);

/**
 * Hook para verificar se pode criar artigos
 */
export const useCanCreateArticles = () => useAuthStore(selectCanCreateArticles);
