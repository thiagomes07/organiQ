// ============================================
// APP CONFIG
// ============================================

export const APP_NAME = 'organiQ'
export const APP_TAGLINE = 'Naturalmente Inteligente'
export const APP_URL = process.env.NEXT_PUBLIC_APP_URL || 'https://organiq.com.br'

// ============================================
// PAGINATION
// ============================================

export const DEFAULT_PAGE_SIZE = 10
export const MAX_PAGE_SIZE = 100

// ============================================
// FILE UPLOAD
// ============================================

export const MAX_FILE_SIZE = 5 * 1024 * 1024 // 5MB
export const ALLOWED_FILE_TYPES = ['application/pdf', 'image/jpeg', 'image/png']
export const ALLOWED_FILE_EXTENSIONS = ['.pdf', '.jpg', '.jpeg', '.png']

// ============================================
// VALIDATION LIMITS
// ============================================

export const BUSINESS_DESCRIPTION_MIN = 10
export const BUSINESS_DESCRIPTION_MAX = 500
export const FEEDBACK_MAX_LENGTH = 500
export const MAX_COMPETITOR_URLS = 10
export const MAX_BLOG_URLS = 20
export const PASSWORD_MIN_LENGTH = 6

// ============================================
// OBJECTIVES (mantém tipagem)
// ============================================

export const OBJECTIVES = [
  { value: 'leads', label: 'Gerar mais leads' },
  { value: 'sales', label: 'Vender mais online' },
  { value: 'branding', label: 'Aumentar reconhecimento da marca' },
] as const

export type ObjectiveValue = typeof OBJECTIVES[number]['value']

// ============================================
// ARTICLE STATUS
// ============================================

export const ARTICLE_STATUS = {
  GENERATING: { value: 'generating', label: 'Gerando...', color: 'yellow' },
  PUBLISHING: { value: 'publishing', label: 'Publicando...', color: 'blue' },
  PUBLISHED: { value: 'published', label: 'Publicado', color: 'green' },
  ERROR: { value: 'error', label: 'Erro', color: 'red' }
} as const

export const ARTICLE_STATUS_OPTIONS = Object.values(ARTICLE_STATUS)

// ============================================
// POLLING INTERVALS
// ============================================

export const POLLING_INTERVAL = {
  IDEAS_STATUS: 3000,      // 3 segundos
  PUBLISH_STATUS: 3000,    // 3 segundos
  ARTICLES_ACTIVE: 5000,   // 5 segundos (quando tem artigos em geração)
  PAYMENT_STATUS: 3000     // 3 segundos
}

// ============================================
// CACHE / STALE TIME
// ============================================

export const STALE_TIME = {
  ARTICLES: 30000,     // 30 segundos
  PLANS: Infinity,     // Nunca fica stale (planos raramente mudam)
  CURRENT_PLAN: 60000, // 1 minuto
  USER: 60000          // 1 minuto
}

// ============================================
// ROUTES
// ============================================

export const ROUTES = {
  PUBLIC: {
    HOME: '/',
    LOGIN: '/login'
  },
  PROTECTED: {
    PLANS: '/app/planos',
    ONBOARDING: '/app/onboarding',
    NEW_ARTICLES: '/app/novo',
    ARTICLES: '/app/materias',
    ACCOUNT: '/app/conta'
  }
} as const

// ============================================
// EXTERNAL LINKS
// ============================================

export const EXTERNAL_LINKS = {
  WORDPRESS_APP_PASSWORD: 'https://wordpress.org/support/article/application-passwords/',
  GOOGLE_SEARCH_CONSOLE: 'https://search.google.com/search-console',
  GOOGLE_ANALYTICS: 'https://analytics.google.com/'
} as const