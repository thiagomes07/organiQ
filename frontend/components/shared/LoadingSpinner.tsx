import { cn } from '@/lib/utils'

interface LoadingSpinnerProps {
  size?: 'sm' | 'md' | 'lg' | 'xl'
  className?: string
  text?: string
}

const sizeMap = {
  sm: 'h-4 w-4',
  md: 'h-8 w-8',
  lg: 'h-12 w-12',
  xl: 'h-16 w-16',
}

export function LoadingSpinner({ 
  size = 'md', 
  className,
  text 
}: LoadingSpinnerProps) {
  return (
    <div className="flex flex-col items-center justify-center gap-3">
      <svg
        className={cn(
          'animate-spin text-[var(--color-primary-purple)]',
          sizeMap[size],
          className
        )}
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
      >
        <circle
          className="opacity-25"
          cx="12"
          cy="12"
          r="10"
          stroke="currentColor"
          strokeWidth="4"
        />
        <path
          className="opacity-75"
          fill="currentColor"
          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
        />
      </svg>
      {text && (
        <p className="text-sm font-onest text-[var(--color-primary-dark)]/70 animate-pulse">
          {text}
        </p>
      )}
    </div>
  )
}

// Variante para overlay de tela cheia
export function LoadingOverlay({ 
  text,
  messages 
}: { 
  text?: string
  messages?: string[]
}) {
  const [currentMessage, setCurrentMessage] = React.useState(0)

  React.useEffect(() => {
    if (!messages || messages.length === 0) return

    const interval = setInterval(() => {
      setCurrentMessage((prev) => (prev + 1) % messages.length)
    }, 3000)

    return () => clearInterval(interval)
  }, [messages])

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-[var(--color-secondary-cream)] backdrop-blur-sm">
      <div className="flex flex-col items-center gap-6 p-8">
        {/* Spinner duplo animado */}
        <div className="relative">
          <div className="h-16 w-16 rounded-full border-4 border-[var(--color-primary-purple)]/20 border-t-[var(--color-primary-purple)] animate-spin" />
          <div className="absolute inset-2 h-12 w-12 rounded-full border-4 border-[var(--color-primary-teal)]/20 border-t-[var(--color-primary-teal)] animate-spin" 
               style={{ animationDirection: 'reverse', animationDuration: '1.5s' }} />
        </div>

        {/* Texto */}
        {messages && messages.length > 0 ? (
          <div className="text-center space-y-2">
            <p className="text-lg font-semibold font-all-round text-[var(--color-primary-dark)] animate-pulse">
              {messages[currentMessage]}
            </p>
            <p className="text-sm font-onest text-[var(--color-primary-dark)]/60">
              Isso pode levar alguns minutos
            </p>
          </div>
        ) : text ? (
          <p className="text-lg font-semibold font-all-round text-[var(--color-primary-dark)] animate-pulse">
            {text}
          </p>
        ) : null}
      </div>
    </div>
  )
}

// Variante inline para conteúdo
export function LoadingContent({ text }: { text?: string }) {
  return (
    <div className="flex items-center justify-center py-12">
      <LoadingSpinner size="lg" text={text} />
    </div>
  )
}

// Variante para botões (já incluída no Button, mas útil standalone)
export function ButtonSpinner() {
  return (
    <svg
      className="animate-spin h-4 w-4"
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 24 24"
    >
      <circle
        className="opacity-25"
        cx="12"
        cy="12"
        r="10"
        stroke="currentColor"
        strokeWidth="4"
      />
      <path
        className="opacity-75"
        fill="currentColor"
        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
      />
    </svg>
  )
}

// Export do React para uso no LoadingOverlay
import * as React from 'react'