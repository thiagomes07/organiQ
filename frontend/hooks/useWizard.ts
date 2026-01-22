import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import api, { getErrorMessage } from "@/lib/axios";
import { useAuthStore } from "@/store/authStore";
import { useAuth } from "@/hooks/useAuth";
import type { ArticleIdea, PublishPayload } from "@/types";
import type { BusinessInput, CompetitorsInput, IntegrationsInput } from "@/lib/validations";

// ============================================
// API FUNCTIONS
// ============================================

const wizardApi = {
  // Onboarding completo
  submitBusiness: async (data: BusinessInput): Promise<{ success: boolean }> => {
    const formData = new FormData();
    formData.append("description", data.description);
    formData.append("primaryObjective", data.primaryObjective);
    if (data.secondaryObjective)
      formData.append("secondaryObjective", data.secondaryObjective);
    formData.append("location", JSON.stringify(data.location));
    if (data.siteUrl) formData.append("siteUrl", data.siteUrl);
    formData.append("hasBlog", String(data.hasBlog));
    formData.append("blogUrls", JSON.stringify(data.blogUrls));
    formData.append("articleCount", String(data.articleCount));
    if (data.brandFile) formData.append("brandFile", data.brandFile);

    const { data: response } = await api.post("/wizard/business", formData, {
      headers: { "Content-Type": "multipart/form-data" },
    });
    return response;
  },

  submitCompetitors: async (data: CompetitorsInput): Promise<{ success: boolean }> => {
    const { data: response } = await api.post("/wizard/competitors", data);
    return response;
  },

  submitIntegrations: async (data: IntegrationsInput): Promise<{ success: boolean }> => {
    const { data: response } = await api.post("/wizard/integrations", data);
    return response;
  },

  generateIdeas: async (): Promise<{
    jobId: string;
    regenerationsRemaining?: number;
    regenerationsLimit?: number;
    nextRegenerationAt?: string;
  }> => {
    const { data } = await api.post("/wizard/generate-ideas");
    return data;
  },

  getIdeasStatus: async (
    jobId: string
  ): Promise<{
    status: "processing" | "completed" | "failed";
    ideas?: ArticleIdea[];
    error?: string;
  }> => {
    const { data } = await api.get(`/wizard/ideas-status/${jobId}`);
    return data;
  },

  publishArticles: async (
    payload: PublishPayload
  ): Promise<{ jobId: string }> => {
    const { data } = await api.post("/wizard/publish", payload);
    return data;
  },

  getPublishStatus: async (
    jobId: string
  ): Promise<{
    status: "processing" | "completed" | "failed" | "queued";
    published?: number;
    total?: number;
    errorMessage?: string;
  }> => {
    const { data } = await api.get(`/wizard/publish-status/${jobId}`);
    return data;
  },

  // Wizard simplificado (novo)
  generateNewIdeas: async (data: {
    articleCount: number;
    competitorUrls?: string[];
  }): Promise<{ jobId: string }> => {
    const { data: response } = await api.post("/articles/generate-ideas", data);
    return response;
  },

  publishNewArticles: async (
    payload: PublishPayload
  ): Promise<{ jobId: string }> => {
    const { data } = await api.post("/articles/publish", payload);
    return data;
  },

  // Buscar dados existentes do wizard
  getWizardData: async (): Promise<{
    onboardingStep: number;
    business?: {
      description: string;
      primaryObjective: string;
      secondaryObjective?: string;
      location?: { country: string; state: string; city: string; hasMultipleUnits?: boolean };
      siteUrl?: string;
      hasBlog: boolean;
      blogUrls?: string[];
    };
    competitors?: string[];
    hasIntegration: boolean;
    pendingIdeas?: Array<{
      id: string;
      title: string;
      summary: string;
      approved: boolean;
      feedback?: string;
    }>;
  }> => {
    const { data } = await api.get("/wizard/data");
    return data;
  },
};

// ============================================
// HOOK
// ============================================

