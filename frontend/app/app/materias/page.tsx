'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Plus, Copy, X } from 'lucide-react'
import { useArticles } from '@/hooks/useArticles'
import { ArticleCard } from '@/components/articles/ArticleCard'
import { ArticleTable } from '@/components/articles/ArticleTable'
import { EmptyArticles } from '@/components/shared/EmptyState'
import { SkeletonTable } from '@/components/ui/skeleton'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Textarea } from '@/components/ui/textarea'
import { copyToClipboard } from '@/lib/utils'
import { toast } from 'sonner'
import type { Article, ArticleStatus } from '@/types'

export default function MateriasPage() {
  const router = useRouter()
  const [statusFilter, setStatusFilter] = useState<ArticleStatus | 'all'>('all')
  const [selectedError, setSelectedError] = useState<Article | null>(null)

  const {
    articles,
    total,
    isLoading,
    isEmpty,
    republishArticle,
    isRepublishing,
    hasActiveArticles,
    refetch,
  } = useArticles({ status: statusFilter })

  const handleCopyContent = async () => {
    if (selectedError?.content) {
      const success = await copyToClipboard(selectedError.content)
      if (success) {
        toast.success('Conteúdo copiado!')
      } else {
        toast.error('Erro ao copiar conteúdo')
      }
    }
  }

  const handleRepublish = (id: string) => {
    republishArticle(id)
    setSelectedError(null)
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold font-all-round text-[var(--color-primary-dark)]">
            Minhas Matérias
          </h1>
          <p className="text-sm font-onest text-[var(--color-primary-dark)]/70 mt-1">
            {total} {total === 1 ? 'matéria' : 'matérias'} no total
          </p>
        </div>

        <Button
          variant="secondary"
          size="lg"
          onClick={() => router.push('/app/novo')}
        >
          <Plus className="h-5 w-5 mr-2" />
          Gerar Novas
        </Button>
      </div>

      {/* Filters */}
      {!isEmpty && (
        <div className="flex items-center gap-4">
          <div className="w-full sm:w-48">
            <Select value={statusFilter} onValueChange={(value) => setStatusFilter(value as any)}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Todos os status</SelectItem>
                <SelectItem value="published">Publicadas</SelectItem>
                <SelectItem value="generating">Gerando</SelectItem>
                <SelectItem value="publishing">Publicando</SelectItem>
                <SelectItem value="error">Com erro</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {hasActiveArticles && (
            <div className="flex items-center gap-2 text-sm font-onest text-[var(--color-primary-dark)]/70">
              <div className="h-2 w-2 rounded-full bg-[var(--color-primary-purple)] animate-pulse" />
              <span>Atualizando automaticamente...</span>
            </div>
          )}
        </div>
      )}

      {/* Loading State */}
      {isLoading && (
        <div className="bg-white rounded-[var(--radius-md)] shadow-sm p-6">
          <SkeletonTable rows={5} />
        </div>
      )}

      {/* Empty State */}
      {!isLoading && isEmpty && (
        <EmptyArticles onCreate={() => router.push('/app/novo')} />
      )}

      {/* Articles List */}
      {!isLoading && !isEmpty && (
        <>
          {/* Desktop: Table */}
          <div className="hidden lg:block bg-white rounded-[var(--radius-md)] shadow-sm overflow-hidden">
            <ArticleTable
              articles={articles}
              onViewError={setSelectedError}
              onRepublish={republishArticle}
              isRepublishing={isRepublishing}
            />
          </div>

          {/* Mobile: Cards */}
          <div className="lg:hidden grid gap-4">
            {articles.map((article) => (
              <ArticleCard
                key={article.id}
                article={article}
                onViewError={setSelectedError}
                onRepublish={republishArticle}
                isRepublishing={isRepublishing}
              />
            ))}
          </div>
        </>
      )}

      {/* Error Modal */}
      <Dialog open={!!selectedError} onOpenChange={() => setSelectedError(null)}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>{selectedError?.title}</DialogTitle>
            <DialogDescription>
              Detalhes do erro ocorrido durante a publicação
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            {/* Error Message */}
            {selectedError?.errorMessage && (
              <div className="bg-[var(--color-error)]/10 border border-[var(--color-error)]/20 rounded-[var(--radius-sm)] p-4">
                <p className="text-sm font-onest text-[var(--color-error)]">
                  {selectedError.errorMessage}
                </p>
              </div>
            )}

            {/* Content */}
            {selectedError?.content && (
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <label className="text-sm font-medium font-onest text-[var(--color-primary-dark)]">
                    Conteúdo gerado
                  </label>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={handleCopyContent}
                  >
                    <Copy className="h-4 w-4 mr-2" />
                    Copiar
                  </Button>
                </div>
                <Textarea
                  value={selectedError.content}
                  readOnly
                  className="min-h-[200px] font-mono text-xs"
                />
              </div>
            )}
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setSelectedError(null)}
            >
              Fechar
            </Button>
            {selectedError && (
              <Button
                variant="primary"
                onClick={() => handleRepublish(selectedError.id)}
                isLoading={isRepublishing}
              >
                Tentar Republicar
              </Button>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}