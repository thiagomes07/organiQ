import { useQuery, useMutation } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import api, { getErrorMessage } from "@/lib/axios";
import axios from "axios";
import { useHasCompletedOnboarding } from "@/store/authStore";
import type { Plan, AccountPlanResponse, CheckoutResponse, PaymentStatus } from "@/types";

// ============================================
// API FUNCTIONS
// ============================================

const plansApi = {
  getPlans: async (): Promise<Plan[]> => {
    const { data } = await api.get<Plan[]>("/plans");
    return data;
  },

  getCurrentPlan: async (): Promise<AccountPlanResponse> => {
    const { data } = await api.get<{ plan: AccountPlanResponse }>("/account/plan");
    return data.plan;
  },

  createCheckout: async (planId: string): Promise<CheckoutResponse> => {
    const { data } = await api.post<CheckoutResponse>(
      "/payments/create-checkout",
      { planId }
    );
    return data;
  },

  getPaymentStatus: async (sessionId: string): Promise<PaymentStatus> => {
    const { data } = await api.get<PaymentStatus>(
      `/payments/status/${sessionId}`
    );
    return data;
  },

  createPortalSession: async (): Promise<{ url: string }> => {
    const { data } = await api.post<{ url: string }>(
      "/payments/create-portal-session"
    );
    return data;
  },
};

// ============================================
// QUERY KEYS
// ============================================

const planKeys = {
  all: ["plans"] as const,
  lists: () => [...planKeys.all, "list"] as const,
  current: () => [...planKeys.all, "current"] as const,
  payment: (sessionId: string) =>
    [...planKeys.all, "payment", sessionId] as const,
};

// ============================================
// HOOK
// ============================================

export function usePlans() {
  const router = useRouter();
  const hasCompletedOnboarding = useHasCompletedOnboarding();

  // ============================================
  // GET PLANS QUERY
  // ============================================

  const plansQuery = useQuery({
    queryKey: planKeys.lists(),
    queryFn: plansApi.getPlans,
    staleTime: Infinity, // Planos raramente mudam
  });

  // ============================================
  // GET CURRENT PLAN QUERY
  // ============================================

  const currentPlanQuery = useQuery({
    queryKey: planKeys.current(),
    queryFn: plansApi.getCurrentPlan,
    staleTime: 60000, // 1 minuto
  });

  // ============================================
  // CREATE CHECKOUT MUTATION
  // ============================================

  const checkoutMutation = useMutation({
    mutationFn: plansApi.createCheckout,
    onSuccess: (data) => {
      console.log('Checkout created, redirecting to checkoutUrl:', data.checkoutUrl);
      // Use a tiny delay to avoid potential click/form side-effects
      setTimeout(() => {
        window.location.href = data.checkoutUrl;
      }, 50);
    },
    onError: (error) => {
      console.error('Checkout error:', error);
      // Se backend indicou que o usuário já possui o plano, redireciona
      if (axios.isAxiosError(error) && (error as any).response?.data?.error === "user_already_has_plan") {
          const target = hasCompletedOnboarding ? "/app" : "/app/onboarding";
        console.log('User already has plan, redirecting (internal) to:', target);
        try {
          router.replace(target);
        } catch (e) {
          console.warn('router.replace failed in onError, will fallback', e);
        }
        setTimeout(() => {
          if (window.location.pathname !== target) {
            console.log('router did not navigate (onError), forcing full navigation to', target);
            window.location.href = target;
          } else {
            console.log('router.replace succeeded (onError) to', target);
          }
        }, 300);
        return
      }

      const message = getErrorMessage(error);
      toast.error(message || "Erro ao criar checkout");
    },
  });

  // ============================================
  // CREATE PORTAL MUTATION
  // ============================================

  const portalMutation = useMutation({
    mutationFn: plansApi.createPortalSession,
    onSuccess: (data) => {
      console.log('Portal session created, redirecting to:', data.url);
      setTimeout(() => {
        window.location.href = data.url;
      }, 50);
    },
    onError: (error) => {
      const message = getErrorMessage(error);
      toast.error(message || "Erro ao abrir portal de pagamentos");
    },
  });

  // ============================================
  // PAYMENT STATUS POLLING
  // ============================================

  const usePaymentStatus = (
    sessionId: string | null,
    enabled: boolean = true
  ) => {
    return useQuery({
      queryKey: planKeys.payment(sessionId || ""),
      queryFn: () => plansApi.getPaymentStatus(sessionId!),
      enabled: enabled && !!sessionId,
      refetchInterval: (query) => {
        // Parar polling se status for 'paid' ou 'failed'
        if (
          query.state.data?.status === "paid" ||
          query.state.data?.status === "failed"
        ) {
          return false;
        }
        return 3000; // Poll a cada 3 segundos
      },
      refetchOnWindowFocus: false,
    });
  };

  // ============================================
  // HELPERS
  // ============================================

  const selectPlan = async (planId: string) => {
    console.log('selectPlan called with planId:', planId);
    console.log('hasCompletedOnboarding:', hasCompletedOnboarding);
    
    // Always fetch current plan to ensure we have latest data
    let currentPlan;
    try {
      const result = await currentPlanQuery.refetch();
      currentPlan = result.data;
      console.log('currentPlan:', currentPlan);
    } catch (error) {
      console.error('Failed to fetch current plan:', error);
      // Proceed to checkout anyway, let backend handle it
    }

    // Se já é o plano atual, avançar no fluxo de onboarding/dashboard
    if (currentPlan?.id === planId) {
      const target = hasCompletedOnboarding ? "/app" : "/app/onboarding";
      console.log('Redirecting (internal) to:', target);
      try {
        router.replace(target);
      } catch (e) {
        console.warn('router.replace failed, will fallback to full navigation', e);
      }
      setTimeout(() => {
        if (window.location.pathname !== target) {
          console.log('router did not navigate, forcing full navigation to', target);
          window.location.href = target;
        } else {
          console.log('router.replace succeeded to', target);
        }
      }, 300);
      return;
    }

    console.log('Creating checkout for planId:', planId);
    // Caso contrário, iniciar checkout
    checkoutMutation.mutate(planId);
  };

  const openPortal = () => {
    portalMutation.mutate();
  };

  const getPlanById = (planId: string) => {
    return plansQuery.data?.find((plan) => plan.id === planId);
  };

  const getRecommendedPlan = () => {
    return plansQuery.data?.find((plan) => plan.recommended);
  };

  const isCurrentPlan = (planId: string) => {
    return currentPlanQuery.data?.id === planId;
  };

  const canUpgrade = (targetPlanId: string) => {
    const currentPlan = plansQuery.data?.find(
      (plan) => plan.name === currentPlanQuery.data?.name
    );
    const targetPlan = getPlanById(targetPlanId);

    if (!currentPlan || !targetPlan) return false;

    return targetPlan.maxArticles > currentPlan.maxArticles;
  };

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
    usePaymentStatus,
  };
}
