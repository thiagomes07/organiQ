# Documentação do Código Frontend

**Total de arquivos com conteúdo:** 60
**Total de arquivos vazios:** 0

---

## 📁 .

### middleware.ts

```ts
// middleware.ts
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const publicPaths = ["/", "/login"];
const onboardingPaths = ["/app/planos", "/app/onboarding"];

function matchesPath(pathname: string, paths: string[]): boolean {
  return paths.some(
    (path) => pathname === path || pathname.startsWith(path + "/")
  );
}

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Permitir assets e API routes
  if (
    pathname.startsWith("/_next") ||
    pathname.startsWith("/api") ||
    pathname.startsWith("/static") ||
    pathname.includes(".")
  ) {
    return NextResponse.next();
  }

  const token = request.cookies.get("accessToken")?.value;
  const isPublicPath = matchesPath(pathname, publicPaths);
  const isOnboardingPath = matchesPath(pathname, onboardingPaths);
  const isProtectedPath = pathname.startsWith("/app");

  // CASO 1: Rota pública
  if (isPublicPath) {
    if (pathname === "/login" && token) {
      // ⚠️ MUDANÇA CRÍTICA: Validar token no BACKEND
      try {
        const response = await fetch(
          `${process.env.NEXT_PUBLIC_API_URL}/auth/me`,
          {
            headers: {
              Cookie: `accessToken=${token}`,
            },
          }
        );

        if (response.ok) {
          const { user } = await response.json();
          const redirectTo = user.hasCompletedOnboarding
            ? "/app/materias"
            : "/app/planos";
          return NextResponse.redirect(new URL(redirectTo, request.url));
        }
      } catch {
        // Token inválido, deixa continuar para login
      }
    }

    return NextResponse.next();
  }

  // CASO 2: Rota protegida sem token
  if (isProtectedPath && !token) {
    const loginUrl = new URL("/login", request.url);
    loginUrl.searchParams.set("redirect", pathname);
    return NextResponse.redirect(loginUrl);
  }

  // CASO 3: Rota protegida com token
  if (isProtectedPath && token) {
    // ⚠️ MUDANÇA CRÍTICA: Validar no backend
    try {
      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/auth/me`,
        {
          headers: {
            Cookie: `accessToken=${token}`,
          },
        }
      );

      if (!response.ok) {
        return NextResponse.redirect(new URL("/login", request.url));
      }

      const { user } = await response.json();
      const hasCompletedOnboarding = user.hasCompletedOnboarding ?? false;

      // Se NÃO completou onboarding
      if (!hasCompletedOnboarding) {
        if (!isOnboardingPath) {
          return NextResponse.redirect(new URL("/app/planos", request.url));
        }
      }

      // Se JÁ completou onboarding
      if (hasCompletedOnboarding) {
        if (isOnboardingPath) {
          return NextResponse.redirect(new URL("/app/materias", request.url));
        }
      }
    } catch {
      return NextResponse.redirect(new URL("/login", request.url));
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    "/((?!_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)",
  ],
};
```

---

### next-env.d.ts

```ts
/// <reference types="next" />
/// <reference types="next/image-types/global" />
import "./.next/dev/types/routes.d.ts";

// NOTE: This file should not be edited
// see https://nextjs.org/docs/app/api-reference/config/typescript for more information.
```

---

### next.config.ts

```ts
import type { NextConfig } from 'next'

const nextConfig: NextConfig = {
  // Para SSG (quando for fazer deploy)
  // output: 'export',
  
  images: {
    // unoptimized: true, // Descomentar quando usar output: 'export'
    remotePatterns: [
      {
        protocol: 'https',
        hostname: '**.organiq.com.br',
      },
    ],
  },
  
  // Otimizações
  compiler: {
    removeConsole: process.env.NODE_ENV === 'production',
  },
  
  // Headers de segurança
  async headers() {
    return [
      {
        source: '/:path*',
        headers: [
          {
            key: 'X-DNS-Prefetch-Control',
            value: 'on'
          },
          {
            key: 'Strict-Transport-Security',
            value: 'max-age=63072000; includeSubDomains; preload'
          },
          {
            key: 'X-Frame-Options',
            value: 'SAMEORIGIN'
          },
          {
            key: 'X-Content-Type-Options',
            value: 'nosniff'
          },
          {
            key: 'X-XSS-Protection',
            value: '1; mode=block'
          },
          {
            key: 'Referrer-Policy',
            value: 'origin-when-cross-origin'
          }
        ]
      }
    ]
  }
}

export default nextConfig
```

---

## 📁 app

### app/error.tsx

```tsx
"use client";

import { useEffect } from "react";
import { AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

/**
 * Error Page - Next.js App Router
 *
 * Renderizado quando ocorre um erro não capturado
 * Automaticamente em Client Component
 */
export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    // Log do erro para serviço de monitoramento (ex: Sentry)
    console.error("App Error:", error);
  }, [error]);

  return (
    <div className="flex min-h-screen items-center justify-center p-4 bg-[var(--color-secondary-cream)]">
      <Card className="max-w-md w-full">
        <CardHeader>
          <div className="flex items-center gap-3">
            <div className="rounded-full bg-[var(--color-error)]/10 p-2">
              <AlertTriangle className="h-6 w-6 text-[var(--color-error)]" />
            </div>
            <CardTitle>Algo deu errado</CardTitle>
          </div>
        </CardHeader>

        <CardContent className="space-y-4">
          <p className="text-sm font-onest text-[var(--color-primary-dark)]/70">
            Ocorreu um erro inesperado. Você pode tentar recarregar a página ou
            voltar ao início.
          </p>

          {/* Mostrar detalhes do erro apenas em desenvolvimento */}
          {process.env.NODE_ENV === "development" && (
            <details className="rounded-[var(--radius-sm)] bg-[var(--color-error)]/5 p-3">
              <summary className="cursor-pointer text-xs font-semibold font-onest text-[var(--color-error)] mb-2">
                Detalhes do erro (visível apenas em desenvolvimento)
              </summary>
              <pre className="text-xs overflow-auto font-mono text-[var(--color-error)]/80 mt-2">
                {error.message}
              </pre>
              {error.digest && (
                <p className="text-xs text-[var(--color-error)]/60 mt-2">
                  Digest: {error.digest}
                </p>
              )}
            </details>
          )}
        </CardContent>

        <CardFooter className="flex gap-2">
          <Button
            variant="outline"
            onClick={() => window.location.reload()}
            className="flex-1"
          >
            Recarregar Página
          </Button>
          <Button variant="primary" onClick={reset} className="flex-1">
            Tentar Novamente
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
}
```

---

### app/globals.css

```css
@import "tailwindcss";

@theme {
  --color-primary-dark: #001d47;
  --color-primary-purple: #551bfa;
  --color-primary-teal: #004563;
  --color-secondary-yellow: #faf01b;
  --color-secondary-dark: #282828;
  --color-secondary-cream: #fffde1;
  --color-success: #10b981;
  --color-error: #ef4444;
  --color-warning: #f59e0b;
  --color-background: #fffde1;
  --color-foreground: #001d47;
  --color-border: #d4d4d8;
  --color-input: #d4d4d8;
  --color-ring: #551bfa;
  --font-all-round: var(--font-all-round-gothic), ui-sans-serif, system-ui,
    sans-serif;
  --font-onest: var(--font-onest-var), ui-sans-serif, system-ui, sans-serif;
  --font-sans: var(--font-onest-var), ui-sans-serif, system-ui, sans-serif;
  --radius-sm: 6px;
  --radius-md: 8px;
  --radius-lg: 12px;
  --spacing-18: 4.5rem;
  --spacing-88: 22rem;
  --animate-accordion-down: accordion-down 0.2s ease-out;
  --animate-accordion-up: accordion-up 0.2s ease-out;

  @keyframes accordion-down {
    from {
      height: 0;
    }
    to {
      height: var(--radix-accordion-content-height);
    }
  }

  @keyframes accordion-up {
    from {
      height: var(--radix-accordion-content-height);
    }
    to {
      height: 0;
    }
  }
}

@media (prefers-color-scheme: dark) {
  :root {
    /* Modo escuro desabilitado por enquanto */
  }
}

@layer base {
  * {
    border-color: var(--color-border);
  }

  body {
    background-color: var(--color-background);
    color: var(--color-foreground);
    font-family: var(--font-sans);
  }

  h1,
  h2,
  h3,
  h4,
  h5,
  h6 {
    font-family: var(--font-all-round);
  }

  /* Cursor pointer para elementos interativos */
  button:not(:disabled),
  a:not([aria-disabled="true"]),
  [role="button"]:not([aria-disabled="true"]),
  [role="tab"],
  [role="menuitem"],
  label[for],
  select:not(:disabled),
  summary {
    cursor: pointer;
  }

  button:disabled,
  a[aria-disabled="true"],
  [role="button"][aria-disabled="true"],
  select:disabled {
    cursor: not-allowed;
  }
}

@layer utilities {
  .text-balance {
    text-wrap: balance;
  }
}
```

---

### app/layout.tsx

```tsx
import type { Metadata } from 'next'
import { Toaster } from 'sonner'
import { Providers } from './providers'
import './globals.css'

// Fonts
import localFont from 'next/font/local'

const allRoundGothic = localFont({
  src: './fonts/AllRoundGothic-Medium.woff2',
  variable: '--font-all-round-gothic',
  weight: '500',
  display: 'swap',
})

const onest = localFont({
  src: [
    { path: './fonts/Onest-Thin.woff2', weight: '100', style: 'normal' },
    { path: './fonts/Onest-ExtraLight.woff2', weight: '200', style: 'normal' },
    { path: './fonts/Onest-Light.woff2', weight: '300', style: 'normal' },
    { path: './fonts/Onest-Regular.woff2', weight: '400', style: 'normal' },
    { path: './fonts/Onest-Medium.woff2', weight: '500', style: 'normal' },
    { path: './fonts/Onest-SemiBold.woff2', weight: '600', style: 'normal' },
    { path: './fonts/Onest-Bold.woff2', weight: '700', style: 'normal' },
    { path: './fonts/Onest-ExtraBold.woff2', weight: '800', style: 'normal' },
    { path: './fonts/Onest-Black.woff2', weight: '900', style: 'normal' },
  ],
  variable: '--font-onest',
  display: 'swap',
})

export const metadata: Metadata = {
  title: {
    default: 'organiQ - Aumente seu tráfego orgânico com IA',
    template: '%s | organiQ',
  },
  description: 'Matérias de blog que geram autoridade e SEO. Naturalmente Inteligente.',
  keywords: ['SEO', 'Marketing de Conteúdo', 'IA', 'WordPress', 'Blog', 'Tráfego Orgânico'],
  authors: [{ name: 'organiQ' }],
  creator: 'organiQ',
  publisher: 'organiQ',
  metadataBase: new URL(process.env.NEXT_PUBLIC_APP_URL || 'https://organiq.com.br'),
  
  openGraph: {
    type: 'website',
    locale: 'pt_BR',
    url: '/',
    siteName: 'organiQ',
    title: 'organiQ - Aumente seu tráfego orgânico com IA',
    description: 'Matérias de blog que geram autoridade e SEO. Naturalmente Inteligente.',
    images: [
      {
        url: '/og-image.jpg',
        width: 1200,
        height: 630,
        alt: 'organiQ - Naturalmente Inteligente',
      },
    ],
  },
  
  twitter: {
    card: 'summary_large_image',
    title: 'organiQ - Aumente seu tráfego orgânico com IA',
    description: 'Matérias de blog que geram autoridade e SEO. Naturalmente Inteligente.',
    images: ['/twitter-image.jpg'],
  },
  
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      'max-video-preview': -1,
      'max-image-preview': 'large',
      'max-snippet': -1,
    },
  },
  
  manifest: '/manifest.json',
  
  icons: {
    icon: [
      { url: '/favicon-16x16.png', sizes: '16x16', type: 'image/png' },
      { url: '/favicon-32x32.png', sizes: '32x32', type: 'image/png' },
    ],
    apple: [
      { url: '/apple-touch-icon.png', sizes: '180x180', type: 'image/png' },
    ],
  },
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="pt-BR" className={`${allRoundGothic.variable} ${onest.variable}`}>
      <body className="antialiased">
        <Providers>
          {children}
          <Toaster position="top-right" richColors closeButton />
        </Providers>
      </body>
    </html>
  )
}
```

---

### app/not-found.tsx

```tsx
import Link from "next/link";
import { FileQuestion } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

/**
 * 404 Not Found Page
 *
 * Renderizado quando uma rota não é encontrada
 */
export default function NotFound() {
  return (
    <div className="flex min-h-screen items-center justify-center p-4 bg-[var(--color-secondary-cream)]">
      <Card className="max-w-md w-full">
        <CardHeader>
          <div className="flex flex-col items-center gap-4">
            <div className="rounded-full bg-[var(--color-primary-purple)]/10 p-6">
              <FileQuestion className="h-12 w-12 text-[var(--color-primary-purple)]" />
            </div>
            <div className="text-center">
              <CardTitle className="text-3xl mb-2">404</CardTitle>
              <p className="text-sm font-onest text-[var(--color-primary-dark)]/70">
                Página não encontrada
              </p>
            </div>
          </div>
        </CardHeader>

        <CardContent>
          <p className="text-center text-sm font-onest text-[var(--color-primary-dark)]/70">
            A página que você está procurando não existe ou foi movida para
            outro endereço.
          </p>
        </CardContent>

        <CardFooter className="flex flex-col gap-2">
          <Link href="/app/materias" className="w-full">
            <Button variant="primary" className="w-full">
              Ir para Dashboard
            </Button>
          </Link>
          <Link href="/" className="w-full">
            <Button variant="outline" className="w-full">
              Voltar ao Início
            </Button>
          </Link>
        </CardFooter>
      </Card>
    </div>
  );
}
```

---

### app/page.tsx

```tsx
import { Metadata } from 'next'
import Link from 'next/link'
import { Sparkles, Target, Zap } from 'lucide-react'
import { Header } from '@/components/layouts/Header'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

export const metadata: Metadata = {
  title: 'organiQ - Aumente seu tráfego orgânico com IA',
  description: 'Matérias de blog que geram autoridade e SEO. Naturalmente Inteligente. Geração automática de conteúdo otimizado para WordPress.',
}

const features = [
  {
    icon: Sparkles,
    title: 'Geração Automática de Conteúdo',
    description: 'IA avançada cria matérias de blog completas, otimizadas para seu nicho e público-alvo.',
  },
  {
    icon: Target,
    title: 'SEO Otimizado',
    description: 'Conteúdo estratégico baseado em análise de concorrentes e palavras-chave relevantes.',
  },
  {
    icon: Zap,
    title: 'Publicação Direta no WordPress',
    description: 'Integração nativa que publica suas matérias automaticamente no seu blog.',
  },
]

export default function LandingPage() {
  return (
    <div className="min-h-screen bg-[var(--color-secondary-cream)]">
      <Header />

      {/* Hero Section */}
      <section className="container mx-auto px-4 pt-20 pb-16 md:pt-32 md:pb-24">
        <div className="max-w-4xl mx-auto text-center space-y-8">
          {/* Tagline Badge */}
          <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-white border border-[var(--color-primary-teal)]/20 shadow-sm">
            <span className="text-sm font-semibold font-all-round text-[var(--color-primary-purple)]">
              ✨ Naturalmente Inteligente
            </span>
          </div>

          {/* Main Heading */}
          <h1 className="text-4xl md:text-5xl lg:text-6xl font-bold font-all-round text-[var(--color-primary-dark)] leading-tight">
            Aumente seu tráfego orgânico{' '}
            <span className="text-[var(--color-primary-purple)]">com IA</span>
          </h1>

          {/* Subtitle */}
          <p className="text-lg md:text-xl font-onest text-[var(--color-primary-teal)] max-w-2xl mx-auto">
            Matérias de blog que geram autoridade e SEO automaticamente
          </p>

          {/* CTA Button */}
          <div className="pt-4">
            <Link href="/login">
              <Button
                variant="secondary"
                size="lg"
                className="text-base md:text-lg px-8 md:px-12 h-12 md:h-14"
              >
                Criar minha conta grátis
              </Button>
            </Link>
          </div>

          {/* Social Proof */}
          <p className="text-sm font-onest text-[var(--color-primary-dark)]/60">
            Junte-se a centenas de empresas que já aumentaram seu tráfego orgânico
          </p>
        </div>
      </section>

      {/* Features Section */}
      <section id="features" className="container mx-auto px-4 py-16 md:py-24">
        <div className="max-w-6xl mx-auto">
          {/* Section Title */}
          <div className="text-center mb-12 md:mb-16">
            <h2 className="text-3xl md:text-4xl font-bold font-all-round text-[var(--color-primary-dark)] mb-4">
              Como funciona
            </h2>
            <p className="text-lg font-onest text-[var(--color-primary-dark)]/70 max-w-2xl mx-auto">
              Três recursos poderosos que transformam sua estratégia de conteúdo
            </p>
          </div>

          {/* Features Grid */}
          <div className="grid md:grid-cols-3 gap-6 md:gap-8">
            {features.map((feature, index) => {
              const Icon = feature.icon

              return (
                <Card
                  key={index}
                  className="group hover:shadow-lg hover:scale-[1.02] transition-all duration-200 border-[var(--color-primary-teal)]/20"
                >
                  <CardHeader>
                    {/* Icon */}
                    <div className="mb-4 inline-flex items-center justify-center h-12 w-12 rounded-full bg-[var(--color-primary-purple)]/10 text-[var(--color-primary-purple)] group-hover:bg-[var(--color-primary-purple)] group-hover:text-white transition-colors duration-200">
                      <Icon className="h-6 w-6" />
                    </div>

                    {/* Title */}
                    <CardTitle className="text-xl">
                      {feature.title}
                    </CardTitle>
                  </CardHeader>

                  <CardContent>
                    {/* Description */}
                    <CardDescription className="text-base">
                      {feature.description}
                    </CardDescription>
                  </CardContent>
                </Card>
              )
            })}
          </div>
        </div>
      </section>

      {/* How It Works Section */}
      <section id="how-it-works" className="container mx-auto px-4 py-16 md:py-24">
        <div className="max-w-4xl mx-auto">
          <div className="bg-white rounded-[var(--radius-lg)] shadow-md p-8 md:p-12">
            <h2 className="text-3xl md:text-4xl font-bold font-all-round text-[var(--color-primary-dark)] mb-8 text-center">
              Simples e rápido
            </h2>

            <div className="space-y-8">
              {/* Step 1 */}
              <div className="flex gap-4 md:gap-6">
                <div className="flex-shrink-0 flex items-center justify-center h-10 w-10 rounded-full bg-[var(--color-primary-purple)] text-white font-bold font-all-round">
                  1
                </div>
                <div>
                  <h3 className="text-xl font-semibold font-all-round text-[var(--color-primary-dark)] mb-2">
                    Conecte seu WordPress
                  </h3>
                  <p className="font-onest text-[var(--color-primary-dark)]/70">
                    Integração segura em poucos cliques. Suporte para qualquer site WordPress.
                  </p>
                </div>
              </div>

              {/* Step 2 */}
              <div className="flex gap-4 md:gap-6">
                <div className="flex-shrink-0 flex items-center justify-center h-10 w-10 rounded-full bg-[var(--color-primary-purple)] text-white font-bold font-all-round">
                  2
                </div>
                <div>
                  <h3 className="text-xl font-semibold font-all-round text-[var(--color-primary-dark)] mb-2">
                    Configure seus objetivos
                  </h3>
                  <p className="font-onest text-[var(--color-primary-dark)]/70">
                    Defina seu nicho, público-alvo e concorrentes. Nossa IA analisa tudo para criar a melhor estratégia.
                  </p>
                </div>
              </div>

              {/* Step 3 */}
              <div className="flex gap-4 md:gap-6">
                <div className="flex-shrink-0 flex items-center justify-center h-10 w-10 rounded-full bg-[var(--color-primary-purple)] text-white font-bold font-all-round">
                  3
                </div>
                <div>
                  <h3 className="text-xl font-semibold font-all-round text-[var(--color-primary-dark)] mb-2">
                    Aprove e publique
                  </h3>
                  <p className="font-onest text-[var(--color-primary-dark)]/70">
                    Revise as ideias geradas, adicione seu toque pessoal e publique automaticamente no seu blog.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="container mx-auto px-4 py-16 md:py-24">
        <div className="max-w-4xl mx-auto text-center bg-gradient-to-br from-[var(--color-primary-purple)] to-[var(--color-primary-teal)] rounded-[var(--radius-lg)] shadow-xl p-8 md:p-12">
          <h2 className="text-3xl md:text-4xl font-bold font-all-round text-white mb-4">
            Pronto para aumentar seu tráfego?
          </h2>
          <p className="text-lg font-onest text-white/90 mb-8 max-w-2xl mx-auto">
            Comece gratuitamente e veja os resultados em poucos dias
          </p>
          <Link href="/login">
            <Button
              variant="secondary"
              size="lg"
              className="text-base md:text-lg px-8 md:px-12 h-12 md:h-14"
            >
              Criar minha conta grátis
            </Button>
          </Link>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-[var(--color-border)] bg-white">
        <div className="container mx-auto px-4 py-8">
          <div className="flex flex-col md:flex-row items-center justify-between gap-4">
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-bold font-all-round text-[var(--color-primary-purple)]">
                organiQ
              </h1>
              <span className="text-sm font-onest text-[var(--color-primary-dark)]/60">
                © 2024 Todos os direitos reservados
              </span>
            </div>
            <div className="flex items-center gap-6">
              <a
                href="#"
                className="text-sm font-onest text-[var(--color-primary-dark)]/70 hover:text-[var(--color-primary-dark)] transition-colors"
              >
                Termos de Uso
              </a>
              <a
                href="#"
                className="text-sm font-onest text-[var(--color-primary-dark)]/70 hover:text-[var(--color-primary-dark)] transition-colors"
              >
                Política de Privacidade
              </a>
              <a
                href="#"
                className="text-sm font-onest text-[var(--color-primary-dark)]/70 hover:text-[var(--color-primary-dark)] transition-colors"
              >
                Contato
              </a>
            </div>
          </div>
        </div>
      </footer>
    </div>
  )
}
```

---

### app/providers.tsx

```tsx
'use client'

import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useState } from 'react'

export function Providers({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            // Configurações globais do React Query
            staleTime: 60 * 1000, // 1 minuto
            refetchOnWindowFocus: false,
            retry: 1,
          },
        },
      })
  )

  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  )
}
```

---

### app/robots.ts

```ts
import { MetadataRoute } from 'next'

export default function robots(): MetadataRoute.Robots {
  return {
    rules: [
      {
        userAgent: '*',
        allow: '/',
        disallow: ['/app/', '/api/'],
      },
    ],
    sitemap: 'https://organiq.com.br/sitemap.xml',
  }
}
```

---

### app/sitemap.ts

```ts
import { MetadataRoute } from 'next'

/**
 * Sitemap Generator
 * 
 * Gera sitemap.xml automaticamente para SEO
 * Next.js 14+ suporta geração dinâmica de sitemap
 */
export default function sitemap(): MetadataRoute.Sitemap {
  const baseUrl = process.env.NEXT_PUBLIC_APP_URL || 'https://organiq.com.br'
  
  // Data atual para lastModified
  const now = new Date()

  return [
    {
      url: baseUrl,
      lastModified: now,
      changeFrequency: 'weekly',
      priority: 1,
    },
    {
      url: `${baseUrl}/login`,
      lastModified: now,
      changeFrequency: 'monthly',
      priority: 0.8,
    },
    // Rotas protegidas não devem estar no sitemap público
    // pois requerem autenticação
  ]
}
```

---

## 📁 app\api\health

### app/api/health/route.ts

```ts
import { NextResponse } from "next/server";

/**
 * Health Check Endpoint
 *
 * Retorna o status da aplicação para monitoramento
 * Útil para load balancers, Kubernetes, etc.
 */
