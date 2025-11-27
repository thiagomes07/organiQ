'use client'

import { ExternalLink, AlertCircle, Loader2 } from 'lucide-react'
import { formatDateTime } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardFooter, CardHeader } from '@/components/ui/card'
import { cn } from '@/lib/utils'
import type { Article } from '@/types'

interface ArticleCardProps {
  article: Article
  onViewError?: (article: Article) => void
  onRepublish?: (id: string) => void
  isRepublishing?: boolean
}

const statusConfig = {
  generating: {
    color: 'bg-[var(--color-warning)]',
    textColor: 'text-[var(--color-warning)]',
    label: 'Gerando...',
    icon: Loader2,
  },
  publishing: {
    color: 'bg-blue-500',
    textColor: 'text-blue-500',
    label: 'Publicando...',
    icon: Loader2,
  },
  published: {
    color: 'bg-[var(--color-success)]',
    textColor: 'text-[var(--color-success)]',
    label: 'Publicado',
    icon: ExternalLink,
  },
  error: {
    color: 'bg-[var(--color-error)]',
    textColor: 'text-[var(--color-error)]',
    label: 'Erro',
    icon: AlertCircle,
  },
}

export function ArticleCard({
  article,
  onViewError,
  onRepublish,
  isRepublishing,
}: ArticleCardProps) {
  const status = statusConfig[article.status]
  const StatusIcon = status.icon

  return (
    <Card className="hover:shadow-md transition-shadow duration-200">
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between gap-3">
          <h3 className="text-lg font-semibold font-all-round text-[var(--color-primary-dark)] line-clamp-2">
            {article.title}
          </h3>
          <div
            className={cn(
              'flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium font-onest shrink-0',
              status.color,
              'text-white'
            )}
          >
            <StatusIcon
              className={cn('h-3.5 w-3.5', {
                'animate-spin': article.status === 'generating' || article.status === 'publishing',
              })}
            />
            {status.label}
          </div>
        </div>
      </CardHeader>

      <CardContent className="pb-3">
        <div className="flex items-center gap-2 text-sm font-onest text-[var(--color-primary-dark)]/60">
          <span>{formatDateTime(article.createdAt)}</span>
        </div>
      </CardContent>

      <CardFooter className="pt-3 border-t border-[var(--color-border)]">
        {article.status === 'published' && article.postUrl && (
          <a
            href={article.postUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="w-full"
          >
            <Button variant="outline" size="sm" className="w-full">
              <ExternalLink className="h-4 w-4 mr-2" />
              Ver Publicação
            </Button>
          </a>
        )}

        {article.status === 'error' && (
          <div className="w-full space-y-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => onViewError?.(article)}
              className="w-full"
            >
              <AlertCircle className="h-4 w-4 mr-2" />
              Ver Detalhes
            </Button>
            {onRepublish && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onRepublish(article.id)}
                disabled={isRepublishing}
                className="w-full text-[var(--color-primary-purple)]"
              >
                Tentar Republicar
              </Button>
            )}
          </div>
        )}

        {(article.status === 'generating' || article.status === 'publishing') && (
          <div className="w-full text-center text-sm font-onest text-[var(--color-primary-dark)]/60">
            Aguarde...
          </div>
        )}
      </CardFooter>
    </Card>
  )
}