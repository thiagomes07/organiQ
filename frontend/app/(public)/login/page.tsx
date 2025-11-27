import { Metadata } from 'next'
import Link from 'next/link'
import { ArrowLeft } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { LoginForm } from '@/components/forms/LoginForm'
import { RegisterForm } from '@/components/forms/RegisterForm'

export const metadata: Metadata = {
  title: 'Login',
  description: 'Faça login ou crie sua conta no organiQ',
  robots: {
    index: false,
    follow: false,
  },
}

export default function LoginPage() {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center p-4">
      {/* Back to Home */}
      <div className="w-full max-w-md mb-8">
        <Link
          href="/"
          className="inline-flex items-center gap-2 text-sm font-medium font-onest text-[var(--color-primary-dark)]/70 hover:text-[var(--color-primary-dark)] transition-colors"
        >
          <ArrowLeft className="h-4 w-4" />
          Voltar para o início
        </Link>
      </div>

      {/* Logo */}
      <div className="mb-8 text-center">
        <h1 className="text-3xl font-bold font-all-round text-[var(--color-primary-purple)] mb-2">
          organiQ
        </h1>
        <p className="text-sm font-onest text-[var(--color-primary-teal)]">
          Naturalmente Inteligente
        </p>
      </div>

      {/* Login/Register Card */}
      <Card className="w-full max-w-md shadow-lg">
        <CardHeader className="space-y-1 pb-4">
          <CardTitle className="text-2xl text-center">
            Bem-vindo
          </CardTitle>
          <CardDescription className="text-center">
            Entre na sua conta ou crie uma nova para começar
          </CardDescription>
        </CardHeader>

        <CardContent>
          <Tabs defaultValue="login" className="w-full">
            <TabsList className="grid w-full grid-cols-2 mb-6">
              <TabsTrigger value="login">Entrar</TabsTrigger>
              <TabsTrigger value="register">Cadastrar</TabsTrigger>
            </TabsList>

            <TabsContent value="login">
              <LoginForm />
            </TabsContent>

            <TabsContent value="register">
              <RegisterForm />
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>

      {/* Footer Note */}
      <div className="mt-8 text-center">
        <p className="text-xs font-onest text-[var(--color-primary-dark)]/60">
          Ao continuar, você concorda com nossos{' '}
          <a href="#" className="text-[var(--color-primary-purple)] hover:underline">
            Termos de Uso
          </a>
          {' e '}
          <a href="#" className="text-[var(--color-primary-purple)] hover:underline">
            Política de Privacidade
          </a>
        </p>
      </div>
    </div>
  )
}