# organiQ - Especificação Frontend Completa

## Stack Técnica

### Framework e Bibliotecas
- **Next.js 14+** (App Router)
- **TypeScript**
- **TailwindCSS** + **shadcn/ui** (customizado)
- **React Hook Form** + **Zod**
- **TanStack Query** (React Query)
- **Zustand** (estado global de auth)
- **Axios** (com interceptors)

### Autenticação
- **Cookies httpOnly, secure, sameSite: strict**
- Access token (15min) + Refresh token (7d)
- Middleware Next.js para proteção de rotas
- Interceptor Axios para refresh automático

---

## Identidade Visual

### Tipografia
- **Primária:** "All Round Gothic" (headings, CTAs)
- **Secundária:** "Onest" (corpo, labels)

### Paleta de Cores
```css
/* Primárias */
--primary-dark: #001d47;
--primary-purple: #551bfa;
--primary-teal: #004563;

/* Secundárias */
--secondary-yellow: #faf01b;
--secondary-dark: #282828;
--secondary-cream: #fffde1;

/* Sistema */
--success: #10b981;
--error: #ef4444;
--warning: #f59e0b;
```

### Princípios de Design
- Espaçamentos: 16px, 24px, 32px
- Border radius: 6px (inputs), 8px (cards), 12px (modais)
- Sombras sutis (shadow-sm, shadow-md)
- Transições suaves (duration-200)
- Background principal: `secondary-cream`

---

## Estrutura de Rotas

```
/ (público)                    → Landing page
/login (público)               → Login/Cadastro

/app (protegido)               → Layout com sidebar
  /app/planos (onboarding)     → Escolha de plano (primeiro acesso)
  /app/onboarding (onboarding) → Wizard completo (primeiro acesso)
  /app/novo (recorrente)       → Wizard simplificado
  /app/materias                → Dashboard de publicações
  /app/conta                   → Configurações e plano
```

---

## Tipos TypeScript Globais

```typescript
// Auth
interface User {
  id: string;
  name: string;
  email: string;
  planId: string;
  planName: string;
  maxArticles: number;
  articlesUsed: number;
  hasCompletedOnboarding: boolean;
  createdAt: string;
}

interface LoginCredentials {
  email: string;
  password: string;
}

interface RegisterData {
  name: string;
  email: string;
  password: string;
}

// Wizard/Business
interface BusinessInfo {
  description: string;
  primaryObjective: 'leads' | 'sales' | 'branding';
  secondaryObjective?: 'leads' | 'sales' | 'branding';
  siteUrl?: string;
  hasBlog: boolean;
  blogUrls: string[];
  brandFile?: File;
}

interface CompetitorData {
  competitorUrls: string[];
}

interface IntegrationsData {
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

// Articles
interface ArticleIdea {
  id: string;
  title: string;
  summary: string;
  approved: boolean;
  feedback?: string;
}

interface Article {
  id: string;
  title: string;
  createdAt: string;
  status: 'generating' | 'publishing' | 'published' | 'error';
  postUrl?: string;
  errorMessage?: string;
  content?: string;
}

// Plans
interface Plan {
  id: string;
  name: string;
  maxArticles: number;
  price: number;
  features: string[];
}
```

---

## 1. Landing Page (`/`)

### Layout
Hero section centralizado + Features + CTA

### Componentes
- **Header**: Logo "organiQ" + Button "Entrar" (link para /login)
- **Hero**:
  - H1: "Aumente seu tráfego orgânico com IA" (All Round Gothic, `primary-dark`)
  - Subtitle: "Matérias de blog que geram autoridade e SEO" (Onest, `primary-teal`)
  - Tagline: "Naturalmente Inteligente"
- **Features Grid** (3 cards):
  - "Geração Automática de Conteúdo"
  - "SEO Otimizado"
  - "Publicação Direta no WordPress"
  - Design: background branco, borda `primary-teal` opacity-20
- **CTA Button**: "Criar minha conta grátis" → `/login`
  - Background: `secondary-yellow`
  - Text: `primary-dark` (All Round Gothic)
  - Hover: scale(1.05) + shadow-lg

