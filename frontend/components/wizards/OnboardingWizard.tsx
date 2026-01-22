'use client'

import { useState } from 'react'
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
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
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
    isLoadingWizardData,
    isInitialized,
    isSubmittingBusiness,
    isSubmittingCompetitors,
    isSubmittingIntegrations,
    isGeneratingIdeas,
    isPublishing,
    approvedCount,
    canPublish,
    regenerationsRemaining,
    regenerationsLimit,
    nextRegenerationAt,
    hasGeneratedIdeas,
    regenerateIdeas,
    canRegenerateIdeas,
    allApproved,
  } = useWizard(true) // true = isOnboarding

  // Estado para modal de confirmação
  const [showConfirmDialog, setShowConfirmDialog] = useState(false)
  const [showRegenerateDialog, setShowRegenerateDialog] = useState(false)

  // Loading state inicial enquanto busca dados do wizard
  if (isLoadingWizardData || !isInitialized) {
    return <LoadingOverlay messages={['Carregando seus dados...', 'Verificando progresso...']} />
  }

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
      setShowConfirmDialog(false)
    }

    const handleRegenerateConfirm = () => {
      regenerateIdeas()
      setShowRegenerateDialog(false)
    }

    const unapprovedWithFeedback = articleIdeas.filter(
      (idea) => !idea.approved && idea.feedback && idea.feedback.trim().length > 0
    ).length

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
        <div className="columns-1 md:columns-2 gap-4 space-y-4">
          {articleIdeas.map((idea) => (
            <div key={idea.id} className="break-inside-avoid mb-4">
              <ArticleIdeaCard
                idea={idea}
                onUpdate={updateArticleIdea}
              />
            </div>
          ))}
        </div>

        {/* Footer Actions */}
        <Card>
          <CardFooter className="flex flex-col sm:flex-row items-center justify-between gap-4 py-4">
            <div className="flex flex-col sm:items-start gap-1 w-full sm:w-auto">
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
              {/* Regenerate Info */}
              {!allApproved && (
                <div className="text-xs text-[var(--color-primary-dark)]/60 font-onest mt-1">
                  Regenerações: {regenerationsRemaining}/{regenerationsLimit}
                  {!canRegenerateIdeas && nextRegenerationAt && (
                    <span className="ml-2 text-[var(--color-warning)]">
                      (Disponível em {new Date(nextRegenerationAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })})
                    </span>
                  )}
                </div>
              )}
            </div>


            <div className="flex items-center gap-3 flex-wrap justify-center sm:justify-end">
              {!allApproved && (
                <Button
                  variant="outline"
                  onClick={() => setShowRegenerateDialog(true)}
                  disabled={!canRegenerateIdeas || isGeneratingIdeas}
                  className="text-[var(--color-primary-purple)] border-[var(--color-primary-purple)] hover:bg-[var(--color-primary-purple)]/5"
                >
                  Gerar Novas Ideias
                </Button>
              )}
              <Button
                variant="outline"
                onClick={previousStep}
              >
                Voltar
              </Button>
              <Button
                variant="primary"
                size="lg"
                onClick={() => setShowConfirmDialog(true)}
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

        {/* Regenerate Confirmation Dialog */}
        <Dialog open={showRegenerateDialog} onOpenChange={setShowRegenerateDialog}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Regenerar Ideias de Matérias?</DialogTitle>
              <DialogDescription>
                Esta ação irá substituir as ideias não aprovadas por novas sugestões.
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4 py-4">
              <div className="bg-[var(--color-warning)]/10 border border-[var(--color-warning)]/30 rounded-[var(--radius-md)] p-4 space-y-2">
                <p className="text-sm font-medium font-onest text-[var(--color-warning)]">
                  ⚠️ Atenção
                </p>
                <p className="text-sm font-onest text-[var(--color-primary-dark)]/80">
                  As ideias <strong>não aprovadas</strong> serão <strong>permanentemente removidas</strong>, incluindo quaisquer sugestões ou direcionamentos que você tenha adicionado.
                </p>
                {unapprovedWithFeedback > 0 && (
                  <p className="text-sm font-semibold font-onest text-[var(--color-error)] mt-2">
                    Você perderá {unapprovedWithFeedback} sugestão{unapprovedWithFeedback !== 1 ? 'ões' : ''} que adicionou!
                  </p>
                )}
              </div>

              <div className="bg-[var(--color-success)]/10 border border-[var(--color-success)]/30 rounded-[var(--radius-md)] p-4">
                <p className="text-sm font-onest text-[var(--color-primary-dark)]/80">
                  ✅ As ideias <strong>aprovadas</strong> serão mantidas e combinadas com as novas sugestões.
                </p>
              </div>

              <p className="text-xs font-onest text-[var(--color-primary-dark)]/60">
                Regenerações restantes: {regenerationsRemaining}/{regenerationsLimit}
              </p>
            </div>

            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => setShowRegenerateDialog(false)}
              >
                Cancelar
              </Button>
              <Button
                variant="primary"
                onClick={handleRegenerateConfirm}
                className="bg-[var(--color-primary-purple)] hover:bg-[var(--color-primary-purple)]/90"
              >
                Sim, Regenerar Ideias
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Confirmation Dialog */}
        <Dialog open={showConfirmDialog} onOpenChange={setShowConfirmDialog}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Confirmar Publicação</DialogTitle>
              <DialogDescription>
                Você está prestes a publicar {approvedCount} matéria{approvedCount !== 1 ? 's' : ''} no seu WordPress.
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4 py-4">
              <div className="bg-[var(--color-secondary-cream)]/50 rounded-[var(--radius-md)] p-4 space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-onest text-[var(--color-primary-dark)]/70">
                    Matérias aprovadas:
                  </span>
                  <span className="text-sm font-semibold font-all-round text-[var(--color-primary-dark)]">
                    {approvedCount}
                  </span>
                </div>
                {feedbackCount > 0 && (
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-onest text-[var(--color-primary-dark)]/70">
                      Com direcionamentos:
                    </span>
                    <span className="text-sm font-semibold font-all-round text-[var(--color-primary-purple)]">
                      {feedbackCount}
                    </span>
                  </div>
                )}
              </div>

              <p className="text-sm font-onest text-[var(--color-primary-dark)]/70">
                As matérias serão escritas com IA e publicadas automaticamente no seu blog WordPress.
                Este processo pode levar alguns minutos.
                <br />
                <span className="text-xs text-[var(--color-warning)] mt-2 block">
                  Nota: As ideias não aprovadas serão descartadas.
                </span>
              </p>
            </div>

            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => setShowConfirmDialog(false)}
              >
                Cancelar
              </Button>
              <Button
                variant="primary"
                onClick={handlePublish}
                className="bg-[var(--color-primary-purple)] hover:bg-[var(--color-primary-purple)]/90"
              >
                Confirmar Publicação
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
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
              hasGeneratedIdeas={hasGeneratedIdeas}
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