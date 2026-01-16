'use client'

import { Check } from 'lucide-react'
import { cn } from '@/lib/utils'

interface Step {
  number: number
  label: string
}

interface StepIndicatorProps {
  currentStep: number
  steps: Step[]
}

export function StepIndicator({ currentStep, steps }: StepIndicatorProps) {
  return (
    <div className="w-full mb-8 px-6">
      
      {/* Desktop: Horizontal */}
      <div className="hidden md:flex items-center">
        {steps.map((step, index) => {
          const isCompleted = currentStep > step.number
          const isCurrent = currentStep === step.number
          const isLast = index === steps.length - 1

          return (
            <div key={step.number} className="flex items-center" style={{ flex: isLast ? '0 0 auto' : '1 1 0%' }}>
              {/* Step Circle & Label Wrapper */}
              <div className="flex flex-col items-center">
                <div
                  className={cn(
                    'flex items-center justify-center h-10 w-10 rounded-full border-2 transition-all duration-200',
                    isCompleted && 'bg-[var(--color-success)] border-[var(--color-success)]',
                    isCurrent && 'bg-[var(--color-primary-purple)] border-[var(--color-primary-purple)]',
                    !isCompleted && !isCurrent && 'bg-white border-[var(--color-border)]'
                  )}
                >
                  {isCompleted ? (
                    <Check className="h-5 w-5 text-white" />
                  ) : (
                    <span
                      className={cn(
                        'text-sm font-bold font-all-round',
                        isCurrent && 'text-white',
                        !isCurrent && 'text-[var(--color-primary-dark)]/60'
                      )}
                    >
                      {step.number}
                    </span>
                  )}
                </div>
                <span
                  className={cn(
                    'mt-2 text-xs font-medium font-onest text-center',
                    isCurrent && 'text-[var(--color-primary-purple)]',
                    isCompleted && 'text-[var(--color-success)]',
                    !isCurrent && !isCompleted && 'text-[var(--color-primary-dark)]/60'
                  )}
                >
                  {step.label}
                </span>
              </div>

              {/* Connector Line */}
              {!isLast && (
                <div
                  className={cn(
                    'h-0.5 transition-all duration-200',
                    isCompleted ? 'bg-[var(--color-success)]' : 'bg-[var(--color-border)]'
                  )}
                  style={{ flex: '1 1 0%', margin: '0 16px' }}
                />
              )}
            </div>
          )
        })}
      </div>

      {/* Mobile: Vertical Compact */}
      <div className="md:hidden">
        <div className="flex items-center gap-4">
          {steps.map((step, index) => {
            const isCompleted = currentStep > step.number
            const isCurrent = currentStep === step.number

            return (
              <div key={step.number} className="flex items-center">
                <div
                  className={cn(
                    'flex items-center justify-center h-8 w-8 rounded-full border-2 transition-all duration-200',
                    isCompleted && 'bg-[var(--color-success)] border-[var(--color-success)]',
                    isCurrent && 'bg-[var(--color-primary-purple)] border-[var(--color-primary-purple)]',
                    !isCompleted && !isCurrent && 'bg-white border-[var(--color-border)]'
                  )}
                >
                  {isCompleted ? (
                    <Check className="h-4 w-4 text-white" />
                  ) : (
                    <span
                      className={cn(
                        'text-xs font-bold font-all-round',
                        isCurrent && 'text-white',
                        !isCurrent && 'text-[var(--color-primary-dark)]/60'
                      )}
                    >
                      {step.number}
                    </span>
                  )}
                </div>
                {index < steps.length - 1 && (
                  <div
                    className={cn(
                      'w-8 h-0.5 mx-1',
                      isCompleted ? 'bg-[var(--color-success)]' : 'bg-[var(--color-border)]'
                    )}
                  />
                )}
              </div>
            )
          })}
        </div>
        {/* Current Step Label */}
        <p className="mt-3 text-sm font-medium font-onest text-[var(--color-primary-purple)]">
          {steps.find((s) => s.number === currentStep)?.label}
        </p>
      </div>
    </div>
  )
}