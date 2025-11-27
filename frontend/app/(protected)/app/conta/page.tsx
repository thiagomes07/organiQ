'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { HelpCircle, Check, Calendar } from 'lucide-react'
import * as Accordion from '@radix-ui/react-accordion'
import { profileUpdateSchema, integrationsUpdateSchema, type ProfileUpdateInput, type IntegrationsUpdateInput } from '@/lib/validations'
import { usePlans } from '@/hooks/usePlans'
import { useUser } from '@/store/authStore'
import { formatDate, formatCurrency } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'

export default function ContaPage() {
  const user = useUser()
  const { currentPlan, openPortal, isOpeningPortal } = usePlans()
  const [isSavingProfile, setIsSavingProfile] = useState(false)
  const [isSavingIntegrations, setIsSavingIntegrations] = useState(false)

  // Profile Form
  const profileForm = useForm<ProfileUpdateInput>({
    resolver: zodResolver(profileUpdateSchema),
    defaultValues: {
      name: user?.name || '',
    },
  })

  // Integrations Form
  const integrationsForm = useForm<IntegrationsUpdateInput>({
    resolver: zodResolver(integrationsUpdateSchema),
    defaultValues: {
      wordpress: {
        siteUrl: '',
        username: '',
        appPassword: '',
      },
      searchConsole: {
        enabled: false,
      },
      analytics: {
        enabled: false,
      },
    },
  })

  const watchSearchConsoleEnabled = integrationsForm.watch('searchConsole.enabled')
  const watchAnalyticsEnabled = integrationsForm.watch('analytics.enabled')

  const handleUpdateProfile = async (data: ProfileUpdateInput) => {
    setIsSavingProfile(true)
    try {
      // TODO: Implement API call
      await new Promise(resolve => setTimeout(resolve, 1000))
      toast.success('Perfil atualizado com sucesso!')
    } catch (error) {
      toast.error('Erro ao atualizar perfil')
    } finally {
      setIsSavingProfile(false)
    }
  }

  const handleUpdateIntegrations = async (data: IntegrationsUpdateInput) => {
    setIsSavingIntegrations(true)
    try {
      // TODO: Implement API call
      await new Promise(resolve => setTimeout(resolve, 1000))
      toast.success('Integrações atualizadas com sucesso!')
    } catch (error) {
      toast.error('Erro ao atualizar integrações')
    } finally {
      setIsSavingIntegrations(false)
    }
  }

  const usagePercentage = user ? (user.articlesUsed / user.maxArticles) * 100 : 0

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold font-all-round text-[var(--color-primary-dark)]">
          Minha Conta
        </h1>
        <p className="text-sm font-onest text-[var(--color-primary-dark)]/70 mt-1">
          Gerencie suas configurações e integrações
        </p>
      </div>

      <div className="grid lg:grid-cols-2 gap-6">
        {/* Card 1: Perfil */}
        <Card>
          <CardHeader>
            <CardTitle>Perfil</CardTitle>
            <CardDescription>Informações da sua conta</CardDescription>
          </CardHeader>

          <form onSubmit={profileForm.handleSubmit(handleUpdateProfile)}>
            <CardContent className="space-y-4">
              {/* Nome */}
              <div className="space-y-2">
                <Label htmlFor="name" required>
                  Nome completo
                </Label>
                <Input
                  id="name"
                  type="text"
                  error={profileForm.formState.errors.name?.message}
                  {...profileForm.register('name')}
                />
              </div>

              {/* Email (disabled) */}
              <div className="space-y-2">
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  type="email"
                  value={user?.email || ''}
                  disabled
                />
                <p className="text-xs text-[var(--color-primary-dark)]/60 font-onest">
                  O email não pode ser alterado
                </p>
              </div>
            </CardContent>

            <CardFooter>
              <Button
                type="submit"
                variant="primary"
                isLoading={isSavingProfile}
                disabled={isSavingProfile}
              >
                Salvar Alterações
              </Button>
            </CardFooter>
          </form>
        </Card>

        {/* Card 2: Meu Plano */}
        <Card>
          <CardHeader>
            <CardTitle>Meu Plano</CardTitle>
            <CardDescription>Informações da sua assinatura</CardDescription>
          </CardHeader>

          <CardContent className="space-y-6">
            {/* Badge do Plano */}
            <div className="flex items-center gap-3">
              <div className="px-4 py-2 rounded-full bg-[var(--color-primary-purple)] text-white font-bold font-all-round text-lg">
                {user?.planName}
              </div>
              <div className="text-sm font-onest text-[var(--color-primary-dark)]/70">
                {currentPlan && formatCurrency(currentPlan.price)}/mês
              </div>
            </div>

            {/* Uso de Matérias */}
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium font-onest text-[var(--color-primary-dark)]">
                  Matérias usadas
                </span>
                <span className="text-sm font-semibold font-all-round text-[var(--color-primary-purple)]">
                  {user?.articlesUsed} / {user?.maxArticles}
                </span>
              </div>
              <Progress value={usagePercentage} showLabel />
              <p className="text-xs text-[var(--color-primary-dark)]/60 font-onest">
                {user && user.maxArticles - user.articlesUsed} matérias restantes este mês
              </p>
            </div>

            {/* Próxima Cobrança */}
            {currentPlan?.nextBillingDate && (
              <div className="flex items-center gap-2 text-sm font-onest text-[var(--color-primary-dark)]/70">
                <Calendar className="h-4 w-4" />
                <span>
                  Próxima cobrança: {formatDate(currentPlan.nextBillingDate)}
                </span>
              </div>
            )}
          </CardContent>

          <CardFooter className="flex gap-2">
            <Button
              variant="outline"
              onClick={openPortal}
              isLoading={isOpeningPortal}
              disabled={isOpeningPortal}
            >
              Gerenciar Assinatura
            </Button>
            <Button variant="primary">
              Fazer Upgrade
            </Button>
          </CardFooter>
        </Card>
      </div>

      {/* Card 3: Integrações (Full Width) */}
      <Card>
        <CardHeader>
          <CardTitle>Integrações</CardTitle>
          <CardDescription>Configure suas conexões com WordPress e Google</CardDescription>
        </CardHeader>

        <form onSubmit={integrationsForm.handleSubmit(handleUpdateIntegrations)}>
          <CardContent>
            <Accordion.Root type="multiple" className="space-y-4">
              {/* WordPress */}
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
                            Publicação automática de matérias
                          </p>
                        </div>
                      </div>
                      <Check className="h-5 w-5 text-[var(--color-success)]" />
                    </Accordion.Trigger>
                  </Accordion.Header>

                  <Accordion.Content className="p-4 pt-0 space-y-4">
                    <div className="space-y-2">
                      <Label htmlFor="wp-siteUrl">URL do site</Label>
                      <Input
                        id="wp-siteUrl"
                        type="url"
                        placeholder="https://seusite.com.br"
                        {...integrationsForm.register('wordpress.siteUrl')}
                      />
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="wp-username">Nome de usuário</Label>
                      <Input
                        id="wp-username"
                        type="text"
                        placeholder="seu_usuario"
                        {...integrationsForm.register('wordpress.username')}
                      />
                    </div>

                    <div className="space-y-2">
                      <div className="flex items-center justify-between">
                        <Label htmlFor="wp-appPassword">Senha de aplicativo</Label>
                        <Dialog>
                          <DialogTrigger asChild>
                            <button type="button" className="text-[var(--color-primary-teal)] hover:text-[var(--color-primary-purple)]">
                              <HelpCircle className="h-4 w-4" />
                            </button>
                          </DialogTrigger>
                          <DialogContent>
                            <DialogHeader>
                              <DialogTitle>Como obter a senha?</DialogTitle>
                              <DialogDescription className="space-y-2 text-left">
                                <p>1. WordPress → Usuários → Perfil</p>
                                <p>2. Role até "Senhas de aplicativo"</p>
                                <p>3. Adicione uma nova senha</p>
                                <p>4. Copie e cole aqui</p>
                              </DialogDescription>
                            </DialogHeader>
                          </DialogContent>
                        </Dialog>
                      </div>
                      <Input
                        id="wp-appPassword"
                        type="password"
                        placeholder="••••••••••••"
                        {...integrationsForm.register('wordpress.appPassword')}
                      />
                    </div>
                  </Accordion.Content>
                </div>
              </Accordion.Item>

              {/* Google Search Console */}
              <Accordion.Item value="searchConsole">
                <div className={cn(
                  "border-2 rounded-[var(--radius-md)] overflow-hidden",
                  watchSearchConsoleEnabled ? "border-[var(--color-primary-teal)]" : "border-[var(--color-border)]"
                )}>
                  <Accordion.Header>
                    <Accordion.Trigger className="flex items-center justify-between w-full p-4 hover:bg-[var(--color-primary-teal)]/5">
                      <div className="flex items-center gap-3">
                        <div className="h-10 w-10 rounded-full bg-[var(--color-primary-teal)]/10 flex items-center justify-center">
                          <span className="text-xl">📊</span>
                        </div>
                        <div className="text-left">
                          <h4 className="text-base font-semibold font-all-round text-[var(--color-primary-dark)]">
                            Google Search Console
                          </h4>
                          <p className="text-xs font-onest text-[var(--color-primary-dark)]/60">
                            Análise de palavras-chave
                          </p>
                        </div>
                      </div>
                      <div className="flex items-center gap-2">
                        {watchSearchConsoleEnabled && <Check className="h-5 w-5 text-[var(--color-success)]" />}
                        <input
                          type="checkbox"
                          {...integrationsForm.register('searchConsole.enabled')}
                          onClick={(e) => e.stopPropagation()}
                          className="h-4 w-4 rounded"
                        />
                      </div>
                    </Accordion.Trigger>
                  </Accordion.Header>

                  {watchSearchConsoleEnabled && (
                    <Accordion.Content className="p-4 pt-0">
                      <Input
                        type="url"
                        placeholder="https://seusite.com.br"
                        {...integrationsForm.register('searchConsole.propertyUrl')}
                      />
                    </Accordion.Content>
                  )}
                </div>
              </Accordion.Item>

              {/* Google Analytics */}
              <Accordion.Item value="analytics">
                <div className={cn(
                  "border-2 rounded-[var(--radius-md)] overflow-hidden",
                  watchAnalyticsEnabled ? "border-[var(--color-primary-teal)]" : "border-[var(--color-border)]"
                )}>
                  <Accordion.Header>
                    <Accordion.Trigger className="flex items-center justify-between w-full p-4 hover:bg-[var(--color-primary-teal)]/5">
                      <div className="flex items-center gap-3">
                        <div className="h-10 w-10 rounded-full bg-[var(--color-primary-teal)]/10 flex items-center justify-center">
                          <span className="text-xl">📈</span>
                        </div>
                        <div className="text-left">
                          <h4 className="text-base font-semibold font-all-round text-[var(--color-primary-dark)]">
                            Google Analytics
                          </h4>
                          <p className="text-xs font-onest text-[var(--color-primary-dark)]/60">
                            Análise de tráfego
                          </p>
                        </div>
                      </div>
                      <div className="flex items-center gap-2">
                        {watchAnalyticsEnabled && <Check className="h-5 w-5 text-[var(--color-success)]" />}
                        <input
                          type="checkbox"
                          {...integrationsForm.register('analytics.enabled')}
                          onClick={(e) => e.stopPropagation()}
                          className="h-4 w-4 rounded"
                        />
                      </div>
                    </Accordion.Trigger>
                  </Accordion.Header>

                  {watchAnalyticsEnabled && (
                    <Accordion.Content className="p-4 pt-0">
                      <Input
                        type="text"
                        placeholder="G-XXXXXXXXXX"
                        {...integrationsForm.register('analytics.measurementId')}
                      />
                    </Accordion.Content>
                  )}
                </div>
              </Accordion.Item>
            </Accordion.Root>
          </CardContent>

          <CardFooter>
            <Button
              type="submit"
              variant="primary"
              isLoading={isSavingIntegrations}
              disabled={isSavingIntegrations}
            >
              Atualizar Integrações
            </Button>
          </CardFooter>
        </form>
      </Card>
    </div>
  )
}