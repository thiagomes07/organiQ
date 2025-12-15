"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Plus, Copy, ChevronLeft, ChevronRight } from "lucide-react";
import { useArticles } from "@/hooks/useArticles";
import { ArticleCard } from "@/components/articles/ArticleCard";
import { ArticleTable } from "@/components/articles/ArticleTable";
import { EmptyArticles } from "@/components/shared/EmptyState";
import { SkeletonTable } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Textarea } from "@/components/ui/textarea";
import { copyToClipboard } from "@/lib/utils";
import { toast } from "sonner";
import type { Article, ArticleStatus } from "@/types";

const ITEMS_PER_PAGE = 10;

export default function MateriasPage() {
  const router = useRouter();
  const [statusFilter, setStatusFilter] = useState<ArticleStatus | "all">(
    "all"
  );
  const [currentPage, setCurrentPage] = useState(1);
  const [selectedError, setSelectedError] = useState<Article | null>(null);

  const {
    articles,
    total,
    page,
    limit,
    isLoading,
    isEmpty,
    republishArticle,
    isRepublishing,
    hasActiveArticles,
  } = useArticles({
    status: statusFilter,
    page: currentPage,
    limit: ITEMS_PER_PAGE,
  });

  const totalPages = Math.ceil(total / limit);

  const handleCopyContent = async () => {
    if (selectedError?.content) {
      const success = await copyToClipboard(selectedError.content);
      if (success) {
        toast.success("Conteúdo copiado!");
      } else {
        toast.error("Erro ao copiar conteúdo");
      }
    }
  };

  const handleRepublish = (id: string) => {
    republishArticle(id);
    setSelectedError(null);
  };

  const handlePageChange = (newPage: number) => {
    if (newPage >= 1 && newPage <= totalPages) {
      setCurrentPage(newPage);
    }
  };

  // Reset page when filter changes
  const handleFilterChange = (value: string) => {
    setStatusFilter(value as ArticleStatus | "all");
    setCurrentPage(1);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold font-all-round text-[var(--color-primary-dark)]">
            Minhas Matérias
          </h1>
          <p className="text-sm font-onest text-[var(--color-primary-dark)]/70 mt-1">
            {total} {total === 1 ? "matéria" : "matérias"} no total
          </p>
        </div>

        <Button
          variant="secondary"
          size="lg"
          onClick={() => router.push("/app/novo")}
        >
          <Plus className="h-5 w-5 mr-2" />
          Gerar Novas
        </Button>
      </div>

      {/* Filters */}
      {!isEmpty && (
        <div className="flex items-center gap-4">
          <div className="w-full sm:w-48">
            <Select value={statusFilter} onValueChange={handleFilterChange}>
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
        <EmptyArticles onCreate={() => router.push("/app/novo")} />
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

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex items-center justify-between pt-4 border-t border-[var(--color-border)]">
              <p className="text-sm font-onest text-[var(--color-primary-dark)]/70">
                Página {page} de {totalPages}
              </p>

              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handlePageChange(currentPage - 1)}
                  disabled={currentPage === 1}
                >
                  <ChevronLeft className="h-4 w-4 mr-1" />
                  Anterior
                </Button>

                <div className="hidden sm:flex items-center gap-1">
                  {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                    let pageNum: number;
                    if (totalPages <= 5) {
                      pageNum = i + 1;
                    } else if (currentPage <= 3) {
                      pageNum = i + 1;
                    } else if (currentPage >= totalPages - 2) {
                      pageNum = totalPages - 4 + i;
                    } else {
                      pageNum = currentPage - 2 + i;
                    }

                    return (
                      <Button
                        key={pageNum}
                        variant={currentPage === pageNum ? "primary" : "ghost"}
                        size="sm"
                        onClick={() => handlePageChange(pageNum)}
                        className="w-10"
                      >
                        {pageNum}
                      </Button>
                    );
                  })}
                </div>

                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handlePageChange(currentPage + 1)}
                  disabled={currentPage === totalPages}
                >
                  Próxima
                  <ChevronRight className="h-4 w-4 ml-1" />
                </Button>
              </div>
            </div>
          )}
        </>
      )}

      {/* Error Modal */}
      <Dialog
        open={!!selectedError}
        onOpenChange={() => setSelectedError(null)}
      >
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
                  <Button variant="ghost" size="sm" onClick={handleCopyContent}>
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
            <Button variant="outline" onClick={() => setSelectedError(null)}>
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
  );
}
