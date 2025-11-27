"use client";

import { usePathname } from "next/navigation";
import Link from "next/link";
import { FileText, PlusCircle, Settings, LogOut } from "lucide-react";
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

  const handleLogout = () => {
    if (window.confirm("Tem certeza que deseja sair?")) {
      logout();
    }
  };

  return (
    <aside className="hidden lg:flex lg:flex-col w-[280px] h-[calc(100vh-32px)] m-4 bg-white rounded-[var(--radius-lg)] shadow-md">
      {/* Logo */}
      <div className="flex items-center justify-center h-20 border-b border-[var(--color-border)]">
        <h1 className="text-2xl font-bold font-all-round text-[var(--color-primary-purple)]">
          organiQ
        </h1>
      </div>

      {/* User Info */}
      {user && (
        <div className="px-4 py-4 border-b border-[var(--color-border)]">
          <div className="flex items-center gap-3">
            <div className="flex items-center justify-center h-10 w-10 rounded-full bg-[var(--color-primary-purple)]/10 text-[var(--color-primary-purple)] font-semibold font-all-round">
              {user.name.charAt(0).toUpperCase()}
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium font-all-round text-[var(--color-primary-dark)] truncate">
                {user.name}
              </p>
              <p className="text-xs font-onest text-[var(--color-primary-dark)]/60 truncate">
                {user.email}
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Navigation */}
      <nav className="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
        {navItems.map((item) => {
          const Icon = item.icon;
          const isActive = pathname === item.href;

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
                {user.articlesUsed}/{user.maxArticles}
              </span>
            </div>
            <div className="w-full bg-[var(--color-primary-dark)]/10 rounded-full h-2">
              <div
                className="bg-[var(--color-primary-purple)] h-2 rounded-full transition-all duration-300"
                style={{
                  width: `${(user.articlesUsed / user.maxArticles) * 100}%`,
                }}
              />
            </div>
            <p className="text-xs font-onest text-[var(--color-primary-dark)]/60">
              {user.maxArticles - user.articlesUsed} matérias restantes
            </p>
          </div>
        </div>
      )}

      {/* Logout */}
      <div className="px-3 py-3 border-t border-[var(--color-border)]">
        <button
          onClick={handleLogout}
          disabled={isLoggingOut}
          className={cn(
            "flex items-center gap-3 w-full px-3 py-2.5 rounded-[var(--radius-sm)] text-sm font-medium font-onest transition-colors duration-200",
            "text-[var(--color-error)] hover:bg-[var(--color-error)]/10",
            "disabled:opacity-50 disabled:cursor-not-allowed"
          )}
        >
          <LogOut className="h-5 w-5" />
          <span>{isLoggingOut ? "Saindo..." : "Sair"}</span>
        </button>
      </div>
    </aside>
  );
}
