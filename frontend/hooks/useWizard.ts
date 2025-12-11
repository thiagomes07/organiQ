import { useState } from "react";
import { useRouter } from "next/navigation";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import api, { getErrorMessage } from "@/lib/axios";
import { useAuthStore } from "@/store/authStore";
import type {
  BusinessInfo,
  CompetitorData,
  IntegrationsData,
  ArticleIdea,
  PublishPayload,
} from "@/types";

// ============================================
// API FUNCTIONS
// ============================================

const wizardApi = {
  // Onboarding completo
  submitBusiness: async (data: BusinessInfo): Promise<{ success: boolean }> => {
    const formData = new FormData();
    formData.append("description", data.description);
    formData.append("primaryObjective", data.primaryObjective);
    if (data.secondaryObjective)
      formData.append("secondaryObjective", data.secondaryObjective);
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

  submitCompetitors: async (
    data: CompetitorData
  ): Promise<{ success: boolean }> => {
    const { data: response } = await api.post("/wizard/competitors", data);
    return response;
  },

  submitIntegrations: async (
    data: IntegrationsData
  ): Promise<{ success: boolean }> => {
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
};

// ============================================
// HOOK
// ============================================

export function useWizard(isOnboarding: boolean = true) {
  const router = useRouter();
  const queryClient = useQueryClient();
  const { updateUser } = useAuthStore();

  // Local state para wizard steps
  const [currentStep, setCurrentStep] = useState(1);
  const [businessData, setBusinessData] = useState<BusinessInfo | null>(null);
  const [competitorData, setCompetitorData] = useState<CompetitorData | null>(
    null
  );
  const [integrationsData, setIntegrationsData] =
    useState<IntegrationsData | null>(null);
  const [articleIdeas, setArticleIdeas] = useState<ArticleIdea[]>([]);
  const [jobId, setJobId] = useState<string | null>(null);

  // ============================================
  // STEP 1: BUSINESS INFO
  // ============================================

  const businessMutation = useMutation({
    mutationFn: wizardApi.submitBusiness,
    onSuccess: (_, variables) => {
      setBusinessData(variables);
      setCurrentStep(2);
      toast.success("Informações salvas!");
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
        generateIdeasMutation.mutate();
      } else {
        toast.success("Concorrentes salvos!");
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
      generateIdeasMutation.mutate();
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
    mutationFn: isOnboarding
      ? wizardApi.generateIdeas
      : () =>
          wizardApi.generateNewIdeas({
            articleCount: businessData?.articleCount || 1,
            competitorUrls: competitorData?.competitorUrls,
          }),
    onSuccess: (data) => {
      setJobId(data.jobId);
      // O polling vai começar automaticamente via useQuery
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

  const submitBusinessInfo = (data: BusinessInfo) => {
    businessMutation.mutate(data);
  };

  const submitCompetitors = (data: CompetitorData) => {
    competitorsMutation.mutate(data);
  };

  const submitIntegrations = (data: IntegrationsData) => {
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

    // Loading states
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
