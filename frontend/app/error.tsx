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
