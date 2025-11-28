'use client'

import { useWizard } from '@/hooks/useWizard'
import { useUser } from '@/store/authStore'
import { StepIndicator } from '@/components/wizards/StepIndicator'
import { CompetitorsForm } from '@/components/forms/CompetitorsForm'
import { LoadingOverlay } from '@/components/shared/LoadingSpinner'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Slider } from '@/components/ui/slider'
import { Label } from '@/components/ui/label'
import { AlertCircle } from 'lucide-react'
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

export default function NovoPage() {
  const user = useUser()
  const {
    currentStep,
    businessData,
    competitorData,
    submitBusinessInfo,
    submitCompetitors,
    previousStep,
    isSubmittingBusiness,
    isSubmittingCompetitors,
    isGeneratingIdeas,
  } = useWizard(false) // false = não é onboarding

  const articlesRemaining = user ? user.maxArticles - user.articlesUsed : 0
  const canCreate = articlesRemaining > 0

  // Loading state para geração de ideias
  if (currentStep === 999 || isGeneratingIdeas) {
    return <LoadingOverlay messages={loadingMessages} />
  }

  // TODO: Implement step 3 (Approval) and step 1000 (Publishing)
  // These will be added in the next phase

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
                    const articleCount = businessData?.articleCount || 1
                    submitBusinessInfo({
                      description: '', // Dados já existem do onboarding
                      primaryObjective: 'leads', // Placeholder
                      hasBlog: false,
                      blogUrls: [],
                      articleCount,
                    } as any)
                  }}
                  className="space-y-6"
                >
                  {/* Slider */}
                  <div className="space-y-2">
                    <Label required>Quantas matérias deseja criar?</Label>
                    <Slider
                      min={1}
                      max={articlesRemaining}
                      step={1}
                      value={[businessData?.articleCount || 1]}
                      onValueChange={(value) => {
                        // Atualizar estado do wizard
                        submitBusinessInfo({
                          ...businessData,
                          articleCount: value[0],
                        } as any)
                      }}
                      showValue
                      formatValue={(value) => `${value} ${value === 1 ? 'matéria' : 'matérias'}`}
                    />
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
                      isLoading={isSubmittingBusiness}
                      disabled={isSubmittingBusiness}
                    >
                      Próximo
                    </Button>
                  </div>
                </form>
              )}

              {/* Step 2: Competitors */}
              {currentStep === 2 && (
                <CompetitorsForm
                  onSubmit={(data: CompetitorsInput) => submitCompetitors(data as any)}
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