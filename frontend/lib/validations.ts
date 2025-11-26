import { z } from 'zod'

// ============================================
// AUTH SCHEMAS
// ============================================

export const loginSchema = z.object({
  email: z
    .string()
    .min(1, 'Email é obrigatório')
    .email('Email inválido'),
  password: z
    .string()
    .min(6, 'Senha deve ter no mínimo 6 caracteres')
    .max(100, 'Senha muito longa')
})

export const registerSchema = z.object({
  name: z
    .string()
    .min(2, 'Nome deve ter no mínimo 2 caracteres')
    .max(100, 'Nome muito longo')
    .regex(/^[a-zA-ZÀ-ÿ\s]+$/, 'Nome deve conter apenas letras'),
  email: z
    .string()
    .min(1, 'Email é obrigatório')
    .email('Email inválido'),
  password: z
    .string()
    .min(6, 'Senha deve ter no mínimo 6 caracteres')
    .max(100, 'Senha muito longa')
})

// ============================================
// WIZARD SCHEMAS
// ============================================

const objectiveEnum = z.enum(['leads', 'sales', 'branding'], {
  errorMap: () => ({ message: 'Selecione um objetivo válido' })
})

export const businessSchema = z.object({
  description: z
    .string()
    .min(10, 'Descrição deve ter no mínimo 10 caracteres')
    .max(500, 'Descrição deve ter no máximo 500 caracteres'),
  
  primaryObjective: objectiveEnum,
  
  secondaryObjective: objectiveEnum.optional(),
  
  siteUrl: z
    .string()
    .url('URL inválida')
    .optional()
    .or(z.literal('')),
  
  hasBlog: z.boolean().default(false),
  
  blogUrls: z
    .array(z.string().url('URL inválida'))
    .default([]),
  
  articleCount: z
    .number()
    .min(1, 'Selecione pelo menos 1 matéria')
    .max(50, 'Máximo de 50 matérias'),
  
  brandFile: z
    .instanceof(File)
    .refine(
      (file) => file.size <= 5 * 1024 * 1024,
      'Arquivo deve ter no máximo 5MB'
    )
    .refine(
      (file) => ['application/pdf', 'image/jpeg', 'image/png'].includes(file.type),
      'Formato inválido. Use PDF, JPG ou PNG'
    )
    .optional()
}).refine(
  (data) => {
    // Validar que o objetivo secundário é diferente do primário
    if (data.secondaryObjective && data.secondaryObjective === data.primaryObjective) {
      return false
    }
    return true
  },
  {
    message: 'Objetivo secundário deve ser diferente do primário',
    path: ['secondaryObjective']
  }
).refine(
  (data) => {
    // Se tem blog, deve ter pelo menos uma URL
    if (data.hasBlog && data.blogUrls.length === 0) {
      return false
    }
    return true
  },
  {
    message: 'Adicione pelo menos uma URL do blog',
    path: ['blogUrls']
  }
)

export const competitorsSchema = z.object({
  competitorUrls: z
    .array(z.string().url('URL inválida'))
    .max(10, 'Máximo de 10 concorrentes')
    .default([])
})

export const integrationsSchema = z.object({
  wordpress: z.object({
    siteUrl: z
      .string()
      .min(1, 'URL do site é obrigatória')
      .url('URL inválida'),
    username: z
      .string()
      .min(1, 'Nome de usuário é obrigatório')
      .max(100, 'Nome de usuário muito longo'),
    appPassword: z
      .string()
      .min(1, 'Senha de aplicativo é obrigatória')
      .max(100, 'Senha muito longa')
  }),
  
  searchConsole: z.object({
    enabled: z.boolean().default(false),
    propertyUrl: z
      .string()
      .url('URL inválida')
      .optional()
  }).optional(),
  
  analytics: z.object({
    enabled: z.boolean().default(false),
    measurementId: z
      .string()
      .regex(/^G-[A-Z0-9]+$/, 'ID de medição inválido (formato: G-XXXXXXXXXX)')
      .optional()
  }).optional()
}).refine(
  (data) => {
    // Se Search Console ativado, deve ter URL
    if (data.searchConsole?.enabled && !data.searchConsole.propertyUrl) {
      return false
    }
    return true
  },
  {
    message: 'URL da propriedade é obrigatória',
    path: ['searchConsole', 'propertyUrl']
  }
).refine(
  (data) => {
    // Se Analytics ativado, deve ter ID
    if (data.analytics?.enabled && !data.analytics.measurementId) {
      return false
    }
    return true
  },
  {
    message: 'ID de medição é obrigatório',
    path: ['analytics', 'measurementId']
  }
)