---

## 2. Login/Cadastro (`/login`)

### Layout
Card centralizado (440px) em fundo `secondary-cream`

### Tabs
- **"Entrar"** | **"Cadastrar"**
- Ativa: `primary-purple`, underline animado

### Form "Entrar"
```typescript
interface LoginForm {
  email: string;    // Zod: email válido
  password: string; // Zod: min 6 chars
}
```
- Button "Entrar" (background `primary-purple`)
- Link "Esqueci minha senha" (disabled, `primary-teal`)

### Form "Cadastrar"
```typescript
interface RegisterForm {
  name: string;     // Zod: min 2 chars
  email: string;    // Zod: email válido
  password: string; // Zod: min 6 chars
}
```
- Button "Criar conta" (background `secondary-yellow`)

### Fluxo Pós-Login
```
Login bem-sucedido
  ├─ Se hasCompletedOnboarding: false → /app/planos
  └─ Se hasCompletedOnboarding: true → /app/materias
```

---

## 3. Layout Protegido (`/app`)

### Middleware (middleware.ts)
```typescript
// Verificar cookie de auth
// Se não autenticado → redirect /login
// Se não completou onboarding → redirect /app/planos ou /app/onboarding
```

### Sidebar (Desktop)
- Width: 280px
- Background: branco, border-radius: 12px
- Margin: 16px, shadow-md
- **Logo** no topo
- **Menu Items**:
  - "Gerar Matérias" → `/app/novo`
  - "Minhas Matérias" → `/app/materias`
  - "Minha Conta" → `/app/conta`
  - "Sair" (logout)
- Item ativo: background `primary-purple` opacity-10, border-left 3px

### Mobile
- Sidebar vira bottom navigation fixa

---

## 4. Escolha de Plano (`/app/planos`)

### Quando Exibir
- Apenas no **primeiro acesso** após cadastro
- Flag `hasCompletedOnboarding: false`

### Layout
- Grid de 3 cards de planos (horizontal em desktop, vertical em mobile)

### Card de Plano
```typescript
// Exemplo de planos
const plans: Plan[] = [
  {
    id: 'starter',
    name: 'Starter',
    maxArticles: 5,
    price: 49.90,
    features: ['5 matérias/mês', 'SEO básico', 'Suporte email']
  },
  {
    id: 'pro',
    name: 'Pro',
    maxArticles: 15,
    price: 99.90,
    features: ['15 matérias/mês', 'SEO avançado', 'Suporte prioritário']
  },
  {
    id: 'enterprise',
    name: 'Enterprise',
    maxArticles: 50,
    price: 249.90,
    features: ['50 matérias/mês', 'SEO premium', 'Suporte dedicado']
  }
]
```

**Design do Card:**
- Border: 2px solid `primary-teal` (plano recomendado tem badge "Recomendado")
- Hover: shadow-lg + scale(1.02)
- Button "Escolher Plano" (background `secondary-yellow`)

### Fluxo de Pagamento
1. Clicar "Escolher Plano" → POST `/api/payments/create-checkout`
2. Backend retorna URL de checkout (Stripe/Mercado Pago)
3. Redirect para gateway
4. Webhook confirma pagamento
5. Frontend faz polling em `/api/payments/status/:id`
6. Quando `status: 'paid'` → Redirect `/app/onboarding`

---

## 5. Wizard Onboarding Completo (`/app/onboarding`)

### Quando Exibir
- Apenas no **primeiro acesso**, após pagamento do plano

### Stepper
```
[1] Negócio → [2] Concorrentes → [3] Integrações → [Loading] → [4] Aprovação → [Loading]
```

### Passo 1: Informações do Negócio

**Campos:**
```typescript
interface BusinessForm {
  description: string;              // Textarea, max 500 chars
  primaryObjective: string;         // Select: 'leads' | 'sales' | 'branding'
  secondaryObjective?: string;      // Select: mesmas opções (exceto a primária)
  siteUrl?: string;                // Input, validação URL
  hasBlog: boolean;                // Checkbox
  blogUrls: string[];              // Array de inputs (se hasBlog)
  articleCount: number;            // Slider (1 a maxArticles)
  brandFile?: File;                // Upload (.pdf, .jpg, .png, max 5MB)
}
```

