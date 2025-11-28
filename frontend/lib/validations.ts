import { z } from "zod";

// ============================================
// PASSWORD VALIDATION
// ============================================

const passwordValidation = z
  .string()
  .min(8, "Senha deve ter no mínimo 8 caracteres")
  .max(100, "Senha muito longa")
  .regex(/[A-Z]/, "Senha deve conter pelo menos uma letra maiúscula")
  .regex(/[a-z]/, "Senha deve conter pelo menos uma letra minúscula")
  .regex(/[0-9]/, "Senha deve conter pelo menos um número");

// ============================================
// AUTH SCHEMAS
// ============================================

export const loginSchema = z.object({
  email: z.string().min(1, "Email é obrigatório").email("Email inválido"),
  password: z.string().min(1, "Senha é obrigatória"),
});

export const registerSchema = z
  .object({
    name: z
      .string()
      .min(2, "Nome deve ter no mínimo 2 caracteres")
      .max(100, "Nome muito longo")
      .regex(/^[a-zA-ZÀ-ÿ\s]+$/, "Nome deve conter apenas letras"),
    email: z.string().min(1, "Email é obrigatório").email("Email inválido"),
    password: passwordValidation,
    confirmPassword: z.string().min(1, "Confirmação de senha é obrigatória"),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "As senhas não coincidem",
    path: ["confirmPassword"],
  });

// ============================================
// WIZARD SCHEMAS
// ============================================

const objectiveEnum = z.enum(["leads", "sales", "branding"], {
  errorMap: () => ({ message: "Selecione um objetivo válido" }),
});

export const businessSchema = z
  .object({
    description: z
      .string()
      .min(10, "Descrição deve ter no mínimo 10 caracteres")
      .max(500, "Descrição deve ter no máximo 500 caracteres")
      .refine(
        (val) => val.trim().split(/\s+/).length >= 5,
        "Descrição deve ter pelo menos 5 palavras"
      ),

    primaryObjective: objectiveEnum,

    secondaryObjective: objectiveEnum.optional(),

    siteUrl: z.string().url("URL inválida").optional().or(z.literal("")),

    hasBlog: z.boolean().default(false),

    blogUrls: z.array(z.string().url("URL inválida")).refine((urls) => {
      const domains = urls.map((url) => new URL(url).hostname);
      return new Set(domains).size === domains.length;
    }, "URLs devem ser de domínios diferentes"),

    articleCount: z
      .number()
      .min(1, "Selecione pelo menos 1 matéria")
      .max(50, "Máximo de 50 matérias"),

    brandFile: z
      .instanceof(File)
      .refine(
        (file) => file.size <= 5 * 1024 * 1024,
        "Arquivo deve ter no máximo 5MB"
      )
      .refine(
        (file) =>
          ["application/pdf", "image/jpeg", "image/png"].includes(file.type),
        "Formato inválido. Use PDF, JPG ou PNG"
      )
      .refine((file) => {
        const ext = file.name.split(".").pop()?.toLowerCase();
        return ["pdf", "jpg", "jpeg", "png"].includes(ext || "");
      }, "Extensão de arquivo inválida")
      .optional(),
  })
  .refine(
    (data) => {
      if (
        data.secondaryObjective &&
        data.secondaryObjective === data.primaryObjective
      ) {
        return false;
      }
      return true;
    },
    {
      message: "Objetivo secundário deve ser diferente do primário",
      path: ["secondaryObjective"],
    }
  )
  .refine(
    (data) => {
      if (data.hasBlog && data.blogUrls.length === 0) {
        return false;
      }
      return true;
    },
    {
      message: "Adicione pelo menos uma URL do blog",
      path: ["blogUrls"],
    }
  );

export const competitorsSchema = z.object({
  competitorUrls: z
    .array(z.string().url("URL inválida"))
    .max(10, "Máximo de 10 concorrentes")
    .default([]),
});

