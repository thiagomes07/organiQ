import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

type JwtPayload = {
  hasCompletedOnboarding?: boolean
}

function safeParseJwt(token: string): JwtPayload | null {
  try {
    const parts = token.split('.')
    if (parts.length < 2) return null

    const base64Url = parts[1]
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
    const padded = base64.padEnd(base64.length + ((4 - (base64.length % 4)) % 4), '=')

    const binary = typeof atob === 'function' ? atob(padded) : Buffer.from(padded, 'base64').toString('binary')
    const bytes = Uint8Array.from(binary, (char) => char.charCodeAt(0))
    const json = new TextDecoder().decode(bytes)

    return JSON.parse(json) as JwtPayload
  } catch {
    return null
  }
}

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  const publicPaths = ['/', '/login']
  const isPublicPath = publicPaths.includes(pathname)

  const token = request.cookies.get('accessToken')?.value

  if (!token && !isPublicPath) {
    return NextResponse.redirect(new URL('/login', request.url))
  }

  // Se tem token, aplica regras de onboarding somente em rotas protegidas
  if (token && !isPublicPath) {
    const payload = safeParseJwt(token)
    const hasCompletedOnboarding = payload?.hasCompletedOnboarding

    // Token inválido/inesperado: trata como não autenticado
    if (typeof hasCompletedOnboarding !== 'boolean') {
      return NextResponse.redirect(new URL('/login', request.url))
    }

    if (!hasCompletedOnboarding) {
      const allowed = ['/app/planos', '/app/onboarding']
      if (!allowed.includes(pathname)) {
        return NextResponse.redirect(new URL('/app/planos', request.url))
      }
    } else {
      // Onboarding completo: evita voltar para telas de onboarding
      const blocked = ['/app/planos', '/app/onboarding']
      if (blocked.includes(pathname)) {
        return NextResponse.redirect(new URL('/app/materias', request.url))
      }
    }
  }

  return NextResponse.next()
}

export const config = {
  matcher: ['/((?!api|_next/static|_next/image|favicon.ico|robots.txt|sitemap.xml|manifest.json).*)'],
}
