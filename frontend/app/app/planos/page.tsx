'use client'

import { Check } from 'lucide-react'
import { usePlans } from '@/hooks/usePlans'
import { PlanCard } from '@/components/plans/PlanCard'
import { SkeletonCard } from '@/components/ui/skeleton'

export default function PlanosPage() {
  const { plans, selectPlan, isLoadingPlans, isCreatingCheckout, getRecommendedPlan } = usePlans()

  const recommendedPlan = getRecommendedPlan()

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="text-center space-y-3">
        <h1 className="text-3xl md:text-4xl font-bold font-all-round text-[var(--color-primary-dark)]">
          Escolha seu Plano
        </h1>
        <p className="text-lg font-onest text-[var(--color-primary-dark)]/70 max-w-2xl mx-auto">
          Comece a gerar conteúdo de qualidade para seu blog hoje mesmo
        </p>
      </div>

      {/* Loading State */}
      {isLoadingPlans && (
        <div className="grid md:grid-cols-3 gap-6">
          {[1, 2, 3].map((i) => (
            <SkeletonCard key={i} />
          ))}
        </div>
      )}

      {/* Plans Grid */}
      {!isLoadingPlans && (
        <div className="grid md:grid-cols-3 gap-6 max-w-6xl mx-auto">
          {plans.map((plan) => (
            <PlanCard
              key={plan.id}
              plan={plan}
              onSelect={() => selectPlan(plan.id)}
              isRecommended={plan.id === recommendedPlan?.id}
              isLoading={isCreatingCheckout}
            />
          ))}
        </div>
      )}

      {/* Garantia */}
      <div className="max-w-3xl mx-auto text-center space-y-3 pt-8">
        <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-[var(--color-success)]/10 text-[var(--color-success)]">
          <Check className="h-5 w-5" />
          <span className="text-sm font-semibold font-onest">
            Garantia de 7 dias - 100% do seu dinheiro de volta
          </span>
        </div>
        <p className="text-sm font-onest text-[var(--color-primary-dark)]/60">
          Todos os planos incluem suporte via email e atualizações automáticas
        </p>
      </div>
    </div>
  )
}