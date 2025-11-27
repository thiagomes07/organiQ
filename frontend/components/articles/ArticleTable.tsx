'use client'

import { ExternalLink, AlertCircle, Loader2 } from 'lucide-react'
import { formatDateTime, truncate } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import type { Article } from '@/types'

interface ArticleTableProps {
  articles: Article[]
  onViewError?: (article: Article) => void
  onRepublish?: (id: string) => void
  isRepublishing?: boolean
}

const statusConfig = {
  generating: {
    color: 'bg-[var(--color-warning)]/10',
    textColor: 'text-[var(--color-warning)]',
    label: 'Gerando...',
    icon: Loader2,
  },
  publishing: {
    color: 'bg-blue-500/10',
    textColor: 'text-blue-500',
    label: 'Publicando...',
    icon: Loader2,
  },
  published: {
    color: 'bg-[var(--color-success)]/10',
    textColor: 'text-[var(--color-success)]',
    label: 'Publicado',
    icon: ExternalLink,
  },
  error: {
    color: 'bg-[var(--color-error)]/10',
    textColor: 'text-[var(--color-error)]',
    label: 'Erro',
    icon: AlertCircle,
  },
}

export function ArticleTable({
  articles,
  onViewError,
  onRepublish,
  isRepublishing,
}: ArticleTableProps) {
  return (
    <div className="w-full overflow-x-auto">
      <table className="w-full">
        <thead>
          <tr className="border-b border-[var(--color-border)]">
            <th className="text-left py-3 px-4 text-sm font-semibold font-all-round text-[var(--color-primary-dark)]">
              Título
            </th>
            <th className="text-left py-3 px-4 text-sm font-semibold font-all-round text-[var(--color-primary-dark)] min-w-[150px]">
              Data
            </th>
            <th className="text-left py-3 px-4 text-sm font-semibold font-all-round text-[var(--color-primary-dark)] min-w-[120px]">
              Status
            </th>
            <th className="text-right py-3 px-4 text-sm font-semibold font-all-round text-[var(--color-primary-dark)] min-w-[140px]">
              Ações
            </th>
          </tr>
        </thead>
        <tbody>
          {articles.map((article) => {
            const status = statusConfig[article.status]
            const StatusIcon = status.icon

            return (
              <tr
                key={article.id}
                className="border-b border-[var(--color-border)] hover:bg-[var(--color-primary-dark)]/5 transition-colors"
              >
                {/* Título */}
                <td className="py-4 px-4">
                  <div className="flex items-center gap-2">
                    <span
                      className="font-medium font-onest text-[var(--color-primary-dark)]"
                      title={article.title}
                    >
                      {truncate(article.title, 60)}
                    </span>
                  </div>
                </td>

                {/* Data */}
                <td className="py-4 px-4">
                  <span className="text-sm font-onest text-[var(--color-primary-dark)]/70">
                    {formatDateTime(article.createdAt)}
                  </span>
                </td>

                {/* Status */}
                <td className="py-4 px-4">
                  <div
                    className={cn(
                      'inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium font-onest',
                      status.color,
                      status.textColor
                    )}
                  >
                    <StatusIcon
                      className={cn('h-3.5 w-3.5', {
                        'animate-spin':
                          article.status === 'generating' || article.status === 'publishing',
                      })}
                    />
                    {status.label}
                  </div>
                </td>

                {/* Ações */}
                <td className="py-4 px-4">
                  <div className="flex items-center justify-end gap-2">
                    {article.status === 'published' && article.postUrl && (
                      <a
                        href={article.postUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        <Button variant="outline" size="sm">
                          <ExternalLink className="h-3.5 w-3.5 mr-1.5" />
                          Ver Post
                        </Button>
                      </a>
                    )}

                    {article.status === 'error' && (
                      <>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => onViewError?.(article)}
                        >
                          <AlertCircle className="h-3.5 w-3.5 mr-1.5" />
                          Detalhes
                        </Button>
                        {onRepublish && (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => onRepublish(article.id)}
                            disabled={isRepublishing}
                            className="text-[var(--color-primary-purple)]"
                          >
                            Republicar
                          </Button>
                        )}
                      </>
                    )}

                    {(article.status === 'generating' || article.status === 'publishing') && (
                      <span className="text-sm font-onest text-[var(--color-primary-dark)]/60">
                        Aguarde...
                      </span>
                    )}
                  </div>
                </td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}