export async function GET() {
  try {
    return NextResponse.json(
      {
        status: "healthy",
        timestamp: new Date().toISOString(),
        service: "organiQ Frontend",
        version: process.env.NEXT_PUBLIC_APP_VERSION || "1.0.0",
        environment: process.env.NODE_ENV,
      },
      { status: 200 }
    );
  } catch (error) {
    return NextResponse.json(
      {
        status: "unhealthy",
        timestamp: new Date().toISOString(),
        error: error instanceof Error ? error.message : "Unknown error",
      },
      { status: 503 }
    );
  }
}

// Permite apenas GET
export async function POST() {
  return NextResponse.json({ error: "Method not allowed" }, { status: 405 });
}
```

---

## 📁 app\app

### app/app/layout.tsx

```tsx
'use client'

import { useEffect } from 'react'
import { useRouter, usePathname } from 'next/navigation'
import { useAuthStore } from '@/store/authStore'
import { Sidebar } from '@/components/layouts/Sidebar'
import { MobileNav } from '@/components/layouts/MobileNav'
import { LoadingSpinner } from '@/components/shared/LoadingSpinner'

export default function ProtectedLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const router = useRouter()
  const pathname = usePathname()
  const { isAuthenticated, isLoading, user } = useAuthStore()

  useEffect(() => {
    // Aguardar hydration do Zustand
    if (isLoading) return

    // Não autenticado: redirecionar para login
    if (!isAuthenticated) {
      router.push('/login')
      return
    }

    // Verificar onboarding
    if (user && !user.hasCompletedOnboarding) {
      const allowedPaths = ['/app/planos', '/app/onboarding']
      
      // Se não está em uma rota permitida, redirecionar
      if (!allowedPaths.includes(pathname)) {
        router.push('/app/planos')
      }
    }
  }, [isAuthenticated, isLoading, user, pathname, router])

  // Loading state durante hydration
  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-[var(--color-secondary-cream)]">
        <LoadingSpinner size="lg" text="Carregando..." />
      </div>
    )
  }

  // Não autenticado: não renderizar nada (vai redirecionar)
  if (!isAuthenticated) {
    return null
  }

  return (
    <div className="flex min-h-screen bg-[var(--color-secondary-cream)]">
      {/* Sidebar Desktop */}
      <Sidebar />

      {/* Main Content */}
      <main className="flex-1 lg:ml-0 pb-20 lg:pb-0 p-4 md:p-8">
        <div className="max-w-7xl mx-auto">
          {children}
        </div>
      </main>

      {/* Mobile Navigation */}
      <MobileNav />
    </div>
  )
}
```

---

## 📁 app\app\conta

### app/app/conta/page.tsx

```tsx
'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { HelpCircle, Check, Calendar } from 'lucide-react'
import * as Accordion from '@radix-ui/react-accordion'
import { profileUpdateSchema, integrationsUpdateSchema, type ProfileUpdateInput, type IntegrationsUpdateInput } from '@/lib/validations'
import { usePlans } from '@/hooks/usePlans'
import { useUser } from '@/store/authStore'
import { formatDate, formatCurrency } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'

export default function ContaPage() {
  const user = useUser()
  const { currentPlan, openPortal, isOpeningPortal } = usePlans()
  const [isSavingProfile, setIsSavingProfile] = useState(false)
  const [isSavingIntegrations, setIsSavingIntegrations] = useState(false)

  // Profile Form
  const profileForm = useForm<ProfileUpdateInput>({
    resolver: zodResolver(profileUpdateSchema),
    defaultValues: {
      name: user?.name || '',
    },
  })

  // Integrations Form
  const integrationsForm = useForm<IntegrationsUpdateInput>({
    resolver: zodResolver(integrationsUpdateSchema),
    defaultValues: {
      wordpress: {
        siteUrl: '',
        username: '',
        appPassword: '',
      },
      searchConsole: {
        enabled: false,
      },
      analytics: {
        enabled: false,
      },
    },
  })

  const watchSearchConsoleEnabled = integrationsForm.watch('searchConsole.enabled')
  const watchAnalyticsEnabled = integrationsForm.watch('analytics.enabled')

  const handleUpdateProfile = async (data: ProfileUpdateInput) => {
    setIsSavingProfile(true)
    try {
      // TODO: Implement API call
      await new Promise(resolve => setTimeout(resolve, 1000))
      toast.success('Perfil atualizado com sucesso!')
    } catch (error) {
      toast.error('Erro ao atualizar perfil')
    } finally {
      setIsSavingProfile(false)
    }
  }

  const handleUpdateIntegrations = async (data: IntegrationsUpdateInput) => {
    setIsSavingIntegrations(true)
    try {
      // TODO: Implement API call
      await new Promise(resolve => setTimeout(resolve, 1000))
      toast.success('Integrações atualizadas com sucesso!')
    } catch (error) {
      toast.error('Erro ao atualizar integrações')
    } finally {
      setIsSavingIntegrations(false)
    }
  }

  const usagePercentage = user ? (user.articlesUsed / user.maxArticles) * 100 : 0

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold font-all-round text-[var(--color-primary-dark)]">
          Minha Conta
        </h1>
        <p className="text-sm font-onest text-[var(--color-primary-dark)]/70 mt-1">
          Gerencie suas configurações e integrações
        </p>
      </div>

      <div className="grid lg:grid-cols-2 gap-6">
        {/* Card 1: Perfil */}
        <Card>
          <CardHeader>
            <CardTitle>Perfil</CardTitle>
            <CardDescription>Informações da sua conta</CardDescription>
          </CardHeader>

          <form onSubmit={profileForm.handleSubmit(handleUpdateProfile)}>
            <CardContent className="space-y-4">
              {/* Nome */}
              <div className="space-y-2">
                <Label htmlFor="name" required>
                  Nome completo
                </Label>
                <Input
                  id="name"
                  type="text"
                  error={profileForm.formState.errors.name?.message}
                  {...profileForm.register('name')}
                />
              </div>

              {/* Email (disabled) */}
              <div className="space-y-2">
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  type="email"
                  value={user?.email || ''}
                  disabled
                />
                <p className="text-xs text-[var(--color-primary-dark)]/60 font-onest">
                  O email não pode ser alterado
                </p>
              </div>
            </CardContent>

            <CardFooter>
              <Button
                type="submit"
                variant="primary"
                isLoading={isSavingProfile}
                disabled={isSavingProfile}
              >
                Salvar Alterações
              </Button>
            </CardFooter>
          </form>
        </Card>

        {/* Card 2: Meu Plano */}
        <Card>
          <CardHeader>
            <CardTitle>Meu Plano</CardTitle>
            <CardDescription>Informações da sua assinatura</CardDescription>
          </CardHeader>

          <CardContent className="space-y-6">
            {/* Badge do Plano */}
            <div className="flex items-center gap-3">
              <div className="px-4 py-2 rounded-full bg-[var(--color-primary-purple)] text-white font-bold font-all-round text-lg">
                {user?.planName}
              </div>
              <div className="text-sm font-onest text-[var(--color-primary-dark)]/70">
                {currentPlan && formatCurrency(currentPlan.price)}/mês
              </div>
            </div>

            {/* Uso de Matérias */}
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium font-onest text-[var(--color-primary-dark)]">
                  Matérias usadas
                </span>
                <span className="text-sm font-semibold font-all-round text-[var(--color-primary-purple)]">
                  {user?.articlesUsed} / {user?.maxArticles}
                </span>
              </div>
              <Progress value={usagePercentage} showLabel />
              <p className="text-xs text-[var(--color-primary-dark)]/60 font-onest">
                {user && user.maxArticles - user.articlesUsed} matérias restantes este mês
              </p>
            </div>

            {/* Próxima Cobrança */}
            {currentPlan?.nextBillingDate && (
              <div className="flex items-center gap-2 text-sm font-onest text-[var(--color-primary-dark)]/70">
                <Calendar className="h-4 w-4" />
                <span>
                  Próxima cobrança: {formatDate(currentPlan.nextBillingDate)}
                </span>
              </div>
            )}
          </CardContent>

          <CardFooter className="flex gap-2">
            <Button
              variant="outline"
              onClick={openPortal}
              isLoading={isOpeningPortal}
              disabled={isOpeningPortal}
            >
              Gerenciar Assinatura
            </Button>
            <Button variant="primary">
              Fazer Upgrade
            </Button>
          </CardFooter>
        </Card>
      </div>

      {/* Card 3: Integrações (Full Width) */}
      <Card>
        <CardHeader>
          <CardTitle>Integrações</CardTitle>
          <CardDescription>Configure suas conexões com WordPress e Google</CardDescription>
        </CardHeader>

        <form onSubmit={integrationsForm.handleSubmit(handleUpdateIntegrations)}>
          <CardContent>
            <Accordion.Root type="multiple" className="space-y-4">
              {/* WordPress */}
              <Accordion.Item value="wordpress">
                <div className="border-2 border-[var(--color-primary-purple)] rounded-[var(--radius-md)] overflow-hidden">
                  <Accordion.Header>
                    <Accordion.Trigger className="flex items-center justify-between w-full p-4 hover:bg-[var(--color-primary-purple)]/5 transition-colors">
                      <div className="flex items-center gap-3">
                        <div className="flex items-center justify-center h-10 w-10 rounded-full bg-[var(--color-primary-purple)]/10">
                          <span className="text-xl">🔌</span>
                        </div>
                        <div className="text-left">
                          <h4 className="text-base font-semibold font-all-round text-[var(--color-primary-dark)]">
                            WordPress
                          </h4>
                          <p className="text-xs font-onest text-[var(--color-primary-dark)]/60">
                            Publicação automática de matérias
                          </p>
                        </div>
                      </div>
                      <Check className="h-5 w-5 text-[var(--color-success)]" />
                    </Accordion.Trigger>
                  </Accordion.Header>

                  <Accordion.Content className="p-4 pt-0 space-y-4">
                    <div className="space-y-2">
                      <Label htmlFor="wp-siteUrl">URL do site</Label>
                      <Input
                        id="wp-siteUrl"
                        type="url"
                        placeholder="https://seusite.com.br"
                        {...integrationsForm.register('wordpress.siteUrl')}
                      />
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="wp-username">Nome de usuário</Label>
                      <Input
                        id="wp-username"
                        type="text"
                        placeholder="seu_usuario"
                        {...integrationsForm.register('wordpress.username')}
                      />
                    </div>

                    <div className="space-y-2">
                      <div className="flex items-center justify-between">
                        <Label htmlFor="wp-appPassword">Senha de aplicativo</Label>
                        <Dialog>
                          <DialogTrigger asChild>
                            <button type="button" className="text-[var(--color-primary-teal)] hover:text-[var(--color-primary-purple)]">
                              <HelpCircle className="h-4 w-4" />
                            </button>
                          </DialogTrigger>
                          <DialogContent>
                            <DialogHeader>
                              <DialogTitle>Como obter a senha?</DialogTitle>
                              <DialogDescription className="space-y-2 text-left">
                                <p>1. WordPress → Usuários → Perfil</p>
                                <p>2. Role até "Senhas de aplicativo"</p>
                                <p>3. Adicione uma nova senha</p>
                                <p>4. Copie e cole aqui</p>
                              </DialogDescription>
                            </DialogHeader>
                          </DialogContent>
                        </Dialog>
                      </div>
                      <Input
                        id="wp-appPassword"
                        type="password"
                        placeholder="••••••••••••"
                        {...integrationsForm.register('wordpress.appPassword')}
                      />
                    </div>
                  </Accordion.Content>
                </div>
              </Accordion.Item>

              {/* Google Search Console */}
              <Accordion.Item value="searchConsole">
                <div className={cn(
                  "border-2 rounded-[var(--radius-md)] overflow-hidden",
                  watchSearchConsoleEnabled ? "border-[var(--color-primary-teal)]" : "border-[var(--color-border)]"
                )}>
                  <Accordion.Header>
                    <Accordion.Trigger className="flex items-center justify-between w-full p-4 hover:bg-[var(--color-primary-teal)]/5">
                      <div className="flex items-center gap-3">
                        <div className="h-10 w-10 rounded-full bg-[var(--color-primary-teal)]/10 flex items-center justify-center">
                          <span className="text-xl">📊</span>
                        </div>
                        <div className="text-left">
                          <h4 className="text-base font-semibold font-all-round text-[var(--color-primary-dark)]">
                            Google Search Console
                          </h4>
                          <p className="text-xs font-onest text-[var(--color-primary-dark)]/60">
                            Análise de palavras-chave
                          </p>
                        </div>
                      </div>
                      <div className="flex items-center gap-2">
                        {watchSearchConsoleEnabled && <Check className="h-5 w-5 text-[var(--color-success)]" />}
                        <input
                          type="checkbox"
                          {...integrationsForm.register('searchConsole.enabled')}
                          onClick={(e) => e.stopPropagation()}
                          className="h-4 w-4 rounded"
                        />
                      </div>
                    </Accordion.Trigger>
                  </Accordion.Header>

                  {watchSearchConsoleEnabled && (
                    <Accordion.Content className="p-4 pt-0">
                      <Input
                        type="url"
                        placeholder="https://seusite.com.br"
                        {...integrationsForm.register('searchConsole.propertyUrl')}
                      />
                    </Accordion.Content>
                  )}
                </div>
              </Accordion.Item>

              {/* Google Analytics */}
              <Accordion.Item value="analytics">
                <div className={cn(
                  "border-2 rounded-[var(--radius-md)] overflow-hidden",
                  watchAnalyticsEnabled ? "border-[var(--color-primary-teal)]" : "border-[var(--color-border)]"
                )}>
                  <Accordion.Header>
                    <Accordion.Trigger className="flex items-center justify-between w-full p-4 hover:bg-[var(--color-primary-teal)]/5">
                      <div className="flex items-center gap-3">
                        <div className="h-10 w-10 rounded-full bg-[var(--color-primary-teal)]/10 flex items-center justify-center">
                          <span className="text-xl">📈</span>
                        </div>
                        <div className="text-left">
                          <h4 className="text-base font-semibold font-all-round text-[var(--color-primary-dark)]">
                            Google Analytics
                          </h4>
                          <p className="text-xs font-onest text-[var(--color-primary-dark)]/60">
                            Análise de tráfego
                          </p>
                        </div>
                      </div>
                      <div className="flex items-center gap-2">
                        {watchAnalyticsEnabled && <Check className="h-5 w-5 text-[var(--color-success)]" />}
                        <input
                          type="checkbox"
                          {...integrationsForm.register('analytics.enabled')}
                          onClick={(e) => e.stopPropagation()}
                          className="h-4 w-4 rounded"
                        />
                      </div>
                    </Accordion.Trigger>
                  </Accordion.Header>

                  {watchAnalyticsEnabled && (
                    <Accordion.Content className="p-4 pt-0">
                      <Input
                        type="text"
                        placeholder="G-XXXXXXXXXX"
                        {...integrationsForm.register('analytics.measurementId')}
                      />
                    </Accordion.Content>
                  )}
                </div>
              </Accordion.Item>
            </Accordion.Root>
          </CardContent>

          <CardFooter>
            <Button
              type="submit"
              variant="primary"
              isLoading={isSavingIntegrations}
              disabled={isSavingIntegrations}
            >
              Atualizar Integrações
            </Button>
          </CardFooter>
        </form>
      </Card>
    </div>
  )
}
```

---

## 📁 app\app\materias

### app/app/materias/page.tsx

```tsx
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
```

---

## 📁 app\app\novo

### app/app/novo/page.tsx

```tsx
'use client'

import { useWizard } from '@/hooks/useWizard'
import { useUser } from '@/store/authStore'
import { StepIndicator } from '@/components/wizards/StepIndicator'
import { CompetitorsForm } from '@/components/forms/CompetitorsForm'
import { LoadingOverlay } from '@/components/shared/LoadingSpinner'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Slider } from '@/components/ui/slider'
import { Label } from '@/components/ui/label'
import { AlertCircle } from 'lucide-react'
import Link from 'next/link'
import type { CompetitorsInput } from '@/lib/validations'

const steps = [
  { number: 1, label: 'Quantidade' },
  { number: 2, label: 'Concorrentes' },
  { number: 3, label: 'Aprovação' },
]

const loadingMessages = [
  'Analisando seus concorrentes...',
  'Mapeando tópicos de autoridade...',
  'Gerando ideias de matérias...',
  'Isso pode levar alguns minutos',
]

export default function NovoPage() {
  const user = useUser()
  const {
    currentStep,
    businessData,
    competitorData,
    submitBusinessInfo,
    submitCompetitors,
    previousStep,
    isSubmittingBusiness,
    isSubmittingCompetitors,
    isGeneratingIdeas,
  } = useWizard(false) // false = não é onboarding

  const articlesRemaining = user ? user.maxArticles - user.articlesUsed : 0
  const canCreate = articlesRemaining > 0

  // Loading state para geração de ideias
  if (currentStep === 999 || isGeneratingIdeas) {
    return <LoadingOverlay messages={loadingMessages} />
  }

  // TODO: Implement step 3 (Approval) and step 1000 (Publishing)
  // These will be added in the next phase

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="text-center space-y-2">
        <h1 className="text-3xl font-bold font-all-round text-[var(--color-primary-dark)]">
          Gerar Novas Matérias
        </h1>
        <p className="text-lg font-onest text-[var(--color-primary-dark)]/70">
          Crie mais conteúdo otimizado para seu blog
        </p>
      </div>

      {/* Limit Warning */}
      {!canCreate && (
        <Card className="border-[var(--color-warning)]">
          <CardContent className="flex items-start gap-3 p-4">
            <AlertCircle className="h-5 w-5 text-[var(--color-warning)] mt-0.5" />
            <div className="flex-1">
              <p className="font-medium font-onest text-[var(--color-primary-dark)]">
                Limite de matérias atingido
              </p>
              <p className="text-sm font-onest text-[var(--color-primary-dark)]/70 mt-1">
                Você já utilizou todas as {user?.maxArticles} matérias do seu plano este mês.
              </p>
              <Link href="/app/conta">
                <Button variant="outline" size="sm" className="mt-3">
                  Fazer Upgrade
                </Button>
              </Link>
            </div>
          </CardContent>
        </Card>
      )}

      {canCreate && (
        <>
          {/* Step Indicator */}
          <StepIndicator currentStep={currentStep} steps={steps} />

          {/* Form Card */}
          <Card>
            <CardHeader>
              <CardTitle>
                {currentStep === 1 && 'Quantidade de Matérias'}
                {currentStep === 2 && 'Análise de Concorrentes'}
              </CardTitle>
              <CardDescription>
                {currentStep === 1 && `Você tem ${articlesRemaining} matérias disponíveis este mês`}
                {currentStep === 2 && 'Adicione URLs de concorrentes para melhorar a estratégia (opcional)'}
              </CardDescription>
            </CardHeader>

            <CardContent>
              {/* Step 1: Article Count */}
              {currentStep === 1 && (
                <form
                  onSubmit={(e) => {
                    e.preventDefault()
                    const articleCount = businessData?.articleCount || 1
                    submitBusinessInfo({
                      description: '', // Dados já existem do onboarding
                      primaryObjective: 'leads', // Placeholder
                      hasBlog: false,
                      blogUrls: [],
                      articleCount,
                    } as any)
                  }}
                  className="space-y-6"
                >
                  {/* Slider */}
                  <div className="space-y-2">
                    <Label required>Quantas matérias deseja criar?</Label>
                    <Slider
                      min={1}
                      max={articlesRemaining}
                      step={1}
                      value={[businessData?.articleCount || 1]}
                      onValueChange={(value) => {
                        // Atualizar estado do wizard
                        submitBusinessInfo({
                          ...businessData,
                          articleCount: value[0],
                        } as any)
                      }}
                      showValue
                      formatValue={(value) => `${value} ${value === 1 ? 'matéria' : 'matérias'}`}
                    />
                  </div>

                  {/* Info */}
                  <div className="bg-[var(--color-primary-purple)]/5 border border-[var(--color-primary-purple)]/20 rounded-[var(--radius-md)] p-4">
                    <p className="text-sm font-onest text-[var(--color-primary-dark)]/80">
                      💡 <strong>Dica:</strong> Você pode gerar várias matérias de uma vez para economizar tempo.
                    </p>
                  </div>

                  {/* Submit Button */}
                  <div className="flex justify-end pt-4">
                    <Button
                      type="submit"
                      variant="secondary"
                      size="lg"
                      isLoading={isSubmittingBusiness}
                      disabled={isSubmittingBusiness}
                    >
                      Próximo
                    </Button>
                  </div>
                </form>
              )}

              {/* Step 2: Competitors */}
              {currentStep === 2 && (
                <CompetitorsForm
                  onSubmit={(data: CompetitorsInput) => submitCompetitors(data as any)}
                  onBack={previousStep}
                  isLoading={isSubmittingCompetitors}
                  defaultValues={competitorData || undefined}
                />
              )}
            </CardContent>
          </Card>

          {/* Progress Info */}
          <div className="text-center">
            <p className="text-sm font-onest text-[var(--color-primary-dark)]/60">
              Passo {currentStep} de {steps.length}
            </p>
          </div>
        </>
      )}
    </div>
  )
}
```

---

## 📁 app\app\onboarding

### app/app/onboarding/page.tsx

```tsx
import { Metadata } from 'next'
import { OnboardingWizard } from '@/components/wizards/OnboardingWizard'

export const metadata: Metadata = {
  title: 'Configuração Inicial',
  description: 'Configure sua conta organiQ',
  robots: {
    index: false,
    follow: false,
  },
}

export default function OnboardingPage() {
  return (
    <div className="min-h-screen p-4 md:p-8">
      <div className="max-w-4xl mx-auto">
        <OnboardingWizard />
      </div>
    </div>
  )
}
```

---

## 📁 app\app\planos

### app/app/planos/page.tsx

```tsx
'use client'

import { Check } from 'lucide-react'
import { usePlans } from '@/hooks/usePlans'
import { PlanCard } from '@/components/plans/PlanCard'
import { SkeletonCard } from '@/components/ui/skeleton'
import { formatCurrency } from '@/lib/utils'

export default function PlanosPage() {
  const { plans, selectPlan, isLoadingPlans, isCreatingCheckout, getRecommendedPlan } = usePlans()

  const recommendedPlan = getRecommendedPlan()

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="text-center space-y-3">
        <h1 className="text-3xl md:text-4xl font-bold font-all-round text-[var(--color-primary-dark)]">
          Escolha seu Plano
        </h1>
        <p className="text-lg font-onest text-[var(--color-primary-dark)]/70 max-w-2xl mx-auto">
          Comece a gerar conteúdo de qualidade para seu blog hoje mesmo
        </p>
      </div>

      {/* Loading State */}
      {isLoadingPlans && (
        <div className="grid md:grid-cols-3 gap-6">
          {[1, 2, 3].map((i) => (
            <SkeletonCard key={i} />
          ))}
        </div>
      )}

      {/* Plans Grid */}
      {!isLoadingPlans && (
        <div className="grid md:grid-cols-3 gap-6 max-w-6xl mx-auto">
          {plans.map((plan) => (
            <PlanCard
              key={plan.id}
              plan={plan}
              onSelect={() => selectPlan(plan.id)}
              isRecommended={plan.id === recommendedPlan?.id}
              isLoading={isCreatingCheckout}
            />
          ))}
        </div>
      )}

      {/* Garantia */}
      <div className="max-w-3xl mx-auto text-center space-y-3 pt-8">
        <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-[var(--color-success)]/10 text-[var(--color-success)]">
          <Check className="h-5 w-5" />
          <span className="text-sm font-semibold font-onest">
            Garantia de 7 dias - 100% do seu dinheiro de volta
          </span>
        </div>
        <p className="text-sm font-onest text-[var(--color-primary-dark)]/60">
          Todos os planos incluem suporte via email e atualizações automáticas
        </p>
      </div>
    </div>
  )
}
```

---

## 📁 app\login

### app/login/page.tsx

```tsx
import { Metadata } from 'next'
import Link from 'next/link'
import { ArrowLeft } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { LoginForm } from '@/components/forms/LoginForm'
import { RegisterForm } from '@/components/forms/RegisterForm'

export const metadata: Metadata = {
  title: 'Login',
  description: 'Faça login ou crie sua conta no organiQ',
  robots: {
    index: false,
    follow: false,
  },
}

