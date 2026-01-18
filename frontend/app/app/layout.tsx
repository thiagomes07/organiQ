"use client";

import { useEffect } from "react";
import { useRouter, usePathname } from "next/navigation";
import { useAuthStore } from "@/store/authStore";
import { Sidebar } from "@/components/layouts/Sidebar";
import { MobileNav } from "@/components/layouts/MobileNav";
import { LoadingSpinner } from "@/components/shared/LoadingSpinner";

export default function ProtectedLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();
  const pathname = usePathname();
  const { isAuthenticated, isLoading, user, isHydrated, hydrate } =
    useAuthStore();

  // Hydrate auth state from server on mount
  useEffect(() => {
    hydrate();
  }, [hydrate]);

  useEffect(() => {
    // Aguardar hydration do store
    if (!isHydrated || isLoading) return;

    // Não autenticado: redirecionar para login
    if (!isAuthenticated) {
      router.push("/login");
      return;
    }

    // Verificar onboarding
    if (user && !user.hasCompletedOnboarding) {
      const allowedPaths = ["/app/planos", "/app/onboarding"];

      // Se não está em uma rota permitida, redirecionar
      if (!allowedPaths.includes(pathname)) {
        router.push("/app/planos");
      }
    } else if (user && user.hasCompletedOnboarding && pathname === "/app/onboarding") {
      // Se já completou onboarding e tenta acessar, redirecionar para dashboard
      router.push("/app/materias");
    }
  }, [isAuthenticated, isLoading, isHydrated, user, pathname, router]);

  // Loading state durante hydration
  if (!isHydrated || isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-[var(--color-secondary-cream)]">
        <LoadingSpinner size="lg" text="Carregando..." />
      </div>
    );
  }

  // Não autenticado: não renderizar nada (vai redirecionar)
  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="flex min-h-screen bg-[var(--color-secondary-cream)]">
      {/* Sidebar Desktop */}
      <Sidebar />

      {/* Main Content */}
      <main className="flex-1 lg:ml-0 pb-20 lg:pb-0 p-4 md:p-8">
        <div className="max-w-7xl mx-auto">{children}</div>
      </main>

      {/* Mobile Navigation */}
      <MobileNav />
    </div>
  );
}
