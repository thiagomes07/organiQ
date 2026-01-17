import { z } from "zod";

// ============================================
// PASSWORD VALIDATION
// ============================================

const passwordValidation = z
  .string()
  .min(6, "Senha deve ter no mínimo 6 caracteres")
  .max(100, "Senha muito longa")
  ;

// ============================================
// AUTH SCHEMAS
// ============================================

export const loginSchema = z.object({
  email: z.string().min(1, "Email é obrigatório").email("Email inválido"),
  password: z
    .string()
    .min(1, "Senha é obrigatória")
    .min(6, "Senha deve ter no mínimo 6 caracteres"),
});

export const registerSchema = z.object({
  name: z
    .string()
    .min(2, "Nome deve ter no mínimo 2 caracteres")
    .max(100, "Nome muito longo")
    .regex(/^[a-zA-ZÀ-ÿ\s]+$/, "Nome deve conter apenas letras"),
  email: z.string().min(1, "Email é obrigatório").email("Email inválido"),
  password: passwordValidation,
});

// ============================================
// LOCATION SCHEMAS (CORRIGIDO - country agora é obrigatório)
// ============================================

export const businessUnitSchema = z.object({
  id: z.string().uuid("ID inválido"),
  name: z.string().optional(),
  country: z.string().min(1, "Selecione o país desta unidade"),
  state: z.string().min(1, "Selecione o estado desta unidade"),
  city: z.string().min(1, "Selecione a cidade desta unidade"),
  isPrimary: z.boolean().optional(),
});

export const locationSchema = z
  .object({
    country: z.string().min(1, "Selecione o país onde seu negócio atua"),
    state: z.string().optional(),
    city: z.string().optional(),
    hasMultipleUnits: z.boolean(),
    units: z.array(businessUnitSchema).optional(),
  })
  .superRefine((data, ctx) => {
    // 1. Validação para SINGLE UNIT (Digital ou Física)
    if (!data.hasMultipleUnits) {
      // Se preencheu cidade, OBRIGATORIAMENTE tem que ter estado
      if (data.city && !data.state) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: "Para informar a cidade, você precisa selecionar o estado primeiro",
          path: ["state"],
        });
      }

      // Se preencheu estado, OBRIGATORIAMENTE tem que ter cidade (Físico)
      // "Se seu negócio é 100% digital, preencha apenas o País"
      // Se não é digital (tem estado), então é físico (precisa de cidade)
      if (data.state && !data.city) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: "Você selecionou um estado. Por favor, selecione também a cidade onde seu negócio atua",
          path: ["city"],
        });
      }
      return;
    }

    // 2. Validação para MÚLTIPLAS UNIDADES
    if (data.hasMultipleUnits) {
      if (!data.units || data.units.length === 0) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: "Você marcou que tem múltiplas unidades. Adicione pelo menos uma unidade clicando no botão abaixo",
          path: ["units"],
        });
        return;
      }

      if (data.units.length > 10) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: "Você pode cadastrar no máximo 10 unidades. Remova algumas para continuar",
          path: ["units"],
        });
      }

      // Valida se existe no máximo 1 unidade principal
      const primaryCount = data.units.filter((u) => u.isPrimary).length;
      if (primaryCount > 1) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: "Você marcou mais de uma unidade como principal. Selecione apenas uma",
          path: ["units"],
        });
      }
    }
  });

// ============================================
// WIZARD SCHEMAS
// ============================================

const objectiveEnum = z.enum(["leads", "sales", "branding"], {
  message: "Selecione um objetivo válido",
});

const preprocessUrl = (val: unknown) => {
  if (typeof val !== "string") return val;
  const trimmed = val.trim();
  if (trimmed === "") return "";
  if (!trimmed.match(/^https?:\/\//)) {
    return `https://${trimmed}`;
  }
  return trimmed;
};

export const businessSchema = z
  .object({
    description: z
      .string()
      .min(10, "A descrição do seu negócio precisa ter pelo menos 10 caracteres")
      .max(500, "A descrição está muito longa. Use no máximo 500 caracteres")
      .refine(
        (val) => val.trim().split(/\s+/).length >= 5,
        "Descreva seu negócio com pelo menos 5 palavras para gerar conteúdo de qualidade"
      ),

    primaryObjective: objectiveEnum,

    secondaryObjective: objectiveEnum.optional(),

    location: locationSchema,

    siteUrl: z.preprocess(
      preprocessUrl,
      z.string().url("Digite uma URL válida (exemplo: seusite.com.br)").optional().or(z.literal(""))
    ),

    hasBlog: z.boolean(),

    blogUrls: z.array(
      z.preprocess(preprocessUrl, z.string().url("Digite uma URL válida para o blog"))
    ),

    articleCount: z
      .number()
      .min(1, "Selecione pelo menos 1 matéria para gerar")
      .max(50, "Você pode gerar no máximo 50 matérias por vez"),

    brandFile: z
      .instanceof(File)
      .refine(
        (file) => file.size <= 5 * 1024 * 1024,
        "O arquivo é muito grande. O tamanho máximo é 5MB"
      )
      .refine(
        (file) =>
          ["application/pdf", "image/jpeg", "image/png"].includes(file.type),
        "Formato de arquivo não suportado. Use PDF, JPG ou PNG"
      )
      .refine((file) => {
        const ext = file.name.split(".").pop()?.toLowerCase();
        return ["pdf", "jpg", "jpeg", "png"].includes(ext || "");
      }, "Extensão de arquivo inválida. Use .pdf, .jpg ou .png")
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
      message: "O objetivo secundário precisa ser diferente do objetivo principal",
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
      message: "Você marcou que tem um blog. Adicione pelo menos uma URL do blog",
      path: ["blogUrls"],
    }
  );

export const competitorsSchema = z.object({
  competitorUrls: z
    .array(z.string().url("URL inválida"))
    .max(10, "Máximo de 10 concorrentes"),
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
        enabled: z.boolean(),
        propertyUrl: z.string().url("URL inválida").optional(),
      })
      .optional(),

    analytics: z
      .object({
        enabled: z.boolean(),
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
export type PublishPayload = z.infer<typeof publishPayloadSchema>;
export type ProfileUpdateInput = z.infer<typeof profileUpdateSchema>;
export type IntegrationsUpdateInput = z.infer<typeof integrationsUpdateSchema>;
export type ArticlesQueryInput = z.infer<typeof articlesQuerySchema>;
export type BusinessUnitInput = z.infer<typeof businessUnitSchema>;
export type LocationInput = z.infer<typeof locationSchema>;