export default function LoginPage() {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center p-4">
      {/* Back to Home */}
      <div className="w-full max-w-md mb-8">
        <Link
          href="/"
          className="inline-flex items-center gap-2 text-sm font-medium font-onest text-[var(--color-primary-dark)]/70 hover:text-[var(--color-primary-dark)] transition-colors"
        >
          <ArrowLeft className="h-4 w-4" />
          Voltar para o início
        </Link>
      </div>

      {/* Logo */}
      <div className="mb-8 text-center">
        <h1 className="text-3xl font-bold font-all-round text-[var(--color-primary-purple)] mb-2">
          organiQ
        </h1>
        <p className="text-sm font-onest text-[var(--color-primary-teal)]">
          Naturalmente Inteligente
        </p>
      </div>

      {/* Login/Register Card */}
      <Card className="w-full max-w-md shadow-lg">
        <CardHeader className="space-y-1 pb-4">
          <CardTitle className="text-2xl text-center">
            Bem-vindo
          </CardTitle>
          <CardDescription className="text-center">
            Entre na sua conta ou crie uma nova para começar
          </CardDescription>
        </CardHeader>

        <CardContent>
          <Tabs defaultValue="login" className="w-full">
            <TabsList className="grid w-full grid-cols-2 mb-6">
              <TabsTrigger value="login">Entrar</TabsTrigger>
              <TabsTrigger value="register">Cadastrar</TabsTrigger>
            </TabsList>

            <TabsContent value="login">
              <LoginForm />
            </TabsContent>

            <TabsContent value="register">
              <RegisterForm />
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>

      {/* Footer Note */}
      <div className="mt-8 text-center">
        <p className="text-xs font-onest text-[var(--color-primary-dark)]/60">
          Ao continuar, você concorda com nossos{' '}
          <a href="#" className="text-[var(--color-primary-purple)] hover:underline">
            Termos de Uso
          </a>
          {' e '}
          <a href="#" className="text-[var(--color-primary-purple)] hover:underline">
            Política de Privacidade
          </a>
        </p>
      </div>
    </div>
  )
}
```

---

## 📁 components\articles

### components/articles/ArticleCard.tsx

```tsx
'use client'

import { ExternalLink, AlertCircle, Loader2 } from 'lucide-react'
import { formatDateTime } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardFooter, CardHeader } from '@/components/ui/card'
import { cn } from '@/lib/utils'
import type { Article } from '@/types'

interface ArticleCardProps {
  article: Article
  onViewError?: (article: Article) => void
  onRepublish?: (id: string) => void
  isRepublishing?: boolean
}

const statusConfig = {
  generating: {
    color: 'bg-[var(--color-warning)]',
    textColor: 'text-[var(--color-warning)]',
    label: 'Gerando...',
    icon: Loader2,
  },
  publishing: {
    color: 'bg-blue-500',
    textColor: 'text-blue-500',
    label: 'Publicando...',
    icon: Loader2,
  },
  published: {
    color: 'bg-[var(--color-success)]',
    textColor: 'text-[var(--color-success)]',
    label: 'Publicado',
    icon: ExternalLink,
  },
  error: {
    color: 'bg-[var(--color-error)]',
    textColor: 'text-[var(--color-error)]',
    label: 'Erro',
    icon: AlertCircle,
  },
}

export function ArticleCard({
  article,
  onViewError,
  onRepublish,
  isRepublishing,
}: ArticleCardProps) {
  const status = statusConfig[article.status]
  const StatusIcon = status.icon

  return (
    <Card className="hover:shadow-md transition-shadow duration-200">
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between gap-3">
          <h3 className="text-lg font-semibold font-all-round text-[var(--color-primary-dark)] line-clamp-2">
            {article.title}
          </h3>
          <div
            className={cn(
              'flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium font-onest shrink-0',
              status.color,
              'text-white'
            )}
          >
            <StatusIcon
              className={cn('h-3.5 w-3.5', {
                'animate-spin': article.status === 'generating' || article.status === 'publishing',
              })}
            />
            {status.label}
          </div>
        </div>
      </CardHeader>

      <CardContent className="pb-3">
        <div className="flex items-center gap-2 text-sm font-onest text-[var(--color-primary-dark)]/60">
          <span>{formatDateTime(article.createdAt)}</span>
        </div>
      </CardContent>

      <CardFooter className="pt-3 border-t border-[var(--color-border)]">
        {article.status === 'published' && article.postUrl && (
          <a
            href={article.postUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="w-full"
          >
            <Button variant="outline" size="sm" className="w-full">
              <ExternalLink className="h-4 w-4 mr-2" />
              Ver Publicação
            </Button>
          </a>
        )}

        {article.status === 'error' && (
          <div className="w-full space-y-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => onViewError?.(article)}
              className="w-full"
            >
              <AlertCircle className="h-4 w-4 mr-2" />
              Ver Detalhes
            </Button>
            {onRepublish && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onRepublish(article.id)}
                disabled={isRepublishing}
                className="w-full text-[var(--color-primary-purple)]"
              >
                Tentar Republicar
              </Button>
            )}
          </div>
        )}

        {(article.status === 'generating' || article.status === 'publishing') && (
          <div className="w-full text-center text-sm font-onest text-[var(--color-primary-dark)]/60">
            Aguarde...
          </div>
        )}
      </CardFooter>
    </Card>
  )
}
```

---

### components/articles/ArticleIdeaCard.tsx

```tsx
'use client'

import { useState, useEffect } from 'react'
import { Check, X, MessageSquare } from 'lucide-react'
import { Card, CardContent, CardFooter, CardHeader } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { cn } from '@/lib/utils'
import { debounce } from '@/lib/utils'
import type { ArticleIdea } from '@/types'

interface ArticleIdeaCardProps {
  idea: ArticleIdea
  onUpdate: (id: string, updates: Partial<ArticleIdea>) => void
}

