"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useWizard } from "@/hooks/useWizard";
import { useUser } from "@/store/authStore";
import { StepIndicator } from "./StepIndicator";
import { CompetitorsForm } from "@/components/forms/CompetitorsForm";
import { ArticleIdeaCard } from "@/components/articles/ArticleIdeaCard";
import { LoadingOverlay } from "@/components/shared/LoadingSpinner";
import { EmptyIdeas } from "@/components/shared/EmptyState";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Slider } from "@/components/ui/slider";
import { Label } from "@/components/ui/label";
import { AlertCircle, MessageSquare } from "lucide-react";
import Link from "next/link";
import type { CompetitorsInput, PublishPayload } from "@/lib/validations";

const steps = [
  { number: 1, label: "Quantidade" },
  { number: 2, label: "Concorrentes" },
  { number: 3, label: "Aprovação" },
];

const loadingMessagesGenerate = [
  "Analisando seus concorrentes...",
  "Mapeando tópicos de autoridade...",
  "Gerando ideias de matérias...",
  "Isso pode levar alguns minutos",
];

const loadingMessagesPublish = [
  "Escrevendo matérias...",
  "Otimizando SEO...",
  "Publicando no WordPress...",
  "Aguarde, estamos finalizando",
];

/**
 * Wizard Simplificado para Gerar Novas Matérias
 *
 * Usado após o onboarding completo para criar matérias adicionais
 */
