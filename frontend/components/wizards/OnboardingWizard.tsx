'use client'

import { useWizard } from '@/hooks/useWizard'
import { StepIndicator } from './StepIndicator'
import { BusinessInfoForm } from '@/components/forms/BusinessInfoForm'
import { CompetitorsForm } from '@/components/forms/CompetitorsForm'
import { IntegrationsForm } from '@/components/forms/IntegrationsForm'
import { ArticleIdeaCard } from '@/components/articles/ArticleIdeaCard'
import { LoadingOverlay } from '@/components/shared/LoadingSpinner'
import { Card, CardContent, CardHeader, CardTitle, CardFooter } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { MessageSquare, FileCheck } from 'lucide-react'
import type { BusinessInput, CompetitorsInput, IntegrationsInput } from '@/lib/validations'

const steps = [
  { number: 1, label: 'Negócio' },
  { number: 2, label: 'Concorrentes' },
  { number: 3, label: 'Integrações' },
  { number: 4, label: 'Aprovação' },
]

const loadingMessages = [
  'Analisando seus concorrentes...',
  'Mapeando tópicos de autoridade...',
  'Gerando ideias de matérias...',
  'Isso pode levar alguns minutos',
]

const publishingMessages = [
  'Escrevendo suas matérias com IA...',
  'Otimizando conteúdo para SEO...',
  'Publicando no WordPress...',
  'Quase lá...',
]

export function OnboardingWizard() {
  const {
    currentStep,
    businessData,
    competitorData,
    integrationsData,
    articleIdeas,
    submitBusinessInfo,
    submitCompetitors,
    submitIntegrations,
    publishArticles,
    updateArticleIdea,
    previousStep,
    isSubmittingBusiness,
    isSubmittingCompetitors,
    isSubmittingIntegrations,
    isGeneratingIdeas,
    isPublishing,
    approvedCount,
    canPublish,
  } = useWizard(true) // true = isOnboarding

  // Loading state para geração de ideias
  if (currentStep === 999 || isGeneratingIdeas) {
    return <LoadingOverlay messages={loadingMessages} />
  }

  // Loading state para publicação
  if (currentStep === 1000 || isPublishing) {
    return <LoadingOverlay messages={publishingMessages} />
  }

  // Step 4: Approval
  if (currentStep === 4) {
    const feedbackCount = articleIdeas.filter((idea) => idea.feedback && idea.feedback.trim().length > 0).length

    const handlePublish = () => {
      const approvedArticles = articleIdeas
        .filter((idea) => idea.approved)
        .map((idea) => ({
          id: idea.id,
          feedback: idea.feedback || undefined,
        }))

      publishArticles({ articles: approvedArticles })
    }

    return (
      <div className="space-y-6">
        {/* Header */}
        <div className="text-center space-y-2">
          <h1 className="text-3xl font-bold font-all-round text-[var(--color-primary-dark)]">
            Aprove suas Matérias
          </h1>
          <p className="text-lg font-onest text-[var(--color-primary-dark)]/70">
            Revise as ideias geradas e adicione direcionamentos se desejar
          </p>
        </div>

        {/* Step Indicator */}
        <StepIndicator currentStep={currentStep} steps={steps} />

        {/* Articles Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {articleIdeas.map((idea) => (
            <ArticleIdeaCard
              key={idea.id}
              idea={idea}
              onUpdate={updateArticleIdea}
            />
          ))}
        </div>

        {/* Footer Actions */}
        <Card>
          <CardFooter className="flex flex-col sm:flex-row items-center justify-between gap-4 py-4">
            <div className="flex items-center gap-4 flex-wrap justify-center sm:justify-start">
              <div className="flex items-center gap-2">
                <FileCheck className="h-5 w-5 text-[var(--color-success)]" />
                <span className="text-sm font-medium font-onest text-[var(--color-primary-dark)]">
                  {approvedCount} matéria{approvedCount !== 1 ? 's' : ''} aprovada{approvedCount !== 1 ? 's' : ''}
                </span>
              </div>
              {feedbackCount > 0 && (
                <div className="flex items-center gap-2">
                  <MessageSquare className="h-5 w-5 text-[var(--color-primary-purple)]" />
                  <span className="text-sm font-medium font-onest text-[var(--color-primary-dark)]/70">
                    {feedbackCount} feedback{feedbackCount !== 1 ? 's' : ''} adicionado{feedbackCount !== 1 ? 's' : ''}
                  </span>
                </div>
              )}
            </div>

            <div className="flex items-center gap-3">
              <Button
                variant="outline"
                onClick={previousStep}
              >
                Voltar
              </Button>
              <Button
                variant="primary"
                size="lg"
                onClick={handlePublish}
                disabled={!canPublish}
                title={!canPublish ? 'Aprove pelo menos uma matéria' : undefined}
                className="bg-[var(--color-primary-purple)] hover:bg-[var(--color-primary-purple)]/90"
              >
                Publicar {approvedCount} Matéria{approvedCount !== 1 ? 's' : ''}
              </Button>
            </div>
          </CardFooter>
        </Card>

        {/* Progress Info */}
        <div className="text-center">
          <p className="text-sm font-onest text-[var(--color-primary-dark)]/60">
            Passo {currentStep} de {steps.length}
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="text-center space-y-2">
        <h1 className="text-3xl font-bold font-all-round text-[var(--color-primary-dark)]">
          Configuração Inicial
        </h1>
        <p className="text-lg font-onest text-[var(--color-primary-dark)]/70">
          Vamos configurar sua conta para começar a gerar conteúdo
        </p>
      </div>

      {/* Step Indicator */}
      <StepIndicator currentStep={currentStep} steps={steps} />

      {/* Form Card */}
      <Card>
        <CardHeader>
          <CardTitle>
            {currentStep === 1 && 'Informações do Negócio'}
            {currentStep === 2 && 'Análise de Concorrentes'}
            {currentStep === 3 && 'Configurar Integrações'}
          </CardTitle>
        </CardHeader>

        <CardContent>
          {/* Step 1: Business Info */}
          {currentStep === 1 && (
            <BusinessInfoForm
              onSubmit={(data: BusinessInput) => submitBusinessInfo(data)}
              isLoading={isSubmittingBusiness}
              defaultValues={businessData || undefined}
            />
          )}

          {/* Step 2: Competitors */}
          {currentStep === 2 && (
            <CompetitorsForm
              onSubmit={(data: CompetitorsInput) => submitCompetitors(data)}
              onBack={previousStep}
              isLoading={isSubmittingCompetitors}
              defaultValues={competitorData || undefined}
            />
          )}

          {/* Step 3: Integrations */}
          {currentStep === 3 && (
            <IntegrationsForm
              onSubmit={(data: IntegrationsInput) => submitIntegrations(data)}
              onBack={previousStep}
              isLoading={isSubmittingIntegrations}
              defaultValues={integrationsData || undefined}
            />
          )}
        </CardContent>
      </Card>

      {/* Progress Info */}
      <div className="text-center">
        <p className="text-sm font-onest text-[var(--color-primary-dark)]/60">
          Passo {currentStep} de {steps.length}
        </p>
      </div>
    </div>
  )
}