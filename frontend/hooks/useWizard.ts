import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import api, { getErrorMessage } from "@/lib/axios";
import { useAuthStore } from "@/store/authStore";
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

  generateIdeas: async (): Promise<{ jobId: string }> => {
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
    status: "processing" | "completed" | "failed";
    published?: number;
    total?: number;
    error?: string;
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

  const ideasStartedAtRef = useRef<number | null>(null);
  const publishStartedAtRef = useRef<number | null>(null);
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
      // onboarding_step 1 = precisa preencher business (wizard step 1)
      // onboarding_step 2 = business feito, preencher competitors (wizard step 2)
      // onboarding_step 3 = competitors feito, preencher integrations (wizard step 3)
      // onboarding_step 4 = integrations feito, aprovar artigos (wizard step 4)
      const wizardStep = Math.min(Math.max(data.onboardingStep, 1), 4);
      console.log('[useWizard] Definindo currentStep para:', wizardStep);
      setCurrentStep(wizardStep);

      // Preencher dados do business se existir
      if (data.business) {
        console.log('[useWizard] Preenchendo businessData:', data.business);
        // Filtrar blogUrls inválidas (backend pode retornar strings como "[]")
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
          hasBlog: data.business.hasBlog && validBlogUrls.length > 0, // Se não tem URLs válidas, não tem blog
          blogUrls: validBlogUrls,
          articleCount: 1, // Default para 1 (mínimo) - não é persistido no banco
        });
      }

      // Preencher dados de competitors se existir
      if (data.competitors && data.competitors.length > 0) {
        console.log('[useWizard] Preenchendo competitorData:', data.competitors);
        setCompetitorData({
          competitorUrls: data.competitors,
        });
      }

      // Preencher dados de integrations se existir
      if (data.hasIntegration) {
        console.log('[useWizard] Preenchendo integrationsData');
        setIntegrationsData({
          wordpress: { siteUrl: '', username: '', appPassword: '' }, // Dados sensiveis nao sao retornados
        });
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
      setCurrentStep(isOnboarding ? 3 : 999); // Se não é onboarding, pula para loading
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
      setCurrentStep(999); // Vai para loading
      generateIdeasMutation.mutate({ competitorUrls: competitorData?.competitorUrls });
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
    mutationFn: async (variables?: { competitorUrls?: string[] }) => {
      if (isOnboarding) {
        return wizardApi.generateIdeas();
      }
      return wizardApi.generateNewIdeas({
        articleCount,
        competitorUrls: variables?.competitorUrls,
      });
    },
    onSuccess: (data) => {
      ideasStartedAtRef.current = Date.now();
      setJobId(data.jobId);
    },
    onError: (error) => {
      const message = getErrorMessage(error);
      toast.error(message || "Erro ao gerar ideias");
      setCurrentStep(isOnboarding ? 3 : 2); // Volta para step anterior
    },
  });

  // ============================================
  // POLLING: IDEAS STATUS
  // ============================================

  const ideasStatusQuery = useQuery({
    queryKey: ["ideas-status", jobId],
    queryFn: () => wizardApi.getIdeasStatus(jobId!),
    enabled: !!jobId && currentStep === 999,
    refetchInterval: (query) => {
      if (ideasStartedAtRef.current && Date.now() - ideasStartedAtRef.current > IDEAS_TIMEOUT_MS) {
        toast.error("Tempo limite ao gerar ideias. Tente novamente.");
        setCurrentStep(isOnboarding ? 3 : 2);
        return false;
      }
      if (query.state.data?.status === "completed") {
        setArticleIdeas(query.state.data.ideas || []);
        setCurrentStep(isOnboarding ? 4 : 3); // Vai para aprovação
        return false;
      }
      if (query.state.data?.status === "failed") {
        toast.error(query.state.data.error || "Erro ao gerar ideias");
        setCurrentStep(isOnboarding ? 3 : 2);
        return false;
      }
      return 3000; // Poll a cada 3 segundos
    },
    refetchOnWindowFocus: false,
  });

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
    refetchInterval: (query) => {
      if (publishStartedAtRef.current && Date.now() - publishStartedAtRef.current > PUBLISH_TIMEOUT_MS) {
        toast.error("Tempo limite ao publicar matérias. Tente novamente.");
        setCurrentStep(isOnboarding ? 4 : 3);
        return false;
      }
      if (query.state.data?.status === "completed") {
        // Atualizar usuário
        if (isOnboarding) {
          updateUser({ hasCompletedOnboarding: true });
        }

        // Invalidar cache de artigos
        queryClient.invalidateQueries({ queryKey: ["articles"] });

        toast.success(
          `${query.state.data.published} matérias publicadas com sucesso!`
        );
        router.push("/app/materias");
        return false;
      }
      if (query.state.data?.status === "failed") {
        toast.error(query.state.data.error || "Erro ao publicar matérias");
        setCurrentStep(isOnboarding ? 4 : 3);
        return false;
      }
      return 3000;
    },
    refetchOnWindowFocus: false,
  });

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

  const updateArticleIdea = (id: string, updates: Partial<ArticleIdea>) => {
    setArticleIdeas((prev) =>
      prev.map((idea) => (idea.id === id ? { ...idea, ...updates } : idea))
    );
  };

  // ============================================
  // RETURN
  // ============================================

  return {
    // Current state
    currentStep,
    businessData,
    competitorData,
    integrationsData,
    articleIdeas,
    articleCount,
    isInitialized,

    // Navigation
    goToStep,
    nextStep,
    previousStep,

    // Actions
    submitBusinessInfo,
    submitCompetitors,
    submitIntegrations,
    publishArticles,
    updateArticleIdea,
    setArticleCount,

    // Loading states
    isLoadingWizardData: wizardDataQuery.isLoading,
    isSubmittingBusiness: businessMutation.isPending,
    isSubmittingCompetitors: competitorsMutation.isPending,
    isSubmittingIntegrations: integrationsMutation.isPending,
    isGeneratingIdeas:
      generateIdeasMutation.isPending || ideasStatusQuery.isFetching,
    isPublishing: publishMutation.isPending || publishStatusQuery.isFetching,

    // Progress info
    ideasProgress: ideasStatusQuery.data?.status,
    publishProgress: publishStatusQuery.data,

    // Computed
    approvedCount: articleIdeas.filter((idea) => idea.approved).length,
    canPublish: articleIdeas.some((idea) => idea.approved),
    isLoading: currentStep === 999 || currentStep === 1000,
  };
}