**Campo "Objetivos":**
- Label: "Quais são seus objetivos?" com asterisco
- **Objetivo Principal (obrigatório):**
  - Select com opções:
    - "Gerar mais leads"
    - "Vender mais online"
    - "Aumentar reconhecimento da marca"
  - Placeholder: "Selecione seu objetivo principal"
  
- **Objetivo Secundário (opcional):**
  - Aparece após selecionar o primário
  - Select com as mesmas opções (exceto a selecionada no primário)
  - Placeholder: "Selecione um objetivo secundário (opcional)"
  - Texto auxiliar: "Um objetivo secundário ajuda a criar conteúdo mais diversificado"

**Validações Zod:**
- description: min 10, max 500 chars
- primaryObjective: required, enum
- secondaryObjective: optional, enum, different from primaryObjective
- siteUrl: URL válida ou undefined
- blogUrls: array de URLs válidas
- articleCount: between 1 and user.maxArticles

**Button:** "Próximo" (background `secondary-yellow`)

### Passo 2: Concorrentes

**Campos:**
```typescript
interface CompetitorsForm {
  competitorUrls: string[]; // Min 0, max 10
}
```

- Lista dinâmica: "+ Adicionar concorrente"
- Cada URL com botão "Remover"
- Texto auxiliar: "Opcional, mas recomendado para melhor estratégia de SEO"

**Buttons:** "Voltar" | "Próximo"

### Passo 3: Integrações

**Campos:**
```typescript
interface IntegrationsForm {
  // WordPress (obrigatório)
  wordpress: {
    siteUrl: string;      // Input, validação URL
    username: string;     // Input, required
    appPassword: string;  // Input, type password, required
  };
  
  // Google Search Console (opcional)
  searchConsole: {
    enabled: boolean;     // Toggle
    propertyUrl?: string; // Input, validação URL, aparece se enabled
  };
  
  // Google Analytics (opcional)
  analytics: {
    enabled: boolean;        // Toggle
    measurementId?: string;  // Input, formato GA4 (G-XXXXXXXXXX)
  };
}
```

**Layout:**
Três seções expansíveis (accordion) no mesmo passo:

**1. WordPress (Obrigatório - Sempre expandido)**
- Mesmos campos anteriores
- Dialog de Ajuda para appPassword

