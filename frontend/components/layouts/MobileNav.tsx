'use client'

import { usePathname } from 'next/navigation'
import Link from 'next/link'
import { FileText, PlusCircle, Settings, LogOut } from 'lucide-react'
import { useAuth } from '@/hooks/useAuth'
import { cn } from '@/lib/utils'

interface NavItem {
  label: string
  href: string
  icon: React.ComponentType<{ className?: string }>
}

const navItems: NavItem[] = [
  {
    label: 'Matérias',
    href: '/app/materias',
    icon: FileText,
  },
  {
    label: 'Criar',
    href: '/app/novo',
    icon: PlusCircle,
  },
  {
    label: 'Conta',
    href: '/app/conta',
    icon: Settings,
  },
]

export function MobileNav() {
  const pathname = usePathname()
  const { logout, isLoggingOut } = useAuth()

  const handleLogout = () => {
    if (window.confirm('Tem certeza que deseja sair?')) {
      logout()
    }
  }

  return (
    <nav className="lg:hidden fixed bottom-0 left-0 right-0 z-50 bg-white border-t border-[var(--color-border)] shadow-lg">
      <div className="grid grid-cols-4 h-16">
        {navItems.map((item) => {
          const Icon = item.icon
          const isActive = pathname === item.href

          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                'flex flex-col items-center justify-center gap-1 transition-colors duration-200',
                isActive
                  ? 'text-[var(--color-primary-purple)]'
                  : 'text-[var(--color-primary-dark)]/60 hover:text-[var(--color-primary-dark)]'
              )}
            >
              <Icon className="h-5 w-5" />
              <span className="text-xs font-medium font-onest">{item.label}</span>
              {isActive && (
                <div className="absolute top-0 left-0 right-0 h-1 bg-[var(--color-primary-purple)]" />
              )}
            </Link>
          )
        })}

        {/* Logout Button */}
        <button
          onClick={handleLogout}
          disabled={isLoggingOut}
          className={cn(
            'flex flex-col items-center justify-center gap-1 transition-colors duration-200',
            'text-[var(--color-error)] hover:text-[var(--color-error)]/80',
            'disabled:opacity-50 disabled:cursor-not-allowed'
          )}
        >
          <LogOut className="h-5 w-5" />
          <span className="text-xs font-medium font-onest">
            {isLoggingOut ? 'Saindo...' : 'Sair'}
          </span>
        </button>
      </div>
    </nav>
  )
}