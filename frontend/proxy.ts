// proxy.ts
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const publicPaths = ["/", "/login"];
const onboardingPaths = ["/app/planos", "/app/onboarding"];

function matchesPath(pathname: string, paths: string[]): boolean {
  return paths.some(
    (path) => pathname === path || pathname.startsWith(path + "/")
  );
}

export default async function proxy(request: NextRequest) {
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
      // Validar token no BACKEND
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
    // Validar no backend
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
