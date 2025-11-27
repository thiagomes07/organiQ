'use client'

import { useWizard } from '@/hooks/useWizard'
import { StepIndicator } from './StepIndicator'
import { BusinessInfoForm } from '@/components/forms/BusinessInfoForm'
import { CompetitorsForm } from '@/components/forms/CompetitorsForm'
import { IntegrationsForm } from '@/components/forms/IntegrationsForm'
import { LoadingOverlay } from '@/components/shared/LoadingSpinner'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
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

export function OnboardingWizard() {
  const {
    currentStep,
    businessData,
    competitorData,
    integrationsData,
    submitBusinessInfo,
    submitCompetitors,
    submitIntegrations,
    previousStep,
    isSubmittingBusiness,
    isSubmittingCompetitors,
    isSubmittingIntegrations,
    isGeneratingIdeas,
  } = useWizard(true) // true = isOnboarding

  // Loading state para geração de ideias
  if (currentStep === 999 || isGeneratingIdeas) {
    return <LoadingOverlay messages={loadingMessages} />
  }

  // TODO: Implement steps 4 (Approval) and 1000 (Publishing)
  // These will be added in the next phase

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
              onSubmit={(data: BusinessInput) => submitBusinessInfo(data as any)}
              isLoading={isSubmittingBusiness}
              defaultValues={businessData || undefined}
            />
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

          {/* Step 3: Integrations */}
          {currentStep === 3 && (
            <IntegrationsForm
              onSubmit={(data: IntegrationsInput) => submitIntegrations(data as any)}
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