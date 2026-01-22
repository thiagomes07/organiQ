import { useRouter } from 'next/navigation'
import { useMutation } from '@tanstack/react-query'
import { toast } from 'sonner'
import { useAuthStore } from '@/store/authStore'
import api, { getErrorMessage } from '@/lib/axios'
import type { LoginCredentials, RegisterData, AuthResponse } from '@/types'

// ============================================
// API FUNCTIONS
// ============================================

const authApi = {
  login: async (credentials: LoginCredentials): Promise<AuthResponse> => {
    const { data } = await api.post<AuthResponse>('/auth/login', credentials)
    return data
  },

  register: async (userData: RegisterData): Promise<AuthResponse> => {
    const { data } = await api.post<AuthResponse>('/auth/register', userData)
    return data
  },

  logout: async (): Promise<void> => {
    await api.post('/auth/logout')
  },

  refreshToken: async (): Promise<void> => {
    await api.post('/auth/refresh')
  },

  getCurrentUser: async (): Promise<AuthResponse> => {
    const { data } = await api.get<AuthResponse>('/auth/me')
    return data
  }
}

// ============================================
// HOOK
// ============================================

export function useAuth() {
  const router = useRouter()
  const { user, setUser, clearUser, isAuthenticated, isLoading } = useAuthStore()

  // ============================================
  // LOGIN MUTATION
  // ============================================

  const loginMutation = useMutation({
    mutationFn: authApi.login,
    onSuccess: (data) => {
      setUser(data.user)
      // Redirecionar baseado no status do onboarding
      if (!data.user.hasCompletedOnboarding) {
        router.push('/app/planos')
      } else {
        router.push('/app/materias')
      }
    },
    onError: (error) => {
      const message = getErrorMessage(error)
      toast.error(message || 'Erro ao fazer login')
    }
  })

  // ============================================
  // REGISTER MUTATION
  // ============================================

  const registerMutation = useMutation({
    mutationFn: authApi.register,
    onSuccess: (data) => {
      setUser(data.user)
      toast.success('Conta criada com sucesso!')

      // Sempre redireciona para planos no primeiro acesso
      router.push('/app/planos')
    },
    onError: (error) => {
      const message = getErrorMessage(error)
      toast.error(message || 'Erro ao criar conta')
    }
  })

  // ============================================
  // LOGOUT MUTATION
  // ============================================

  const logoutMutation = useMutation({
    mutationFn: authApi.logout,
    onSuccess: () => {
      clearUser()
      router.push('/login')
    },
    onError: (error) => {
      // Limpa localmente mesmo se API falhar
      clearUser()
      router.push('/login')

      const message = getErrorMessage(error)
      toast.error(message || 'Erro ao fazer logout')
    }
  })

  // ============================================
  // HELPERS
  // ============================================

  const login = (credentials: LoginCredentials) => {
    loginMutation.mutate(credentials)
  }

  const register = (userData: RegisterData) => {
    registerMutation.mutate(userData)
  }

  const logout = () => {
    logoutMutation.mutate()
  }

  // ============================================
  // REFRESH AUTH MUTATION
  // ============================================

  const refreshAuthMutation = useMutation({
    mutationFn: async () => {
      // Call refresh endpoint to get new token
      await authApi.refreshToken()
      // Then get updated user data
      return authApi.getCurrentUser()
    },
    onSuccess: (data) => {
      setUser(data.user)
    },
    onError: (error) => {
      const message = getErrorMessage(error)
      console.error('Error refreshing auth:', message)
    }
  })

  const refreshAuth = async () => {
    return refreshAuthMutation.mutateAsync()
  }

  // ============================================
  // RETURN
  // ============================================

  return {
    // State
    user,
    isAuthenticated,
    isLoading,

    // Actions
    login,
    register,
    logout,
    refreshAuth,

    // Mutation states
    isLoggingIn: loginMutation.isPending,
    isRegistering: registerMutation.isPending,
    isLoggingOut: logoutMutation.isPending,
    isRefreshing: refreshAuthMutation.isPending,

    // Helpers
    hasCompletedOnboarding: user?.hasCompletedOnboarding ?? false,
    canCreateArticles: user ? user.articlesUsed < user.maxArticles : false,
    articlesRemaining: user ? user.maxArticles - user.articlesUsed : 0
  }
}