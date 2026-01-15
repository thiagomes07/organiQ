'use client'

import { Check } from 'lucide-react'
import { formatCurrency } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { cn } from '@/lib/utils'
import type { Plan } from '@/types'

interface PlanCardProps {
  plan: Plan
  onSelect: () => void
  isRecommended?: boolean
  isLoading?: boolean
}

export function PlanCard({ plan, onSelect, isRecommended, isLoading }: PlanCardProps) {
  return (
    <Card
      className={cn(
        'relative hover:shadow-lg hover:scale-[1.02] transition-all duration-200',
        isRecommended && 'border-2 border-[var(--color-primary-teal)]'
      )}
    >
      {/* Recommended Badge */}
      {isRecommended && (
        <div className="absolute -top-3 left-1/2 -translate-x-1/2">
          <div className="px-4 py-1 rounded-full bg-[var(--color-primary-teal)] text-white text-xs font-bold font-all-round shadow-md">
            Recomendado
          </div>
        </div>
      )}

      <CardHeader className="text-center pt-8">
        <CardTitle className="text-2xl">{plan.name}</CardTitle>
        <div className="mt-4">
          <span className="text-4xl font-bold font-all-round text-[var(--color-primary-dark)]">
            {formatCurrency(plan.price)}
          </span>
          <span className="text-sm font-onest text-[var(--color-primary-dark)]/60">/mês</span>
        </div>
      </CardHeader>

      <CardContent className="space-y-4">
        {/* Destaque Principal */}
        <div className="text-center py-3 px-4 rounded-[var(--radius-md)] bg-[var(--color-primary-purple)]/10">
          <p className="text-lg font-bold font-all-round text-[var(--color-primary-purple)]">
            {plan.maxArticles} matérias/mês
          </p>
        </div>

        {/* Features List */}
        <ul className="space-y-3">
          {plan.features.map((feature, index) => (
            <li key={index} className="flex items-start gap-2">
              <Check className="h-5 w-5 text-[var(--color-success)] mt-0.5 flex-shrink-0" />
              <span className="text-sm font-onest text-[var(--color-primary-dark)]/80">
                {feature}
              </span>
            </li>
          ))}
        </ul>
      </CardContent>

      <CardFooter>
        <Button
          variant="secondary"
          className="w-full"
          size="lg"
          onClick={onSelect}
          isLoading={isLoading}
          disabled={isLoading}
        >
          Escolher Plano
        </Button>
      </CardFooter>
    </Card>
  )
}