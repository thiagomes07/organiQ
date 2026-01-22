'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { HelpCircle, Check, Eye, EyeOff } from 'lucide-react'
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
  hasGeneratedIdeas?: boolean
}

export function IntegrationsForm({
  onSubmit,
  onBack,
  isLoading,
  defaultValues,
  hasGeneratedIdeas,
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

  const [showPassword, setShowPassword] = useState(false)

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
                          <p>2. Role até <strong>&ldquo;Senhas de aplicativo&rdquo;</strong></p>
                          <p>3. Digite um nome (ex: &ldquo;organiQ&rdquo;) e clique em <strong>&ldquo;Adicionar&rdquo;</strong></p>
                          <p>4. Copie a senha gerada e cole aqui</p>
                          <p className="text-xs text-[var(--color-warning)] mt-4">
                            ⚠️ A senha só é exibida uma vez. Guarde-a em local seguro.
                          </p>
                        </DialogDescription>
                      </DialogHeader>
                    </DialogContent>
                  </Dialog>
                </div>
                <div className="relative">
                  <Input
                    id="wp-appPassword"
                    type="text"
                    className={!showPassword ? "text-security-disc" : ""}
                    placeholder="xxxx xxxx xxxx xxxx xxxx xxxx"
                    autoComplete="off"
                    error={errors.wordpress?.appPassword?.message}
                    {...register('wordpress.appPassword')}
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-[var(--color-primary-dark)]/40 hover:text-[var(--color-primary-dark)]/70 transition-colors"
                  >
                    {showPassword ? (
                      <EyeOff className="h-4 w-4" />
                    ) : (
                      <Eye className="h-4 w-4" />
                    )}
                  </button>
                </div>
              </div>
            </Accordion.Content>
          </div>
        </Accordion.Item>

        {/* Google Search Console (Opcional) - BREVE */}
        <Accordion.Item value="searchConsole" className="cursor-not-allowed opacity-60 relative" title="Em breve será implementado">
          <div className={cn(
            "border-2 rounded-[var(--radius-md)] overflow-hidden transition-colors border-[var(--color-border)] pointer-events-none select-none bg-gray-50"
          )}>
            <Accordion.Header>
              <div className="flex items-center justify-between w-full p-4">
                <div className="flex items-center gap-3">
                  <div className="flex items-center justify-center h-10 w-10 rounded-full bg-[var(--color-primary-teal)]/10 grayscale">
                    <span className="text-xl">📊</span>
                  </div>
                  <div className="text-left">
                    <h4 className="text-base font-semibold font-all-round text-[var(--color-primary-dark)] flex items-center gap-2">
                       Google Search Console
                       <span className="px-2 py-0.5 rounded-full bg-gray-200 text-[10px] text-gray-500 font-bold uppercase tracking-wider">Breve</span>
                    </h4>
                    <p className="text-xs font-onest text-[var(--color-primary-dark)]/60">
                      Análise de palavras-chave e rankings
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                   {/* Checkbox disabled visually */}
                   <div className="h-4 w-4 rounded border border-gray-300 bg-gray-100"></div>
                </div>
              </div>
            </Accordion.Header>
          </div>
        </Accordion.Item>

        {/* Google Analytics (Opcional) - BREVE */}
        <Accordion.Item value="analytics" className="cursor-not-allowed opacity-60 relative" title="Em breve será implementado">
          <div className={cn(
            "border-2 rounded-[var(--radius-md)] overflow-hidden transition-colors border-[var(--color-border)] pointer-events-none select-none bg-gray-50"
          )}>
            <Accordion.Header>
              <div className="flex items-center justify-between w-full p-4">
                <div className="flex items-center gap-3">
                  <div className="flex items-center justify-center h-10 w-10 rounded-full bg-[var(--color-primary-teal)]/10 grayscale">
                    <span className="text-xl">📈</span>
                  </div>
                  <div className="text-left">
                    <h4 className="text-base font-semibold font-all-round text-[var(--color-primary-dark)] flex items-center gap-2">
                      Google Analytics
                      <span className="px-2 py-0.5 rounded-full bg-gray-200 text-[10px] text-gray-500 font-bold uppercase tracking-wider">Breve</span>
                    </h4>
                    <p className="text-xs font-onest text-[var(--color-primary-dark)]/60">
                      Análise de tráfego e conversões
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                   {/* Checkbox disabled visually */}
                   <div className="h-4 w-4 rounded border border-gray-300 bg-gray-100"></div>
                </div>
              </div>
            </Accordion.Header>
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
          {hasGeneratedIdeas ? 'Ver Ideias' : 'Gerar Ideias'}
        </Button>
      </div>
    </form>
  )
}