export function NewArticlesWizard() {
  const router = useRouter();
  const user = useUser();

  const {
    currentStep,
    businessData,
    competitorData,
    articleIdeas,
    submitBusinessInfo,
    submitCompetitors,
    publishArticles,
    updateArticleIdea,
    previousStep,
    isSubmittingBusiness,
    isSubmittingCompetitors,
    isGeneratingIdeas,
    isPublishing,
    approvedCount,
    canPublish,
  } = useWizard(false); // false = não é onboarding

  const [articleCount, setArticleCount] = useState(1);
  const articlesRemaining = user ? user.maxArticles - user.articlesUsed : 0;
  const canCreate = articlesRemaining > 0;

  // ============================================
  // LOADING STATES
  // ============================================

  // Loading: Gerando ideias
  if (currentStep === 999 || isGeneratingIdeas) {
    return <LoadingOverlay messages={loadingMessagesGenerate} />;
  }

  // Loading: Publicando
  if (currentStep === 1000 || isPublishing) {
    return <LoadingOverlay messages={loadingMessagesPublish} />;
  }

  // ============================================
  // STEP 1: QUANTIDADE
  // ============================================

  const renderStepQuantity = () => (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        submitBusinessInfo({
          description: "", // Dados já existem do onboarding
          primaryObjective: "leads",
          hasBlog: false,
          blogUrls: [],
          articleCount,
        } as any);
      }}
      className="space-y-6"
    >
      {/* Alerta de Limite */}
      {!canCreate && (
        <div className="bg-[var(--color-warning)]/10 border border-[var(--color-warning)] rounded-[var(--radius-md)] p-4 flex items-start gap-3">
          <AlertCircle className="h-5 w-5 text-[var(--color-warning)] mt-0.5" />
          <div className="flex-1">
            <p className="font-medium font-onest text-[var(--color-primary-dark)]">
              Limite de matérias atingido
            </p>
            <p className="text-sm font-onest text-[var(--color-primary-dark)]/70 mt-1">
              Você já utilizou todas as {user?.maxArticles} matérias do seu
              plano este mês.
            </p>
            <Link href="/app/conta">
              <Button variant="outline" size="sm" className="mt-3">
                Fazer Upgrade
              </Button>
            </Link>
          </div>
        </div>
      )}

      {canCreate && (
        <>
          {/* Slider de Quantidade */}
          <div className="space-y-2">
            <Label required>Quantas matérias deseja criar?</Label>
            <Slider
              min={1}
              max={articlesRemaining}
              step={1}
              value={[articleCount]}
              onValueChange={(value) => setArticleCount(value[0])}
              showValue
              formatValue={(value) =>
                `${value} ${value === 1 ? "matéria" : "matérias"}`
              }
            />
            <p className="text-xs text-[var(--color-primary-dark)]/60 font-onest">
              Você tem {articlesRemaining}{" "}
              {articlesRemaining === 1 ? "matéria" : "matérias"} disponível
              {articlesRemaining === 1 ? "" : "is"} este mês
            </p>
          </div>

          {/* Info Box */}
          <div className="bg-[var(--color-primary-purple)]/5 border border-[var(--color-primary-purple)]/20 rounded-[var(--radius-md)] p-4">
            <p className="text-sm font-onest text-[var(--color-primary-dark)]/80">
              💡 <strong>Dica:</strong> Você pode gerar várias matérias de uma
              vez para economizar tempo.
            </p>
          </div>

          {/* Submit Button */}
          <div className="flex justify-end pt-4">
            <Button
              type="submit"
              variant="secondary"
              size="lg"
              isLoading={isSubmittingBusiness}
              disabled={isSubmittingBusiness}
            >
              Próximo
            </Button>
          </div>
        </>
      )}
    </form>
  );

  // ============================================
  // STEP 2: CONCORRENTES
  // ============================================

  const renderStepCompetitors = () => (
    <CompetitorsForm
      onSubmit={(data: CompetitorsInput) => submitCompetitors(data as any)}
      onBack={previousStep}
      isLoading={isSubmittingCompetitors}
      defaultValues={competitorData || undefined}
    />
  );

  // ============================================
  // STEP 3: APROVAÇÃO
  // ============================================

  const renderStepApproval = () => {
    const feedbackCount = articleIdeas.filter(
      (idea) => idea.feedback && idea.feedback.length > 0
    ).length;

    const handlePublish = () => {
      const approvedArticles = articleIdeas
        .filter((idea) => idea.approved)
        .map((idea) => ({
          id: idea.id,
          feedback: idea.feedback,
        }));

      publishArticles({
        articles: approvedArticles,
      } as PublishPayload);
    };

    return (
      <div className="space-y-6">
        {/* Header Info */}
        <div className="bg-[var(--color-primary-purple)]/5 border border-[var(--color-primary-purple)]/20 rounded-[var(--radius-md)] p-4">
          <p className="text-sm font-onest text-[var(--color-primary-dark)]/80">
            Revise as ideias geradas e aprove as que deseja publicar. Você pode
            adicionar feedbacks para direcionar o conteúdo.
          </p>
        </div>

        {/* Empty State */}
        {articleIdeas.length === 0 && (
          <EmptyIdeas onRegenerate={() => router.push("/app/novo")} />
        )}

        {/* Grid de Ideias */}
        {articleIdeas.length > 0 && (
          <div className="grid md:grid-cols-2 gap-4">
            {articleIdeas.map((idea) => (
              <ArticleIdeaCard
                key={idea.id}
                idea={idea}
                onUpdate={updateArticleIdea}
              />
            ))}
          </div>
        )}

        {/* Footer com Contador e Ação */}
        {articleIdeas.length > 0 && (
          <div className="border-t border-[var(--color-border)] pt-6 space-y-4">
            {/* Contadores */}
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium font-onest text-[var(--color-primary-dark)]">
                    {approvedCount}{" "}
                    {approvedCount === 1
                      ? "matéria aprovada"
                      : "matérias aprovadas"}
                  </span>
                </div>
                {feedbackCount > 0 && (
                  <div className="flex items-center gap-2 text-[var(--color-primary-purple)]">
                    <MessageSquare className="h-4 w-4" />
                    <span className="text-sm font-medium font-onest">
                      {feedbackCount}{" "}
                      {feedbackCount === 1
                        ? "feedback adicionado"
                        : "feedbacks adicionados"}
                    </span>
                  </div>
                )}
              </div>
            </div>

            {/* Botões de Ação */}
            <div className="flex items-center justify-between gap-4">
              <Button variant="outline" onClick={previousStep}>
                Voltar
              </Button>

              <Button
                variant="primary"
                size="lg"
                onClick={handlePublish}
                disabled={!canPublish}
                title={
                  !canPublish ? "Aprove pelo menos uma matéria" : undefined
                }
              >
                Publicar{" "}
                {approvedCount > 0
                  ? `${approvedCount} ${
                      approvedCount === 1 ? "Matéria" : "Matérias"
                    }`
                  : "Matérias"}
              </Button>
            </div>
          </div>
        )}
      </div>
    );
  };

  // ============================================
  // RENDER
  // ============================================

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="text-center space-y-2">
        <h1 className="text-3xl font-bold font-all-round text-[var(--color-primary-dark)]">
          Gerar Novas Matérias
        </h1>
        <p className="text-lg font-onest text-[var(--color-primary-dark)]/70">
          Crie mais conteúdo otimizado para seu blog
        </p>
      </div>

      {canCreate && (
        <>
          {/* Step Indicator */}
          <StepIndicator currentStep={currentStep} steps={steps} />

          {/* Form Card */}
          <Card>
            <CardHeader>
              <CardTitle>
                {currentStep === 1 && "Quantidade de Matérias"}
                {currentStep === 2 && "Análise de Concorrentes"}
                {currentStep === 3 && "Aprovação de Ideias"}
              </CardTitle>
              <CardDescription>
                {currentStep === 1 &&
                  `Você tem ${articlesRemaining} matérias disponíveis este mês`}
                {currentStep === 2 &&
                  "Adicione URLs de concorrentes para melhorar a estratégia (opcional)"}
                {currentStep === 3 &&
                  "Revise e aprove as matérias que deseja publicar"}
              </CardDescription>
            </CardHeader>

            <CardContent>
              {currentStep === 1 && renderStepQuantity()}
              {currentStep === 2 && renderStepCompetitors()}
              {currentStep === 3 && renderStepApproval()}
            </CardContent>
          </Card>

          {/* Progress Info */}
          <div className="text-center">
            <p className="text-sm font-onest text-[var(--color-primary-dark)]/60">
              Passo {currentStep} de {steps.length}
            </p>
          </div>
        </>
      )}
    </div>
  );
}
