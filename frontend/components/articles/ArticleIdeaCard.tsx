'use client'

import { useState, useEffect } from 'react'
import { Check, X, MessageSquare } from 'lucide-react'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { cn } from '@/lib/utils'
import { debounce } from '@/lib/utils'
import type { ArticleIdea } from '@/types'

interface ArticleIdeaCardProps {
  idea: ArticleIdea
  onUpdate: (id: string, updates: Partial<ArticleIdea>) => void
}

export function ArticleIdeaCard({ idea, onUpdate }: ArticleIdeaCardProps) {
  const [localFeedback, setLocalFeedback] = useState(idea.feedback || '')

  // Debounced update para feedback
  useEffect(() => {
    const debouncedUpdate = debounce(() => {
      if (localFeedback !== idea.feedback) {
        onUpdate(idea.id, { feedback: localFeedback })
      }
    }, 1000)

    debouncedUpdate()
  }, [localFeedback, idea.id, idea.feedback, onUpdate])

  const handleToggleApprove = (approved: boolean) => {
    onUpdate(idea.id, { approved })
  }

  return (
    <Card
      className={cn(
        'transition-all duration-200',
        idea.approved && 'border-l-4 border-l-[var(--color-success)]',
        !idea.approved && 'opacity-60'
      )}
    >
      <CardHeader>
        <div className="flex items-start justify-between gap-3">
          <h3 className="text-lg font-semibold font-all-round text-[var(--color-primary-dark)] line-clamp-2 flex-1">
            {idea.title}
          </h3>
          {idea.approved && (
            <div className="flex items-center gap-1 px-2 py-1 rounded-full bg-[var(--color-success)]/10 text-[var(--color-success)] text-xs font-medium shrink-0">
              <Check className="h-3.5 w-3.5" />
              Aprovado
            </div>
          )}
        </div>
      </CardHeader>

      <CardContent className="space-y-4">
        {/* Summary */}
        <p className="text-sm font-onest text-[var(--color-primary-dark)]/80 line-clamp-3">
          {idea.summary}
        </p>

        {/* Toggle Buttons */}
        <div className="flex gap-2">
          <Button
            variant={idea.approved ? 'success' : 'outline'}
            size="sm"
            onClick={() => handleToggleApprove(true)}
            className={cn(
              'flex-1',
              idea.approved && 'bg-[var(--color-success)] text-white hover:bg-[var(--color-success)]/90'
            )}
          >
            <Check className="h-4 w-4 mr-2" />
            Aprovar
          </Button>
          <Button
            variant={!idea.approved ? 'outline' : 'ghost'}
            size="sm"
            onClick={() => handleToggleApprove(false)}
            className={cn(
              'flex-1',
              !idea.approved && 'border-[var(--color-primary-dark)]/20'
            )}
          >
            <X className="h-4 w-4 mr-2" />
            Rejeitar
          </Button>
        </div>

        {/* Feedback Field */}
        <div className="space-y-2 pt-2 border-t border-[var(--color-border)]">
          <div className="flex items-center gap-2">
            <MessageSquare className="h-4 w-4 text-[var(--color-primary-teal)]" />
            <Label htmlFor={`feedback-${idea.id}`} className="text-xs">
              Sugestões ou direcionamentos (opcional)
            </Label>
          </div>
          <Textarea
            id={`feedback-${idea.id}`}
            value={localFeedback}
            onChange={(e) => setLocalFeedback(e.target.value)}
            placeholder="Ex: Foque em pequenas empresas, adicione exemplos práticos..."
            className="min-h-[60px] max-h-[100px] text-sm bg-[var(--color-secondary-cream)]/50 border-[var(--color-primary-teal)]/30 focus:border-[var(--color-primary-purple)]"
            maxLength={500}
            showCount
          />
        </div>

        {/* Badges */}
        <div className="flex items-center gap-2 flex-wrap">
          {idea.approved && localFeedback && (
            <div className="px-2 py-1 rounded-full bg-[var(--color-primary-purple)]/10 text-[var(--color-primary-purple)] text-xs font-medium">
              Com direcionamento
            </div>
          )}
          {!idea.approved && localFeedback && (
            <div className="px-2 py-1 rounded-full bg-[var(--color-primary-dark)]/10 text-[var(--color-primary-dark)]/60 text-xs font-medium">
              Feedback enviado
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}