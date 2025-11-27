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