export function ArticleIdeaCard({ idea, onUpdate }: ArticleIdeaCardProps) {
  const [localFeedback, setLocalFeedback] = useState(idea.feedback || '')

  // Debounced update para feedback
  useEffect(() => {
    const debouncedUpdate = debounce(() => {
      if (localFeedback !== idea.feedback) {
        onUpdate(idea.id, { feedback: localFeedback })
      }
    }, 1000)

    debouncedUpdate()
  }, [localFeedback, idea.id, idea.feedback, onUpdate])

  const handleToggleApprove = (approved: boolean) => {
    onUpdate(idea.id, { approved })
  }

  return (
    <Card
      className={cn(
        'transition-all duration-200',
        idea.approved && 'border-l-4 border-l-[var(--color-success)]',
        !idea.approved && 'opacity-60'
      )}
    >
      <CardHeader>
        <div className="flex items-start justify-between gap-3">
          <h3 className="text-lg font-semibold font-all-round text-[var(--color-primary-dark)] line-clamp-2 flex-1">
            {idea.title}
          </h3>
          {idea.approved && (
            <div className="flex items-center gap-1 px-2 py-1 rounded-full bg-[var(--color-success)]/10 text-[var(--color-success)] text-xs font-medium shrink-0">
              <Check className="h-3.5 w-3.5" />
              Aprovado
            </div>
          )}
        </div>
      </CardHeader>

      <CardContent className="space-y-4">
        {/* Summary */}
        <p className="text-sm font-onest text-[var(--color-primary-dark)]/80 line-clamp-3">
          {idea.summary}
        </p>

        {/* Toggle Buttons */}
        <div className="flex gap-2">
          <Button
            variant={idea.approved ? 'success' : 'outline'}
            size="sm"
            onClick={() => handleToggleApprove(true)}
            className={cn(
              'flex-1',
              idea.approved && 'bg-[var(--color-success)] text-white hover:bg-[var(--color-success)]/90'
            )}
          >
            <Check className="h-4 w-4 mr-2" />
            Aprovar
          </Button>
          <Button
            variant={!idea.approved ? 'outline' : 'ghost'}
            size="sm"
            onClick={() => handleToggleApprove(false)}
            className={cn(
              'flex-1',
              !idea.approved && 'border-[var(--color-primary-dark)]/20'
            )}
          >
            <X className="h-4 w-4 mr-2" />
            Rejeitar
          </Button>
        </div>

        {/* Feedback Field */}
        <div className="space-y-2 pt-2 border-t border-[var(--color-border)]">
          <div className="flex items-center gap-2">
            <MessageSquare className="h-4 w-4 text-[var(--color-primary-teal)]" />
            <Label htmlFor={`feedback-${idea.id}`} className="text-xs">
              Sugestões ou direcionamentos (opcional)
            </Label>
          </div>
          <Textarea
            id={`feedback-${idea.id}`}
            value={localFeedback}
            onChange={(e) => setLocalFeedback(e.target.value)}
            placeholder="Ex: Foque em pequenas empresas, adicione exemplos práticos..."
            className="min-h-[60px] max-h-[100px] text-sm bg-[var(--color-secondary-cream)]/50 border-[var(--color-primary-teal)]/30 focus:border-[var(--color-primary-purple)]"
            maxLength={500}
            showCount
          />
        </div>

        {/* Badges */}
        <div className="flex items-center gap-2 flex-wrap">
          {idea.approved && localFeedback && (
            <div className="px-2 py-1 rounded-full bg-[var(--color-primary-purple)]/10 text-[var(--color-primary-purple)] text-xs font-medium">
              Com direcionamento
            </div>
          )}
          {!idea.approved && localFeedback && (
            <div className="px-2 py-1 rounded-full bg-[var(--color-primary-dark)]/10 text-[var(--color-primary-dark)]/60 text-xs font-medium">
              Feedback enviado
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
```

---

### components/articles/ArticleTable.tsx

```tsx
'use client'

import { ExternalLink, AlertCircle, Loader2 } from 'lucide-react'
import { formatDateTime, truncate } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import type { Article } from '@/types'

interface ArticleTableProps {
  articles: Article[]
  onViewError?: (article: Article) => void
  onRepublish?: (id: string) => void
  isRepublishing?: boolean
}

const statusConfig = {
  generating: {
    color: 'bg-[var(--color-warning)]/10',
    textColor: 'text-[var(--color-warning)]',
    label: 'Gerando...',
    icon: Loader2,
  },
  publishing: {
    color: 'bg-blue-500/10',
    textColor: 'text-blue-500',
    label: 'Publicando...',
    icon: Loader2,
  },
  published: {
    color: 'bg-[var(--color-success)]/10',
    textColor: 'text-[var(--color-success)]',
    label: 'Publicado',
    icon: ExternalLink,
  },
  error: {
    color: 'bg-[var(--color-error)]/10',
    textColor: 'text-[var(--color-error)]',
    label: 'Erro',
    icon: AlertCircle,
  },
}

export function ArticleTable({
  articles,
  onViewError,
  onRepublish,
  isRepublishing,
}: ArticleTableProps) {
  return (
    <div className="w-full overflow-x-auto">
      <table className="w-full">
        <thead>
          <tr className="border-b border-[var(--color-border)]">
            <th className="text-left py-3 px-4 text-sm font-semibold font-all-round text-[var(--color-primary-dark)]">
              Título
            </th>
            <th className="text-left py-3 px-4 text-sm font-semibold font-all-round text-[var(--color-primary-dark)] min-w-[150px]">
              Data
            </th>
            <th className="text-left py-3 px-4 text-sm font-semibold font-all-round text-[var(--color-primary-dark)] min-w-[120px]">
              Status
            </th>
            <th className="text-right py-3 px-4 text-sm font-semibold font-all-round text-[var(--color-primary-dark)] min-w-[140px]">
              Ações
            </th>
          </tr>
        </thead>
        <tbody>
          {articles.map((article) => {
            const status = statusConfig[article.status]
            const StatusIcon = status.icon

            return (
              <tr
                key={article.id}
                className="border-b border-[var(--color-border)] hover:bg-[var(--color-primary-dark)]/5 transition-colors"
              >
                {/* Título */}
                <td className="py-4 px-4">
                  <div className="flex items-center gap-2">
                    <span
                      className="font-medium font-onest text-[var(--color-primary-dark)]"
                      title={article.title}
                    >
                      {truncate(article.title, 60)}
                    </span>
                  </div>
                </td>

                {/* Data */}
                <td className="py-4 px-4">
                  <span className="text-sm font-onest text-[var(--color-primary-dark)]/70">
                    {formatDateTime(article.createdAt)}
                  </span>
                </td>

                {/* Status */}
                <td className="py-4 px-4">
                  <div
                    className={cn(
                      'inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium font-onest',
                      status.color,
                      status.textColor
                    )}
                  >
                    <StatusIcon
                      className={cn('h-3.5 w-3.5', {
                        'animate-spin':
                          article.status === 'generating' || article.status === 'publishing',
                      })}
                    />
                    {status.label}
                  </div>
                </td>

                {/* Ações */}
                <td className="py-4 px-4">
                  <div className="flex items-center justify-end gap-2">
                    {article.status === 'published' && article.postUrl && (
                      <a
                        href={article.postUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        <Button variant="outline" size="sm">
                          <ExternalLink className="h-3.5 w-3.5 mr-1.5" />
                          Ver Post
                        </Button>
                      </a>
                    )}

                    {article.status === 'error' && (
                      <>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => onViewError?.(article)}
                        >
                          <AlertCircle className="h-3.5 w-3.5 mr-1.5" />
                          Detalhes
                        </Button>
                        {onRepublish && (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => onRepublish(article.id)}
                            disabled={isRepublishing}
                            className="text-[var(--color-primary-purple)]"
                          >
                            Republicar
                          </Button>
                        )}
                      </>
                    )}

                    {(article.status === 'generating' || article.status === 'publishing') && (
                      <span className="text-sm font-onest text-[var(--color-primary-dark)]/60">
                        Aguarde...
                      </span>
                    )}
                  </div>
                </td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}
```

---

## 📁 components\forms

### components/forms/BusinessInfoForm.tsx

```tsx
'use client'

import { useForm, useFieldArray } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Plus, X, Upload } from 'lucide-react'
import { businessSchema, type BusinessInput } from '@/lib/validations'
import { OBJECTIVES } from '@/lib/constants'
import { useUser } from '@/store/authStore'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Slider } from '@/components/ui/slider'
import { useState } from 'react'

interface BusinessInfoFormProps {
  onSubmit: (data: BusinessInput) => void
  isLoading?: boolean
  defaultValues?: Partial<BusinessInput>
}

export function BusinessInfoForm({ onSubmit, isLoading, defaultValues }: BusinessInfoFormProps) {
  const user = useUser()
  const [selectedFile, setSelectedFile] = useState<File | null>(null)

  const {
    register,
    handleSubmit,
    watch,
    setValue,
    control,
    formState: { errors },
  } = useForm<BusinessInput>({
    resolver: zodResolver(businessSchema),
    defaultValues: {
      description: '',
      hasBlog: false,
      blogUrls: [],
      articleCount: 1,
      ...defaultValues,
    },
  })

  const { fields, append, remove } = useFieldArray({
    control,
    name: 'blogUrls',
  })

  const watchPrimaryObjective = watch('primaryObjective')
  const watchHasBlog = watch('hasBlog')
  const watchArticleCount = watch('articleCount')

  const availableSecondaryObjectives = OBJECTIVES.filter(
    (obj) => obj.value !== watchPrimaryObjective
  )

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      setSelectedFile(file)
      setValue('brandFile', file)
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {/* Descrição do Negócio */}
      <div className="space-y-2">
        <Label htmlFor="description" required>
          Descreva seu negócio
        </Label>
        <Textarea
          id="description"
          placeholder="Ex: Somos uma agência de marketing digital especializada em pequenas empresas..."
          maxLength={500}
          showCount
          error={errors.description?.message}
          {...register('description')}
        />
        <p className="text-xs text-[var(--color-primary-dark)]/60 font-onest">
          Quanto mais detalhes, melhor será o conteúdo gerado
        </p>
      </div>

      {/* Objetivos */}
      <div className="space-y-4">
        <Label required>Quais são seus objetivos?</Label>

        {/* Objetivo Principal */}
        <div className="space-y-2">
          <Label htmlFor="primaryObjective">Objetivo Principal</Label>
          <Select
            value={watchPrimaryObjective}
            onValueChange={(value) => setValue('primaryObjective', value as any)}
          >
            <SelectTrigger error={errors.primaryObjective?.message}>
              <SelectValue placeholder="Selecione seu objetivo principal" />
            </SelectTrigger>
            <SelectContent>
              {OBJECTIVES.map((obj) => (
                <SelectItem key={obj.value} value={obj.value}>
                  {obj.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* Objetivo Secundário */}
        {watchPrimaryObjective && (
          <div className="space-y-2">
            <Label htmlFor="secondaryObjective">Objetivo Secundário (opcional)</Label>
            <Select
              value={watch('secondaryObjective') || ''}
              onValueChange={(value) => setValue('secondaryObjective', value as any || undefined)}
            >
              <SelectTrigger error={errors.secondaryObjective?.message}>
                <SelectValue placeholder="Selecione um objetivo secundário (opcional)" />
              </SelectTrigger>
              <SelectContent>
                {availableSecondaryObjectives.map((obj) => (
                  <SelectItem key={obj.value} value={obj.value}>
                    {obj.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <p className="text-xs text-[var(--color-primary-dark)]/60 font-onest">
              Um objetivo secundário ajuda a criar conteúdo mais diversificado
            </p>
          </div>
        )}
      </div>

      {/* URL do Site */}
      <div className="space-y-2">
        <Label htmlFor="siteUrl">URL do seu site (opcional)</Label>
        <Input
          id="siteUrl"
          type="url"
          placeholder="https://seusite.com.br"
          error={errors.siteUrl?.message}
          {...register('siteUrl')}
        />
      </div>

      {/* Tem Blog? */}
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <input
            type="checkbox"
            id="hasBlog"
            className="h-4 w-4 rounded border-[var(--color-border)] text-[var(--color-primary-purple)] focus:ring-[var(--color-primary-purple)]"
            {...register('hasBlog')}
          />
          <Label htmlFor="hasBlog" className="cursor-pointer">
            Meu site já tem um blog
          </Label>
        </div>
      </div>

      {/* URLs do Blog */}
      {watchHasBlog && (
        <div className="space-y-2">
          <Label>URLs do blog</Label>
          <div className="space-y-2">
            {fields.map((field, index) => (
              <div key={field.id} className="flex gap-2">
                <Input
                  type="url"
                  placeholder="https://seusite.com.br/blog"
                  error={errors.blogUrls?.[index]?.message}
                  {...register(`blogUrls.${index}` as const)}
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  onClick={() => remove(index)}
                >
                  <X className="h-4 w-4" />
                </Button>
              </div>
            ))}
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => append('')}
            >
              <Plus className="h-4 w-4 mr-2" />
              Adicionar URL
            </Button>
          </div>
        </div>
      )}

      {/* Quantidade de Matérias */}
      <div className="space-y-2">
        <Label required>Quantas matérias deseja criar?</Label>
        <Slider
          min={1}
          max={user?.maxArticles || 50}
          step={1}
          value={[watchArticleCount || 1]}
          onValueChange={(value) => setValue('articleCount', value[0])}
          showValue
          formatValue={(value) => `${value} ${value === 1 ? 'matéria' : 'matérias'}`}
        />
        {errors.articleCount && (
          <p className="text-xs text-[var(--color-error)] font-onest">
            {errors.articleCount.message}
          </p>
        )}
      </div>

      {/* Upload Manual da Marca */}
      <div className="space-y-2">
        <Label htmlFor="brandFile">Manual da marca (opcional)</Label>
        <div className="flex items-center gap-4">
          <label
            htmlFor="brandFile"
            className="flex items-center gap-2 px-4 py-2 rounded-[var(--radius-sm)] border-2 border-dashed border-[var(--color-border)] hover:border-[var(--color-primary-purple)] transition-colors cursor-pointer"
          >
            <Upload className="h-4 w-4" />
            <span className="text-sm font-onest">
              {selectedFile ? selectedFile.name : 'Escolher arquivo'}
            </span>
          </label>
          <input
            id="brandFile"
            type="file"
            accept=".pdf,.jpg,.jpeg,.png"
            className="hidden"
            onChange={handleFileChange}
          />
        </div>
        <p className="text-xs text-[var(--color-primary-dark)]/60 font-onest">
          PDF, JPG ou PNG (máx. 5MB)
        </p>
        {errors.brandFile && (
          <p className="text-xs text-[var(--color-error)] font-onest">
            {errors.brandFile.message}
          </p>
        )}
      </div>

      {/* Submit Button */}
      <div className="flex justify-end pt-4">
        <Button
          type="submit"
          variant="secondary"
          size="lg"
          isLoading={isLoading}
          disabled={isLoading}
        >
          Próximo
        </Button>
      </div>
    </form>
  )
}
```

---

### components/forms/CompetitorsForm.tsx

```tsx
'use client'

import { useForm, useFieldArray } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Plus, X } from 'lucide-react'
import { competitorsSchema, type CompetitorsInput } from '@/lib/validations'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

interface CompetitorsFormProps {
  onSubmit: (data: CompetitorsInput) => void
  onBack: () => void
  isLoading?: boolean
  defaultValues?: Partial<CompetitorsInput>
}

export function CompetitorsForm({
  onSubmit,
  onBack,
  isLoading,
  defaultValues,
}: CompetitorsFormProps) {
  const {
    register,
    handleSubmit,
    control,
    formState: { errors },
  } = useForm<CompetitorsInput>({
    resolver: zodResolver(competitorsSchema),
    defaultValues: {
      competitorUrls: defaultValues?.competitorUrls || [],
    },
  })

  const { fields, append, remove } = useFieldArray({
    control,
    name: 'competitorUrls',
  })

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {/* Header */}
      <div className="space-y-2">
        <h3 className="text-xl font-semibold font-all-round text-[var(--color-primary-dark)]">
          Concorrentes (Opcional)
        </h3>
        <p className="text-sm font-onest text-[var(--color-primary-dark)]/70">
          Adicione URLs de concorrentes para criar uma estratégia de SEO mais competitiva. Esta etapa é opcional, mas recomendada.
        </p>
      </div>

      {/* Lista de URLs */}
      <div className="space-y-3">
        {fields.length === 0 ? (
          <div className="text-center py-8 px-4 border-2 border-dashed border-[var(--color-border)] rounded-[var(--radius-md)]">
            <p className="text-sm font-onest text-[var(--color-primary-dark)]/60 mb-4">
              Nenhum concorrente adicionado ainda
            </p>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => append('')}
            >
              <Plus className="h-4 w-4 mr-2" />
              Adicionar primeiro concorrente
            </Button>
          </div>
        ) : (
          <>
            {fields.map((field, index) => (
              <div key={field.id} className="space-y-2">
                <div className="flex items-start gap-2">
                  <div className="flex-1">
                    <Label htmlFor={`competitor-${index}`}>
                      Concorrente {index + 1}
                    </Label>
                    <div className="mt-1 flex gap-2">
                      <Input
                        id={`competitor-${index}`}
                        type="url"
                        placeholder="https://concorrente.com.br"
                        error={errors.competitorUrls?.[index]?.message}
                        {...register(`competitorUrls.${index}` as const)}
                      />
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        onClick={() => remove(index)}
                        className="flex-shrink-0"
                      >
                        <X className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                </div>
              </div>
            ))}

            {/* Add More Button */}
            {fields.length < 10 && (
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => append('')}
                className="w-full"
              >
                <Plus className="h-4 w-4 mr-2" />
                Adicionar concorrente ({fields.length}/10)
              </Button>
            )}

            {fields.length >= 10 && (
              <p className="text-xs text-[var(--color-warning)] font-onest text-center">
                Limite de 10 concorrentes atingido
              </p>
            )}
          </>
        )}
      </div>

      {/* Info Box */}
      <div className="bg-[var(--color-primary-purple)]/5 border border-[var(--color-primary-purple)]/20 rounded-[var(--radius-md)] p-4">
        <p className="text-sm font-onest text-[var(--color-primary-dark)]/80">
          💡 <strong>Dica:</strong> Adicione sites que produzem conteúdo similar ao seu. Nossa IA analisará suas estratégias de SEO para criar matérias ainda melhores.
        </p>
      </div>

      {/* Action Buttons */}
      <div className="flex items-center justify-between pt-4">
        <Button
          type="button"
          variant="outline"
          onClick={onBack}
          disabled={isLoading}
        >
          Voltar
        </Button>

        <Button
          type="submit"
          variant="secondary"
          size="lg"
          isLoading={isLoading}
          disabled={isLoading}
        >
          {fields.length === 0 ? 'Pular esta etapa' : 'Próximo'}
        </Button>
      </div>
    </form>
  )
}
```

---

### components/forms/IntegrationsForm.tsx

```tsx
'use client'

import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { HelpCircle, Check } from 'lucide-react'
import * as Accordion from '@radix-ui/react-accordion'
import { integrationsSchema, type IntegrationsInput } from '@/lib/validations'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { cn } from '@/lib/utils'

interface IntegrationsFormProps {
  onSubmit: (data: IntegrationsInput) => void
  onBack: () => void
  isLoading?: boolean
  defaultValues?: Partial<IntegrationsInput>
}

export function IntegrationsForm({
  onSubmit,
  onBack,
  isLoading,
  defaultValues,
}: IntegrationsFormProps) {
  const {
    register,
    handleSubmit,
    watch,
    setValue,
    formState: { errors },
  } = useForm<IntegrationsInput>({
    resolver: zodResolver(integrationsSchema),
    defaultValues: {
      wordpress: {
        siteUrl: '',
        username: '',
        appPassword: '',
        ...defaultValues?.wordpress,
      },
      searchConsole: {
        enabled: false,
        ...defaultValues?.searchConsole,
      },
      analytics: {
        enabled: false,
        ...defaultValues?.analytics,
      },
    },
  })

  const watchSearchConsoleEnabled = watch('searchConsole.enabled')
  const watchAnalyticsEnabled = watch('analytics.enabled')

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {/* Header */}
      <div className="space-y-2">
        <h3 className="text-xl font-semibold font-all-round text-[var(--color-primary-dark)]">
          Integrações
        </h3>
        <p className="text-sm font-onest text-[var(--color-primary-dark)]/70">
          Conecte suas ferramentas para publicação automática e análise de resultados
        </p>
      </div>

      {/* Accordion */}
      <Accordion.Root type="multiple" defaultValue={['wordpress']} className="space-y-4">
        {/* WordPress (Obrigatório) */}
        <Accordion.Item value="wordpress">
          <div className="border-2 border-[var(--color-primary-purple)] rounded-[var(--radius-md)] overflow-hidden">
            <Accordion.Header>
              <Accordion.Trigger className="flex items-center justify-between w-full p-4 hover:bg-[var(--color-primary-purple)]/5 transition-colors">
                <div className="flex items-center gap-3">
                  <div className="flex items-center justify-center h-10 w-10 rounded-full bg-[var(--color-primary-purple)]/10">
                    <span className="text-xl">🔌</span>
                  </div>
                  <div className="text-left">
                    <h4 className="text-base font-semibold font-all-round text-[var(--color-primary-dark)]">
                      WordPress
                    </h4>
                    <p className="text-xs font-onest text-[var(--color-primary-dark)]/60">
                      Obrigatório para publicação automática
                    </p>
                  </div>
                </div>
                <span className="px-3 py-1 rounded-full bg-[var(--color-primary-purple)]/10 text-xs font-semibold font-onest text-[var(--color-primary-purple)]">
                  Obrigatório
                </span>
              </Accordion.Trigger>
            </Accordion.Header>

            <Accordion.Content className="p-4 pt-0 space-y-4">
              {/* Site URL */}
              <div className="space-y-2">
                <Label htmlFor="wp-siteUrl" required>
                  URL do site WordPress
                </Label>
                <Input
                  id="wp-siteUrl"
                  type="url"
                  placeholder="https://seusite.com.br"
                  error={errors.wordpress?.siteUrl?.message}
                  {...register('wordpress.siteUrl')}
                />
              </div>

              {/* Username */}
              <div className="space-y-2">
                <Label htmlFor="wp-username" required>
                  Nome de usuário
                </Label>
                <Input
                  id="wp-username"
                  type="text"
                  placeholder="seu_usuario"
                  error={errors.wordpress?.username?.message}
                  {...register('wordpress.username')}
                />
              </div>

              {/* App Password */}
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label htmlFor="wp-appPassword" required>
                    Senha de aplicativo
                  </Label>
                  <Dialog>
                    <DialogTrigger asChild>
                      <button
                        type="button"
                        className="text-[var(--color-primary-teal)] hover:text-[var(--color-primary-purple)] transition-colors"
                      >
                        <HelpCircle className="h-4 w-4" />
                      </button>
                    </DialogTrigger>
                    <DialogContent>
                      <DialogHeader>
                        <DialogTitle>Como obter a senha de aplicativo?</DialogTitle>
                        <DialogDescription className="space-y-2 text-left">
                          <p>1. Acesse seu WordPress: <strong>Usuários → Perfil</strong></p>
                          <p>2. Role até <strong>"Senhas de aplicativo"</strong></p>
                          <p>3. Digite um nome (ex: "organiQ") e clique em <strong>"Adicionar"</strong></p>
                          <p>4. Copie a senha gerada e cole aqui</p>
                          <p className="text-xs text-[var(--color-warning)] mt-4">
                            ⚠️ A senha só é exibida uma vez. Guarde-a em local seguro.
                          </p>
                        </DialogDescription>
                      </DialogHeader>
                    </DialogContent>
                  </Dialog>
                </div>
                <Input
                  id="wp-appPassword"
                  type="password"
                  placeholder="xxxx xxxx xxxx xxxx xxxx xxxx"
                  error={errors.wordpress?.appPassword?.message}
                  {...register('wordpress.appPassword')}
                />
              </div>
            </Accordion.Content>
          </div>
        </Accordion.Item>

        {/* Google Search Console (Opcional) */}
        <Accordion.Item value="searchConsole">
          <div className={cn(
            "border-2 rounded-[var(--radius-md)] overflow-hidden transition-colors",
            watchSearchConsoleEnabled 
              ? "border-[var(--color-primary-teal)]" 
              : "border-[var(--color-border)]"
          )}>
            <Accordion.Header>
              <Accordion.Trigger className="flex items-center justify-between w-full p-4 hover:bg-[var(--color-primary-teal)]/5 transition-colors">
                <div className="flex items-center gap-3">
                  <div className="flex items-center justify-center h-10 w-10 rounded-full bg-[var(--color-primary-teal)]/10">
                    <span className="text-xl">📊</span>
                  </div>
                  <div className="text-left">
                    <h4 className="text-base font-semibold font-all-round text-[var(--color-primary-dark)]">
                      Google Search Console
                    </h4>
                    <p className="text-xs font-onest text-[var(--color-primary-dark)]/60">
                      Análise de palavras-chave e rankings
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  {watchSearchConsoleEnabled && (
                    <Check className="h-5 w-5 text-[var(--color-success)]" />
                  )}
                  <input
                    type="checkbox"
                    checked={watchSearchConsoleEnabled}
                    onChange={(e) => setValue('searchConsole.enabled', e.target.checked)}
                    onClick={(e) => e.stopPropagation()}
                    className="h-4 w-4 rounded border-[var(--color-border)] text-[var(--color-primary-teal)] focus:ring-[var(--color-primary-teal)]"
                  />
                </div>
              </Accordion.Trigger>
            </Accordion.Header>

            {watchSearchConsoleEnabled && (
              <Accordion.Content className="p-4 pt-0 space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="sc-propertyUrl" required>
                    URL da propriedade
                  </Label>
                  <Input
                    id="sc-propertyUrl"
                    type="url"
                    placeholder="https://seusite.com.br"
                    error={errors.searchConsole?.propertyUrl?.message}
                    {...register('searchConsole.propertyUrl')}
                  />
                  <p className="text-xs text-[var(--color-primary-dark)]/60 font-onest">
                    Use a mesma URL cadastrada no Search Console
                  </p>
                </div>
              </Accordion.Content>
            )}
          </div>
        </Accordion.Item>

        {/* Google Analytics (Opcional) */}
        <Accordion.Item value="analytics">
          <div className={cn(
            "border-2 rounded-[var(--radius-md)] overflow-hidden transition-colors",
            watchAnalyticsEnabled 
              ? "border-[var(--color-primary-teal)]" 
              : "border-[var(--color-border)]"
          )}>
            <Accordion.Header>
              <Accordion.Trigger className="flex items-center justify-between w-full p-4 hover:bg-[var(--color-primary-teal)]/5 transition-colors">
                <div className="flex items-center gap-3">
                  <div className="flex items-center justify-center h-10 w-10 rounded-full bg-[var(--color-primary-teal)]/10">
                    <span className="text-xl">📈</span>
                  </div>
                  <div className="text-left">
                    <h4 className="text-base font-semibold font-all-round text-[var(--color-primary-dark)]">
                      Google Analytics
                    </h4>
                    <p className="text-xs font-onest text-[var(--color-primary-dark)]/60">
                      Análise de tráfego e conversões
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  {watchAnalyticsEnabled && (
                    <Check className="h-5 w-5 text-[var(--color-success)]" />
                  )}
                  <input
                    type="checkbox"
                    checked={watchAnalyticsEnabled}
                    onChange={(e) => setValue('analytics.enabled', e.target.checked)}
                    onClick={(e) => e.stopPropagation()}
                    className="h-4 w-4 rounded border-[var(--color-border)] text-[var(--color-primary-teal)] focus:ring-[var(--color-primary-teal)]"
                  />
                </div>
              </Accordion.Trigger>
            </Accordion.Header>

            {watchAnalyticsEnabled && (
              <Accordion.Content className="p-4 pt-0 space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="ga-measurementId" required>
                    ID de Medição GA4
                  </Label>
                  <Input
                    id="ga-measurementId"
                    type="text"
                    placeholder="G-XXXXXXXXXX"
                    error={errors.analytics?.measurementId?.message}
                    {...register('analytics.measurementId')}
                  />
                  <p className="text-xs text-[var(--color-primary-dark)]/60 font-onest">
                    Formato: G-XXXXXXXXXX (encontrado em Admin → Fluxos de dados)
                  </p>
                </div>
              </Accordion.Content>
            )}
          </div>
        </Accordion.Item>
      </Accordion.Root>

      {/* Action Buttons */}
      <div className="flex items-center justify-between pt-4">
        <Button
          type="button"
          variant="outline"
          onClick={onBack}
          disabled={isLoading}
        >
          Voltar
        </Button>

        <Button
          type="submit"
          variant="primary"
          size="lg"
          isLoading={isLoading}
          disabled={isLoading}
        >
          Gerar Ideias
        </Button>
      </div>
    </form>
  )
}
```

---

### components/forms/LoginForm.tsx

```tsx
'use client'

import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { loginSchema, type LoginInput } from '@/lib/validations'
import { useAuth } from '@/hooks/useAuth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

export function LoginForm() {
  const { login, isLoggingIn } = useAuth()

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginInput>({
    resolver: zodResolver(loginSchema),
  })

  const onSubmit = (data: LoginInput) => {
    login(data)
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      {/* Email */}
      <div className="space-y-2">
        <Label htmlFor="login-email" required>
          Email
        </Label>
        <Input
          id="login-email"
          type="email"
          placeholder="seu@email.com"
          error={errors.email?.message}
          {...register('email')}
        />
      </div>

      {/* Senha */}
      <div className="space-y-2">
        <Label htmlFor="login-password" required>
          Senha
        </Label>
        <Input
          id="login-password"
          type="password"
          placeholder="••••••••"
          error={errors.password?.message}
          {...register('password')}
        />
      </div>

      {/* Link Esqueci Senha */}
      <div className="flex justify-end">
        <button
          type="button"
          disabled
          className="text-sm text-[var(--color-primary-teal)] opacity-50 cursor-not-allowed"
        >
          Esqueci minha senha
        </button>
      </div>

      {/* Submit Button */}
      <Button
        type="submit"
        variant="primary"
        className="w-full"
        isLoading={isLoggingIn}
        disabled={isLoggingIn}
      >
        {isLoggingIn ? 'Entrando...' : 'Entrar'}
      </Button>
    </form>
  )
}
```

---

### components/forms/RegisterForm.tsx

```tsx
'use client'

import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { registerSchema, type RegisterInput } from '@/lib/validations'
import { useAuth } from '@/hooks/useAuth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

export function RegisterForm() {
  const { register: registerUser, isRegistering } = useAuth()

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<RegisterInput>({
    resolver: zodResolver(registerSchema),
  })

  const onSubmit = (data: RegisterInput) => {
    registerUser(data)
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      {/* Nome */}
      <div className="space-y-2">
        <Label htmlFor="register-name" required>
          Nome completo
        </Label>
        <Input
          id="register-name"
          type="text"
          placeholder="Seu nome"
          error={errors.name?.message}
          {...register('name')}
        />
      </div>

      {/* Email */}
      <div className="space-y-2">
        <Label htmlFor="register-email" required>
          Email
        </Label>
        <Input
          id="register-email"
          type="email"
          placeholder="seu@email.com"
          error={errors.email?.message}
          {...register('email')}
        />
      </div>

      {/* Senha */}
      <div className="space-y-2">
        <Label htmlFor="register-password" required>
          Senha
        </Label>
        <Input
          id="register-password"
          type="password"
          placeholder="••••••••"
          error={errors.password?.message}
          {...register('password')}
        />
      </div>

      {/* Submit Button */}
      <Button
        type="submit"
        variant="secondary"
        className="w-full"
        isLoading={isRegistering}
        disabled={isRegistering}
      >
        {isRegistering ? 'Criando conta...' : 'Criar conta'}
      </Button>
    </form>
  )
}
```

---

## 📁 components\layouts

### components/layouts/Header.tsx

```tsx
"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";

export function Header() {
  return (
    <header className="sticky top-0 z-50 w-full border-b border-[var(--color-border)] bg-white/95 backdrop-blur supports-[backdrop-filter]:bg-white/60">
      <div className="container mx-auto flex h-16 items-center justify-between px-4">
        {/* Logo */}
        <Link href="/" className="flex items-center gap-2">
          <h1 className="text-2xl font-bold font-all-round text-[var(--color-primary-purple)]">
            organiQ
          </h1>
          <span className="hidden sm:inline text-sm font-onest text-[var(--color-primary-teal)]">
            Naturalmente Inteligente
          </span>
        </Link>

        {/* Navigation */}
        <nav className="hidden md:flex items-center gap-6">
          <a
            href="#features"
            className="text-sm font-medium font-onest text-[var(--color-primary-dark)]/70 hover:text-[var(--color-primary-dark)] transition-colors"
          >
            Recursos
          </a>
          <a
            href="#how-it-works"
            className="text-sm font-medium font-onest text-[var(--color-primary-dark)]/70 hover:text-[var(--color-primary-dark)] transition-colors"
          >
            Como Funciona
          </a>
          <a
            href="#pricing"
            className="text-sm font-medium font-onest text-[var(--color-primary-dark)]/70 hover:text-[var(--color-primary-dark)] transition-colors"
          >
            Preços
          </a>
        </nav>

        {/* CTA Button */}
        <Link href="/login">
          <Button variant="primary" size="md">
            Entrar
          </Button>
        </Link>
      </div>
    </header>
  );
}
```

---

### components/layouts/MobileNav.tsx

```tsx
'use client'

import { usePathname } from 'next/navigation'
import Link from 'next/link'
import { FileText, PlusCircle, Settings, LogOut } from 'lucide-react'
import { useAuth } from '@/hooks/useAuth'
import { cn } from '@/lib/utils'

interface NavItem {
  label: string
  href: string
  icon: React.ComponentType<{ className?: string }>
}

const navItems: NavItem[] = [
  {
    label: 'Matérias',
    href: '/app/materias',
    icon: FileText,
  },
  {
    label: 'Criar',
    href: '/app/novo',
    icon: PlusCircle,
  },
  {
    label: 'Conta',
    href: '/app/conta',
    icon: Settings,
  },
]

export function MobileNav() {
  const pathname = usePathname()
  const { logout, isLoggingOut } = useAuth()

  const handleLogout = () => {
    if (window.confirm('Tem certeza que deseja sair?')) {
      logout()
    }
  }

  return (
    <nav className="lg:hidden fixed bottom-0 left-0 right-0 z-50 bg-white border-t border-[var(--color-border)] shadow-lg">
      <div className="grid grid-cols-4 h-16">
        {navItems.map((item) => {
          const Icon = item.icon
          const isActive = pathname === item.href

          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                'flex flex-col items-center justify-center gap-1 transition-colors duration-200',
                isActive
                  ? 'text-[var(--color-primary-purple)]'
                  : 'text-[var(--color-primary-dark)]/60 hover:text-[var(--color-primary-dark)]'
              )}
            >
              <Icon className="h-5 w-5" />
              <span className="text-xs font-medium font-onest">{item.label}</span>
              {isActive && (
                <div className="absolute top-0 left-0 right-0 h-1 bg-[var(--color-primary-purple)]" />
              )}
            </Link>
          )
        })}

        {/* Logout Button */}
        <button
          onClick={handleLogout}
          disabled={isLoggingOut}
          className={cn(
            'flex flex-col items-center justify-center gap-1 transition-colors duration-200',
            'text-[var(--color-error)] hover:text-[var(--color-error)]/80',
            'disabled:opacity-50 disabled:cursor-not-allowed'
          )}
        >
          <LogOut className="h-5 w-5" />
          <span className="text-xs font-medium font-onest">
            {isLoggingOut ? 'Saindo...' : 'Sair'}
          </span>
        </button>
      </div>
    </nav>
  )
}
```

---

### components/layouts/Sidebar.tsx

```tsx
"use client";

import { usePathname } from "next/navigation";
import Link from "next/link";
import { FileText, PlusCircle, Settings, LogOut } from "lucide-react";
import { useAuth } from "@/hooks/useAuth";
import { cn } from "@/lib/utils";

interface NavItem {
  label: string;
  href: string;
  icon: React.ComponentType<{ className?: string }>;
}

const navItems: NavItem[] = [
  {
    label: "Gerar Matérias",
    href: "/app/novo",
    icon: PlusCircle,
  },
  {
    label: "Minhas Matérias",
    href: "/app/materias",
    icon: FileText,
  },
  {
    label: "Minha Conta",
    href: "/app/conta",
    icon: Settings,
  },
];

export function Sidebar() {
  const pathname = usePathname();
  const { logout, isLoggingOut, user } = useAuth();

  const handleLogout = () => {
    if (window.confirm("Tem certeza que deseja sair?")) {
      logout();
    }
  };

  return (
    <aside className="hidden lg:flex lg:flex-col w-[280px] h-[calc(100vh-32px)] m-4 bg-white rounded-[var(--radius-lg)] shadow-md">
      {/* Logo */}
      <div className="flex items-center justify-center h-20 border-b border-[var(--color-border)]">
        <h1 className="text-2xl font-bold font-all-round text-[var(--color-primary-purple)]">
          organiQ
        </h1>
      </div>

      {/* User Info */}
      {user && (
        <div className="px-4 py-4 border-b border-[var(--color-border)]">
          <div className="flex items-center gap-3">
            <div className="flex items-center justify-center h-10 w-10 rounded-full bg-[var(--color-primary-purple)]/10 text-[var(--color-primary-purple)] font-semibold font-all-round">
              {user.name.charAt(0).toUpperCase()}
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium font-all-round text-[var(--color-primary-dark)] truncate">
                {user.name}
              </p>
              <p className="text-xs font-onest text-[var(--color-primary-dark)]/60 truncate">
                {user.email}
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Navigation */}
      <nav className="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
        {navItems.map((item) => {
          const Icon = item.icon;
          const isActive = pathname === item.href;

          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex items-center gap-3 px-3 py-2.5 rounded-[var(--radius-sm)] text-sm font-medium font-onest transition-colors duration-200",
                isActive
                  ? "bg-[var(--color-primary-purple)]/10 text-[var(--color-primary-purple)] border-l-3 border-[var(--color-primary-purple)]"
                  : "text-[var(--color-primary-dark)]/70 hover:bg-[var(--color-primary-dark)]/5 hover:text-[var(--color-primary-dark)]"
              )}
            >
              <Icon className="h-5 w-5" />
              <span>{item.label}</span>
            </Link>
          );
        })}
      </nav>

      {/* Plan Info */}
      {user && (
        <div className="px-4 py-3 border-t border-[var(--color-border)]">
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium font-onest text-[var(--color-primary-dark)]/70">
                Plano {user.planName}
              </span>
              <span className="text-xs font-semibold font-all-round text-[var(--color-primary-purple)]">
                {user.articlesUsed}/{user.maxArticles}
              </span>
            </div>
            <div className="w-full bg-[var(--color-primary-dark)]/10 rounded-full h-2">
              <div
                className="bg-[var(--color-primary-purple)] h-2 rounded-full transition-all duration-300"
                style={{
                  width: `${(user.articlesUsed / user.maxArticles) * 100}%`,
                }}
              />
            </div>
            <p className="text-xs font-onest text-[var(--color-primary-dark)]/60">
              {user.maxArticles - user.articlesUsed} matérias restantes
            </p>
          </div>
        </div>
      )}

      {/* Logout */}
      <div className="px-3 py-3 border-t border-[var(--color-border)]">
        <button
          onClick={handleLogout}
          disabled={isLoggingOut}
          className={cn(
            "flex items-center gap-3 w-full px-3 py-2.5 rounded-[var(--radius-sm)] text-sm font-medium font-onest transition-colors duration-200",
            "text-[var(--color-error)] hover:bg-[var(--color-error)]/10",
            "disabled:opacity-50 disabled:cursor-not-allowed"
          )}
        >
          <LogOut className="h-5 w-5" />
          <span>{isLoggingOut ? "Saindo..." : "Sair"}</span>
        </button>
      </div>
    </aside>
  );
}
```

---

## 📁 components\plans

### components/plans/PlanCard.tsx

```tsx
'use client'

import { Check } from 'lucide-react'
import { formatCurrency } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { cn } from '@/lib/utils'
import type { Plan } from '@/types'

interface PlanCardProps {
  plan: Plan
  onSelect: () => void
  isRecommended?: boolean
  isLoading?: boolean
}

export function PlanCard({ plan, onSelect, isRecommended, isLoading }: PlanCardProps) {
  return (
    <Card
      className={cn(
        'relative hover:shadow-lg hover:scale-[1.02] transition-all duration-200',
        isRecommended && 'border-2 border-[var(--color-primary-teal)]'
      )}
    >
      {/* Recommended Badge */}
      {isRecommended && (
        <div className="absolute -top-3 left-1/2 -translate-x-1/2">
          <div className="px-4 py-1 rounded-full bg-[var(--color-primary-teal)] text-white text-xs font-bold font-all-round shadow-md">
            Recomendado
          </div>
        </div>
      )}

      <CardHeader className="text-center pt-8">
        <CardTitle className="text-2xl">{plan.name}</CardTitle>
        <CardDescription>
          <div className="mt-4">
            <span className="text-4xl font-bold font-all-round text-[var(--color-primary-dark)]">
              {formatCurrency(plan.price)}
            </span>
            <span className="text-sm font-onest text-[var(--color-primary-dark)]/60">/mês</span>
          </div>
        </CardDescription>
      </CardHeader>

      <CardContent className="space-y-4">
        {/* Destaque Principal */}
        <div className="text-center py-3 px-4 rounded-[var(--radius-md)] bg-[var(--color-primary-purple)]/10">
          <p className="text-lg font-bold font-all-round text-[var(--color-primary-purple)]">
            {plan.maxArticles} matérias/mês
          </p>
        </div>

        {/* Features List */}
        <ul className="space-y-3">
          {plan.features.map((feature, index) => (
            <li key={index} className="flex items-start gap-2">
              <Check className="h-5 w-5 text-[var(--color-success)] mt-0.5 flex-shrink-0" />
              <span className="text-sm font-onest text-[var(--color-primary-dark)]/80">
                {feature}
              </span>
            </li>
          ))}
        </ul>
      </CardContent>

      <CardFooter>
        <Button
          variant="secondary"
          className="w-full"
          size="lg"
          onClick={onSelect}
          isLoading={isLoading}
          disabled={isLoading}
        >
          Escolher Plano
        </Button>
      </CardFooter>
    </Card>
  )
}
```

---

## 📁 components\shared

### components/shared/EmptyState.tsx

```tsx
import { FileText, Search, Inbox, AlertCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface EmptyStateProps {
  icon?: 'article' | 'search' | 'inbox' | 'alert'
  title: string
  description?: string
  action?: {
    label: string
    onClick: () => void
  }
  className?: string
}

const iconMap = {
  article: FileText,
  search: Search,
  inbox: Inbox,
  alert: AlertCircle,
}

export function EmptyState({
  icon = 'inbox',
  title,
  description,
  action,
  className,
}: EmptyStateProps) {
  const Icon = iconMap[icon]

  return (
    <div
      className={cn(
        'flex flex-col items-center justify-center py-12 px-4 text-center',
        className
      )}
    >
      {/* Ícone */}
      <div className="mb-4 rounded-full bg-[var(--color-primary-purple)]/10 p-6">
        <Icon className="h-12 w-12 text-[var(--color-primary-purple)]" />
      </div>

      {/* Título */}
      <h3 className="mb-2 text-xl font-semibold font-all-round text-[var(--color-primary-dark)]">
        {title}
      </h3>

      {/* Descrição */}
      {description && (
        <p className="mb-6 max-w-md text-sm font-onest text-[var(--color-primary-dark)]/70">
          {description}
        </p>
      )}

      {/* Ação */}
      {action && (
        <Button onClick={action.onClick} variant="secondary">
          {action.label}
        </Button>
      )}
    </div>
  )
}

// Variantes pré-definidas para casos comuns
export function EmptyArticles({ onCreate }: { onCreate?: () => void }) {
  return (
    <EmptyState
      icon="article"
      title="Nenhuma matéria criada ainda"
      description="Comece criando sua primeira matéria para aumentar seu tráfego orgânico."
      action={
        onCreate
          ? {
              label: 'Criar Primeira Matéria',
              onClick: onCreate,
            }
          : undefined
      }
    />
  )
}

export function EmptySearch({ query }: { query?: string }) {
  return (
    <EmptyState
      icon="search"
      title="Nenhum resultado encontrado"
      description={
        query
          ? `Não encontramos resultados para "${query}". Tente ajustar sua busca.`
          : 'Nenhum resultado corresponde aos filtros aplicados.'
      }
    />
  )
}

export function EmptyIdeas({ onRegenerate }: { onRegenerate?: () => void }) {
  return (
    <EmptyState
      icon="alert"
      title="Nenhuma ideia gerada"
      description="Não foi possível gerar ideias de matérias. Tente novamente com informações diferentes."
      action={
        onRegenerate
          ? {
              label: 'Tentar Novamente',
              onClick: onRegenerate,
            }
          : undefined
      }
    />
  )
}
```

---

### components/shared/ErrorBoundary.tsx

```tsx
'use client'

import { Component, ReactNode } from 'react'
import { AlertTriangle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'

interface Props {
  children: ReactNode
  fallback?: ReactNode
  onReset?: () => void
}

interface State {
  hasError: boolean
  error: Error | null
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false, error: null }
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    // Log do erro para serviço de monitoramento (ex: Sentry)
    console.error('ErrorBoundary caught an error:', error, errorInfo)
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null })
    this.props.onReset?.()
  }

  render() {
    if (this.state.hasError) {
      // Usar fallback customizado se fornecido
      if (this.props.fallback) {
        return this.props.fallback
      }

      // Fallback padrão
      return (
        <div className="flex min-h-screen items-center justify-center p-4 bg-[var(--color-secondary-cream)]">
          <Card className="max-w-md w-full">
            <CardHeader>
              <div className="flex items-center gap-3">
                <div className="rounded-full bg-[var(--color-error)]/10 p-2">
                  <AlertTriangle className="h-6 w-6 text-[var(--color-error)]" />
                </div>
                <CardTitle>Algo deu errado</CardTitle>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <p className="text-sm font-onest text-[var(--color-primary-dark)]/70">
                Ocorreu um erro inesperado. Você pode tentar recarregar a página ou voltar ao início.
              </p>
              {process.env.NODE_ENV === 'development' && this.state.error && (
                <details className="rounded-[var(--radius-sm)] bg-[var(--color-error)]/5 p-3">
                  <summary className="cursor-pointer text-xs font-semibold font-onest text-[var(--color-error)] mb-2">
                    Detalhes do erro (visível apenas em desenvolvimento)
                  </summary>
                  <pre className="text-xs overflow-auto font-mono text-[var(--color-error)]/80">
                    {this.state.error.message}
                  </pre>
                </details>
              )}
            </CardContent>
            <CardFooter className="flex gap-2">
              <Button
                variant="outline"
                onClick={() => window.location.reload()}
                className="flex-1"
              >
                Recarregar Página
              </Button>
              <Button
                variant="primary"
                onClick={this.handleReset}
                className="flex-1"
              >
                Tentar Novamente
              </Button>
            </CardFooter>
          </Card>
        </div>
      )
    }

    return this.props.children
  }
}