// ============================================
// NEW ARTICLES SCHEMA
// ============================================

export const newArticlesSchema = z.object({
  articleCount: z
    .number()
    .min(1, 'Selecione pelo menos 1 matéria')
    .max(50, 'Máximo de 50 matérias')
})

// ============================================
// ARTICLE IDEA SCHEMA
// ============================================

export const articleIdeaSchema = z.object({
  id: z.string(),
  title: z.string().min(1, 'Título é obrigatório'),
  summary: z.string().min(1, 'Resumo é obrigatório'),
  approved: z.boolean().default(false),
  feedback: z
    .string()
    .max(500, 'Feedback deve ter no máximo 500 caracteres')
    .optional()
})

export const publishPayloadSchema = z.object({
  articles: z
    .array(
      z.object({
        id: z.string(),
        feedback: z.string().max(500).optional()
      })
    )
    .min(1, 'Selecione pelo menos uma matéria para publicar')
})

// ============================================
// ACCOUNT SCHEMAS
// ============================================

export const profileUpdateSchema = z.object({
  name: z
    .string()
    .min(2, 'Nome deve ter no mínimo 2 caracteres')
    .max(100, 'Nome muito longo')
    .regex(/^[a-zA-ZÀ-ÿ\s]+$/, 'Nome deve conter apenas letras')
})

export const integrationsUpdateSchema = z.object({
  wordpress: z.object({
    siteUrl: z.string().url('URL inválida'),
    username: z.string().min(1, 'Nome de usuário é obrigatório'),
    appPassword: z.string().min(1, 'Senha de aplicativo é obrigatória')
  }).optional(),
  
  searchConsole: z.object({
    enabled: z.boolean(),
    propertyUrl: z.string().url('URL inválida').optional()
  }).optional(),
  
  analytics: z.object({
    enabled: z.boolean(),
    measurementId: z
      .string()
      .regex(/^G-[A-Z0-9]+$/, 'ID de medição inválido')
      .optional()
  }).optional()
})

// ============================================
// QUERY PARAMS SCHEMAS
// ============================================

export const articlesQuerySchema = z.object({
  page: z.coerce.number().min(1).default(1),
  limit: z.coerce.number().min(1).max(100).default(10),
  status: z.enum(['all', 'generating', 'publishing', 'published', 'error']).default('all')
})

// ============================================
// HELPER TYPES
// ============================================

export type LoginInput = z.infer<typeof loginSchema>
export type RegisterInput = z.infer<typeof registerSchema>
export type BusinessInput = z.infer<typeof businessSchema>
export type CompetitorsInput = z.infer<typeof competitorsSchema>
export type IntegrationsInput = z.infer<typeof integrationsSchema>
export type NewArticlesInput = z.infer<typeof newArticlesSchema>
export type ArticleIdeaInput = z.infer<typeof articleIdeaSchema>
export type PublishPayloadInput = z.infer<typeof publishPayloadSchema>
export type ProfileUpdateInput = z.infer<typeof profileUpdateSchema>
export type IntegrationsUpdateInput = z.infer<typeof integrationsUpdateSchema>
export type ArticlesQueryInput = z.infer<typeof articlesQuerySchema>