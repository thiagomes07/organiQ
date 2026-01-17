// ============================================
// AUTH TYPES
// ============================================

export interface User {
  id: string;
  name: string;
  email: string;
  planId: string;
  planName: string;
  maxArticles: number;
  articlesUsed: number;
  hasCompletedOnboarding: boolean;
  onboardingStep: number;
  createdAt: string;
}

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface RegisterData {
  name: string;
  email: string;
  password: string;
}

export interface AuthResponse {
  user: User;
  message?: string;
}

// ============================================
// LOCATION TYPES
// ============================================

export interface BusinessUnit {
  id: string;
  name?: string;
  country: string;
  state?: string;
  city?: string;
}

export interface BusinessLocation {
  country: string;
  state?: string;
  city?: string;
  hasMultipleUnits: boolean;
  units?: BusinessUnit[];
}

// ============================================
// WIZARD / BUSINESS TYPES
// ============================================

export type ObjectiveType = "leads" | "sales" | "branding";

export interface BusinessInfo {
  description: string;
  primaryObjective: ObjectiveType;
  secondaryObjective?: ObjectiveType;
  location: BusinessLocation;
  siteUrl?: string;
  hasBlog: boolean;
  blogUrls: string[]; // REVERTIDO: deve ser obrigatório mas pode ser array vazio
  articleCount: number;
  brandFile?: File;
}

export interface CompetitorData {
  competitorUrls: string[];
}

export interface IntegrationsData {
  wordpress: {
    siteUrl: string;
    username: string;
    appPassword: string;
  };
  searchConsole?: {
    enabled: boolean;
    propertyUrl?: string;
  };
  analytics?: {
    enabled: boolean;
    measurementId?: string;
  };
}

export interface WizardState {
  currentStep: number;
  businessInfo?: BusinessInfo;
  competitorData?: CompetitorData;
  integrationsData?: IntegrationsData;
  articleIdeas?: ArticleIdea[];
}

// ============================================
// ARTICLES TYPES
// ============================================

export interface ArticleIdea {
  id: string;
  title: string;
  summary: string;
  approved: boolean;
  feedback?: string;
}

export type ArticleStatus = "generating" | "publishing" | "published" | "error";

export interface Article {
  id: string;
  title: string;
  createdAt: string;
  status: ArticleStatus;
  postUrl?: string;
  errorMessage?: string;
  content?: string;
}

export interface ArticlesResponse {
  articles: Article[];
  total: number;
  page: number;
  limit: number;
}

export interface ArticleFilters {
  page?: number;
  limit?: number;
  status?: ArticleStatus | "all";
}

export interface PublishPayload {
  articles: Array<{
    id: string;
    feedback?: string;
  }>;
}

// ============================================
// PLANS TYPES
// ============================================

export interface Plan {
  id: string;
  name: string;
  maxArticles: number;
  price: number;
  features: string[];
  recommended?: boolean;
}

export interface PlanInfo {
  name: string;
  maxArticles: number;
  articlesUsed: number;
  nextBillingDate: string;
  price: number;
}

export interface CheckoutResponse {
  checkoutUrl: string;
  sessionId: string;
}

export interface PaymentStatus {
  id: string;
  status: "pending" | "paid" | "failed";
  planId: string;
}

// ============================================
// ACCOUNT TYPES
// ============================================

export interface AccountPlanResponse {
  id: string;
  name: string;
  maxArticles: number;
  articlesUsed: number;
  remainingArticles: number;
  limitReached: boolean;
  price: number;
  active: boolean;
  features: string[];
  nextBillingDate?: string;
}

export interface ProfileUpdateData {
  name: string;
}

export interface IntegrationsUpdateData {
  wordpress?: {
    siteUrl: string;
    username: string;
    appPassword: string;
  };
  searchConsole?: {
    enabled: boolean;
    propertyUrl?: string;
  };
  analytics?: {
    enabled: boolean;
    measurementId?: string;
  };
}

// ============================================
// API RESPONSE TYPES
// ============================================

export interface ApiError {
  message: string;
  code?: string;
  field?: string;
}

export interface ApiResponse<T = unknown> {
  data?: T;
  error?: ApiError;
  success: boolean;
}

// ============================================
// FORM TYPES
// ============================================

export interface LoginForm {
  email: string;
  password: string;
}

export interface RegisterForm {
  name: string;
  email: string;
  password: string;
}

export interface BusinessForm {
  description: string;
  primaryObjective: ObjectiveType;
  secondaryObjective?: ObjectiveType;
  location: BusinessLocation;
  siteUrl?: string;
  hasBlog: boolean;
  blogUrls: string[]; // REVERTIDO: deve ser obrigatório mas pode ser array vazio
  articleCount: number;
  brandFile?: File;
}

export interface CompetitorsForm {
  competitorUrls: string[];
}

export interface IntegrationsForm {
  wordpress: {
    siteUrl: string;
    username: string;
    appPassword: string;
  };
  searchConsole: {
    enabled: boolean;
    propertyUrl?: string;
  };
  analytics: {
    enabled: boolean;
    measurementId?: string;
  };
}

export interface NewArticlesForm {
  articleCount: number;
}

// ============================================
// UTILITY TYPES
// ============================================

export type LoadingState = "idle" | "loading" | "success" | "error";

export interface PaginationState {
  page: number;
  limit: number;
  total: number;
}