export const integrationsSchema = z
  .object({
    wordpress: z.object({
      siteUrl: z
        .string()
        .min(1, "URL do site é obrigatória")
        .url("URL inválida"),
      username: z
        .string()
        .min(1, "Nome de usuário é obrigatório")
        .max(100, "Nome de usuário muito longo"),
      appPassword: z
        .string()
        .min(1, "Senha de aplicativo é obrigatória")
        .max(100, "Senha muito longa"),
    }),

    searchConsole: z
      .object({
        enabled: z.boolean().default(false),
        propertyUrl: z.string().url("URL inválida").optional(),
      })
      .optional(),

    analytics: z
      .object({
        enabled: z.boolean().default(false),
        measurementId: z
          .string()
          .regex(
            /^G-[A-Z0-9]+$/,
            "ID de medição inválido (formato: G-XXXXXXXXXX)"
          )
          .optional(),
      })
      .optional(),
  })
  .refine(
    (data) => {
      if (data.searchConsole?.enabled && !data.searchConsole.propertyUrl) {
        return false;
      }
      return true;
    },
    {
      message: "URL da propriedade é obrigatória",
      path: ["searchConsole", "propertyUrl"],
    }
  )
  .refine(
    (data) => {
      if (data.analytics?.enabled && !data.analytics.measurementId) {
        return false;
      }
      return true;
    },
    {
      message: "ID de medição é obrigatório",
      path: ["analytics", "measurementId"],
    }
  );

// ============================================
// NEW ARTICLES SCHEMA
// ============================================

export const newArticlesSchema = z.object({
  articleCount: z
    .number()
    .min(1, "Selecione pelo menos 1 matéria")
    .max(50, "Máximo de 50 matérias"),
});

// ============================================
// ARTICLE IDEA SCHEMA
// ============================================

export const articleIdeaSchema = z.object({
  id: z.string(),
  title: z.string().min(1, "Título é obrigatório"),
  summary: z.string().min(1, "Resumo é obrigatório"),
  approved: z.boolean().default(false),
  feedback: z
    .string()
    .max(500, "Feedback deve ter no máximo 500 caracteres")
    .optional(),
});

export const publishPayloadSchema = z.object({
  articles: z
    .array(
      z.object({
        id: z.string(),
        feedback: z.string().max(500).optional(),
      })
    )
    .min(1, "Selecione pelo menos uma matéria para publicar"),
});

// ============================================
// ACCOUNT SCHEMAS
// ============================================

export const profileUpdateSchema = z.object({
  name: z
    .string()
    .min(2, "Nome deve ter no mínimo 2 caracteres")
    .max(100, "Nome muito longo")
    .regex(/^[a-zA-ZÀ-ÿ\s]+$/, "Nome deve conter apenas letras"),
});

export const integrationsUpdateSchema = z.object({
  wordpress: z
    .object({
      siteUrl: z.string().url("URL inválida"),
      username: z.string().min(1, "Nome de usuário é obrigatório"),
      appPassword: z.string().min(1, "Senha de aplicativo é obrigatória"),
    })
    .optional(),

  searchConsole: z
    .object({
      enabled: z.boolean(),
      propertyUrl: z.string().url("URL inválida").optional(),
    })
    .optional(),

  analytics: z
    .object({
      enabled: z.boolean(),
      measurementId: z
        .string()
        .regex(/^G-[A-Z0-9]+$/, "ID de medição inválido")
        .optional(),
    })
    .optional(),
});

// ============================================
// QUERY PARAMS SCHEMAS
// ============================================

export const articlesQuerySchema = z.object({
  page: z.coerce.number().min(1).default(1),
  limit: z.coerce.number().min(1).max(100).default(10),
  status: z
    .enum(["all", "generating", "publishing", "published", "error"])
    .default("all"),
});

// ============================================
// HELPER TYPES
// ============================================

export type LoginInput = z.infer<typeof loginSchema>;
export type RegisterInput = z.infer<typeof registerSchema>;
export type BusinessInput = z.infer<typeof businessSchema>;
export type CompetitorsInput = z.infer<typeof competitorsSchema>;
export type IntegrationsInput = z.infer<typeof integrationsSchema>;
export type NewArticlesInput = z.infer<typeof newArticlesSchema>;
export type ArticleIdeaInput = z.infer<typeof articleIdeaSchema>;
export type PublishPayloadInput = z.infer<typeof publishPayloadSchema>;
export type ProfileUpdateInput = z.infer<typeof profileUpdateSchema>;
export type IntegrationsUpdateInput = z.infer<typeof integrationsUpdateSchema>;
export type ArticlesQueryInput = z.infer<typeof articlesQuerySchema>;