export function useWizard(isOnboarding: boolean = true) {
  const router = useRouter();
  const queryClient = useQueryClient();
  const { updateUser } = useAuthStore();
  const { refreshAuth } = useAuth();

  const ideasStartedAtRef = useRef<number | null>(null);
  const publishStartedAtRef = useRef<number | null>(null);
  const onboardingCompletedRef = useRef<boolean>(false);
  const IDEAS_TIMEOUT_MS = 5 * 60 * 1000;
  const PUBLISH_TIMEOUT_MS = 5 * 60 * 1000;

  // Local state para wizard steps
  const [currentStep, setCurrentStep] = useState(1);
  const [isInitialized, setIsInitialized] = useState(false);
  const [businessData, setBusinessData] = useState<BusinessInput | null>(null);
  const [competitorData, setCompetitorData] = useState<CompetitorsInput | null>(
    null
  );
  const [integrationsData, setIntegrationsData] = useState<IntegrationsInput | null>(null);
  const [articleIdeas, setArticleIdeas] = useState<ArticleIdea[]>([]);
  const [jobId, setJobId] = useState<string | null>(null);
  const [articleCount, setArticleCount] = useState(1);

  // Estados de regeneração
  const [hasGeneratedIdeas, setHasGeneratedIdeas] = useState(false);
  const [regenerationsRemaining, setRegenerationsRemaining] = useState<number>(0);
  const [regenerationsLimit, setRegenerationsLimit] = useState<number>(0);
  const [nextRegenerationAt, setNextRegenerationAt] = useState<string | null>(null);

  // ============================================
  // FETCH EXISTING DATA
  // ============================================

  const wizardDataQuery = useQuery({
    queryKey: ["wizard-data"],
    queryFn: wizardApi.getWizardData,
    enabled: isOnboarding,
    staleTime: 0, // Sempre considerar dados como stale
    gcTime: 0, // Não cachear
    refetchOnMount: 'always', // Sempre buscar ao montar
    refetchOnWindowFocus: false,
  });

  // Inicializar states baseado nos dados existentes
  useEffect(() => {
    if (wizardDataQuery.isSuccess && wizardDataQuery.data && !isInitialized) {
      const data = wizardDataQuery.data;

      console.log('[useWizard] Inicializando com dados:', data);

      // Converter onboardingStep (1-5) para currentStep do wizard (1-4)
      const wizardStep = Math.min(Math.max(data.onboardingStep, 1), 4);
      console.log('[useWizard] Definindo currentStep para:', wizardStep);
      setCurrentStep(wizardStep);

      // Preencher dados do business se existir
      if (data.business) {
        console.log('[useWizard] Preenchendo businessData:', data.business);
        const validBlogUrls = (data.business.blogUrls || []).filter((url: string) => {
          if (!url || url === '[]' || url === '""' || url.trim() === '') return false;
          try {
            new URL(url.startsWith('http') ? url : `https://${url}`);
            return true;
          } catch {
            return false;
          }
        });

        setBusinessData({
          description: data.business.description,
          primaryObjective: data.business.primaryObjective as 'leads' | 'sales' | 'branding',
          secondaryObjective: data.business.secondaryObjective as 'leads' | 'sales' | 'branding' | undefined,
          location: data.business.location ? { ...data.business.location, hasMultipleUnits: data.business.location.hasMultipleUnits || false } : { country: 'Brasil', state: '', city: '', hasMultipleUnits: false },
          siteUrl: data.business.siteUrl || '',
          hasBlog: data.business.hasBlog && validBlogUrls.length > 0,
          blogUrls: validBlogUrls,
          articleCount: 1,
        });
      }

      // Preencher dados de competitors se existir
      if (data.competitors && data.competitors.length > 0) {
        setCompetitorData({
          competitorUrls: data.competitors,
        });
      }

      // Preencher dados de integrations
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const responseData = data as any;
      if (responseData.integrationData) {
        console.log('[useWizard] Preenchendo integrationData:', responseData.integrationData);
        setIntegrationsData({
          wordpress: {
            siteUrl: responseData.integrationData.wordpress?.siteUrl || '',
            username: responseData.integrationData.wordpress?.username || '',
            appPassword: responseData.integrationData.wordpress?.appPassword || '',
          },
          searchConsole: {
            enabled: !!responseData.integrationData.searchConsole?.propertyUrl,
            propertyUrl: responseData.integrationData.searchConsole?.propertyUrl || '',
          },
          analytics: {
            enabled: !!responseData.integrationData.analytics?.measurementId,
            measurementId: responseData.integrationData.analytics?.measurementId || '',
          }
        });
      } else if (data.hasIntegration) {
         // Fallback antigo (apenas se nao tiver integrationData mas tiver hasIntegration)
         setIntegrationsData({
           wordpress: { siteUrl: '', username: '', appPassword: '' },
           searchConsole: { enabled: false },
           analytics: { enabled: false }
         });
      }

      // Preencher ideias de artigos pendentes
      if (data.pendingIdeas && data.pendingIdeas.length > 0) {
        setArticleIdeas(data.pendingIdeas.map(idea => ({
          id: idea.id,
          title: idea.title,
          summary: idea.summary,
          approved: idea.approved,
          feedback: idea.feedback || '',
        })));
      }

      // Preencher estados de regeneração (se disponíveis no tipo retornado pela API, que atualizamos no backend)
      // Ajuste de tipagem necessário aqui pois o tipo retornado pela API no hook pode não estar atualizado
      // Vamos assumir que os dados vêm no objeto data, casting if needed
      if (typeof responseData.hasGeneratedIdeas === 'boolean') {
        setHasGeneratedIdeas(responseData.hasGeneratedIdeas);
      }
      if (typeof responseData.regenerationsRemaining === 'number') {
        setRegenerationsRemaining(responseData.regenerationsRemaining);
      }
      if (typeof responseData.regenerationsLimit === 'number') {
        setRegenerationsLimit(responseData.regenerationsLimit);
      }
      if (responseData.nextRegenerationAt) {
        setNextRegenerationAt(responseData.nextRegenerationAt);
      }


      setIsInitialized(true);
      console.log('[useWizard] Inicialização completa');
    }
  }, [wizardDataQuery.isSuccess, wizardDataQuery.data, isInitialized]);

  // ============================================
  // STEP 1: BUSINESS INFO
  // ============================================

  const businessMutation = useMutation({
    mutationFn: wizardApi.submitBusiness,
    onSuccess: (_, variables) => {
      setBusinessData(variables);
      setCurrentStep(2);
    },
    onError: (error) => {
      const message = getErrorMessage(error);
      toast.error(message || "Erro ao salvar informações");
    },
  });

  // ============================================
  // STEP 2: COMPETITORS
  // ============================================

  const competitorsMutation = useMutation({
    mutationFn: wizardApi.submitCompetitors,
    onSuccess: (_, variables) => {
      setCompetitorData(variables);
      setCurrentStep(isOnboarding ? 3 : 999);
      if (!isOnboarding) {
        generateIdeasMutation.mutate({ competitorUrls: variables.competitorUrls });
      }
    },
    onError: (error) => {
      const message = getErrorMessage(error);
      toast.error(message || "Erro ao salvar concorrentes");
    },
  });

  // ============================================
  // STEP 3: INTEGRATIONS (só no onboarding)
  // ============================================

  const integrationsMutation = useMutation({
    mutationFn: wizardApi.submitIntegrations,
    onSuccess: (_, variables) => {
      setIntegrationsData(variables);
      // Se já gerou ideias antes, vai direto para o passo de aprovação
      if (hasGeneratedIdeas) {
        console.log('[useWizard] Ideias já geradas, indo para aprovação (Step 4)');
        setCurrentStep(4);
      } else {
        // Se não, gera novas
        setCurrentStep(999); // Vai para loading
        generateIdeasMutation.mutate({ competitorUrls: competitorData?.competitorUrls });
      }
    },
    onError: (error) => {
      const message = getErrorMessage(error);
      toast.error(message || "Erro ao salvar integrações");
    },
  });

  // ============================================
  // LOADING: GENERATE IDEAS
  // ============================================

  const generateIdeasMutation = useMutation({
    mutationFn: async (variables?: { competitorUrls?: string[], isRegeneration?: boolean }) => {
      if (isOnboarding) {
        // Atualizar api helper para aceitar isRegeneration
        const { data } = await api.post("/wizard/generate-ideas", {
          isRegeneration: variables?.isRegeneration
        });
        return data as { jobId: string, regenerationsRemaining?: number, regenerationsLimit?: number, nextRegenerationAt?: string };
      }
      return wizardApi.generateNewIdeas({
        articleCount,
        competitorUrls: variables?.competitorUrls,
      });
    },
    onSuccess: (data, variables) => {
      ideasStartedAtRef.current = Date.now();
      setJobId(data.jobId);

      // Se foi regeneração, atualizar stats
      if (variables?.isRegeneration) {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const result = data as any;
        // Se a API retornou stats atualizados, usar
        if (typeof result.regenerationsRemaining === 'number') {
          setRegenerationsRemaining(result.regenerationsRemaining);
        }
        if (typeof result.regenerationsLimit === 'number') {
          setRegenerationsLimit(result.regenerationsLimit);
        }
        if (result.nextRegenerationAt) {
          setNextRegenerationAt(result.nextRegenerationAt);
        } else {
          setNextRegenerationAt(null);
        }
      }
    },
    onError: (error: unknown) => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const err = error as any; // Cast safely for now to access response
      const message = getErrorMessage(err);

      // Checar erro de limite (429)
      if (err?.response?.status === 429) {
        toast.error("Limite de regeneração excedido. Tente novamente mais tarde.");
        const responseData = err.response.data;
        if (responseData) {
          if (typeof responseData.regenerationsRemaining === 'number') setRegenerationsRemaining(responseData.regenerationsRemaining);
          if (typeof responseData.regenerationsLimit === 'number') setRegenerationsLimit(responseData.regenerationsLimit);
          if (responseData.nextRegenerationAt) setNextRegenerationAt(responseData.nextRegenerationAt);
        }
      } else {
        toast.error(message || "Erro ao gerar ideias");
      }

      setCurrentStep(isOnboarding ? 4 : 2); // Volta para step 4 (approval) se falhar regeneração, ou 2 se business
      if (isOnboarding && currentStep === 999 && !hasGeneratedIdeas) {
        // Se estava no loading inicial (step 3 -> 4) e falhou, volta para 3
        setCurrentStep(3);
      } else if (isOnboarding && currentStep === 999 && hasGeneratedIdeas) {
        // Se era regeneração (999 mas já tinha ideias, ou vindo do step 4)
        setCurrentStep(4);
      }
    },
  });

  // ============================================
  // POLLING: IDEAS STATUS
  // ============================================

  const ideasStatusQuery = useQuery({
    queryKey: ["ideas-status", jobId],
    queryFn: () => wizardApi.getIdeasStatus(jobId!),
    enabled: !!jobId && currentStep === 999,
    refetchOnMount: false,
    refetchOnWindowFocus: false,
    refetchOnReconnect: false,
    refetchInterval: (query) => {
      if (query.state.data?.status === "completed" || query.state.data?.status === "failed") {
        return false;
      }
      return 3000;
    },
  });

  // Monitorar status das ideias
  useEffect(() => {
    if (currentStep !== 999 || !ideasStatusQuery.data) return;

    const { status, ideas, error } = ideasStatusQuery.data;

    if (status === "completed") {
      console.log('[useWizard] Ideas generation completed, updating ideas');

      // O backend já retorna TODAS as ideias (aprovadas + novas)
      // Não precisamos mesclar, apenas substituir
      setArticleIdeas(ideas || []);

      setHasGeneratedIdeas(true); // Marcar que temos ideias geradas
      setCurrentStep(isOnboarding ? 4 : 3);
    } else if (status === "failed") {
      toast.error(error || "Erro ao gerar ideias");
      setCurrentStep(isOnboarding ? 4 : 2); // Se falhar regeneração, volta para aprovação
      if (isOnboarding && !hasGeneratedIdeas) setCurrentStep(3); // Se era primeira vez, volta para integrações
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentStep, ideasStatusQuery.data?.status]);

  // Monitorar timeout das ideias
  useEffect(() => {
    if (currentStep !== 999) return;

    const checkTimeout = () => {
      if (ideasStartedAtRef.current && Date.now() - ideasStartedAtRef.current > IDEAS_TIMEOUT_MS) {
        toast.error("Tempo limite ao gerar ideias. Tente novamente.");
        setCurrentStep(isOnboarding ? 4 : 2);
        if (isOnboarding && !hasGeneratedIdeas) setCurrentStep(3);
      }
    };

    const timer = setInterval(checkTimeout, 1000);
    return () => clearInterval(timer);
  }, [currentStep, isOnboarding, IDEAS_TIMEOUT_MS, hasGeneratedIdeas]);

  // ============================================
  // STEP 4: PUBLISH
  // ============================================

  const publishMutation = useMutation({
    mutationFn: isOnboarding
      ? wizardApi.publishArticles
      : wizardApi.publishNewArticles,
    onSuccess: (data) => {
      publishStartedAtRef.current = Date.now();
      setJobId(data.jobId);
      setCurrentStep(1000); // Loading de publicação
    },
    onError: (error) => {
      const message = getErrorMessage(error);
      toast.error(message || "Erro ao publicar matérias");
    },
  });

  // ============================================
  // POLLING: PUBLISH STATUS
  // ============================================

  const publishStatusQuery = useQuery({
    queryKey: ["publish-status", jobId],
    queryFn: () => wizardApi.getPublishStatus(jobId!),
    enabled: !!jobId && currentStep === 1000,
    refetchOnWindowFocus: false,
    refetchInterval: (query) => {
      if (query.state.data?.status === "completed" || query.state.data?.status === "failed") {
        return false;
      }
      return 3000;
    },
  });

  // Monitorar status da publicação
  useEffect(() => {
    if (currentStep !== 1000 || !publishStatusQuery.data) return;

    const { status, published, errorMessage } = publishStatusQuery.data;

    if (status === "completed") {
      // Only process onboarding completion once
      if (isOnboarding && !onboardingCompletedRef.current) {
        onboardingCompletedRef.current = true;
        updateUser({ hasCompletedOnboarding: true });
        // Refresh auth to get new JWT with updated onboardingStep
        refreshAuth().catch((err) => {
          console.error('Failed to refresh auth after onboarding:', err);
        });
      }
      queryClient.invalidateQueries({ queryKey: ["articles"] });
      toast.success(`${published} matérias publicadas com sucesso!`);
      router.push("/app/materias");
    } else if (status === "failed") {
      toast.error(errorMessage || "Erro ao publicar matérias");
      setCurrentStep(isOnboarding ? 4 : 3);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentStep, publishStatusQuery.data, isOnboarding, queryClient, router]);

  // Monitorar timeout da publicação
  useEffect(() => {
    if (currentStep !== 1000) return;

    const checkTimeout = () => {
      if (publishStartedAtRef.current && Date.now() - publishStartedAtRef.current > PUBLISH_TIMEOUT_MS) {
        toast.error("Tempo limite ao publicar matérias. Tente novamente.");
        setCurrentStep(isOnboarding ? 4 : 3);
      }
    };

    const timer = setInterval(checkTimeout, 1000);
    return () => clearInterval(timer);
  }, [currentStep, isOnboarding, PUBLISH_TIMEOUT_MS]);

  // ============================================
  // NAVIGATION HELPERS
  // ============================================

  const goToStep = (step: number) => {
    setCurrentStep(step);
  };

  const nextStep = () => {
    setCurrentStep((prev) => prev + 1);
  };

  const previousStep = () => {
    setCurrentStep((prev) => Math.max(1, prev - 1));
  };

  const submitBusinessInfo = (data: BusinessInput) => {
    businessMutation.mutate(data);
  };

  const submitCompetitors = (data: CompetitorsInput) => {
    if (isOnboarding) {
      competitorsMutation.mutate(data);
      return;
    }

    setCompetitorData(data);
    setCurrentStep(999);
    generateIdeasMutation.mutate({ competitorUrls: data.competitorUrls });
  };

  const submitIntegrations = (data: IntegrationsInput) => {
    integrationsMutation.mutate(data);
  };

  const publishArticles = (payload: PublishPayload) => {
    publishMutation.mutate(payload);
  };

  const regenerateIdeas = () => {
    setCurrentStep(999);
    generateIdeasMutation.mutate({ isRegeneration: true });
  };

  const updateArticleIdea = (id: string, updates: Partial<ArticleIdea>) => {
    setArticleIdeas((prev) =>
      prev.map((idea) => (idea.id === id ? { ...idea, ...updates } : idea))
    );
  };

  // ============================================
  // RETURN
  // ============================================

  const isGeneratingIdeas = generateIdeasMutation.isPending || ideasStatusQuery.isFetching;
  const isPublishing = publishMutation.isPending || publishStatusQuery.isFetching;

  return {
    // Current state
    currentStep,
    businessData,
    competitorData,
    integrationsData,
    articleIdeas,
    articleCount,
    isInitialized,

    // Regeneration State
    hasGeneratedIdeas,
    regenerationsRemaining,
    regenerationsLimit,
    nextRegenerationAt,

    // Navigation
    goToStep,
    nextStep,
    previousStep,

    // Actions
    submitBusinessInfo,
    submitCompetitors,
    submitIntegrations,
    publishArticles,
    regenerateIdeas,
    updateArticleIdea,
    setArticleCount,

    // Loading states
    isLoadingWizardData: wizardDataQuery.isLoading,
    isSubmittingBusiness: businessMutation.isPending,
    isSubmittingCompetitors: competitorsMutation.isPending,
    isSubmittingIntegrations: integrationsMutation.isPending,
    isGeneratingIdeas,
    isPublishing,

    // Progress info
    ideasProgress: ideasStatusQuery.data?.status,
    publishProgress: publishStatusQuery.data,

    // Computed
    approvedCount: articleIdeas.filter((idea) => idea.approved).length,
    allApproved: articleIdeas.length > 0 && articleIdeas.every((idea) => idea.approved),
    canPublish: articleIdeas.some((idea) => idea.approved),
    canRegenerateIdeas: regenerationsRemaining > 0 && !isGeneratingIdeas && !isPublishing && !(articleIdeas.length > 0 && articleIdeas.every((idea) => idea.approved)),
    isLoading: currentStep === 999 || currentStep === 1000,
  };
}
