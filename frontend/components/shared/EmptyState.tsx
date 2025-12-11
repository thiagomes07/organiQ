import { FileText, Search, Inbox, AlertCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface EmptyStateProps {
  icon?: 'article' | 'search' | 'inbox' | 'alert'
  title: string
  description?: string
  action?: {
    label: string
    onClick: () => void
  }
  className?: string
}

const iconMap = {
  article: FileText,
  search: Search,
  inbox: Inbox,
  alert: AlertCircle,
}

export function EmptyState({
  icon = 'inbox',
  title,
  description,
  action,
  className,
}: EmptyStateProps) {
  const Icon = iconMap[icon]

  return (
    <div
      className={cn(
        'flex flex-col items-center justify-center py-12 px-4 text-center',
        className
      )}
    >
      {/* Ícone */}
      <div className="mb-4 rounded-full bg-[var(--color-primary-purple)]/10 p-6">
        <Icon className="h-12 w-12 text-[var(--color-primary-purple)]" />
      </div>

      {/* Título */}
      <h3 className="mb-2 text-xl font-semibold font-all-round text-[var(--color-primary-dark)]">
        {title}
      </h3>

      {/* Descrição */}
      {description && (
        <p className="mb-6 max-w-md text-sm font-onest text-[var(--color-primary-dark)]/70">
          {description}
        </p>
      )}

      {/* Ação */}
      {action && (
        <Button onClick={action.onClick} variant="secondary">
          {action.label}
        </Button>
      )}
    </div>
  )
}

// Variantes pré-definidas para casos comuns
export function EmptyArticles({ onCreate }: { onCreate?: () => void }) {
  return (
    <EmptyState
      icon="article"
      title="Nenhuma matéria criada ainda"
      description="Comece criando sua primeira matéria para aumentar seu tráfego orgânico."
      action={
        onCreate
          ? {
              label: 'Criar Primeira Matéria',
              onClick: onCreate,
            }
          : undefined
      }
    />
  )
}

export function EmptySearch({ query }: { query?: string }) {
  return (
    <EmptyState
      icon="search"
      title="Nenhum resultado encontrado"
      description={
        query
          ? `Não encontramos resultados para "${query}". Tente ajustar sua busca.`
          : 'Nenhum resultado corresponde aos filtros aplicados.'
      }
    />
  )
}

export function EmptyIdeas({ onRegenerate }: { onRegenerate?: () => void }) {
  return (
    <EmptyState
      icon="alert"
      title="Nenhuma ideia gerada"
      description="Não foi possível gerar ideias de matérias. Tente novamente com informações diferentes."
      action={
        onRegenerate
          ? {
              label: 'Tentar Novamente',
              onClick: onRegenerate,
            }
          : undefined
      }
    />
  )
}