// Componente funcional para uso mais simples
export function ErrorFallback({ 
  error, 
  resetErrorBoundary 
}: { 
  error: Error
  resetErrorBoundary: () => void 
}) {
  return (
    <div className="flex min-h-[400px] items-center justify-center p-4">
      <Card className="max-w-md w-full">
        <CardHeader>
          <div className="flex items-center gap-3">
            <div className="rounded-full bg-[var(--color-error)]/10 p-2">
              <AlertTriangle className="h-6 w-6 text-[var(--color-error)]" />
            </div>
            <CardTitle>Erro ao carregar</CardTitle>
          </div>
        </CardHeader>
        <CardContent>
          <p className="text-sm font-onest text-[var(--color-primary-dark)]/70">
            {error.message || 'Ocorreu um erro ao carregar este conteúdo.'}
          </p>
        </CardContent>
        <CardFooter>
          <Button
            variant="primary"
            onClick={resetErrorBoundary}
            className="w-full"
          >
            Tentar Novamente
          </Button>
        </CardFooter>
      </Card>
    </div>
  )
}
```

---

### components/shared/LoadingSpinner.tsx

```tsx
import { cn } from '@/lib/utils'

interface LoadingSpinnerProps {
  size?: 'sm' | 'md' | 'lg' | 'xl'
  className?: string
  text?: string
}

const sizeMap = {
  sm: 'h-4 w-4',
  md: 'h-8 w-8',
  lg: 'h-12 w-12',
  xl: 'h-16 w-16',
}

export function LoadingSpinner({ 
  size = 'md', 
  className,
  text 
}: LoadingSpinnerProps) {
  return (
    <div className="flex flex-col items-center justify-center gap-3">
      <svg
        className={cn(
          'animate-spin text-[var(--color-primary-purple)]',
          sizeMap[size],
          className
        )}
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
      >
        <circle
          className="opacity-25"
          cx="12"
          cy="12"
          r="10"
          stroke="currentColor"
          strokeWidth="4"
        />
        <path
          className="opacity-75"
          fill="currentColor"
          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
        />
      </svg>
      {text && (
        <p className="text-sm font-onest text-[var(--color-primary-dark)]/70 animate-pulse">
          {text}
        </p>
      )}
    </div>
  )
}

// Variante para overlay de tela cheia
export function LoadingOverlay({ 
  text,
  messages 
}: { 
  text?: string
  messages?: string[]
}) {
  const [currentMessage, setCurrentMessage] = React.useState(0)

  React.useEffect(() => {
    if (!messages || messages.length === 0) return

    const interval = setInterval(() => {
      setCurrentMessage((prev) => (prev + 1) % messages.length)
    }, 3000)

    return () => clearInterval(interval)
  }, [messages])

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-[var(--color-secondary-cream)] backdrop-blur-sm">
      <div className="flex flex-col items-center gap-6 p-8">
        {/* Spinner duplo animado */}
        <div className="relative">
          <div className="h-16 w-16 rounded-full border-4 border-[var(--color-primary-purple)]/20 border-t-[var(--color-primary-purple)] animate-spin" />
          <div className="absolute inset-2 h-12 w-12 rounded-full border-4 border-[var(--color-primary-teal)]/20 border-t-[var(--color-primary-teal)] animate-spin" 
               style={{ animationDirection: 'reverse', animationDuration: '1.5s' }} />
        </div>

        {/* Texto */}
        {messages && messages.length > 0 ? (
          <div className="text-center space-y-2">
            <p className="text-lg font-semibold font-all-round text-[var(--color-primary-dark)] animate-pulse">
              {messages[currentMessage]}
            </p>
            <p className="text-sm font-onest text-[var(--color-primary-dark)]/60">
              Isso pode levar alguns minutos
            </p>
          </div>
        ) : text ? (
          <p className="text-lg font-semibold font-all-round text-[var(--color-primary-dark)] animate-pulse">
            {text}
          </p>
        ) : null}
      </div>
    </div>
  )
}

// Variante inline para conteúdo
export function LoadingContent({ text }: { text?: string }) {
  return (
    <div className="flex items-center justify-center py-12">
      <LoadingSpinner size="lg" text={text} />
    </div>
  )
}

// Variante para botões (já incluída no Button, mas útil standalone)
export function ButtonSpinner() {
  return (
    <svg
      className="animate-spin h-4 w-4"
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 24 24"
    >
      <circle
        className="opacity-25"
        cx="12"
        cy="12"
        r="10"
        stroke="currentColor"
        strokeWidth="4"
      />
      <path
        className="opacity-75"
        fill="currentColor"
        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
      />
    </svg>
  )
}

// Export do React para uso no LoadingOverlay
import * as React from 'react'
```

---

## 📁 components\ui

### components/ui/button.tsx

```tsx
import * as React from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/utils'

const buttonVariants = cva(
  'inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium font-all-round transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 cursor-pointer',
  {
    variants: {
      variant: {
        primary: 
          'bg-[var(--color-primary-purple)] text-white hover:opacity-90 hover:shadow-lg hover:scale-105',
        secondary: 
          'bg-[var(--color-secondary-yellow)] text-[var(--color-primary-dark)] hover:opacity-90 hover:shadow-lg hover:scale-105',
        outline: 
          'border-2 border-[var(--color-primary-teal)] text-[var(--color-primary-teal)] hover:bg-[var(--color-primary-teal)] hover:text-white',
        ghost: 
          'text-[var(--color-primary-dark)] hover:bg-[var(--color-primary-teal)]/10',
        danger: 
          'bg-[var(--color-error)] text-white hover:opacity-90',
        success: 
          'bg-[var(--color-success)] text-white hover:opacity-90',
      },
      size: {
        sm: 'h-9 px-3 text-xs',
        md: 'h-10 px-4 py-2',
        lg: 'h-11 px-8 text-base',
        icon: 'h-10 w-10',
      },
    },
    defaultVariants: {
      variant: 'primary',
      size: 'md',
    },
  }
)

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean
  isLoading?: boolean
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, isLoading, children, disabled, ...props }, ref) => {
    return (
      <button
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        disabled={disabled || isLoading}
        aria-busy={isLoading}
        aria-live={isLoading ? 'polite' : undefined}
        {...props}
      >
        {isLoading && (
          <svg
            className="animate-spin h-4 w-4"
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <circle
              className="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              strokeWidth="4"
            />
            <path
              className="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            />
          </svg>
        )}
        {children}
      </button>
    )
  }
)
Button.displayName = 'Button'

export { Button, buttonVariants }
```

---

### components/ui/card.tsx

```tsx
import * as React from 'react'
import { cn } from '@/lib/utils'

const Card = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div
    ref={ref}
    className={cn(
      'rounded-[var(--radius-md)] bg-white shadow-sm border border-[var(--color-border)]',
      'transition-shadow duration-200',
      className
    )}
    {...props}
  />
))
Card.displayName = 'Card'

const CardHeader = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div
    ref={ref}
    className={cn('flex flex-col space-y-1.5 p-6', className)}
    {...props}
  />
))
CardHeader.displayName = 'CardHeader'

const CardTitle = React.forwardRef<
  HTMLParagraphElement,
  React.HTMLAttributes<HTMLHeadingElement>
>(({ className, ...props }, ref) => (
  <h3
    ref={ref}
    className={cn(
      'text-2xl font-semibold font-all-round leading-none tracking-tight text-[var(--color-primary-dark)]',
      className
    )}
    {...props}
  />
))
CardTitle.displayName = 'CardTitle'

const CardDescription = React.forwardRef<
  HTMLParagraphElement,
  React.HTMLAttributes<HTMLParagraphElement>
>(({ className, ...props }, ref) => (
  <p
    ref={ref}
    className={cn(
      'text-sm font-onest text-[var(--color-primary-dark)]/70',
      className
    )}
    {...props}
  />
))
CardDescription.displayName = 'CardDescription'

const CardContent = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div ref={ref} className={cn('p-6 pt-0', className)} {...props} />
))
CardContent.displayName = 'CardContent'

const CardFooter = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div
    ref={ref}
    className={cn('flex items-center p-6 pt-0', className)}
    {...props}
  />
))
CardFooter.displayName = 'CardFooter'

export { Card, CardHeader, CardFooter, CardTitle, CardDescription, CardContent }
```

---

### components/ui/dialog.tsx

```tsx
import * as React from 'react'
import * as DialogPrimitive from '@radix-ui/react-dialog'
import { X } from 'lucide-react'
import { cn } from '@/lib/utils'

const Dialog = DialogPrimitive.Root

const DialogTrigger = DialogPrimitive.Trigger

const DialogPortal = DialogPrimitive.Portal

const DialogClose = DialogPrimitive.Close

const DialogOverlay = React.forwardRef<
  React.ElementRef<typeof DialogPrimitive.Overlay>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Overlay>
>(({ className, ...props }, ref) => (
  <DialogPrimitive.Overlay
    ref={ref}
    className={cn(
      'fixed inset-0 z-50 bg-black/50 backdrop-blur-sm',
      'data-[state=open]:animate-in data-[state=closed]:animate-out',
      'data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0',
      className
    )}
    {...props}
  />
))
DialogOverlay.displayName = DialogPrimitive.Overlay.displayName

const DialogContent = React.forwardRef<
  React.ElementRef<typeof DialogPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Content>
>(({ className, children, ...props }, ref) => (
  <DialogPortal>
    <DialogOverlay />
    <DialogPrimitive.Content
      ref={ref}
      className={cn(
        'fixed left-[50%] top-[50%] z-50 grid w-full max-w-lg translate-x-[-50%] translate-y-[-50%] gap-4',
        'bg-white p-6 shadow-lg duration-200 rounded-[var(--radius-lg)]',
        'data-[state=open]:animate-in data-[state=closed]:animate-out',
        'data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0',
        'data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95',
        'data-[state=closed]:slide-out-to-left-1/2 data-[state=closed]:slide-out-to-top-[48%]',
        'data-[state=open]:slide-in-from-left-1/2 data-[state=open]:slide-in-from-top-[48%]',
        'sm:rounded-[var(--radius-lg)]',
        className
      )}
      {...props}
    >
      {children}
      <DialogPrimitive.Close className="absolute right-4 top-4 rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:pointer-events-none data-[state=open]:bg-accent data-[state=open]:text-muted-foreground">
        <X className="h-4 w-4" />
        <span className="sr-only">Fechar</span>
      </DialogPrimitive.Close>
    </DialogPrimitive.Content>
  </DialogPortal>
))
DialogContent.displayName = DialogPrimitive.Content.displayName

const DialogHeader = ({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) => (
  <div
    className={cn(
      'flex flex-col space-y-1.5 text-center sm:text-left',
      className
    )}
    {...props}
  />
)
DialogHeader.displayName = 'DialogHeader'

const DialogFooter = ({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) => (
  <div
    className={cn(
      'flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2',
      className
    )}
    {...props}
  />
)
DialogFooter.displayName = 'DialogFooter'

const DialogTitle = React.forwardRef<
  React.ElementRef<typeof DialogPrimitive.Title>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Title>
>(({ className, ...props }, ref) => (
  <DialogPrimitive.Title
    ref={ref}
    className={cn(
      'text-lg font-semibold font-all-round leading-none tracking-tight text-[var(--color-primary-dark)]',
      className
    )}
    {...props}
  />
))
DialogTitle.displayName = DialogPrimitive.Title.displayName

const DialogDescription = React.forwardRef<
  React.ElementRef<typeof DialogPrimitive.Description>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Description>
>(({ className, ...props }, ref) => (
  <DialogPrimitive.Description
    ref={ref}
    className={cn('text-sm font-onest text-[var(--color-primary-dark)]/70', className)}
    {...props}
  />
))
DialogDescription.displayName = DialogPrimitive.Description.displayName

export {
  Dialog,
  DialogPortal,
  DialogOverlay,
  DialogClose,
  DialogTrigger,
  DialogContent,
  DialogHeader,
  DialogFooter,
  DialogTitle,
  DialogDescription,
}
```

---

### components/ui/input.tsx

```tsx
import * as React from 'react'
import { cn } from '@/lib/utils'

export interface InputProps
  extends React.InputHTMLAttributes<HTMLInputElement> {
  error?: string
}

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, error, ...props }, ref) => {
    return (
      <div className="w-full">
        <input
          type={type}
          className={cn(
            'flex h-10 w-full rounded-[var(--radius-sm)] border border-input bg-white px-3 py-2 text-sm font-onest',
            'transition-colors duration-200',
            'placeholder:text-[var(--color-primary-dark)]/40',
            'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--color-primary-purple)] focus-visible:border-transparent',
            'disabled:cursor-not-allowed disabled:opacity-50',
            error && 'border-[var(--color-error)] focus-visible:ring-[var(--color-error)]',
            className
          )}
          ref={ref}
          {...props}
        />
        {error && (
          <p className="mt-1 text-xs text-[var(--color-error)] font-onest">
            {error}
          </p>
        )}
      </div>
    )
  }
)
Input.displayName = 'Input'

export { Input }
```

---

### components/ui/label.tsx

```tsx
import * as React from 'react'
import { cn } from '@/lib/utils'

export interface LabelProps extends React.LabelHTMLAttributes<HTMLLabelElement> {
  required?: boolean
}

const Label = React.forwardRef<HTMLLabelElement, LabelProps>(
  ({ className, children, required, ...props }, ref) => {
    return (
      <label
        ref={ref}
        className={cn(
          'text-sm font-medium font-onest text-[var(--color-primary-dark)] leading-none',
          'peer-disabled:cursor-not-allowed peer-disabled:opacity-70',
          className
        )}
        {...props}
      >
        {children}
        {required && (
          <span className="ml-1 text-[var(--color-error)]">*</span>
        )}
      </label>
    )
  }
)
Label.displayName = 'Label'

export { Label }
```

---

### components/ui/progress.tsx

```tsx
import * as React from 'react'
import * as ProgressPrimitive from '@radix-ui/react-progress'
import { cn } from '@/lib/utils'

interface ProgressProps
  extends React.ComponentPropsWithoutRef<typeof ProgressPrimitive.Root> {
  indicatorClassName?: string
  showLabel?: boolean
}

const Progress = React.forwardRef<
  React.ElementRef<typeof ProgressPrimitive.Root>,
  ProgressProps
>(({ className, value, indicatorClassName, showLabel, ...props }, ref) => (
  <div className="w-full">
    <ProgressPrimitive.Root
      ref={ref}
      className={cn(
        'relative h-4 w-full overflow-hidden rounded-full bg-[var(--color-primary-dark)]/10',
        className
      )}
      {...props}
    >
      <ProgressPrimitive.Indicator
        className={cn(
          'h-full w-full flex-1 bg-[var(--color-primary-purple)] transition-all duration-300 ease-in-out',
          indicatorClassName
        )}
        style={{ transform: `translateX(-${100 - (value || 0)}%)` }}
      />
    </ProgressPrimitive.Root>
    {showLabel && (
      <p className="mt-1 text-xs text-right font-onest text-[var(--color-primary-dark)]/70">
        {value}%
      </p>
    )}
  </div>
))
Progress.displayName = ProgressPrimitive.Root.displayName

export { Progress }
```

---

### components/ui/select.tsx

```tsx
import * as React from 'react'
import * as SelectPrimitive from '@radix-ui/react-select'
import { Check, ChevronDown, ChevronUp } from 'lucide-react'
import { cn } from '@/lib/utils'

const Select = SelectPrimitive.Root

const SelectGroup = SelectPrimitive.Group

const SelectValue = SelectPrimitive.Value

const SelectTrigger = React.forwardRef<
  React.ElementRef<typeof SelectPrimitive.Trigger>,
  React.ComponentPropsWithoutRef<typeof SelectPrimitive.Trigger> & {
    error?: string
  }
>(({ className, children, error, ...props }, ref) => (
  <div className="w-full">
    <SelectPrimitive.Trigger
      ref={ref}
      className={cn(
        'flex h-10 w-full items-center justify-between rounded-[var(--radius-sm)] border border-input bg-white px-3 py-2 text-sm font-onest',
        'transition-colors duration-200',
        'placeholder:text-[var(--color-primary-dark)]/40',
        'focus:outline-none focus:ring-2 focus:ring-[var(--color-primary-purple)] focus:border-transparent',
        'disabled:cursor-not-allowed disabled:opacity-50',
        '[&>span]:line-clamp-1',
        error && 'border-[var(--color-error)] focus:ring-[var(--color-error)]',
        className
      )}
      {...props}
    >
      {children}
      <SelectPrimitive.Icon asChild>
        <ChevronDown className="h-4 w-4 opacity-50" />
      </SelectPrimitive.Icon>
    </SelectPrimitive.Trigger>
    {error && (
      <p className="mt-1 text-xs text-[var(--color-error)] font-onest">
        {error}
      </p>
    )}
  </div>
))
SelectTrigger.displayName = SelectPrimitive.Trigger.displayName

const SelectScrollUpButton = React.forwardRef<
  React.ElementRef<typeof SelectPrimitive.ScrollUpButton>,
  React.ComponentPropsWithoutRef<typeof SelectPrimitive.ScrollUpButton>
>(({ className, ...props }, ref) => (
  <SelectPrimitive.ScrollUpButton
    ref={ref}
    className={cn(
      'flex cursor-default items-center justify-center py-1',
      className
    )}
    {...props}
  >
    <ChevronUp className="h-4 w-4" />
  </SelectPrimitive.ScrollUpButton>
))
SelectScrollUpButton.displayName = SelectPrimitive.ScrollUpButton.displayName

const SelectScrollDownButton = React.forwardRef<
  React.ElementRef<typeof SelectPrimitive.ScrollDownButton>,
  React.ComponentPropsWithoutRef<typeof SelectPrimitive.ScrollDownButton>
>(({ className, ...props }, ref) => (
  <SelectPrimitive.ScrollDownButton
    ref={ref}
    className={cn(
      'flex cursor-default items-center justify-center py-1',
      className
    )}
    {...props}
  >
    <ChevronDown className="h-4 w-4" />
  </SelectPrimitive.ScrollDownButton>
))
SelectScrollDownButton.displayName = SelectPrimitive.ScrollDownButton.displayName

const SelectContent = React.forwardRef<
  React.ElementRef<typeof SelectPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof SelectPrimitive.Content>
>(({ className, children, position = 'popper', ...props }, ref) => (
  <SelectPrimitive.Portal>
    <SelectPrimitive.Content
      ref={ref}
      className={cn(
        'relative z-50 max-h-96 min-w-[8rem] overflow-hidden rounded-[var(--radius-md)] bg-white shadow-md',
        'border border-[var(--color-border)]',
        'data-[state=open]:animate-in data-[state=closed]:animate-out',
        'data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0',
        'data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95',
        'data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2',
        'data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2',
        position === 'popper' &&
          'data-[side=bottom]:translate-y-1 data-[side=left]:-translate-x-1 data-[side=right]:translate-x-1 data-[side=top]:-translate-y-1',
        className
      )}
      position={position}
      {...props}
    >
      <SelectScrollUpButton />
      <SelectPrimitive.Viewport
        className={cn(
          'p-1',
          position === 'popper' &&
            'h-[var(--radix-select-trigger-height)] w-full min-w-[var(--radix-select-trigger-width)]'
        )}
      >
        {children}
      </SelectPrimitive.Viewport>
      <SelectScrollDownButton />
    </SelectPrimitive.Content>
  </SelectPrimitive.Portal>
))
SelectContent.displayName = SelectPrimitive.Content.displayName

