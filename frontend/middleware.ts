import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

// Rotas públicas (não requerem autenticação)
const publicPaths = ['/', '/login']

// Rotas de onboarding (requerem auth mas não completedOnboarding)
const onboardingPaths = ['/app/planos', '/app/onboarding']

// Helper para verificar se o path começa com algum dos prefixos
function matchesPath(pathname: string, paths: string[]): boolean {
  return paths.some(path => pathname === path || pathname.startsWith(path + '/'))
}

// Helper para parsear JWT (simplificado - em produção use biblioteca)
function parseJWT(token: string): { hasCompletedOnboarding?: boolean } | null {
  try {
    const base64Url = token.split('.')[1]
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map(c => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    )
    return JSON.parse(jsonPayload)
  } catch {
    return null
  }
}

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl
  
  // Permitir assets e API routes
  if (
    pathname.startsWith('/_next') ||
    pathname.startsWith('/api') ||
    pathname.startsWith('/static') ||
    pathname.includes('.')
  ) {
    return NextResponse.next()
  }

  // Obter token de autenticação do cookie
  const token = request.cookies.get('accessToken')?.value
  const isPublicPath = matchesPath(pathname, publicPaths)
  const isOnboardingPath = matchesPath(pathname, onboardingPaths)
  const isProtectedPath = pathname.startsWith('/app')

  // ============================================
  // CASO 1: Rota pública
  // ============================================
  if (isPublicPath) {
    // Se já está autenticado e tenta acessar login, redirecionar
    if (pathname === '/login' && token) {
      const user = parseJWT(token)
      const redirectTo = user?.hasCompletedOnboarding ? '/app/materias' : '/app/planos'
      return NextResponse.redirect(new URL(redirectTo, request.url))
    }
    
    return NextResponse.next()
  }

  // ============================================
  // CASO 2: Rota protegida sem autenticação
  // ============================================
  if (isProtectedPath && !token) {
    const loginUrl = new URL('/login', request.url)
    loginUrl.searchParams.set('redirect', pathname)
    return NextResponse.redirect(loginUrl)
  }

  // ============================================
  // CASO 3: Rota protegida com autenticação
  // ============================================
  if (isProtectedPath && token) {
    const user = parseJWT(token)
    
    // Token inválido
    if (!user) {
      return NextResponse.redirect(new URL('/login', request.url))
    }

    // Verificar se completou onboarding
    const hasCompletedOnboarding = user.hasCompletedOnboarding ?? false

    // Se NÃO completou onboarding
    if (!hasCompletedOnboarding) {
      // Permitir acesso apenas às rotas de onboarding
      if (!isOnboardingPath) {
        return NextResponse.redirect(new URL('/app/planos', request.url))
      }
    }

    // Se JÁ completou onboarding
    if (hasCompletedOnboarding) {
      // Se tentar acessar rotas de onboarding, redirecionar para dashboard
      if (isOnboardingPath) {
        return NextResponse.redirect(new URL('/app/materias', request.url))
      }
    }
  }

  return NextResponse.next()
}

export const config = {
  matcher: [
    /*
     * Match all request paths except for:
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     * - public files (public folder)
     */
    '/((?!_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)',
  ],
}