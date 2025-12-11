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
