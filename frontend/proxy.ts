import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const publicPaths = ["/", "/login"];
const onboardingPaths = ["/app/planos", "/app/onboarding"];

/**
 * Decode JWT payload without verification (frontend-only validation)
 * The actual signature verification happens on the backend
 */
function parseJWT(token: string): { hasCompletedOnboarding?: boolean } | null {
  try {
    const base64Payload = token.split('.')[1];
    if (!base64Payload) return null;
    
    const payload = Buffer.from(base64Payload, 'base64').toString('utf-8');
    return JSON.parse(payload);
  } catch {
    return null;
  }
}

function matchesPath(pathname: string, paths: string[]): boolean {
  return paths.some(
    (path) => pathname === path || pathname.startsWith(path + "/")
  );
}

// Next.js 16: Exportar como "proxy" ao invés de "middleware"
export function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Permitir assets e API routes
  if (
    pathname.startsWith("/_next") ||
    pathname.startsWith("/api") ||
    pathname.startsWith("/static") ||
    pathname.startsWith("/fonts") ||
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
    // Usuário autenticado tentando acessar login - redirecionar
    if (pathname === "/login" && token) {
      const user = parseJWT(token);
      if (user) {
        const redirectTo = user.hasCompletedOnboarding
          ? "/app/materias"
          : "/app/planos";
        return NextResponse.redirect(new URL(redirectTo, request.url));
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
    const user = parseJWT(token);

    if (!user) {
      // Token inválido - redirecionar para login e limpar cookies
      const response = NextResponse.redirect(new URL("/login", request.url));
      response.cookies.delete("accessToken");
      response.cookies.delete("refreshToken");
      return response;
    }

    const hasCompletedOnboarding = user.hasCompletedOnboarding ?? false;

    // Usuário não completou onboarding - restringir acesso
    if (!hasCompletedOnboarding && !isOnboardingPath) {
      return NextResponse.redirect(new URL("/app/planos", request.url));
    }

    // Usuário completou onboarding mas tenta acessar páginas de onboarding
    if (hasCompletedOnboarding && isOnboardingPath) {
      return NextResponse.redirect(new URL("/app/materias", request.url));
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    "/((?!_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)",
  ],
};