const SelectLabel = React.forwardRef<
  React.ElementRef<typeof SelectPrimitive.Label>,
  React.ComponentPropsWithoutRef<typeof SelectPrimitive.Label>
>(({ className, ...props }, ref) => (
  <SelectPrimitive.Label
    ref={ref}
    className={cn('py-1.5 pl-8 pr-2 text-sm font-semibold font-onest', className)}
    {...props}
  />
))
SelectLabel.displayName = SelectPrimitive.Label.displayName

const SelectItem = React.forwardRef<
  React.ElementRef<typeof SelectPrimitive.Item>,
  React.ComponentPropsWithoutRef<typeof SelectPrimitive.Item>
>(({ className, children, ...props }, ref) => (
  <SelectPrimitive.Item
    ref={ref}
    className={cn(
      'relative flex w-full cursor-default select-none items-center rounded-sm py-1.5 pl-8 pr-2 text-sm font-onest outline-none',
      'transition-colors duration-150',
      'focus:bg-[var(--color-primary-purple)]/10 focus:text-[var(--color-primary-dark)]',
      'data-[disabled]:pointer-events-none data-[disabled]:opacity-50',
      className
    )}
    {...props}
  >
    <span className="absolute left-2 flex h-3.5 w-3.5 items-center justify-center">
      <SelectPrimitive.ItemIndicator>
        <Check className="h-4 w-4 text-[var(--color-primary-purple)]" />
      </SelectPrimitive.ItemIndicator>
    </span>

    <SelectPrimitive.ItemText>{children}</SelectPrimitive.ItemText>
  </SelectPrimitive.Item>
))
SelectItem.displayName = SelectPrimitive.Item.displayName

const SelectSeparator = React.forwardRef<
  React.ElementRef<typeof SelectPrimitive.Separator>,
  React.ComponentPropsWithoutRef<typeof SelectPrimitive.Separator>
>(({ className, ...props }, ref) => (
  <SelectPrimitive.Separator
    ref={ref}
    className={cn('-mx-1 my-1 h-px bg-[var(--color-border)]', className)}
    {...props}
  />
))
SelectSeparator.displayName = SelectPrimitive.Separator.displayName

export {
  Select,
  SelectGroup,
  SelectValue,
  SelectTrigger,
  SelectContent,
  SelectLabel,
  SelectItem,
  SelectSeparator,
  SelectScrollUpButton,
  SelectScrollDownButton,
}
```

---

### components/ui/skeleton.tsx

```tsx
import { cn } from '@/lib/utils'

function Skeleton({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn(
        'animate-pulse rounded-[var(--radius-sm)] bg-[var(--color-primary-dark)]/10',
        className
      )}
      {...props}
    />
  )
}

// Skeleton pré-definidos para casos comuns
const SkeletonCard = () => (
  <div className="space-y-3 p-4">
    <Skeleton className="h-4 w-3/4" />
    <Skeleton className="h-4 w-1/2" />
    <Skeleton className="h-20 w-full" />
  </div>
)

const SkeletonTable = ({ rows = 5 }: { rows?: number }) => (
  <div className="space-y-2">
    {Array.from({ length: rows }).map((_, i) => (
      <div key={i} className="flex items-center gap-4">
        <Skeleton className="h-12 w-full" />
      </div>
    ))}
  </div>
)

const SkeletonButton = () => <Skeleton className="h-10 w-24" />

const SkeletonAvatar = () => <Skeleton className="h-12 w-12 rounded-full" />

const SkeletonInput = () => <Skeleton className="h-10 w-full" />

const SkeletonText = ({ lines = 3 }: { lines?: number }) => (
  <div className="space-y-2">
    {Array.from({ length: lines }).map((_, i) => (
      <Skeleton
        key={i}
        className={cn(
          'h-4',
          i === lines - 1 ? 'w-2/3' : 'w-full' // Última linha mais curta
        )}
      />
    ))}
  </div>
)

export {
  Skeleton,
  SkeletonCard,
  SkeletonTable,
  SkeletonButton,
  SkeletonAvatar,
  SkeletonInput,
  SkeletonText,
}
```

---

### components/ui/slider.tsx

```tsx
import * as React from 'react'
import * as SliderPrimitive from '@radix-ui/react-slider'
import { cn } from '@/lib/utils'

interface SliderProps
  extends React.ComponentPropsWithoutRef<typeof SliderPrimitive.Root> {
  showValue?: boolean
  formatValue?: (value: number) => string
}

const Slider = React.forwardRef<
  React.ElementRef<typeof SliderPrimitive.Root>,
  SliderProps
>(({ className, showValue, formatValue, value, ...props }, ref) => {
  const currentValue = Array.isArray(value) ? value[0] : value || 0
  const displayValue = formatValue ? formatValue(currentValue) : currentValue

  return (
    <div className="w-full space-y-2">
      {showValue && (
        <div className="flex items-center justify-between">
          <span className="text-sm font-medium font-onest text-[var(--color-primary-dark)]">
            Quantidade
          </span>
          <span className="text-sm font-semibold font-all-round text-[var(--color-primary-purple)]">
            {displayValue}
          </span>
        </div>
      )}
      <SliderPrimitive.Root
        ref={ref}
        className={cn(
          'relative flex w-full touch-none select-none items-center',
          className
        )}
        value={value}
        {...props}
      >
        <SliderPrimitive.Track className="relative h-2 w-full grow overflow-hidden rounded-full bg-[var(--color-primary-dark)]/10">
          <SliderPrimitive.Range className="absolute h-full bg-[var(--color-primary-purple)]" />
        </SliderPrimitive.Track>
        <SliderPrimitive.Thumb className="block h-5 w-5 rounded-full border-2 border-[var(--color-primary-purple)] bg-white ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 cursor-pointer hover:scale-110" />
      </SliderPrimitive.Root>
    </div>
  )
})
Slider.displayName = SliderPrimitive.Root.displayName

export { Slider }
```

---

### components/ui/tabs.tsx

```tsx
import * as React from 'react'
import * as TabsPrimitive from '@radix-ui/react-tabs'
import { cn } from '@/lib/utils'

const Tabs = TabsPrimitive.Root

const TabsList = React.forwardRef<
  React.ElementRef<typeof TabsPrimitive.List>,
  React.ComponentPropsWithoutRef<typeof TabsPrimitive.List>
>(({ className, ...props }, ref) => (
  <TabsPrimitive.List
    ref={ref}
    className={cn(
      'inline-flex h-10 items-center justify-center rounded-[var(--radius-md)] bg-[var(--color-primary-dark)]/5 p-1',
      className
    )}
    {...props}
  />
))
TabsList.displayName = TabsPrimitive.List.displayName

const TabsTrigger = React.forwardRef<
  React.ElementRef<typeof TabsPrimitive.Trigger>,
  React.ComponentPropsWithoutRef<typeof TabsPrimitive.Trigger>
>(({ className, ...props }, ref) => (
  <TabsPrimitive.Trigger
    ref={ref}
    className={cn(
      'inline-flex items-center justify-center whitespace-nowrap rounded-sm px-3 py-1.5 text-sm font-medium font-all-round',
      'transition-all duration-200',
      'ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
      'disabled:pointer-events-none disabled:opacity-50',
      'text-[var(--color-primary-dark)]/70',
      'data-[state=active]:bg-white data-[state=active]:text-[var(--color-primary-purple)] data-[state=active]:shadow-sm',
      'hover:text-[var(--color-primary-dark)]',
      className
    )}
    {...props}
  />
))
TabsTrigger.displayName = TabsPrimitive.Trigger.displayName

const TabsContent = React.forwardRef<
  React.ElementRef<typeof TabsPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof TabsPrimitive.Content>
>(({ className, ...props }, ref) => (
  <TabsPrimitive.Content
    ref={ref}
    className={cn(
      'mt-2 ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
      className
    )}
    {...props}
  />
))
TabsContent.displayName = TabsPrimitive.Content.displayName

export { Tabs, TabsList, TabsTrigger, TabsContent }
```

---

### components/ui/textarea.tsx

```tsx
import * as React from 'react'
import { cn } from '@/lib/utils'

export interface TextareaProps
  extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  error?: string
  showCount?: boolean
  maxLength?: number
}

const Textarea = React.forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ className, error, showCount, maxLength, value, ...props }, ref) => {
    const currentLength = typeof value === 'string' ? value.length : 0

    return (
      <div className="w-full">
        <textarea
          className={cn(
            'flex min-h-[80px] w-full rounded-[var(--radius-sm)] border border-input bg-white px-3 py-2 text-sm font-onest',
            'transition-colors duration-200',
            'placeholder:text-[var(--color-primary-dark)]/40',
            'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--color-primary-purple)] focus-visible:border-transparent',
            'disabled:cursor-not-allowed disabled:opacity-50',
            'resize-y',
            error && 'border-[var(--color-error)] focus-visible:ring-[var(--color-error)]',
            className
          )}
          ref={ref}
          maxLength={maxLength}
          value={value}
          {...props}
        />
        <div className="mt-1 flex items-center justify-between">
          {error ? (
            <p className="text-xs text-[var(--color-error)] font-onest">
              {error}
            </p>
          ) : (
            <span />
          )}
          {showCount && maxLength && (
            <p className="text-xs text-[var(--color-primary-dark)]/60 font-onest">
              {currentLength} / {maxLength}
            </p>
          )}
        </div>
      </div>
    )
  }
)
Textarea.displayName = 'Textarea'

export { Textarea }
```

---

### components/ui/toast.tsx

```tsx
/**
 * Toast Configuration for Sonner
 * 
 * Este arquivo configura o Sonner com as cores do projeto organiQ.
 * O Toaster já está incluído no layout principal (app/layout.tsx).
 * 
 * Para usar em qualquer componente:
 * 
 * import { toast } from 'sonner'
 * 
 * toast.success('Matéria publicada!')
 * toast.error('Erro ao salvar')
 * toast.warning('Atenção: limite atingido')
 * toast.info('Nova atualização disponível')
 * 
 * Com loading:
 * const toastId = toast.loading('Salvando...')
 * // ... operação async
 * toast.success('Salvo!', { id: toastId })
 * 
 * Com ação:
 * toast.success('Matéria criada!', {
 *   action: {
 *     label: 'Ver',
 *     onClick: () => router.push('/app/materias')
 *   }
 * })
 */

import { toast as sonnerToast } from 'sonner'

// Configurações padrão do toast
export const toastConfig = {
  position: 'top-right' as const,
  duration: 5000,
  richColors: true,
  closeButton: true,
  
  // Estilos customizados
  style: {
    fontFamily: 'var(--font-onest)',
  },
  
  // Classes para cada tipo
  classNames: {
    toast: 'font-onest',
    title: 'font-semibold',
    description: 'text-sm opacity-90',
    actionButton: 'bg-[var(--color-primary-purple)] text-white hover:opacity-90',
    cancelButton: 'bg-[var(--color-primary-dark)]/10 hover:bg-[var(--color-primary-dark)]/20',
    closeButton: 'bg-white hover:bg-[var(--color-primary-dark)]/5',
  }
}

// Re-exportar o toast do sonner com tipagem
export const toast = {
  success: (message: string, data?: Parameters<typeof sonnerToast.success>[1]) =>
    sonnerToast.success(message, data),
  
  error: (message: string, data?: Parameters<typeof sonnerToast.error>[1]) =>
    sonnerToast.error(message, data),
  
  warning: (message: string, data?: Parameters<typeof sonnerToast.warning>[1]) =>
    sonnerToast.warning(message, data),
  
  info: (message: string, data?: Parameters<typeof sonnerToast.info>[1]) =>
    sonnerToast.info(message, data),
  
  loading: (message: string, data?: Parameters<typeof sonnerToast.loading>[1]) =>
    sonnerToast.loading(message, data),
  
  promise: sonnerToast.promise,
  dismiss: sonnerToast.dismiss,
  custom: sonnerToast.custom,
}

// Helper para toast com promise
export const toastPromise = <T,>(
  promise: Promise<T>,
  messages: {
    loading: string
    success: string | ((data: T) => string)
    error: string | ((error: unknown) => string)
  }
) => {
  return sonnerToast.promise(promise, messages)
}

// Exemplos de uso:
export const toastExamples = {
  // Básico
  basic: () => {
    toast.success('Operação realizada com sucesso!')
  },
  
  // Com descrição
  withDescription: () => {
    toast.success('Matéria publicada!', {
      description: 'A matéria está disponível no seu blog WordPress'
    })
  },
  
  // Com ação
  withAction: () => {
    toast.success('Matéria criada!', {
      action: {
        label: 'Visualizar',
        onClick: () => console.log('Navegando...')
      }
    })
  },
  
  // Loading com atualização
  loadingUpdate: async () => {
    const toastId = toast.loading('Salvando alterações...')
    
    // Simular operação
    await new Promise(resolve => setTimeout(resolve, 2000))
    
    toast.success('Alterações salvas!', { id: toastId })
  },
  
  // Promise
  withPromise: async () => {
    const promise = new Promise((resolve, reject) => {
      setTimeout(() => Math.random() > 0.5 ? resolve('OK') : reject('Erro'), 2000)
    })
    
    toastPromise(promise, {
      loading: 'Processando...',
      success: 'Sucesso!',
      error: 'Falha ao processar'
    })
  }
}

export default toast
```

---

## 📁 components\wizards

### components/wizards/NewArticlesWizard.tsx

```tsx
"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useWizard } from "@/hooks/useWizard";
import { useUser } from "@/store/authStore";
import { StepIndicator } from "./StepIndicator";
import { CompetitorsForm } from "@/components/forms/CompetitorsForm";
import { ArticleIdeaCard } from "@/components/articles/ArticleIdeaCard";
import { LoadingOverlay } from "@/components/shared/LoadingSpinner";
import { EmptyIdeas } from "@/components/shared/EmptyState";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Slider } from "@/components/ui/slider";
import { Label } from "@/components/ui/label";
import { AlertCircle, MessageSquare } from "lucide-react";
import Link from "next/link";
import type { CompetitorsInput, PublishPayload } from "@/lib/validations";

const steps = [
  { number: 1, label: "Quantidade" },
  { number: 2, label: "Concorrentes" },
  { number: 3, label: "Aprovação" },
];

const loadingMessagesGenerate = [
  "Analisando seus concorrentes...",
  "Mapeando tópicos de autoridade...",
  "Gerando ideias de matérias...",
  "Isso pode levar alguns minutos",
];

const loadingMessagesPublish = [
  "Escrevendo matérias...",
  "Otimizando SEO...",
  "Publicando no WordPress...",
  "Aguarde, estamos finalizando",
];

/**
 * Wizard Simplificado para Gerar Novas Matérias
 *
 * Usado após o onboarding completo para criar matérias adicionais
 */
export function NewArticlesWizard() {
  const router = useRouter();
  const user = useUser();

  const {
    currentStep,
    businessData,
    competitorData,
    articleIdeas,
    submitBusinessInfo,
    submitCompetitors,
    publishArticles,
    updateArticleIdea,
    previousStep,
    isSubmittingBusiness,
    isSubmittingCompetitors,
    isGeneratingIdeas,
    isPublishing,
    approvedCount,
    canPublish,
  } = useWizard(false); // false = não é onboarding

  const [articleCount, setArticleCount] = useState(1);
  const articlesRemaining = user ? user.maxArticles - user.articlesUsed : 0;
  const canCreate = articlesRemaining > 0;

  // ============================================
  // LOADING STATES
  // ============================================

  // Loading: Gerando ideias
  if (currentStep === 999 || isGeneratingIdeas) {
    return <LoadingOverlay messages={loadingMessagesGenerate} />;
  }

  // Loading: Publicando
  if (currentStep === 1000 || isPublishing) {
    return <LoadingOverlay messages={loadingMessagesPublish} />;
  }

  // ============================================
  // STEP 1: QUANTIDADE
  // ============================================

  const renderStepQuantity = () => (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        submitBusinessInfo({
          description: "", // Dados já existem do onboarding
          primaryObjective: "leads",
          hasBlog: false,
          blogUrls: [],
          articleCount,
        } as any);
      }}
      className="space-y-6"
    >
      {/* Alerta de Limite */}
      {!canCreate && (
        <div className="bg-[var(--color-warning)]/10 border border-[var(--color-warning)] rounded-[var(--radius-md)] p-4 flex items-start gap-3">
          <AlertCircle className="h-5 w-5 text-[var(--color-warning)] mt-0.5" />
          <div className="flex-1">
            <p className="font-medium font-onest text-[var(--color-primary-dark)]">
              Limite de matérias atingido
            </p>
            <p className="text-sm font-onest text-[var(--color-primary-dark)]/70 mt-1">
              Você já utilizou todas as {user?.maxArticles} matérias do seu
              plano este mês.
            </p>
            <Link href="/app/conta">
              <Button variant="outline" size="sm" className="mt-3">
                Fazer Upgrade
              </Button>
            </Link>
          </div>
        </div>
      )}

      {canCreate && (
        <>
          {/* Slider de Quantidade */}
          <div className="space-y-2">
            <Label required>Quantas matérias deseja criar?</Label>
            <Slider
              min={1}
              max={articlesRemaining}
              step={1}
              value={[articleCount]}
              onValueChange={(value) => setArticleCount(value[0])}
              showValue
              formatValue={(value) =>
                `${value} ${value === 1 ? "matéria" : "matérias"}`
              }
            />
            <p className="text-xs text-[var(--color-primary-dark)]/60 font-onest">
              Você tem {articlesRemaining}{" "}
              {articlesRemaining === 1 ? "matéria" : "matérias"} disponível
              {articlesRemaining === 1 ? "" : "is"} este mês
            </p>
          </div>

          {/* Info Box */}
          <div className="bg-[var(--color-primary-purple)]/5 border border-[var(--color-primary-purple)]/20 rounded-[var(--radius-md)] p-4">
            <p className="text-sm font-onest text-[var(--color-primary-dark)]/80">
              💡 <strong>Dica:</strong> Você pode gerar várias matérias de uma
              vez para economizar tempo.
            </p>
          </div>

          {/* Submit Button */}
          <div className="flex justify-end pt-4">
            <Button
              type="submit"
              variant="secondary"
              size="lg"
              isLoading={isSubmittingBusiness}
              disabled={isSubmittingBusiness}
            >
              Próximo
            </Button>
          </div>
        </>
      )}
    </form>
  );

  // ============================================
  // STEP 2: CONCORRENTES
  // ============================================

  const renderStepCompetitors = () => (
    <CompetitorsForm
      onSubmit={(data: CompetitorsInput) => submitCompetitors(data as any)}
      onBack={previousStep}
      isLoading={isSubmittingCompetitors}
      defaultValues={competitorData || undefined}
    />
  );

  // ============================================
  // STEP 3: APROVAÇÃO
  // ============================================

  const renderStepApproval = () => {
    const feedbackCount = articleIdeas.filter(
      (idea) => idea.feedback && idea.feedback.length > 0
    ).length;

    const handlePublish = () => {
      const approvedArticles = articleIdeas
        .filter((idea) => idea.approved)
        .map((idea) => ({
          id: idea.id,
          feedback: idea.feedback,
        }));

      publishArticles({
        articles: approvedArticles,
      } as PublishPayload);
    };

    return (
      <div className="space-y-6">
        {/* Header Info */}
        <div className="bg-[var(--color-primary-purple)]/5 border border-[var(--color-primary-purple)]/20 rounded-[var(--radius-md)] p-4">
          <p className="text-sm font-onest text-[var(--color-primary-dark)]/80">
            Revise as ideias geradas e aprove as que deseja publicar. Você pode
            adicionar feedbacks para direcionar o conteúdo.
          </p>
        </div>

        {/* Empty State */}
        {articleIdeas.length === 0 && (
          <EmptyIdeas onRegenerate={() => router.push("/app/novo")} />
        )}

        {/* Grid de Ideias */}
        {articleIdeas.length > 0 && (
          <div className="grid md:grid-cols-2 gap-4">
            {articleIdeas.map((idea) => (
              <ArticleIdeaCard
                key={idea.id}
                idea={idea}
                onUpdate={updateArticleIdea}
              />
            ))}
          </div>
        )}

        {/* Footer com Contador e Ação */}
        {articleIdeas.length > 0 && (
          <div className="border-t border-[var(--color-border)] pt-6 space-y-4">
            {/* Contadores */}
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium font-onest text-[var(--color-primary-dark)]">
                    {approvedCount}{" "}
                    {approvedCount === 1
                      ? "matéria aprovada"
                      : "matérias aprovadas"}
                  </span>
                </div>
                {feedbackCount > 0 && (
                  <div className="flex items-center gap-2 text-[var(--color-primary-purple)]">
                    <MessageSquare className="h-4 w-4" />
                    <span className="text-sm font-medium font-onest">
                      {feedbackCount}{" "}
                      {feedbackCount === 1
                        ? "feedback adicionado"
                        : "feedbacks adicionados"}
                    </span>
                  </div>
                )}
              </div>
            </div>

            {/* Botões de Ação */}
            <div className="flex items-center justify-between gap-4">
              <Button variant="outline" onClick={previousStep}>
                Voltar
              </Button>

              <Button
                variant="primary"
                size="lg"
                onClick={handlePublish}
                disabled={!canPublish}
                title={
                  !canPublish ? "Aprove pelo menos uma matéria" : undefined
                }
              >
                Publicar{" "}
                {approvedCount > 0
                  ? `${approvedCount} ${
                      approvedCount === 1 ? "Matéria" : "Matérias"
                    }`
                  : "Matérias"}
              </Button>
            </div>
          </div>
        )}
      </div>
    );
  };

  // ============================================
  // RENDER
  // ============================================

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="text-center space-y-2">
        <h1 className="text-3xl font-bold font-all-round text-[var(--color-primary-dark)]">
          Gerar Novas Matérias
        </h1>
        <p className="text-lg font-onest text-[var(--color-primary-dark)]/70">
          Crie mais conteúdo otimizado para seu blog
        </p>
      </div>

      {canCreate && (
        <>
          {/* Step Indicator */}
          <StepIndicator currentStep={currentStep} steps={steps} />

          {/* Form Card */}
          <Card>
            <CardHeader>
              <CardTitle>
                {currentStep === 1 && "Quantidade de Matérias"}
                {currentStep === 2 && "Análise de Concorrentes"}
                {currentStep === 3 && "Aprovação de Ideias"}
              </CardTitle>
              <CardDescription>
                {currentStep === 1 &&
                  `Você tem ${articlesRemaining} matérias disponíveis este mês`}
                {currentStep === 2 &&
                  "Adicione URLs de concorrentes para melhorar a estratégia (opcional)"}
                {currentStep === 3 &&
                  "Revise e aprove as matérias que deseja publicar"}
              </CardDescription>
            </CardHeader>

            <CardContent>
              {currentStep === 1 && renderStepQuantity()}
              {currentStep === 2 && renderStepCompetitors()}
              {currentStep === 3 && renderStepApproval()}
            </CardContent>
          </Card>

          {/* Progress Info */}
          <div className="text-center">
            <p className="text-sm font-onest text-[var(--color-primary-dark)]/60">
              Passo {currentStep} de {steps.length}
            </p>
          </div>
        </>
      )}
    </div>
  );
}
```

---

### components/wizards/OnboardingWizard.tsx

```tsx
'use client'

import { useWizard } from '@/hooks/useWizard'
import { StepIndicator } from './StepIndicator'
import { BusinessInfoForm } from '@/components/forms/BusinessInfoForm'
import { CompetitorsForm } from '@/components/forms/CompetitorsForm'
import { IntegrationsForm } from '@/components/forms/IntegrationsForm'
import { LoadingOverlay } from '@/components/shared/LoadingSpinner'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import type { BusinessInput, CompetitorsInput, IntegrationsInput } from '@/lib/validations'

const steps = [
  { number: 1, label: 'Negócio' },
  { number: 2, label: 'Concorrentes' },
  { number: 3, label: 'Integrações' },
  { number: 4, label: 'Aprovação' },
]

