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
      // Adicionado /app/conta para permitir acesso à troca de planos/perfil durante onboarding
      const allowedPaths = ["/app/planos", "/app/onboarding", "/app/conta"];

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

  // Banner de Lembrete de Onboarding
  const showOnboardingReminder =
    user &&
    !user.hasCompletedOnboarding &&
    pathname !== "/app/onboarding" &&
    pathname !== "/app/planos";

  return (
    <div className="flex min-h-screen bg-[var(--color-secondary-cream)]">
      {/* Sidebar Desktop */}
      <Sidebar />

      {/* Main Content */}
      <main className="flex-1 lg:ml-0 pb-20 lg:pb-0 p-4 md:p-8">

        {/* Reminder Banner */}
        {showOnboardingReminder && (
          <div className="mb-6 bg-[var(--color-primary-purple)]/10 border border-[var(--color-primary-purple)]/20 rounded-[var(--radius-sm)] p-4 flex flex-col md:flex-row items-center justify-between gap-4">
            <div>
              <h3 className="text-sm font-semibold font-all-round text-[var(--color-primary-purple)]">
                Vamos começar?
              </h3>
              <p className="text-xs text-[var(--color-primary-dark)]/70 font-onest mt-1">
                Finalize o preenchimento das informações para gerar seus primeiros artigos.
              </p>
            </div>
            <button
              onClick={() => router.push('/app/onboarding')}
              className="whitespace-nowrap px-4 py-2 bg-[var(--color-primary-purple)] text-white text-xs font-semibold rounded-[var(--radius-sm)] hover:bg-[var(--color-primary-purple)]/90 transition-colors"
            >
              Finalizar Onboarding →
            </button>
          </div>
        )}

        <div className="max-w-7xl mx-auto">{children}</div>
      </main>

      {/* Mobile Navigation */}
      <MobileNav />
    </div>
  );
}
