"use client";

import { useState, useRef, useEffect } from "react";
import { usePathname } from "next/navigation";
import Link from "next/link";
import { FileText, PlusCircle, Settings, LogOut, ChevronDown } from "lucide-react";
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
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Fecha o dropdown quando clicar fora
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsDropdownOpen(false);
      }
    }

    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const handleLogout = () => {
    setIsDropdownOpen(false);
    logout();
  };

  return (
    <aside className="hidden lg:flex lg:flex-col w-[280px] h-[calc(100vh-32px)] m-4 bg-white rounded-[var(--radius-lg)] shadow-md">
      {/* Logo */}
      <div className="flex items-center justify-center h-20 border-b border-[var(--color-border)]">
        <h1 className="text-2xl font-bold font-all-round text-[var(--color-primary-purple)]">
          organiQ
        </h1>
      </div>

      {/* User Info with Dropdown */}
      {user && (
        <div className="px-4 py-4 border-b border-[var(--color-border)] relative" ref={dropdownRef}>
          <button
            onClick={() => setIsDropdownOpen(!isDropdownOpen)}
            className="flex items-center gap-3 w-full hover:bg-[var(--color-primary-dark)]/5 rounded-[var(--radius-sm)] p-2 -m-2 transition-colors duration-200"
          >
            <div className="flex items-center justify-center h-10 w-10 rounded-full bg-[var(--color-primary-purple)]/10 text-[var(--color-primary-purple)] font-semibold font-all-round">
              {user.name.charAt(0).toUpperCase()}
            </div>
            <div className="flex-1 min-w-0 text-left">
              <p className="text-sm font-medium font-all-round text-[var(--color-primary-dark)] truncate">
                {user.name}
              </p>
              <p className="text-xs font-onest text-[var(--color-primary-dark)]/60 truncate">
                {user.email}
              </p>
            </div>
            <ChevronDown
              className={cn(
                "h-4 w-4 text-[var(--color-primary-dark)]/60 transition-transform duration-200",
                isDropdownOpen && "rotate-180"
              )}
            />
          </button>

          {/* Dropdown Menu */}
          {isDropdownOpen && (
            <div className="absolute top-full left-4 right-4 mt-2 bg-white rounded-[var(--radius-sm)] shadow-lg border border-[var(--color-border)] z-50 overflow-hidden">
              <Link
                href="/app/conta"
                onClick={() => setIsDropdownOpen(false)}
                className="flex items-center gap-3 px-4 py-3 text-sm font-medium font-onest text-[var(--color-primary-dark)] hover:bg-[var(--color-primary-dark)]/5 transition-colors duration-200 border-b border-[var(--color-border)]"
              >
                <Settings className="h-4 w-4" />
                <span>Minha Conta</span>
              </Link>
              <button
                onClick={handleLogout}
                disabled={isLoggingOut}
                className={cn(
                  "flex items-center gap-3 w-full px-4 py-3 text-sm font-medium font-onest transition-colors duration-200",
                  "text-[var(--color-error)] hover:bg-[var(--color-error)]/10",
                  "disabled:opacity-50 disabled:cursor-not-allowed"
                )}
              >
                <LogOut className="h-4 w-4" />
                <span>{isLoggingOut ? "Saindo..." : "Sair da conta"}</span>
              </button>
            </div>
          )}
        </div>
      )}

      {/* Navigation */}
      <nav className="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
        {navItems.map((item) => {
          const Icon = item.icon;
          const isActive = pathname === item.href;

          // Lock logic: Bloquear rotas de matérias se não completou onboarding
          const isLocked = !user?.hasCompletedOnboarding && (
            item.href === '/app/materias' ||
            item.href === '/app/novo'
          );

          if (isLocked) {
            return (
              <div
                key={item.href}
                className="flex items-center gap-3 px-3 py-2.5 rounded-[var(--radius-sm)] text-sm font-medium font-onest text-gray-400 cursor-not-allowed opacity-60"
                title="Finalize o onboarding para acessar"
              >
                <Icon className="h-5 w-5" />
                <span>{item.label}</span>
                <span className="ml-auto">🔒</span>
              </div>
            )
          }

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
                {user.articlesUsed}/{user.maxArticles || 1}
              </span>
            </div>
            <div className="w-full bg-[var(--color-primary-dark)]/10 rounded-full h-2">
              <div
                className="bg-[var(--color-primary-purple)] h-2 rounded-full transition-all duration-300"
                style={{
                  width: `${(user.articlesUsed / (user.maxArticles || 1)) * 100}%`,
                }}
              />
            </div>
            <p className="text-xs font-onest text-[var(--color-primary-dark)]/60">
              {user.maxArticles - user.articlesUsed} matérias restantes
            </p>
          </div>
        </div>
      )}
    </aside>
  );
}       