const loadingMessages = [
  'Analisando seus concorrentes...',
  'Mapeando tópicos de autoridade...',
  'Gerando ideias de matérias...',
  'Isso pode levar alguns minutos',
]

export function OnboardingWizard() {
  const {
    currentStep,
    businessData,
    competitorData,
    integrationsData,
    submitBusinessInfo,
    submitCompetitors,
    submitIntegrations,
    previousStep,
    isSubmittingBusiness,
    isSubmittingCompetitors,
    isSubmittingIntegrations,
    isGeneratingIdeas,
  } = useWizard(true) // true = isOnboarding

  // Loading state para geração de ideias
  if (currentStep === 999 || isGeneratingIdeas) {
    return <LoadingOverlay messages={loadingMessages} />
  }

  // TODO: Implement steps 4 (Approval) and 1000 (Publishing)
  // These will be added in the next phase

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="text-center space-y-2">
        <h1 className="text-3xl font-bold font-all-round text-[var(--color-primary-dark)]">
          Configuração Inicial
        </h1>
        <p className="text-lg font-onest text-[var(--color-primary-dark)]/70">
          Vamos configurar sua conta para começar a gerar conteúdo
        </p>
      </div>

      {/* Step Indicator */}
      <StepIndicator currentStep={currentStep} steps={steps} />

      {/* Form Card */}
      <Card>
        <CardHeader>
          <CardTitle>
            {currentStep === 1 && 'Informações do Negócio'}
            {currentStep === 2 && 'Análise de Concorrentes'}
            {currentStep === 3 && 'Configurar Integrações'}
          </CardTitle>
        </CardHeader>

        <CardContent>
          {/* Step 1: Business Info */}
          {currentStep === 1 && (
            <BusinessInfoForm
              onSubmit={(data: BusinessInput) => submitBusinessInfo(data as any)}
              isLoading={isSubmittingBusiness}
              defaultValues={businessData || undefined}
            />
          )}

          {/* Step 2: Competitors */}
          {currentStep === 2 && (
            <CompetitorsForm
              onSubmit={(data: CompetitorsInput) => submitCompetitors(data as any)}
              onBack={previousStep}
              isLoading={isSubmittingCompetitors}
              defaultValues={competitorData || undefined}
            />
          )}

          {/* Step 3: Integrations */}
          {currentStep === 3 && (
            <IntegrationsForm
              onSubmit={(data: IntegrationsInput) => submitIntegrations(data as any)}
              onBack={previousStep}
              isLoading={isSubmittingIntegrations}
              defaultValues={integrationsData || undefined}
            />
          )}
        </CardContent>
      </Card>

      {/* Progress Info */}
      <div className="text-center">
        <p className="text-sm font-onest text-[var(--color-primary-dark)]/60">
          Passo {currentStep} de {steps.length}
        </p>
      </div>
    </div>
  )
}
```

---

### components/wizards/StepIndicator.tsx

```tsx
'use client'

import { Check } from 'lucide-react'
import { cn } from '@/lib/utils'

interface Step {
  number: number
  label: string
}

interface StepIndicatorProps {
  currentStep: number
  steps: Step[]
}

export function StepIndicator({ currentStep, steps }: StepIndicatorProps) {
  return (
    <div className="w-full mb-8">
      {/* Desktop: Horizontal */}
      <div className="hidden md:flex items-center justify-between">
        {steps.map((step, index) => {
          const isCompleted = currentStep > step.number
          const isCurrent = currentStep === step.number
          const isLast = index === steps.length - 1

          return (
            <div key={step.number} className="flex items-center flex-1">
              {/* Step Circle */}
              <div className="flex flex-col items-center">
                <div
                  className={cn(
                    'flex items-center justify-center h-10 w-10 rounded-full border-2 transition-all duration-200',
                    isCompleted && 'bg-[var(--color-success)] border-[var(--color-success)]',
                    isCurrent && 'bg-[var(--color-primary-purple)] border-[var(--color-primary-purple)]',
                    !isCompleted && !isCurrent && 'bg-white border-[var(--color-border)]'
                  )}
                >
                  {isCompleted ? (
                    <Check className="h-5 w-5 text-white" />
                  ) : (
                    <span
                      className={cn(
                        'text-sm font-bold font-all-round',
                        isCurrent && 'text-white',
                        !isCurrent && 'text-[var(--color-primary-dark)]/60'
                      )}
                    >
                      {step.number}
                    </span>
                  )}
                </div>
                <span
                  className={cn(
                    'mt-2 text-xs font-medium font-onest text-center',
                    isCurrent && 'text-[var(--color-primary-purple)]',
                    isCompleted && 'text-[var(--color-success)]',
                    !isCurrent && !isCompleted && 'text-[var(--color-primary-dark)]/60'
                  )}
                >
                  {step.label}
                </span>
              </div>

              {/* Connector Line */}
              {!isLast && (
                <div
                  className={cn(
                    'flex-1 h-0.5 mx-4 transition-all duration-200',
                    isCompleted ? 'bg-[var(--color-success)]' : 'bg-[var(--color-border)]'
                  )}
                />
              )}
            </div>
          )
        })}
      </div>

      {/* Mobile: Vertical Compact */}
      <div className="md:hidden">
        <div className="flex items-center gap-4">
          {steps.map((step, index) => {
            const isCompleted = currentStep > step.number
            const isCurrent = currentStep === step.number

            return (
              <div key={step.number} className="flex items-center">
                <div
                  className={cn(
                    'flex items-center justify-center h-8 w-8 rounded-full border-2 transition-all duration-200',
                    isCompleted && 'bg-[var(--color-success)] border-[var(--color-success)]',
                    isCurrent && 'bg-[var(--color-primary-purple)] border-[var(--color-primary-purple)]',
                    !isCompleted && !isCurrent && 'bg-white border-[var(--color-border)]'
                  )}
                >
                  {isCompleted ? (
                    <Check className="h-4 w-4 text-white" />
                  ) : (
                    <span
                      className={cn(
                        'text-xs font-bold font-all-round',
                        isCurrent && 'text-white',
                        !isCurrent && 'text-[var(--color-primary-dark)]/60'
                      )}
                    >
                      {step.number}
                    </span>
                  )}
                </div>
                {index < steps.length - 1 && (
                  <div
                    className={cn(
                      'w-8 h-0.5 mx-1',
                      isCompleted ? 'bg-[var(--color-success)]' : 'bg-[var(--color-border)]'
                    )}
                  />
                )}
              </div>
            )
          })}
        </div>
        {/* Current Step Label */}
        <p className="mt-3 text-sm font-medium font-onest text-[var(--color-primary-purple)]">
          {steps.find((s) => s.number === currentStep)?.label}
        </p>
      </div>
    </div>
  )
}
```

---

## 📁 hooks

### hooks/useArticles.ts

```ts
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
```

---

### hooks/useAuth.ts

```ts
import { useRouter } from 'next/navigation'
import { useMutation } from '@tanstack/react-query'
import { toast } from 'sonner'
import { useAuthStore } from '@/store/authStore'
import api, { getErrorMessage } from '@/lib/axios'
import type { LoginCredentials, RegisterData, AuthResponse } from '@/types'

// ============================================
// API FUNCTIONS
// ============================================

const authApi = {
  login: async (credentials: LoginCredentials): Promise<AuthResponse> => {
    const { data } = await api.post<AuthResponse>('/auth/login', credentials)
    return data
  },

  register: async (userData: RegisterData): Promise<AuthResponse> => {
    const { data } = await api.post<AuthResponse>('/auth/register', userData)
    return data
  },

  logout: async (): Promise<void> => {
    await api.post('/auth/logout')
  },

  refreshToken: async (): Promise<void> => {
    await api.post('/auth/refresh')
  },

  getCurrentUser: async (): Promise<AuthResponse> => {
    const { data } = await api.get<AuthResponse>('/auth/me')
    return data
  }
}

// ============================================
// HOOK
// ============================================

export function useAuth() {
  const router = useRouter()
  const { user, setUser, clearUser, isAuthenticated, isLoading } = useAuthStore()

  // ============================================
  // LOGIN MUTATION
  // ============================================

  const loginMutation = useMutation({
    mutationFn: authApi.login,
    onSuccess: (data) => {
      setUser(data.user)
      toast.success('Login realizado com sucesso!')

      // Redirecionar baseado no status do onboarding
      if (!data.user.hasCompletedOnboarding) {
        router.push('/app/planos')
      } else {
        router.push('/app/materias')
      }
    },
    onError: (error) => {
      const message = getErrorMessage(error)
      toast.error(message || 'Erro ao fazer login')
    }
  })

  // ============================================
  // REGISTER MUTATION
  // ============================================

  const registerMutation = useMutation({
    mutationFn: authApi.register,
    onSuccess: (data) => {
      setUser(data.user)
      toast.success('Conta criada com sucesso!')
      
      // Sempre redireciona para planos no primeiro acesso
      router.push('/app/planos')
    },
    onError: (error) => {
      const message = getErrorMessage(error)
      toast.error(message || 'Erro ao criar conta')
    }
  })

  // ============================================
  // LOGOUT MUTATION
  // ============================================

  const logoutMutation = useMutation({
    mutationFn: authApi.logout,
    onSuccess: () => {
      clearUser()
      toast.success('Logout realizado com sucesso!')
      router.push('/login')
    },
    onError: (error) => {
      // Limpa localmente mesmo se API falhar
      clearUser()
      router.push('/login')
      
      const message = getErrorMessage(error)
      toast.error(message || 'Erro ao fazer logout')
    }
  })

  // ============================================
  // HELPERS
  // ============================================

  const login = (credentials: LoginCredentials) => {
    loginMutation.mutate(credentials)
  }

  const register = (userData: RegisterData) => {
    registerMutation.mutate(userData)
  }

  const logout = () => {
    logoutMutation.mutate()
  }

  // ============================================
  // RETURN
  // ============================================

  return {
    // State
    user,
    isAuthenticated,
    isLoading,
    
    // Actions
    login,
    register,
    logout,
    
    // Mutation states
    isLoggingIn: loginMutation.isPending,
    isRegistering: registerMutation.isPending,
    isLoggingOut: logoutMutation.isPending,
    
    // Helpers
    hasCompletedOnboarding: user?.hasCompletedOnboarding ?? false,
    canCreateArticles: user ? user.articlesUsed < user.maxArticles : false,
    articlesRemaining: user ? user.maxArticles - user.articlesUsed : 0
  }
}
```

---

### hooks/usePlans.ts

```ts
import { useQuery, useMutation } from '@tanstack/react-query'
import { useRouter } from 'next/navigation'
import { toast } from 'sonner'
import api, { getErrorMessage } from '@/lib/axios'
import { useAuthStore } from '@/store/authStore'
import type { 
  Plan, 
  PlanInfo, 
  CheckoutResponse, 
  PaymentStatus 
} from '@/types'

// ============================================
// API FUNCTIONS
// ============================================

const plansApi = {
  getPlans: async (): Promise<Plan[]> => {
    const { data } = await api.get<Plan[]>('/plans')
    return data
  },

  getCurrentPlan: async (): Promise<PlanInfo> => {
    const { data } = await api.get<PlanInfo>('/account/plan')
    return data
  },

  createCheckout: async (planId: string): Promise<CheckoutResponse> => {
    const { data } = await api.post<CheckoutResponse>('/payments/create-checkout', { planId })
    return data
  },

  getPaymentStatus: async (sessionId: string): Promise<PaymentStatus> => {
    const { data } = await api.get<PaymentStatus>(`/payments/status/${sessionId}`)
    return data
  },

  createPortalSession: async (): Promise<{ url: string }> => {
    const { data } = await api.post<{ url: string }>('/payments/create-portal-session')
    return data
  }
}

// ============================================
// QUERY KEYS
// ============================================

const planKeys = {
  all: ['plans'] as const,
  lists: () => [...planKeys.all, 'list'] as const,
  current: () => [...planKeys.all, 'current'] as const,
  payment: (sessionId: string) => [...planKeys.all, 'payment', sessionId] as const
}

// ============================================
// HOOK
// ============================================

export function usePlans() {
  const router = useRouter()
  const { updateUser } = useAuthStore()

  // ============================================
  // GET PLANS QUERY
  // ============================================

  const plansQuery = useQuery({
    queryKey: planKeys.lists(),
    queryFn: plansApi.getPlans,
    staleTime: Infinity // Planos raramente mudam
  })

  // ============================================
  // GET CURRENT PLAN QUERY
  // ============================================

  const currentPlanQuery = useQuery({
    queryKey: planKeys.current(),
    queryFn: plansApi.getCurrentPlan,
    staleTime: 60000 // 1 minuto
  })

  // ============================================
  // CREATE CHECKOUT MUTATION
  // ============================================

  const checkoutMutation = useMutation({
    mutationFn: plansApi.createCheckout,
    onSuccess: (data) => {
      // Redirecionar para checkout
      window.location.href = data.checkoutUrl
    },
    onError: (error) => {
      const message = getErrorMessage(error)
      toast.error(message || 'Erro ao criar checkout')
    }
  })

  // ============================================
  // CREATE PORTAL MUTATION
  // ============================================

  const portalMutation = useMutation({
    mutationFn: plansApi.createPortalSession,
    onSuccess: (data) => {
      // Redirecionar para portal
      window.location.href = data.url
    },
    onError: (error) => {
      const message = getErrorMessage(error)
      toast.error(message || 'Erro ao abrir portal de pagamentos')
    }
  })

  // ============================================
  // PAYMENT STATUS POLLING
  // ============================================

  const usePaymentStatus = (sessionId: string | null, enabled: boolean = true) => {
    return useQuery({
      queryKey: planKeys.payment(sessionId || ''),
      queryFn: () => plansApi.getPaymentStatus(sessionId!),
      enabled: enabled && !!sessionId,
      refetchInterval: (data) => {
        // Parar polling se status for 'paid' ou 'failed'
        if (data?.status === 'paid' || data?.status === 'failed') {
          return false
        }
        return 3000 // Poll a cada 3 segundos
      },
      refetchOnWindowFocus: false
    })
  }

  // ============================================
  // HELPERS
  // ============================================

  const selectPlan = (planId: string) => {
    checkoutMutation.mutate(planId)
  }

  const openPortal = () => {
    portalMutation.mutate()
  }

  const getPlanById = (planId: string) => {
    return plansQuery.data?.find((plan) => plan.id === planId)
  }

  const getRecommendedPlan = () => {
    return plansQuery.data?.find((plan) => plan.recommended)
  }

  const isCurrentPlan = (planId: string) => {
    return currentPlanQuery.data?.name === getPlanById(planId)?.name
  }

  const canUpgrade = (targetPlanId: string) => {
    const currentPlan = plansQuery.data?.find(
      (plan) => plan.name === currentPlanQuery.data?.name
    )
    const targetPlan = getPlanById(targetPlanId)
    
    if (!currentPlan || !targetPlan) return false
    
    return targetPlan.maxArticles > currentPlan.maxArticles
  }

  // ============================================
  // RETURN
  // ============================================

  return {
    // Data
    plans: plansQuery.data || [],
    currentPlan: currentPlanQuery.data,
    
    // States
    isLoadingPlans: plansQuery.isLoading,
    isLoadingCurrentPlan: currentPlanQuery.isLoading,
    isError: plansQuery.isError || currentPlanQuery.isError,
    error: plansQuery.error || currentPlanQuery.error,
    
    // Actions
    selectPlan,
    openPortal,
    refetchCurrentPlan: currentPlanQuery.refetch,
    
    // Mutation states
    isCreatingCheckout: checkoutMutation.isPending,
    isOpeningPortal: portalMutation.isPending,
    
    // Helpers
    getPlanById,
    getRecommendedPlan,
    isCurrentPlan,
    canUpgrade,
    
    // Payment status hook
    usePaymentStatus
  }
}
```

---

### hooks/useWizard.ts

```ts
import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import api, { getErrorMessage, uploadFile } from '@/lib/axios'
import { useAuthStore } from '@/store/authStore'
import type { 
  BusinessInfo, 
  CompetitorData, 
  IntegrationsData,
  ArticleIdea,
  PublishPayload
} from '@/types'

// ============================================
// API FUNCTIONS
// ============================================

