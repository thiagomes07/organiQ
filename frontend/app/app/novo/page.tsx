'use client'

import { useState } from 'react'
import { useWizard } from '@/hooks/useWizard'
import { useUser } from '@/store/authStore'
import { StepIndicator } from '@/components/wizards/StepIndicator'
import { CompetitorsForm } from '@/components/forms/CompetitorsForm'
import { ArticleIdeaCard } from '@/components/articles/ArticleIdeaCard'
import { LoadingOverlay } from '@/components/shared/LoadingSpinner'
import { Card, CardContent, CardDescription, CardHeader, CardTitle, CardFooter } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Slider } from '@/components/ui/slider'
import { Label } from '@/components/ui/label'
import { AlertCircle, MessageSquare, FileCheck } from 'lucide-react'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import Link from 'next/link'
import type { CompetitorsInput } from '@/lib/validations'

const steps = [
  { number: 1, label: 'Quantidade' },
  { number: 2, label: 'Concorrentes' },
  { number: 3, label: 'Aprovação' },
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

export default function NovoPage() {
  const user = useUser()

  const {
    currentStep,
    competitorData,
    articleIdeas,
    articleCount,
    nextStep,
    submitCompetitors,
    publishArticles,
    updateArticleIdea,
    setArticleCount,
    previousStep,
    isSubmittingCompetitors,
    isGeneratingIdeas,
    isPublishing,
    approvedCount,
    canPublish,
  } = useWizard(false) // false = não é onboarding

  // Estado para modal de confirmação
  const [showConfirmDialog, setShowConfirmDialog] = useState(false)

  const articlesRemaining = user ? user.maxArticles - user.articlesUsed : 0
  const canCreate = articlesRemaining > 0

  // Loading state para geração de ideias
  if (currentStep === 999 || isGeneratingIdeas) {
    return <LoadingOverlay messages={loadingMessages} />
  }

  // Loading state para publicação
  if (currentStep === 1000 || isPublishing) {
    return <LoadingOverlay messages={publishingMessages} />
  }

  // Step 3: Approval
  if (currentStep === 3) {
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
          Gerar Novas Matérias
        </h1>
        <p className="text-lg font-onest text-[var(--color-primary-dark)]/70">
          Crie mais conteúdo otimizado para seu blog
        </p>
      </div>

      {/* Limit Warning */}
      {!canCreate && (
        <Card className="border-[var(--color-warning)]">
          <CardContent className="flex items-start gap-3 p-4">
            <AlertCircle className="h-5 w-5 text-[var(--color-warning)] mt-0.5" />
            <div className="flex-1">
              <p className="font-medium font-onest text-[var(--color-primary-dark)]">
                Limite de matérias atingido
              </p>
              <p className="text-sm font-onest text-[var(--color-primary-dark)]/70 mt-1">
                Você já utilizou todas as {user?.maxArticles} matérias do seu plano este mês.
              </p>
              <Link href="/app/conta">
                <Button variant="outline" size="sm" className="mt-3">
                  Fazer Upgrade
                </Button>
              </Link>
            </div>
          </CardContent>
        </Card>
      )}

      {canCreate && (
        <>
          {/* Step Indicator */}
          <StepIndicator currentStep={currentStep} steps={steps} />

          {/* Form Card */}
          <Card>
            <CardHeader>
              <CardTitle>
                {currentStep === 1 && 'Quantidade de Matérias'}
                {currentStep === 2 && 'Análise de Concorrentes'}
              </CardTitle>
              <CardDescription>
                {currentStep === 1 && `Você tem ${articlesRemaining} matérias disponíveis este mês`}
                {currentStep === 2 && 'Adicione URLs de concorrentes para melhorar a estratégia (opcional)'}
              </CardDescription>
            </CardHeader>

            <CardContent>
              {/* Step 1: Article Count */}
              {currentStep === 1 && (
                <form
                  onSubmit={(e) => {
                    e.preventDefault()
                    nextStep()
                  }}
                  className="space-y-6"
                >
                  {/* Slider */}
                  <div className="space-y-4">
                    <Label required>Quantas matérias deseja criar?</Label>
                    <Slider
                      min={1}
                      max={articlesRemaining}
                      step={1}
                      value={[articleCount]}
                      onValueChange={(value) => setArticleCount(value[0])}
                      showValue
                      formatValue={(value) => `${value} ${value === 1 ? 'matéria' : 'matérias'}`}
                    />
                    <p className="text-center text-2xl font-bold font-all-round text-[var(--color-primary-purple)]">
                      {articleCount} {articleCount === 1 ? 'matéria' : 'matérias'}
                    </p>
                  </div>

                  {/* Info */}
                  <div className="bg-[var(--color-primary-purple)]/5 border border-[var(--color-primary-purple)]/20 rounded-[var(--radius-md)] p-4">
                    <p className="text-sm font-onest text-[var(--color-primary-dark)]/80">
                      💡 <strong>Dica:</strong> Você pode gerar várias matérias de uma vez para economizar tempo.
                    </p>
                  </div>

                  {/* Submit Button */}
                  <div className="flex justify-end pt-4">
                    <Button
                      type="submit"
                      variant="secondary"
                      size="lg"
                    >
                      Próximo
                    </Button>
                  </div>
                </form>
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
  )
}