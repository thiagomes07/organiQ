'use client'

import { Component, ReactNode } from 'react'
import { AlertTriangle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'

interface Props {
  children: ReactNode
  fallback?: ReactNode
  onReset?: () => void
}

interface State {
  hasError: boolean
  error: Error | null
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false, error: null }
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    // Log do erro para serviço de monitoramento (ex: Sentry)
    console.error('ErrorBoundary caught an error:', error, errorInfo)
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null })
    this.props.onReset?.()
  }

  render() {
    if (this.state.hasError) {
      // Usar fallback customizado se fornecido
      if (this.props.fallback) {
        return this.props.fallback
      }

      // Fallback padrão
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
                Ocorreu um erro inesperado. Você pode tentar recarregar a página ou voltar ao início.
              </p>
              {process.env.NODE_ENV === 'development' && this.state.error && (
                <details className="rounded-[var(--radius-sm)] bg-[var(--color-error)]/5 p-3">
                  <summary className="cursor-pointer text-xs font-semibold font-onest text-[var(--color-error)] mb-2">
                    Detalhes do erro (visível apenas em desenvolvimento)
                  </summary>
                  <pre className="text-xs overflow-auto font-mono text-[var(--color-error)]/80">
                    {this.state.error.message}
                  </pre>
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
              <Button
                variant="primary"
                onClick={this.handleReset}
                className="flex-1"
              >
                Tentar Novamente
              </Button>
            </CardFooter>
          </Card>
        </div>
      )
    }

    return this.props.children
  }
}

// Componente funcional para uso mais simples
export function ErrorFallback({ 
  error, 
  resetErrorBoundary 
}: { 
  error: Error
  resetErrorBoundary: () => void 
}) {
  return (
    <div className="flex min-h-[400px] items-center justify-center p-4">
      <Card className="max-w-md w-full">
        <CardHeader>
          <div className="flex items-center gap-3">
            <div className="rounded-full bg-[var(--color-error)]/10 p-2">
              <AlertTriangle className="h-6 w-6 text-[var(--color-error)]" />
            </div>
            <CardTitle>Erro ao carregar</CardTitle>
          </div>
        </CardHeader>
        <CardContent>
          <p className="text-sm font-onest text-[var(--color-primary-dark)]/70">
            {error.message || 'Ocorreu um erro ao carregar este conteúdo.'}
          </p>
        </CardContent>
        <CardFooter>
          <Button
            variant="primary"
            onClick={resetErrorBoundary}
            className="w-full"
          >
            Tentar Novamente
          </Button>
        </CardFooter>
      </Card>
    </div>
  )
}