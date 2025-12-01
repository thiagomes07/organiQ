import axios, { AxiosError, AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import { toast } from 'sonner'

// ============================================
// TYPES
// ============================================

interface ApiErrorResponse {
  message?: string
  error?: string
  errors?: Record<string, string>
}

// ============================================
// CONFIGURAÇÃO BASE
// ============================================

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3001/api',
  timeout: 30000, // 30 segundos
  withCredentials: true, // Importante para cookies httpOnly
  headers: {
    'Content-Type': 'application/json'
  }
})

// ============================================
// REQUEST INTERCEPTOR
// ============================================

api.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // Adicionar timestamp para evitar cache em requisições específicas
    if (config.method === 'get') {
      config.params = {
        ...config.params,
        _t: Date.now()
      }
    }
    
    return config
  },
  (error: AxiosError) => {
    return Promise.reject(error)
  }
)

// ============================================
// RESPONSE INTERCEPTOR
// ============================================

let isRefreshing = false
let failedQueue: Array<{
  resolve: (value?: unknown) => void
  reject: (reason?: unknown) => void
}> = []

const processQueue = (error: AxiosError | null = null) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error)
    } else {
      prom.resolve()
    }
  })
  
  failedQueue = []
}

api.interceptors.response.use(
  (response: AxiosResponse) => {
    // Resposta bem-sucedida
    return response
  },
  async (error: AxiosError<ApiErrorResponse>) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & {
      _retry?: boolean
    }
    
    // ============================================
    // HANDLE 401 - TOKEN EXPIRADO
    // ============================================
    
    if (error.response?.status === 401 && !originalRequest._retry) {
      if (isRefreshing) {
        // Se já está refreshing, adiciona à fila
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject })
        })
          .then(() => {
            return api(originalRequest)
          })
          .catch((err) => {
            return Promise.reject(err)
          })
      }
      
      originalRequest._retry = true
      isRefreshing = true
      
      try {
        // Tenta fazer refresh do token
        await api.post('/auth/refresh')
        
        processQueue(null)
        
        // Retry a requisição original
        return api(originalRequest)
      } catch (refreshError) {
        processQueue(refreshError as AxiosError)
        
        // Redireciona para login
        if (typeof window !== 'undefined') {
          window.location.href = '/login'
        }
        
        return Promise.reject(refreshError)
      } finally {
        isRefreshing = false
      }
    }
    
    // ============================================
    // HANDLE OUTROS ERROS
    // ============================================
    
    // 403 - Forbidden (sem permissão)
    if (error.response?.status === 403) {
      toast.error('Você não tem permissão para realizar esta ação')
    }
    
    // 404 - Not Found
    if (error.response?.status === 404) {
      toast.error('Recurso não encontrado')
    }
    
    // 422 - Validation Error
    if (error.response?.status === 422) {
      const message = error.response.data?.message || 'Dados inválidos'
      toast.error(message)
    }
    
    // 429 - Rate Limit
    if (error.response?.status === 429) {
      toast.error('Muitas requisições. Tente novamente em alguns instantes')
    }
    
    // 500+ - Server Error
    if (error.response?.status && error.response.status >= 500) {
      toast.error('Erro no servidor. Tente novamente mais tarde')
    }
    
    // Timeout
    if (error.code === 'ECONNABORTED') {
      toast.error('A requisição demorou muito. Tente novamente')
    }
    
    // Network Error
    if (error.message === 'Network Error') {
      toast.error('Erro de conexão. Verifique sua internet')
    }
    
    return Promise.reject(error)
  }
)

// ============================================
// HELPER FUNCTIONS
// ============================================

/**
 * Extrai mensagem de erro da resposta da API
 */
export const getErrorMessage = (error: unknown): string => {
  if (axios.isAxiosError(error)) {
    const axiosError = error as AxiosError<ApiErrorResponse>
    return (
      axiosError.response?.data?.message ||
      axiosError.response?.data?.error ||
      axiosError.message ||
      'Erro desconhecido'
    )
  }
  
  if (error instanceof Error) {
    return error.message
  }
  
  return 'Erro desconhecido'
}

/**
 * Verifica se é erro de validação (422)
 */
export const isValidationError = (error: unknown): boolean => {
  return axios.isAxiosError(error) && error.response?.status === 422
}

/**
 * Extrai erros de campo do erro de validação
 */
export const getFieldErrors = (error: unknown): Record<string, string> => {
  if (axios.isAxiosError(error)) {
    const axiosError = error as AxiosError<ApiErrorResponse>
    if (axiosError.response?.status === 422) {
      return axiosError.response.data?.errors || {}
    }
  }
  return {}
}

/**
 * Faz upload de arquivo com progress
 */
export const uploadFile = async (
  url: string,
  file: File,
  onProgress?: (progress: number) => void
) => {
  const formData = new FormData()
  formData.append('file', file)
  
  return api.post(url, formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    },
    onUploadProgress: (progressEvent) => {
      if (onProgress && progressEvent.total) {
        const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total)
        onProgress(progress)
      }
    }
  })
}

// ============================================
// EXPORT
// ============================================

export default api