const wizardApi = {
  // Onboarding completo
  submitBusiness: async (data: BusinessInfo): Promise<{ success: boolean }> => {
    const formData = new FormData()
    formData.append('description', data.description)
    formData.append('primaryObjective', data.primaryObjective)
    if (data.secondaryObjective) formData.append('secondaryObjective', data.secondaryObjective)
    if (data.siteUrl) formData.append('siteUrl', data.siteUrl)
    formData.append('hasBlog', String(data.hasBlog))
    formData.append('blogUrls', JSON.stringify(data.blogUrls))
    formData.append('articleCount', String(data.articleCount))
    if (data.brandFile) formData.append('brandFile', data.brandFile)

    const { data: response } = await api.post('/wizard/business', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    return response
  },

  submitCompetitors: async (data: CompetitorData): Promise<{ success: boolean }> => {
    const { data: response } = await api.post('/wizard/competitors', data)
    return response
  },

  submitIntegrations: async (data: IntegrationsData): Promise<{ success: boolean }> => {
    const { data: response } = await api.post('/wizard/integrations', data)
    return response
  },

  generateIdeas: async (): Promise<{ jobId: string }> => {
    const { data } = await api.post('/wizard/generate-ideas')
    return data
  },

  getIdeasStatus: async (jobId: string): Promise<{ 
    status: 'processing' | 'completed' | 'failed'
    ideas?: ArticleIdea[]
    error?: string 
  }> => {
    const { data } = await api.get(`/wizard/ideas-status/${jobId}`)
    return data
  },

  publishArticles: async (payload: PublishPayload): Promise<{ jobId: string }> => {
    const { data } = await api.post('/wizard/publish', payload)
    return data
  },

  getPublishStatus: async (jobId: string): Promise<{
    status: 'processing' | 'completed' | 'failed'
    published?: number
    total?: number
    error?: string
  }> => {
    const { data } = await api.get(`/wizard/publish-status/${jobId}`)
    return data
  },

  // Wizard simplificado (novo)
  generateNewIdeas: async (data: { articleCount: number; competitorUrls?: string[] }): Promise<{ jobId: string }> => {
    const { data: response } = await api.post('/articles/generate-ideas', data)
    return response
  },

  publishNewArticles: async (payload: PublishPayload): Promise<{ jobId: string }> => {
    const { data } = await api.post('/articles/publish', payload)
    return data
  }
}

// ============================================
// HOOK
// ============================================

export function useWizard(isOnboarding: boolean = true) {
  const router = useRouter()
  const queryClient = useQueryClient()
  const { updateUser } = useAuthStore()

  // Local state para wizard steps
  const [currentStep, setCurrentStep] = useState(1)
  const [businessData, setBusinessData] = useState<BusinessInfo | null>(null)
  const [competitorData, setCompetitorData] = useState<CompetitorData | null>(null)
  const [integrationsData, setIntegrationsData] = useState<IntegrationsData | null>(null)
  const [articleIdeas, setArticleIdeas] = useState<ArticleIdea[]>([])
  const [jobId, setJobId] = useState<string | null>(null)

  // ============================================
  // STEP 1: BUSINESS INFO
  // ============================================

  const businessMutation = useMutation({
    mutationFn: wizardApi.submitBusiness,
    onSuccess: (_, variables) => {
      setBusinessData(variables)
      setCurrentStep(2)
      toast.success('Informações salvas!')
    },
    onError: (error) => {
      const message = getErrorMessage(error)
      toast.error(message || 'Erro ao salvar informações')
    }
  })

  // ============================================
  // STEP 2: COMPETITORS
  // ============================================

  const competitorsMutation = useMutation({
    mutationFn: wizardApi.submitCompetitors,
    onSuccess: (_, variables) => {
      setCompetitorData(variables)
      setCurrentStep(isOnboarding ? 3 : 999) // Se não é onboarding, pula para loading
      if (!isOnboarding) {
        generateIdeasMutation.mutate()
      } else {
        toast.success('Concorrentes salvos!')
      }
    },
    onError: (error) => {
      const message = getErrorMessage(error)
      toast.error(message || 'Erro ao salvar concorrentes')
    }
  })

  // ============================================
  // STEP 3: INTEGRATIONS (só no onboarding)
  // ============================================

  const integrationsMutation = useMutation({
    mutationFn: wizardApi.submitIntegrations,
    onSuccess: (_, variables) => {
      setIntegrationsData(variables)
      setCurrentStep(999) // Vai para loading
      generateIdeasMutation.mutate()
    },
    onError: (error) => {
      const message = getErrorMessage(error)
      toast.error(message || 'Erro ao salvar integrações')
    }
  })

  // ============================================
  // LOADING: GENERATE IDEAS
  // ============================================

  const generateIdeasMutation = useMutation({
    mutationFn: isOnboarding 
      ? wizardApi.generateIdeas 
      : () => wizardApi.generateNewIdeas({
          articleCount: businessData?.articleCount || 1,
          competitorUrls: competitorData?.competitorUrls
        }),
    onSuccess: (data) => {
      setJobId(data.jobId)
      // O polling vai começar automaticamente via useQuery
    },
    onError: (error) => {
      const message = getErrorMessage(error)
      toast.error(message || 'Erro ao gerar ideias')
      setCurrentStep(isOnboarding ? 3 : 2) // Volta para step anterior
    }
  })

  // ============================================
  // POLLING: IDEAS STATUS
  // ============================================

  const ideasStatusQuery = useQuery({
    queryKey: ['ideas-status', jobId],
    queryFn: () => wizardApi.getIdeasStatus(jobId!),
    enabled: !!jobId && currentStep === 999,
    refetchInterval: (data) => {
      if (data?.status === 'completed') {
        setArticleIdeas(data.ideas || [])
        setCurrentStep(isOnboarding ? 4 : 3) // Vai para aprovação
        return false
      }
      if (data?.status === 'failed') {
        toast.error(data.error || 'Erro ao gerar ideias')
        setCurrentStep(isOnboarding ? 3 : 2)
        return false
      }
      return 3000 // Poll a cada 3 segundos
    },
    refetchOnWindowFocus: false
  })

  // ============================================
  // STEP 4: PUBLISH
  // ============================================

  const publishMutation = useMutation({
    mutationFn: isOnboarding ? wizardApi.publishArticles : wizardApi.publishNewArticles,
    onSuccess: (data) => {
      setJobId(data.jobId)
      setCurrentStep(1000) // Loading de publicação
    },
    onError: (error) => {
      const message = getErrorMessage(error)
      toast.error(message || 'Erro ao publicar matérias')
    }
  })

  // ============================================
  // POLLING: PUBLISH STATUS
  // ============================================

  const publishStatusQuery = useQuery({
    queryKey: ['publish-status', jobId],
    queryFn: () => wizardApi.getPublishStatus(jobId!),
    enabled: !!jobId && currentStep === 1000,
    refetchInterval: (data) => {
      if (data?.status === 'completed') {
        // Atualizar usuário
        if (isOnboarding) {
          updateUser({ hasCompletedOnboarding: true })
        }
        
        // Invalidar cache de artigos
        queryClient.invalidateQueries({ queryKey: ['articles'] })
        
        toast.success(`${data.published} matérias publicadas com sucesso!`)
        router.push('/app/materias')
        return false
      }
      if (data?.status === 'failed') {
        toast.error(data.error || 'Erro ao publicar matérias')
        setCurrentStep(isOnboarding ? 4 : 3)
        return false
      }
      return 3000
    },
    refetchOnWindowFocus: false
  })

  // ============================================
  // NAVIGATION HELPERS
  // ============================================

  const goToStep = (step: number) => {
    setCurrentStep(step)
  }

  const nextStep = () => {
    setCurrentStep((prev) => prev + 1)
  }

  const previousStep = () => {
    setCurrentStep((prev) => Math.max(1, prev - 1))
  }

  const submitBusinessInfo = (data: BusinessInfo) => {
    businessMutation.mutate(data)
  }

  const submitCompetitors = (data: CompetitorData) => {
    competitorsMutation.mutate(data)
  }

  const submitIntegrations = (data: IntegrationsData) => {
    integrationsMutation.mutate(data)
  }

  const publishArticles = (payload: PublishPayload) => {
    publishMutation.mutate(payload)
  }

  const updateArticleIdea = (id: string, updates: Partial<ArticleIdea>) => {
    setArticleIdeas((prev) =>
      prev.map((idea) => (idea.id === id ? { ...idea, ...updates } : idea))
    )
  }

  // ============================================
  // RETURN
  // ============================================

  return {
    // Current state
    currentStep,
    businessData,
    competitorData,
    integrationsData,
    articleIdeas,
    
    // Navigation
    goToStep,
    nextStep,
    previousStep,
    
    // Actions
    submitBusinessInfo,
    submitCompetitors,
    submitIntegrations,
    publishArticles,
    updateArticleIdea,
    
    // Loading states
    isSubmittingBusiness: businessMutation.isPending,
    isSubmittingCompetitors: competitorsMutation.isPending,
    isSubmittingIntegrations: integrationsMutation.isPending,
    isGeneratingIdeas: generateIdeasMutation.isPending || ideasStatusQuery.isFetching,
    isPublishing: publishMutation.isPending || publishStatusQuery.isFetching,
    
    // Progress info
    ideasProgress: ideasStatusQuery.data?.status,
    publishProgress: publishStatusQuery.data,
    
    // Computed
    approvedCount: articleIdeas.filter((idea) => idea.approved).length,
    canPublish: articleIdeas.some((idea) => idea.approved),
    isLoading: currentStep === 999 || currentStep === 1000
  }
}
```

---

## 📁 lib

### lib/axios.ts

```ts
import axios, { AxiosError, AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import { toast } from 'sonner'

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
  async (error: AxiosError) => {
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
    return (
      error.response?.data?.message ||
      error.response?.data?.error ||
      error.message ||
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
  if (axios.isAxiosError(error) && error.response?.status === 422) {
    return error.response.data?.errors || {}
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
```

---

### lib/constantes.ts

```ts
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
```

---

### lib/utils.ts

```ts
import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'
import { format, formatDistanceToNow, parseISO, isValid } from 'date-fns'
import { ptBR } from 'date-fns/locale'

// ============================================
// TAILWIND UTILITIES
// ============================================

/**
 * Combina classes Tailwind evitando conflitos
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

// ============================================
// DATE UTILITIES
// ============================================

/**
 * Formata data no padrão brasileiro (dd/MM/yyyy)
 */
export function formatDate(date: string | Date): string {
  try {
    const dateObj = typeof date === 'string' ? parseISO(date) : date
    if (!isValid(dateObj)) return 'Data inválida'
    return format(dateObj, 'dd/MM/yyyy', { locale: ptBR })
  } catch {
    return 'Data inválida'
  }
}

/**
 * Formata data e hora no padrão brasileiro (dd/MM/yyyy HH:mm)
 */
export function formatDateTime(date: string | Date): string {
  try {
    const dateObj = typeof date === 'string' ? parseISO(date) : date
    if (!isValid(dateObj)) return 'Data inválida'
    return format(dateObj, 'dd/MM/yyyy HH:mm', { locale: ptBR })
  } catch {
    return 'Data inválida'
  }
}

/**
 * Formata data de forma relativa (ex: "há 2 dias")
 */
export function formatRelativeDate(date: string | Date): string {
  try {
    const dateObj = typeof date === 'string' ? parseISO(date) : date
    if (!isValid(dateObj)) return 'Data inválida'
    return formatDistanceToNow(dateObj, {
      addSuffix: true,
      locale: ptBR
    })
  } catch {
    return 'Data inválida'
  }
}

// ============================================
// CURRENCY UTILITIES
// ============================================

/**
 * Formata valor monetário em BRL
 */
export function formatCurrency(value: number): string {
  return new Intl.NumberFormat('pt-BR', {
    style: 'currency',
    currency: 'BRL'
  }).format(value)
}

/**
 * Formata valor monetário compacto (ex: R$ 1,5K)
 */
export function formatCompactCurrency(value: number): string {
  return new Intl.NumberFormat('pt-BR', {
    style: 'currency',
    currency: 'BRL',
    notation: 'compact',
    maximumFractionDigits: 1
  }).format(value)
}

// ============================================
// NUMBER UTILITIES
// ============================================

/**
 * Formata número com separadores
 */
export function formatNumber(value: number): string {
  return new Intl.NumberFormat('pt-BR').format(value)
}

/**
 * Formata porcentagem
 */
export function formatPercentage(value: number, decimals: number = 0): string {
  return new Intl.NumberFormat('pt-BR', {
    style: 'percent',
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals
  }).format(value / 100)
}

// ============================================
// STRING UTILITIES
// ============================================

/**
 * Trunca texto com ellipsis
 */
export function truncate(text: string, maxLength: number): string {
  if (text.length <= maxLength) return text
  return text.substring(0, maxLength) + '...'
}

/**
 * Capitaliza primeira letra
 */
export function capitalize(text: string): string {
  return text.charAt(0).toUpperCase() + text.slice(1)
}

/**
 * Converte para slug (URL-friendly)
 */
export function slugify(text: string): string {
  return text
    .toLowerCase()
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '') // Remove acentos
    .replace(/[^\w\s-]/g, '') // Remove caracteres especiais
    .replace(/\s+/g, '-') // Substitui espaços por hífens
    .replace(/--+/g, '-') // Remove hífens duplicados
    .trim()
}

/**
 * Extrai iniciais do nome (ex: "João Silva" -> "JS")
 */
export function getInitials(name: string): string {
  return name
    .split(' ')
    .map(word => word.charAt(0))
    .slice(0, 2)
    .join('')
    .toUpperCase()
}

// ============================================
// VALIDATION UTILITIES
// ============================================

/**
 * Valida URL
 */
export function isValidUrl(url: string): boolean {
  try {
    new URL(url)
    return true
  } catch {
    return false
  }
}

/**
 * Valida email
 */
export function isValidEmail(email: string): boolean {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(email)
}

/**
 * Valida ID do Google Analytics (GA4)
 */
export function isValidGA4Id(id: string): boolean {
  return /^G-[A-Z0-9]+$/.test(id)
}

// ============================================
// FILE UTILITIES
// ============================================

/**
 * Formata tamanho de arquivo
 */
export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 Bytes'
  
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

/**
 * Valida tipo de arquivo
 */
export function isValidFileType(file: File, allowedTypes: string[]): boolean {
  return allowedTypes.includes(file.type)
}

/**
 * Valida tamanho de arquivo
 */
export function isValidFileSize(file: File, maxSizeInMB: number): boolean {
  const maxSizeInBytes = maxSizeInMB * 1024 * 1024
  return file.size <= maxSizeInBytes
}

// ============================================
// ARRAY UTILITIES
// ============================================

/**
 * Remove duplicatas de array
 */
export function unique<T>(array: T[]): T[] {
  return Array.from(new Set(array))
}

/**
 * Agrupa array por chave
 */
export function groupBy<T>(array: T[], key: keyof T): Record<string, T[]> {
  return array.reduce((acc, item) => {
    const groupKey = String(item[key])
    if (!acc[groupKey]) {
      acc[groupKey] = []
    }
    acc[groupKey].push(item)
    return acc
  }, {} as Record<string, T[]>)
}

// ============================================
// DEBOUNCE & THROTTLE
// ============================================

/**
 * Debounce function
 */
export function debounce<T extends (...args: unknown[]) => unknown>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: NodeJS.Timeout | null = null
  
  return function(...args: Parameters<T>) {
    if (timeout) clearTimeout(timeout)
    timeout = setTimeout(() => func(...args), wait)
  }
}

/**
 * Throttle function
 */
export function throttle<T extends (...args: unknown[]) => unknown>(
  func: T,
  limit: number
): (...args: Parameters<T>) => void {
  let inThrottle: boolean
  
  return function(...args: Parameters<T>) {
    if (!inThrottle) {
      func(...args)
      inThrottle = true
      setTimeout(() => (inThrottle = false), limit)
    }
  }
}

// ============================================
// CLIPBOARD UTILITIES
// ============================================

/**
 * Copia texto para o clipboard
 */
export async function copyToClipboard(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text)
    return true
  } catch {
    // Fallback para navegadores antigos
    try {
      const textarea = document.createElement('textarea')
      textarea.value = text
      textarea.style.position = 'fixed'
      textarea.style.opacity = '0'
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      document.body.removeChild(textarea)
      return true
    } catch {
      return false
    }
  }
}

// ============================================
// LOCAL STORAGE UTILITIES
// ============================================

/**
 * Salva no localStorage com JSON
 */
export function setLocalStorage<T>(key: string, value: T): void {
  try {
    localStorage.setItem(key, JSON.stringify(value))
  } catch (error) {
    console.error('Error saving to localStorage:', error)
  }
}

/**
 * Obtém do localStorage com parse JSON
 */
export function getLocalStorage<T>(key: string, defaultValue: T): T {
  try {
    const item = localStorage.getItem(key)
    return item ? JSON.parse(item) : defaultValue
  } catch (error) {
    console.error('Error reading from localStorage:', error)
    return defaultValue
  }
}

/**
 * Remove do localStorage
 */
export function removeLocalStorage(key: string): void {
  try {
    localStorage.removeItem(key)
  } catch (error) {
    console.error('Error removing from localStorage:', error)
  }
}

// ============================================
// SLEEP UTILITY
// ============================================

/**
 * Promise-based sleep
 */
export function sleep(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms))
}

// ============================================
// RANDOM UTILITIES
// ============================================

/**
 * Gera ID aleatório
 */
export function generateId(length: number = 8): string {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
  let result = ''
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  return result
}
```

---

### lib/validations.ts

```ts
import { z } from "zod";

// ============================================
// AUTH SCHEMAS
// ============================================

export const loginSchema = z.object({
  email: z.string().min(1, "Email é obrigatório").email("Email inválido"),
  password: z
    .string()
    .min(6, "Senha deve ter no mínimo 6 caracteres")
    .max(100, "Senha muito longa"),
});

export const registerSchema = z.object({
  name: z
    .string()
    .min(2, "Nome deve ter no mínimo 2 caracteres")
    .max(100, "Nome muito longo")
    .regex(/^[a-zA-ZÀ-ÿ\s]+$/, "Nome deve conter apenas letras"),
  email: z.string().min(1, "Email é obrigatório").email("Email inválido"),
  password: z
    .string()
    .min(6, "Senha deve ter no mínimo 6 caracteres")
    .max(100, "Senha muito longa"),
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
        // Validar extensão real (não apenas MIME type)
        const ext = file.name.split(".").pop()?.toLowerCase();
        return ["pdf", "jpg", "jpeg", "png"].includes(ext || "");
      }, "Extensão de arquivo inválida")
      .optional(),
  })
  .refine(
    (data) => {
      // Validar que o objetivo secundário é diferente do primário
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
      // Se tem blog, deve ter pelo menos uma URL
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
      // Se Search Console ativado, deve ter URL
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
      // Se Analytics ativado, deve ter ID
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
```

---

## 📁 public

### public/manifest.json

```json
{
  "name": "organiQ - Aumente seu tráfego orgânico com IA",
  "short_name": "organiQ",
  "description": "Matérias de blog que geram autoridade e SEO. Naturalmente Inteligente.",
  "start_url": "/",
  "display": "standalone",
  "background_color": "#fffde1",
  "theme_color": "#001d47",
  "orientation": "portrait-primary",
  "icons": [
    {
      "src": "/favicon-16x16.png",
      "sizes": "16x16",
      "type": "image/png"
    },
    {
      "src": "/favicon-32x32.png",
      "sizes": "32x32",
      "type": "image/png"
    },
    {
      "src": "/apple-touch-icon.png",
      "sizes": "180x180",
      "type": "image/png",
      "purpose": "any maskable"
    },
    {
      "src": "/android-chrome-192x192.png",
      "sizes": "192x192",
      "type": "image/png"
    },
    {
      "src": "/android-chrome-512x512.png",
      "sizes": "512x512",
      "type": "image/png"
    }
  ],
  "categories": ["business", "productivity", "marketing"],
  "lang": "pt-BR",
  "dir": "ltr",
  "scope": "/",
  "shortcuts": [
    {
      "name": "Gerar Matérias",
      "short_name": "Gerar",
      "description": "Criar novas matérias",
      "url": "/app/novo",
      "icons": [
        {
          "src": "/favicon-96x96.png",
          "sizes": "96x96"
        }
      ]
    },
    {
      "name": "Minhas Matérias",
      "short_name": "Matérias",
      "description": "Ver matérias publicadas",
      "url": "/app/materias",
      "icons": [
        {
          "src": "/favicon-96x96.png",
          "sizes": "96x96"
        }
      ]
    }
  ]
}
```

---

## 📁 store

### store/authStore.ts

```ts
import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'
import type { User } from '@/types'

// ============================================
// TYPES
// ============================================

interface AuthState {
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
}

interface AuthActions {
  setUser: (user: User | null) => void
  updateUser: (updates: Partial<User>) => void
  clearUser: () => void
  setLoading: (loading: boolean) => void
}

type AuthStore = AuthState & AuthActions

// ============================================
// INITIAL STATE
// ============================================

const initialState: AuthState = {
  user: null,
  isAuthenticated: false,
  isLoading: true
}

// ============================================
// STORE
// ============================================

export const useAuthStore = create<AuthStore>()(
  persist(
    (set) => ({
      ...initialState,
      
      /**
       * Define o usuário atual
       */
      setUser: (user) =>
        set({
          user,
          isAuthenticated: !!user,
          isLoading: false
        }),
      
      /**
       * Atualiza dados parciais do usuário
       */
      updateUser: (updates) =>
        set((state) => ({
          user: state.user ? { ...state.user, ...updates } : null
        })),
      
      /**
       * Limpa o estado de autenticação
       */
      clearUser: () =>
        set({
          user: null,
          isAuthenticated: false,
          isLoading: false
        }),
      
      /**
       * Define estado de loading
       */
      setLoading: (loading) =>
        set({ isLoading: loading })
    }),
    {
      name: 'organiq-auth', // Nome da chave no localStorage
      storage: createJSONStorage(() => localStorage),
      
      // Particionar o que será persistido
      partialize: (state) => ({
        user: state.user,
        isAuthenticated: state.isAuthenticated
        // isLoading NÃO é persistido
      }),
      
      // Callback após hidratar do localStorage
      onRehydrateStorage: () => (state) => {
        // Finaliza loading após hidratar
        if (state) {
          state.isLoading = false
        }
      }
    }
  )
)

// ============================================
// SELECTORS (Para uso otimizado)
// ============================================

export const selectUser = (state: AuthStore) => state.user
export const selectIsAuthenticated = (state: AuthStore) => state.isAuthenticated
export const selectIsLoading = (state: AuthStore) => state.isLoading
export const selectHasCompletedOnboarding = (state: AuthStore) => 
  state.user?.hasCompletedOnboarding ?? false
export const selectArticlesRemaining = (state: AuthStore) => 
  state.user ? state.user.maxArticles - state.user.articlesUsed : 0
export const selectCanCreateArticles = (state: AuthStore) => 
  state.user ? state.user.articlesUsed < state.user.maxArticles : false

// ============================================
// HELPER HOOKS
// ============================================

/**
 * Hook otimizado que só re-renderiza quando o usuário muda
 */
export const useUser = () => useAuthStore(selectUser)

/**
 * Hook otimizado que só re-renderiza quando isAuthenticated muda
 */
export const useIsAuthenticated = () => useAuthStore(selectIsAuthenticated)

/**
 * Hook otimizado para loading
 */
export const useAuthLoading = () => useAuthStore(selectIsLoading)

/**
 * Hook para verificar se completou onboarding
 */
export const useHasCompletedOnboarding = () => 
  useAuthStore(selectHasCompletedOnboarding)

/**
 * Hook para verificar limite de artigos
 */
export const useArticlesRemaining = () => 
  useAuthStore(selectArticlesRemaining)

/**
 * Hook para verificar se pode criar artigos
 */
export const useCanCreateArticles = () => 
  useAuthStore(selectCanCreateArticles)
```

---

## 📁 types

### types/index.ts

```ts
// ============================================
// AUTH TYPES
// ============================================

export interface User {
  id: string
  name: string
  email: string
  planId: string
  planName: string
  maxArticles: number
  articlesUsed: number
  hasCompletedOnboarding: boolean
  createdAt: string
}

export interface LoginCredentials {
  email: string
  password: string
}

export interface RegisterData {
  name: string
  email: string
  password: string
}

export interface AuthResponse {
  user: User
  message?: string
}

// ============================================
// WIZARD / BUSINESS TYPES
// ============================================

export type ObjectiveType = 'leads' | 'sales' | 'branding'

export interface BusinessInfo {
  description: string
  primaryObjective: ObjectiveType
  secondaryObjective?: ObjectiveType
  siteUrl?: string
  hasBlog: boolean
  blogUrls: string[]
  articleCount: number
  brandFile?: File
}

export interface CompetitorData {
  competitorUrls: string[]
}

export interface IntegrationsData {
  wordpress: {
    siteUrl: string
    username: string
    appPassword: string
  }
  searchConsole?: {
    enabled: boolean
    propertyUrl?: string
  }
  analytics?: {
    enabled: boolean
    measurementId?: string
  }
}

export interface WizardState {
  currentStep: number
  businessInfo?: BusinessInfo
  competitorData?: CompetitorData
  integrationsData?: IntegrationsData
  articleIdeas?: ArticleIdea[]
}

// ============================================
// ARTICLES TYPES
// ============================================

export interface ArticleIdea {
  id: string
  title: string
  summary: string
  approved: boolean
  feedback?: string
}

export type ArticleStatus = 'generating' | 'publishing' | 'published' | 'error'

export interface Article {
  id: string
  title: string
  createdAt: string
  status: ArticleStatus
  postUrl?: string
  errorMessage?: string
  content?: string
}

export interface ArticlesResponse {
  articles: Article[]
  total: number
  page: number
  limit: number
}

export interface ArticleFilters {
  page?: number
  limit?: number
  status?: ArticleStatus | 'all'
}

export interface PublishPayload {
  articles: Array<{
    id: string
    feedback?: string
  }>
}

// ============================================
// PLANS TYPES
// ============================================

export interface Plan {
  id: string
  name: string
  maxArticles: number
  price: number
  features: string[]
  recommended?: boolean
}

export interface PlanInfo {
  name: string
  maxArticles: number
  articlesUsed: number
  nextBillingDate: string
  price: number
}

export interface CheckoutResponse {
  checkoutUrl: string
  sessionId: string
}

export interface PaymentStatus {
  id: string
  status: 'pending' | 'paid' | 'failed'
  planId: string
}

// ============================================
// ACCOUNT TYPES
// ============================================

export interface ProfileUpdateData {
  name: string
}

export interface IntegrationsUpdateData {
  wordpress?: {
    siteUrl: string
    username: string
    appPassword: string
  }
  searchConsole?: {
    enabled: boolean
    propertyUrl?: string
  }
  analytics?: {
    enabled: boolean
    measurementId?: string
  }
}

// ============================================
// API RESPONSE TYPES
// ============================================

export interface ApiError {
  message: string
  code?: string
  field?: string
}

export interface ApiResponse<T = unknown> {
  data?: T
  error?: ApiError
  success: boolean
}

// ============================================
// FORM TYPES
// ============================================

export interface LoginForm {
  email: string
  password: string
}

export interface RegisterForm {
  name: string
  email: string
  password: string
}

export interface BusinessForm {
  description: string
  primaryObjective: ObjectiveType
  secondaryObjective?: ObjectiveType
  siteUrl?: string
  hasBlog: boolean
  blogUrls: string[]
  articleCount: number
  brandFile?: File
}

export interface CompetitorsForm {
  competitorUrls: string[]
}

export interface IntegrationsForm {
  wordpress: {
    siteUrl: string
    username: string
    appPassword: string
  }
  searchConsole: {
    enabled: boolean
    propertyUrl?: string
  }
  analytics: {
    enabled: boolean
    measurementId?: string
  }
}

export interface NewArticlesForm {
  articleCount: number
}

// ============================================
// UTILITY TYPES
// ============================================

export type LoadingState = 'idle' | 'loading' | 'success' | 'error'

export interface PaginationState {
  page: number
  limit: number
  total: number
}
```

---


## 📊 Resumo Final

### ✅ Arquivos com Conteúdo

- `app/api/health/route.ts` (892 caracteres)
- `app/app/conta/page.tsx` (16642 caracteres)
- `app/app/layout.tsx` (1871 caracteres)
- `app/app/materias/page.tsx` (6992 caracteres)
- `app/app/novo/page.tsx` (6845 caracteres)
- `app/app/onboarding/page.tsx` (478 caracteres)
- `app/app/planos/page.tsx` (2179 caracteres)
- `app/error.tsx` (2620 caracteres)
- `app/globals.css` (2059 caracteres)
- `app/layout.tsx` (3174 caracteres)
- `app/login/page.tsx` (2898 caracteres)
- `app/not-found.tsx` (1812 caracteres)
- `app/page.tsx` (10089 caracteres)
- `app/providers.tsx` (615 caracteres)
- `app/robots.ts` (281 caracteres)
- `app/sitemap.ts` (710 caracteres)
- `components/articles/ArticleCard.tsx` (3894 caracteres)
- `components/articles/ArticleIdeaCard.tsx` (4553 caracteres)
- `components/articles/ArticleTable.tsx` (5853 caracteres)
- `components/forms/BusinessInfoForm.tsx` (8687 caracteres)
- `components/forms/CompetitorsForm.tsx` (5088 caracteres)
- `components/forms/IntegrationsForm.tsx` (12773 caracteres)
- `components/forms/LoginForm.tsx` (1965 caracteres)
- `components/forms/RegisterForm.tsx` (2079 caracteres)
- `components/layouts/Header.tsx` (1769 caracteres)
- `components/layouts/MobileNav.tsx` (2487 caracteres)
- `components/layouts/Sidebar.tsx` (4869 caracteres)
- `components/plans/PlanCard.tsx` (2687 caracteres)
- `components/shared/EmptyState.tsx` (2730 caracteres)
- `components/shared/ErrorBoundary.tsx` (4278 caracteres)
- `components/shared/LoadingSpinner.tsx` (3983 caracteres)
- `components/ui/button.tsx` (2845 caracteres)
- `components/ui/card.tsx` (2031 caracteres)
- `components/ui/dialog.tsx` (4049 caracteres)
- `components/ui/input.tsx` (1213 caracteres)
- `components/ui/label.tsx` (770 caracteres)
- `components/ui/progress.tsx` (1245 caracteres)
- `components/ui/select.tsx` (6290 caracteres)
- `components/ui/skeleton.tsx` (1481 caracteres)
- `components/ui/slider.tsx` (1982 caracteres)
- `components/ui/tabs.tsx` (2070 caracteres)
- `components/ui/textarea.tsx` (1764 caracteres)
- `components/ui/toast.tsx` (3583 caracteres)
- `components/wizards/NewArticlesWizard.tsx` (11612 caracteres)
- `components/wizards/OnboardingWizard.tsx` (3687 caracteres)
- `components/wizards/StepIndicator.tsx` (4696 caracteres)
- `hooks/useArticles.ts` (6191 caracteres)
- `hooks/useAuth.ts` (4064 caracteres)
- `hooks/usePlans.ts` (5736 caracteres)
- `hooks/useWizard.ts` (10577 caracteres)
- `lib/axios.ts` (5760 caracteres)
- `lib/constantes.ts` (3457 caracteres)
- `lib/utils.ts` (9035 caracteres)
- `lib/validations.ts` (8124 caracteres)
- `middleware.ts` (3144 caracteres)
- `next-env.d.ts` (250 caracteres)
- `next.config.ts` (1229 caracteres)
- `public/manifest.json` (1545 caracteres)
- `store/authStore.ts` (3877 caracteres)
- `types/index.ts` (4709 caracteres)

**Total:** 60 arquivos

### 📈 Estatísticas por Tipo

- **.css**: 1 arquivo(s), 2,059 caracteres
- **.json**: 1 arquivo(s), 1,545 caracteres
- **.ts**: 16 arquivo(s), 68,036 caracteres
- **.tsx**: 42 arquivo(s), 173,228 caracteres
