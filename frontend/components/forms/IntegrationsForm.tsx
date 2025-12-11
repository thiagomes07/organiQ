'use client'

import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { HelpCircle, Check } from 'lucide-react'
import * as Accordion from '@radix-ui/react-accordion'
import { integrationsSchema, type IntegrationsInput } from '@/lib/validations'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { cn } from '@/lib/utils'

interface IntegrationsFormProps {
  onSubmit: (data: IntegrationsInput) => void
  onBack: () => void
  isLoading?: boolean
  defaultValues?: Partial<IntegrationsInput>
}

export function IntegrationsForm({
  onSubmit,
  onBack,
  isLoading,
  defaultValues,
}: IntegrationsFormProps) {
  const {
    register,
    handleSubmit,
    watch,
    setValue,
    formState: { errors },
  } = useForm<IntegrationsInput>({
    resolver: zodResolver(integrationsSchema),
    defaultValues: {
      wordpress: {
        siteUrl: '',
        username: '',
        appPassword: '',
        ...defaultValues?.wordpress,
      },
      searchConsole: {
        enabled: false,
        ...defaultValues?.searchConsole,
      },
      analytics: {
        enabled: false,
        ...defaultValues?.analytics,
      },
    },
  })

  const watchSearchConsoleEnabled = watch('searchConsole.enabled')
  const watchAnalyticsEnabled = watch('analytics.enabled')

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {/* Header */}
      <div className="space-y-2">
        <h3 className="text-xl font-semibold font-all-round text-[var(--color-primary-dark)]">
          Integrações
        </h3>
        <p className="text-sm font-onest text-[var(--color-primary-dark)]/70">
          Conecte suas ferramentas para publicação automática e análise de resultados
        </p>
      </div>

      {/* Accordion */}
      <Accordion.Root type="multiple" defaultValue={['wordpress']} className="space-y-4">
        {/* WordPress (Obrigatório) */}
        <Accordion.Item value="wordpress">
          <div className="border-2 border-[var(--color-primary-purple)] rounded-[var(--radius-md)] overflow-hidden">
            <Accordion.Header>
              <Accordion.Trigger className="flex items-center justify-between w-full p-4 hover:bg-[var(--color-primary-purple)]/5 transition-colors">
                <div className="flex items-center gap-3">
                  <div className="flex items-center justify-center h-10 w-10 rounded-full bg-[var(--color-primary-purple)]/10">
                    <span className="text-xl">🔌</span>
                  </div>
                  <div className="text-left">
                    <h4 className="text-base font-semibold font-all-round text-[var(--color-primary-dark)]">
                      WordPress
                    </h4>
                    <p className="text-xs font-onest text-[var(--color-primary-dark)]/60">
                      Obrigatório para publicação automática
                    </p>
                  </div>
                </div>
                <span className="px-3 py-1 rounded-full bg-[var(--color-primary-purple)]/10 text-xs font-semibold font-onest text-[var(--color-primary-purple)]">
                  Obrigatório
                </span>
              </Accordion.Trigger>
            </Accordion.Header>

            <Accordion.Content className="p-4 pt-0 space-y-4">
              {/* Site URL */}
              <div className="space-y-2">
                <Label htmlFor="wp-siteUrl" required>
                  URL do site WordPress
                </Label>
                <Input
                  id="wp-siteUrl"
                  type="url"
                  placeholder="https://seusite.com.br"
                  error={errors.wordpress?.siteUrl?.message}
                  {...register('wordpress.siteUrl')}
                />
              </div>

              {/* Username */}
              <div className="space-y-2">
                <Label htmlFor="wp-username" required>
                  Nome de usuário
                </Label>
                <Input
                  id="wp-username"
                  type="text"
                  placeholder="seu_usuario"
                  error={errors.wordpress?.username?.message}
                  {...register('wordpress.username')}
                />
              </div>

              {/* App Password */}
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label htmlFor="wp-appPassword" required>
                    Senha de aplicativo
                  </Label>
                  <Dialog>
                    <DialogTrigger asChild>
                      <button
                        type="button"
                        className="text-[var(--color-primary-teal)] hover:text-[var(--color-primary-purple)] transition-colors"
                      >
                        <HelpCircle className="h-4 w-4" />
                      </button>
                    </DialogTrigger>
                    <DialogContent>
                      <DialogHeader>
                        <DialogTitle>Como obter a senha de aplicativo?</DialogTitle>
                        <DialogDescription className="space-y-2 text-left">
                          <p>1. Acesse seu WordPress: <strong>Usuários → Perfil</strong></p>
                          <p>2. Role até <strong>"Senhas de aplicativo"</strong></p>
                          <p>3. Digite um nome (ex: "organiQ") e clique em <strong>"Adicionar"</strong></p>
                          <p>4. Copie a senha gerada e cole aqui</p>
                          <p className="text-xs text-[var(--color-warning)] mt-4">
                            ⚠️ A senha só é exibida uma vez. Guarde-a em local seguro.
                          </p>
                        </DialogDescription>
                      </DialogHeader>
                    </DialogContent>
                  </Dialog>
                </div>
                <Input
                  id="wp-appPassword"
                  type="password"
                  placeholder="xxxx xxxx xxxx xxxx xxxx xxxx"
                  error={errors.wordpress?.appPassword?.message}
                  {...register('wordpress.appPassword')}
                />
              </div>
            </Accordion.Content>
          </div>
        </Accordion.Item>

        {/* Google Search Console (Opcional) */}
        <Accordion.Item value="searchConsole">
          <div className={cn(
            "border-2 rounded-[var(--radius-md)] overflow-hidden transition-colors",
            watchSearchConsoleEnabled 
              ? "border-[var(--color-primary-teal)]" 
              : "border-[var(--color-border)]"
          )}>
            <Accordion.Header>
              <Accordion.Trigger className="flex items-center justify-between w-full p-4 hover:bg-[var(--color-primary-teal)]/5 transition-colors">
                <div className="flex items-center gap-3">
                  <div className="flex items-center justify-center h-10 w-10 rounded-full bg-[var(--color-primary-teal)]/10">
                    <span className="text-xl">📊</span>
                  </div>
                  <div className="text-left">
                    <h4 className="text-base font-semibold font-all-round text-[var(--color-primary-dark)]">
                      Google Search Console
                    </h4>
                    <p className="text-xs font-onest text-[var(--color-primary-dark)]/60">
                      Análise de palavras-chave e rankings
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  {watchSearchConsoleEnabled && (
                    <Check className="h-5 w-5 text-[var(--color-success)]" />
                  )}
                  <input
                    type="checkbox"
                    checked={watchSearchConsoleEnabled}
                    onChange={(e) => setValue('searchConsole.enabled', e.target.checked)}
                    onClick={(e) => e.stopPropagation()}
                    className="h-4 w-4 rounded border-[var(--color-border)] text-[var(--color-primary-teal)] focus:ring-[var(--color-primary-teal)]"
                  />
                </div>
              </Accordion.Trigger>
            </Accordion.Header>

            {watchSearchConsoleEnabled && (
              <Accordion.Content className="p-4 pt-0 space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="sc-propertyUrl" required>
                    URL da propriedade
                  </Label>
                  <Input
                    id="sc-propertyUrl"
                    type="url"
                    placeholder="https://seusite.com.br"
                    error={errors.searchConsole?.propertyUrl?.message}
                    {...register('searchConsole.propertyUrl')}
                  />
                  <p className="text-xs text-[var(--color-primary-dark)]/60 font-onest">
                    Use a mesma URL cadastrada no Search Console
                  </p>
                </div>
              </Accordion.Content>
            )}
          </div>
        </Accordion.Item>

        {/* Google Analytics (Opcional) */}
        <Accordion.Item value="analytics">
          <div className={cn(
            "border-2 rounded-[var(--radius-md)] overflow-hidden transition-colors",
            watchAnalyticsEnabled 
              ? "border-[var(--color-primary-teal)]" 
              : "border-[var(--color-border)]"
          )}>
            <Accordion.Header>
              <Accordion.Trigger className="flex items-center justify-between w-full p-4 hover:bg-[var(--color-primary-teal)]/5 transition-colors">
                <div className="flex items-center gap-3">
                  <div className="flex items-center justify-center h-10 w-10 rounded-full bg-[var(--color-primary-teal)]/10">
                    <span className="text-xl">📈</span>
                  </div>
                  <div className="text-left">
                    <h4 className="text-base font-semibold font-all-round text-[var(--color-primary-dark)]">
                      Google Analytics
                    </h4>
                    <p className="text-xs font-onest text-[var(--color-primary-dark)]/60">
                      Análise de tráfego e conversões
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  {watchAnalyticsEnabled && (
                    <Check className="h-5 w-5 text-[var(--color-success)]" />
                  )}
                  <input
                    type="checkbox"
                    checked={watchAnalyticsEnabled}
                    onChange={(e) => setValue('analytics.enabled', e.target.checked)}
                    onClick={(e) => e.stopPropagation()}
                    className="h-4 w-4 rounded border-[var(--color-border)] text-[var(--color-primary-teal)] focus:ring-[var(--color-primary-teal)]"
                  />
                </div>
              </Accordion.Trigger>
            </Accordion.Header>

            {watchAnalyticsEnabled && (
              <Accordion.Content className="p-4 pt-0 space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="ga-measurementId" required>
                    ID de Medição GA4
                  </Label>
                  <Input
                    id="ga-measurementId"
                    type="text"
                    placeholder="G-XXXXXXXXXX"
                    error={errors.analytics?.measurementId?.message}
                    {...register('analytics.measurementId')}
                  />
                  <p className="text-xs text-[var(--color-primary-dark)]/60 font-onest">
                    Formato: G-XXXXXXXXXX (encontrado em Admin → Fluxos de dados)
                  </p>
                </div>
              </Accordion.Content>
            )}
          </div>
        </Accordion.Item>
      </Accordion.Root>

      {/* Action Buttons */}
      <div className="flex items-center justify-between pt-4">
        <Button
          type="button"
          variant="outline"
          onClick={onBack}
          disabled={isLoading}
        >
          Voltar
        </Button>

        <Button
          type="submit"
          variant="primary"
          size="lg"
          isLoading={isLoading}
          disabled={isLoading}
        >
          Gerar Ideias
        </Button>
      </div>
    </form>
  )
}