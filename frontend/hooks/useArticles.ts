import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import api, { getErrorMessage } from '@/lib/axios'
import { useAuthStore } from '@/store/authStore'
import type { 
  Article, 
  ArticlesResponse, 
  ArticleFilters,
  ArticleStatus 
} from '@/types'

// ============================================
// API FUNCTIONS
// ============================================

const articlesApi = {
  getArticles: async (filters: ArticleFilters = {}): Promise<ArticlesResponse> => {
    const params = {
      page: filters.page || 1,
      limit: filters.limit || 10,
      status: filters.status || 'all'
    }
    const { data } = await api.get<ArticlesResponse>('/articles', { params })
    return data
  },

  getArticleById: async (id: string): Promise<Article> => {
    const { data } = await api.get<Article>(`/articles/${id}`)
    return data
  },

  republishArticle: async (id: string): Promise<Article> => {
    const { data } = await api.post<Article>(`/articles/${id}/republish`)
    return data
  },

  deleteArticle: async (id: string): Promise<void> => {
    await api.delete(`/articles/${id}`)
  }
}

// ============================================
// QUERY KEYS
// ============================================

const articleKeys = {
  all: ['articles'] as const,
  lists: () => [...articleKeys.all, 'list'] as const,
  list: (filters: ArticleFilters) => [...articleKeys.lists(), filters] as const,
  details: () => [...articleKeys.all, 'detail'] as const,
  detail: (id: string) => [...articleKeys.details(), id] as const
}

// ============================================
// HOOK
// ============================================

export function useArticles(filters: ArticleFilters = {}) {
  const queryClient = useQueryClient()
  const { updateUser } = useAuthStore()

  // ============================================
  // GET ARTICLES QUERY
  // ============================================

  const articlesQuery = useQuery({
    queryKey: articleKeys.list(filters),
    queryFn: () => articlesApi.getArticles(filters),
    staleTime: 30000, // 30 segundos
    refetchInterval: (data) => {
      // Auto-refetch se houver artigos em geração/publicação
      const hasActiveArticles = data?.articles.some(
        (article) => article.status === 'generating' || article.status === 'publishing'
      )
      return hasActiveArticles ? 5000 : false // 5 segundos se ativo, senão não refetch
    }
  })

  // ============================================
  // REPUBLISH MUTATION
  // ============================================

  const republishMutation = useMutation({
    mutationFn: articlesApi.republishArticle,
    onSuccess: (updatedArticle) => {
      // Atualizar cache
      queryClient.invalidateQueries({ queryKey: articleKeys.lists() })
      toast.success('Matéria reenviada para publicação!')
    },
    onError: (error) => {
      const message = getErrorMessage(error)
      toast.error(message || 'Erro ao republicar matéria')
    }
  })

  // ============================================
  // DELETE MUTATION
  // ============================================

  const deleteMutation = useMutation({
    mutationFn: articlesApi.deleteArticle,
    onSuccess: (_, deletedId) => {
      // Atualizar cache removendo o artigo
      queryClient.setQueryData<ArticlesResponse>(
        articleKeys.list(filters),
        (old) => {
          if (!old) return old
          return {
            ...old,
            articles: old.articles.filter((a) => a.id !== deletedId),
            total: old.total - 1
          }
        }
      )
      
      // Atualizar contador do usuário
      updateUser({ articlesUsed: (articlesQuery.data?.articles.length || 0) - 1 })
      
      toast.success('Matéria excluída com sucesso!')
    },
    onError: (error) => {
      const message = getErrorMessage(error)
      toast.error(message || 'Erro ao excluir matéria')
    }
  })

  // ============================================
  // HELPERS
  // ============================================

  const getArticlesByStatus = (status: ArticleStatus) => {
    return articlesQuery.data?.articles.filter((article) => article.status === status) || []
  }

  const hasActiveArticles = () => {
    return articlesQuery.data?.articles.some(
      (article) => article.status === 'generating' || article.status === 'publishing'
    ) || false
  }

  const republishArticle = (id: string) => {
    republishMutation.mutate(id)
  }

  const deleteArticle = (id: string) => {
    deleteMutation.mutate(id)
  }

  // ============================================
  // RETURN
  // ============================================

  return {
    // Data
    articles: articlesQuery.data?.articles || [],
    total: articlesQuery.data?.total || 0,
    page: articlesQuery.data?.page || 1,
    limit: articlesQuery.data?.limit || 10,
    
    // States
    isLoading: articlesQuery.isLoading,
    isError: articlesQuery.isError,
    error: articlesQuery.error,
    isRefetching: articlesQuery.isRefetching,
    
    // Actions
    republishArticle,
    deleteArticle,
    refetch: articlesQuery.refetch,
    
    // Mutation states
    isRepublishing: republishMutation.isPending,
    isDeleting: deleteMutation.isPending,
    
    // Helpers
    getArticlesByStatus,
    hasActiveArticles: hasActiveArticles(),
    isEmpty: articlesQuery.data?.articles.length === 0,
    
    // Computed
    publishedCount: getArticlesByStatus('published').length,
    errorCount: getArticlesByStatus('error').length,
    activeCount: getArticlesByStatus('generating').length + getArticlesByStatus('publishing').length
  }
}

// ============================================
// SINGLE ARTICLE HOOK
// ============================================

export function useArticle(id: string) {
  const articleQuery = useQuery({
    queryKey: articleKeys.detail(id),
    queryFn: () => articlesApi.getArticleById(id),
    enabled: !!id,
    staleTime: 10000 // 10 segundos
  })

  return {
    article: articleQuery.data,
    isLoading: articleQuery.isLoading,
    isError: articleQuery.isError,
    error: articleQuery.error,
    refetch: articleQuery.refetch
  }
}