**2. Google Search Console (Opcional)**
- Toggle "Conectar Search Console"
- Se ativado, mostra:
  - Input "URL da propriedade" (ex: https://seusite.com)
  - Dialog de Ajuda com instruções OAuth/Service Account
  - Texto: "Permite análise de palavras-chave e rankings"

**3. Google Analytics (Opcional)**
- Toggle "Conectar Analytics"
- Se ativado, mostra:
  - Input "ID de Medição GA4" (formato: G-XXXXXXXXXX)
  - Dialog de Ajuda sobre onde encontrar o ID
  - Texto: "Permite análise de tráfego e conversões"

**Visual:**
- Cards com borda para cada integração
- WordPress: borda `primary-purple` (obrigatório)
- Search Console/Analytics: borda `primary-teal` (opcional)
- Ícones de cada serviço ao lado dos títulos

**Buttons:** "Voltar" | "Gerar Ideias"

### Loading: Gerando Ideias

**Layout:** Modal centralizado, não pode fechar

**Design:**
- Spinner animado (`primary-purple` + `primary-teal`)
- Textos alternando a cada 3s:
  - "Analisando seus concorrentes..."
  - "Mapeando tópicos de autoridade..."
  - "Gerando ideias de matérias..."
  - "Isso pode levar alguns minutos"

**Backend:** POST `/api/wizard/generate-ideas`
**Polling:** GET `/api/wizard/ideas-status/:id` a cada 3s

### Passo 4: Aprovação de Matérias

**Layout:** Grid de cards (2 colunas desktop, 1 mobile)

**Card de Ideia:**
```typescript
interface ArticleIdeaCard {
  id: string;
  title: string;
  summary: string;
  approved: boolean;
  feedback?: string;  // Novo campo
}
```

**Design do Card:**
- Título (All Round Gothic, `primary-dark`)
- Resumo (3 linhas, ellipsis)
- Toggle buttons:
  - "Aprovar" (verde, outline) | "Rejeitar" (cinza)
  - Estado selecionado: background filled

**Campo de Feedback (Novo):**
- Aparece sempre, independente do estado
- Textarea expansível (min 2 linhas, max 4 linhas)
- Placeholder: "Adicione sugestões ou direcionamentos para esta matéria (opcional)"
- Character count: "0 / 500"
- Border: `primary-teal` opacity-30
- Focus: border `primary-purple`
- Design:
  - Ícone de mensagem ao lado do label
  - Subtle background `secondary-cream` opacity-50
  - Aparece abaixo dos botões de aprovação

**Comportamento do Feedback:**
- Salvo automaticamente (debounce 1s)
- Persiste mesmo se mudar de aprovado para rejeitado
- Enviado junto com os IDs aprovados no publish
- Se preenchido em matéria rejeitada, mostra badge "Feedback enviado" (para caso queira usar depois)

**Visual do Card:**
- Matéria aprovada + com feedback: borda esquerda verde + ícone de check + badge "Com direcionamento"
- Matéria aprovada sem feedback: borda esquerda verde + ícone de check
- Matéria rejeitada: opacity-60 + borda cinza

**Footer:**
- Contador: "X matérias aprovadas"
- Badge secundário: "Y feedbacks adicionados"
- Button "Publicar X Matérias Aprovadas" (disabled se nenhuma aprovada)
  - Background: `primary-purple`
  - Tooltip se hover e nenhuma aprovada: "Aprove pelo menos uma matéria"

**Payload de Publicação:**
```typescript
interface PublishPayload {
  articles: Array<{
    id: string;
    feedback?: string;
  }>;
}
```

### Loading: Publicando

**Layout:** Similar ao anterior

**Texto:** "Escrevendo e publicando no WordPress..."

**Backend:** POST `/api/wizard/publish`
```typescript
interface PublishPayload {
  approvedIds: string[];
}
```

**Após Conclusão:**
- Atualizar `hasCompletedOnboarding: true`
- Redirect `/app/materias`
- Toast: "X matérias publicadas com sucesso!"

---

## 6. Wizard Simplificado (`/app/novo`)

### Quando Exibir
- **Acessos subsequentes** (hasCompletedOnboarding: true)

### Stepper Reduzido
```
[1] Quantidade → [2] Concorrentes → [Loading] → [3] Aprovação → [Loading]
```

### Passo 1: Quantidade de Matérias

**Campos:**
```typescript
interface NewArticlesForm {
  articleCount: number; // Slider (1 a maxArticles - articlesUsed)
}
```

**Validação:**
- Se `articlesUsed >= maxArticles`: mostrar alerta "Limite atingido" + link para `/app/conta` (upgrade)

**Button:** "Próximo"

### Passo 2: Concorrentes (Opcional)

- Mesmos campos do onboarding
- **Pré-preenchido** com URLs do banco
- Usuário pode editar ou pular

**Buttons:** "Voltar" | "Gerar Ideias"

### Restante do Fluxo
- Igual ao onboarding (Loading → Aprovação → Publicação)

---

## 7. Dashboard de Matérias (`/app/materias`)

### Layout
- Header: "Minhas Matérias" + Button "Gerar Novas" → `/app/novo`
- Filtros: Status (todos, publicadas, erro, gerando)
- Tabela/Cards responsivos
- Paginação (10 itens/página)

### Colunas da Tabela
1. **Título** (truncado com tooltip)
2. **Data** (formato dd/MM/yyyy HH:mm)
3. **Status** (Badge):
   - `generating`: amarelo "Gerando..."
   - `publishing`: azul "Publicando..."
   - `published`: verde "Publicado"
   - `error`: vermelho "Erro"
4. **Ações**:
   - Se `published`: Button "Ver Post" (link externo)
   - Se `error`: Button "Ver Detalhes"

### Modal de Erro

```typescript
interface ErrorModal {
  title: string;
  errorMessage: string;
  content: string; // Conteúdo gerado que falhou
}
```

**Componentes:**
- Textarea readonly com conteúdo
- Button "Copiar Conteúdo" (clipboard API)
- Button "Tentar Republicar" (re-envia para WordPress)
- Button "Fechar"

### Estados
- **Loading:** Skeleton de 5 linhas
- **Empty State:** 
  - Ilustração
  - "Nenhuma matéria criada ainda"
  - Button "Criar Primeira Matéria" → `/app/novo` ou `/app/onboarding`

### Endpoint
```typescript
GET /api/articles?page=1&limit=10&status=all

Response: {
  articles: Article[];
  total: number;
  page: number;
  limit: number;
}
```

---

## 8. Minha Conta (`/app/conta`)

### Layout
Três cards verticais

### Card 1: Perfil

**Campos:**
```typescript
interface ProfileForm {
  name: string;  // Zod: min 2 chars
  email: string; // Disabled (não editável)
}
```

**Button:** "Salvar Alterações" (background `primary-purple`)

**Endpoint:** PATCH `/api/account/profile`

### Card 2: Integrações

**WordPress (obrigatório):**
```typescript
interface WordPressUpdateForm {
  siteUrl: string;
  username: string;
  appPassword: string; // Type password
}
```

**Pre-fill:** Dados salvos (appPassword como bullets "••••••")

**Ícone "?":** Mesmo Dialog de ajuda

**Google Search Console (opcional):**
```typescript
interface SearchConsoleUpdateForm {
  enabled: boolean;
  propertyUrl?: string;
}
```

**Pre-fill:** Se já conectado, mostrar URL da propriedade

**Google Analytics (opcional):**
```typescript
interface AnalyticsUpdateForm {
  enabled: boolean;
  measurementId?: string;
}
```

**Pre-fill:** Se já conectado, mostrar ID de medição

**Layout:**
- Accordion com 3 seções (WordPress, Search Console, Analytics)
- Cada seção pode ser expandida/colapsada
- Indicador visual se integração está ativa (badge verde "Conectado")

**Button:** "Atualizar Integrações"

**Endpoint:** PATCH `/api/account/integrations`

### Card 3: Meu Plano

**Informações Exibidas:**
```typescript
interface PlanInfo {
  name: string;
  maxArticles: number;
  articlesUsed: number;
  nextBillingDate: string;
  price: number;
}
```

**Layout:**
- Badge com nome do plano (background `primary-purple`)
- Progress bar: `articlesUsed / maxArticles`
- Texto: "X de Y matérias usadas este mês"
- Texto: "Próxima cobrança: dd/MM/yyyy"

**Buttons:**
- "Fazer Upgrade" (se não é o maior plano)
- "Gerenciar Assinatura" (abre portal de pagamento)

**Endpoints:**
- GET `/api/account/plan`
- POST `/api/payments/create-portal-session` (retorna URL Stripe/MP)

---

## 9. Componentes Globais

### Toast System
- Biblioteca: `sonner` ou `react-hot-toast`
- Posição: top-right
- Auto-dismiss: 5s
- Tipos: success, error, warning, info

### Loading States
- **Skeleton:** Tabelas e listas
- **Spinner:** Botões e modais
- **Overlay:** Tela completa (wizard loading)

### Error Boundary
```typescript
// app/error.tsx (Next.js)
export default function Error({ error, reset }: ErrorProps) {
  return (
    <div className="error-container">
      <h2>Algo deu errado</h2>
      <Button onClick={reset}>Tentar novamente</Button>
    </div>
  )
}
```

### Dialogs/Modals
- shadcn/ui `Dialog`
- Backdrop: dark com opacity-50
- Animação: fade + scale
- Max-width: 600px

---

## 10. Autenticação e Segurança

### Implementação de Cookies

**middleware.ts:**
```typescript
import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export function middleware(request: NextRequest) {
  const token = request.cookies.get('accessToken');
  
  const publicPaths = ['/', '/login'];
  const isPublicPath = publicPaths.includes(request.nextUrl.pathname);
  
  if (!token && !isPublicPath) {
    return NextResponse.redirect(new URL('/login', request.url));
  }
  
  // Verificar onboarding apenas em rotas protegidas
  if (token && !isPublicPath) {
    const user = parseJWT(token.value); // helper function
    
    if (!user.hasCompletedOnboarding) {
      const allowedPaths = ['/app/planos', '/app/onboarding'];
      if (!allowedPaths.includes(request.nextUrl.pathname)) {
        return NextResponse.redirect(new URL('/app/planos', request.url));
      }
    }
  }
  
  return NextResponse.next();
}

export const config = {
  matcher: ['/((?!api|_next/static|_next/image|favicon.ico).*)']
};
```

### Axios Interceptor

**lib/axios.ts:**
```typescript
import axios from 'axios';

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL,
  withCredentials: true // Importante para cookies
});

// Response interceptor para refresh token
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;
    
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      
      try {
        await api.post('/auth/refresh');
        return api(originalRequest);
      } catch (refreshError) {
        // Redirect para login
        window.location.href = '/login';
        return Promise.reject(refreshError);
      }
    }
    
    return Promise.reject(error);
  }
);

export default api;
```

### Endpoints de Auth

```typescript
// POST /api/auth/register
Request: { name, email, password }
Response: { user: User } + Set-Cookie: accessToken, refreshToken

// POST /api/auth/login
Request: { email, password }
Response: { user: User } + Set-Cookie: accessToken, refreshToken

// POST /api/auth/refresh
Request: Cookie: refreshToken
Response: Set-Cookie: novo accessToken

// POST /api/auth/logout
Request: Cookie: refreshToken
Response: Clear cookies
```

---

## 11. Responsividade

### Breakpoints Tailwind
- `sm`: 640px
- `md`: 768px
- `lg`: 1024px
- `xl`: 1280px

### Adaptações Mobile
- Sidebar → Bottom navigation (4 ícones)
- Tabelas → Cards empilhados
- Grid 2 colunas → 1 coluna
- Stepper horizontal → Vertical compacto
- Modais → Fullscreen em mobile

---

## 12. Padrões de Código

### Estrutura de Pastas (Next.js App Router)
```
/app
  /(public)
    /page.tsx              # Landing
    /login/page.tsx        # Login
  /(protected)
    /app/layout.tsx        # Layout com sidebar
    /app/planos/page.tsx
    /app/onboarding/page.tsx
    /app/novo/page.tsx
    /app/materias/page.tsx
    /app/conta/page.tsx
  /api
    /auth/route.ts
/components
  /ui                      # shadcn components
  /forms
  /layouts
/lib
  /axios.ts
  /validations.ts          # Zod schemas
  /utils.ts
/hooks
  /useAuth.ts
  /useArticles.ts
/store
  /authStore.ts            # Zustand
/types
  /index.ts
```

### Validações Zod (lib/validations.ts)
```typescript
import { z } from 'zod';

export const loginSchema = z.object({
  email: z.string().email('Email inválido'),
  password: z.string().min(6, 'Senha deve ter no mínimo 6 caracteres')
});

export const businessSchema = z.object({
  description: z.string().min(10).max(500),
  objective: z.enum(['leads', 'sales', 'branding']),
  siteUrl: z.string().url().optional().or(z.literal('')),
  hasBlog: z.boolean(),
  blogUrls: z.array(z.string().url()),
  articleCount: z.number().min(1).max(50),
  brandFile: z.instanceof(File).optional()
});

// ... outros schemas
```

### Custom Hooks

**hooks/useAuth.ts:**
```typescript
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/store/authStore';
import api from '@/lib/axios';

export function useAuth() {
  const router = useRouter();
  const { user, setUser, clearUser } = useAuthStore();
  
  const login = async (credentials: LoginCredentials) => {
    const { data } = await api.post('/auth/login', credentials);
    setUser(data.user);
    
    if (!data.user.hasCompletedOnboarding) {
      router.push('/app/planos');
    } else {
      router.push('/app/materias');
    }
  };
  
  const logout = async () => {
    await api.post('/auth/logout');
    clearUser();
    router.push('/login');
  };
  
  return { user, login, logout };
}
```

---

## 13. Testes e Qualidade

### Checklist de Implementação
- [ ] Validação Zod em todos os forms
- [ ] Loading states em todos os botões/forms
- [ ] Error handling com fallback UI
- [ ] Toast notifications consistentes
- [ ] Skeleton loaders
- [ ] Empty states
- [ ] Responsive em todos os breakpoints
- [ ] Keyboard navigation
- [ ] ARIA labels
- [ ] Focus states visíveis
- [ ] Rate limiting no frontend (debounce em buscas)

### Performance
- Next.js Image para todas as imagens
- Lazy loading de componentes pesados
- TanStack Query cache (staleTime, cacheTime)
- Debounce em inputs de busca/filtro

---

## 14. Endpoints Esperados (Resumo)

```typescript
// Auth
POST   /api/auth/register
POST   /api/auth/login
POST   /api/auth/refresh
POST   /api/auth/logout

// Plans
GET    /api/plans
POST   /api/payments/create-checkout
GET    /api/payments/status/:id
POST   /api/payments/create-portal-session

// Wizard (Onboarding)
POST   /api/wizard/business
POST   /api/wizard/competitors
POST   /api/wizard/integrations
POST   /api/wizard/generate-ideas
GET    /api/wizard/ideas-status/:id
POST   /api/wizard/publish

// Wizard (Novo)
POST   /api/articles/generate-ideas
POST   /api/articles/publish

// Articles
GET    /api/articles?page=1&limit=10&status=all
POST   /api/articles/:id/republish

// Account
GET    /api/account
PATCH  /api/account/profile
PATCH  /api/account/integrations
GET    /api/account/plan
```

---

## Arquitetura de Renderização e Build

### Decisão: Renderização Híbrida (SSR + SSG)

O projeto utiliza a arquitetura **Híbrida nativa do Next.js**, descartando a exportação estática pura (`output: 'export'`). Esta decisão é mandatória para sustentar os requisitos de segurança e SEO simultaneamente.

### Justificativa Técnica

**1. Incompatibilidade com Static Site Generation (SSG) Puro**
A exportação estática (`output: 'export'`) gera apenas arquivos HTML/CSS/JS, eliminando a camada de servidor (Runtime). Isso inviabiliza a arquitetura de segurança definida, pois:

  * **Middleware Inoperante:** O Middleware do Next.js (`middleware.ts`), essencial para proteção de rotas, requer um servidor para interceptar requisições HTTP antes da renderização. Em um site estático, ele não é executado.
  * **HttpOnly Cookies:** A validação segura de cookies `HttpOnly` (inacessíveis via JS do cliente) deve ocorrer no servidor. Sem runtime, o frontend não consegue validar a sessão de forma segura antes de montar a tela.

**2. A Estratégia Híbrida**
O Next.js gerencia automaticamente o método de renderização por rota:

  * **Rotas Públicas (Landing Page `/`):** Renderizadas como **Static Site Generation (SSG)**. O HTML é gerado no build, garantindo performance máxima (Time to First Byte) e **SEO otimizado** para indexação no Google.
  * **Rotas Protegidas (`/app/*`) e API:** Renderizadas via **Dynamic Rendering (SSR)**. Isso permite que o servidor (Node.js) verifique os cookies de sessão e execute o Middleware a cada requisição, garantindo que dados sensíveis só sejam enviados para usuários autenticados.

### Configuração de Build (`next.config.ts`)

Para suportar deploy em container (AWS App Runner/Docker) mantendo otimização de tamanho:

```typescript
const nextConfig: NextConfig = {
  // REMOVIDO: output: 'export' (Incompatível com Middleware/Auth)
  
  // ADICIONADO: Otimiza o build para containers Docker (~150MB vs ~1GB)
  output: 'standalone', 
  
  // Mantém otimização de imagens (não suportada em 'export' puro sem loader externo)
  images: { 
    unoptimized: false,
    remotePatterns: [...] 
  },
  // ...
}
```

### Impacto na Infraestrutura (AWS)

  * **Artefato de Deploy:** Imagem Docker (via `Dockerfile` multi-stage).
  * **Hospedagem:** Requer ambiente com suporte a Node.js (ex: AWS App Runner, ECS ou Amplify Hosting SSR), não sendo possível usar apenas hospedagem estática (S3 Bucket simples).

<div align="center">

| Estratégia | Setup & Complexidade | Vantagens (Pros) | Desvantagens (Cons) | Previsibilidade de Custo | Custo Inicial (Mês) | Risco de Custo | Autoscaling / Risco Operacional (DDoS & Downtime) |
|-----------|-----------------------|-------------------|----------------------|---------------------------|----------------------|-----------------|----------------------------------------------------|
| **1. AWS Amplify (Hosting Gen 2)** | Muito Baixa – Conecta ao Git e pronto. | • CI/CD nativo.<br>• Preview URLs.<br>• Dominio + SSL automáticos.<br>• Integração Amplify Backend. | • “Caixa Preta”.<br>• Lock-in.<br>• Cold Starts.<br>• Pouco controle. | ⭐⭐ (Variável) | **$5 – $15** | **Alto – Pode explodir com tráfego intenso.** | **Autoscaling agressivo (serverless)**.<br>• Amplify tenta aguentar tudo o que vier.<br>• Em DDoS de aplicação → **não cai rápido**, mas **cobra por cada request e GB**.<br>• **Maior risco financeiro**, menor risco de queda.<br>• Necessário WAF ou rate limiting para controlar custos. |
| **2. AWS App Runner (Container Gerenciado)** | Média – Precisa Dockerfile + ECR. | • Docker padrão.<br>• Portável.<br>• Auto-scaling sólido.<br>• Estável para produção. | • Pipeline manual.<br>• Sem CDN nativo.<br>• Preço base maior. | ⭐⭐⭐⭐ (Controlado) | $15 – $25 | **Médio – Controlável pelos limites.** | **Autoscaling limitado por configuração**.<br>• Pode definir ex: “máximo 3 instâncias”.<br>• Em DDoS → escala até o teto → **custo previsível**.<br>• **Site não cai imediatamente**, só quando bater o max.<br>• Com WAF → muito seguro.<br>• Sem WAF → ainda controlável. |
| **3. Amazon Lightsail (Container Service)** | Baixa/Média – Docker simplificado. | • Preço fixo.<br>• Transferência inclusa.<br>• DNS incluso.<br>• Simplicidade máxima. | • Menos performance.<br>• Escalabilidade limitada.<br>• Poucas integrações AWS. | ⭐⭐⭐⭐⭐ (Fixo) | $7 – $10 | **Muito Baixo – custo fixo real.** | **Não faz autoscaling automático**.<br>• Em DDoS → CPU 100% → **site cai**, mas **custo não aumenta**.<br>• Melhor proteção contra “DDoS Financeiro”.<br>• Você paga com *downtime*, não com dinheiro.<br>• Rate limiting via Next.js é obrigatório para proteger disponibilidade. |

</div>


---

## 15. Melhorias Futuras (v2)

- [ ] Edição de matérias antes de publicar
- [ ] Agendamento de publicações
- [ ] Analytics de performance (pageviews, rankings) - **Integrado com Google Analytics e Search Console**
- [ ] Dashboard de métricas SEO (posições, impressões, CTR)
- [ ] Sugestões automáticas baseadas em dados do Search Console
- [ ] Templates customizáveis
- [ ] Multi-idioma
- [ ] Sugestões de otimização SEO em tempo real
- [ ] Análise de concorrentes com dados do Search Console
- [ ] Relatórios mensais de performance automatizados