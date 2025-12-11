import { useQuery, useMutation } from '@tanstack/react-query'
import { useRouter } from 'next/navigation'
import { toast } from 'sonner'
import api, { getErrorMessage } from '@/lib/axios'
import { useAuthStore } from '@/store/authStore'
import type { 
  Plan, 
  PlanInfo, 
  CheckoutResponse, 
  PaymentStatus 
} from '@/types'

// ============================================
// API FUNCTIONS
// ============================================

const plansApi = {
  getPlans: async (): Promise<Plan[]> => {
    const { data } = await api.get<Plan[]>('/plans')
    return data
  },

  getCurrentPlan: async (): Promise<PlanInfo> => {
    const { data } = await api.get<PlanInfo>('/account/plan')
    return data
  },

  createCheckout: async (planId: string): Promise<CheckoutResponse> => {
    const { data } = await api.post<CheckoutResponse>('/payments/create-checkout', { planId })
    return data
  },

  getPaymentStatus: async (sessionId: string): Promise<PaymentStatus> => {
    const { data } = await api.get<PaymentStatus>(`/payments/status/${sessionId}`)
    return data
  },

  createPortalSession: async (): Promise<{ url: string }> => {
    const { data } = await api.post<{ url: string }>('/payments/create-portal-session')
    return data
  }
}

// ============================================
// QUERY KEYS
// ============================================

const planKeys = {
  all: ['plans'] as const,
  lists: () => [...planKeys.all, 'list'] as const,
  current: () => [...planKeys.all, 'current'] as const,
  payment: (sessionId: string) => [...planKeys.all, 'payment', sessionId] as const
}

// ============================================
// HOOK
// ============================================

export function usePlans() {
  const router = useRouter()
  const { updateUser } = useAuthStore()

  // ============================================
  // GET PLANS QUERY
  // ============================================

  const plansQuery = useQuery({
    queryKey: planKeys.lists(),
    queryFn: plansApi.getPlans,
    staleTime: Infinity // Planos raramente mudam
  })

  // ============================================
  // GET CURRENT PLAN QUERY
  // ============================================

  const currentPlanQuery = useQuery({
    queryKey: planKeys.current(),
    queryFn: plansApi.getCurrentPlan,
    staleTime: 60000 // 1 minuto
  })

  // ============================================
  // CREATE CHECKOUT MUTATION
  // ============================================

  const checkoutMutation = useMutation({
    mutationFn: plansApi.createCheckout,
    onSuccess: (data) => {
      // Redirecionar para checkout
      window.location.href = data.checkoutUrl
    },
    onError: (error) => {
      const message = getErrorMessage(error)
      toast.error(message || 'Erro ao criar checkout')
    }
  })

  // ============================================
  // CREATE PORTAL MUTATION
  // ============================================

  const portalMutation = useMutation({
    mutationFn: plansApi.createPortalSession,
    onSuccess: (data) => {
      // Redirecionar para portal
      window.location.href = data.url
    },
    onError: (error) => {
      const message = getErrorMessage(error)
      toast.error(message || 'Erro ao abrir portal de pagamentos')
    }
  })

  // ============================================
  // PAYMENT STATUS POLLING
  // ============================================

  const usePaymentStatus = (sessionId: string | null, enabled: boolean = true) => {
    return useQuery({
      queryKey: planKeys.payment(sessionId || ''),
      queryFn: () => plansApi.getPaymentStatus(sessionId!),
      enabled: enabled && !!sessionId,
      refetchInterval: (query) => {
        // Parar polling se status for 'paid' ou 'failed'
        if (query.state.data?.status === 'paid' || query.state.data?.status === 'failed') {
          return false
        }
        return 3000 // Poll a cada 3 segundos
      },
      refetchOnWindowFocus: false
    })
  }

  // ============================================
  // HELPERS
  // ============================================

  const selectPlan = (planId: string) => {
    checkoutMutation.mutate(planId)
  }

  const openPortal = () => {
    portalMutation.mutate()
  }

  const getPlanById = (planId: string) => {
    return plansQuery.data?.find((plan) => plan.id === planId)
  }

  const getRecommendedPlan = () => {
    return plansQuery.data?.find((plan) => plan.recommended)
  }

  const isCurrentPlan = (planId: string) => {
    return currentPlanQuery.data?.name === getPlanById(planId)?.name
  }

  const canUpgrade = (targetPlanId: string) => {
    const currentPlan = plansQuery.data?.find(
      (plan) => plan.name === currentPlanQuery.data?.name
    )
    const targetPlan = getPlanById(targetPlanId)
    
    if (!currentPlan || !targetPlan) return false
    
    return targetPlan.maxArticles > currentPlan.maxArticles
  }

  // ============================================
  // RETURN
  // ============================================

  return {
    // Data
    plans: plansQuery.data || [],
    currentPlan: currentPlanQuery.data,
    
    // States
    isLoadingPlans: plansQuery.isLoading,
    isLoadingCurrentPlan: currentPlanQuery.isLoading,
    isError: plansQuery.isError || currentPlanQuery.isError,
    error: plansQuery.error || currentPlanQuery.error,
    
    // Actions
    selectPlan,
    openPortal,
    refetchCurrentPlan: currentPlanQuery.refetch,
    
    // Mutation states
    isCreatingCheckout: checkoutMutation.isPending,
    isOpeningPortal: portalMutation.isPending,
    
    // Helpers
    getPlanById,
    getRecommendedPlan,
    isCurrentPlan,
    canUpgrade,
    
    // Payment status hook
    usePaymentStatus
  }
}