import { Metadata } from 'next'
import Link from 'next/link'
import { Sparkles, Target, Zap } from 'lucide-react'
import { Header } from '@/components/layouts/Header'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

export const metadata: Metadata = {
  title: 'organiQ - Aumente seu tráfego orgânico com IA',
  description: 'Matérias de blog que geram autoridade e SEO. Naturalmente Inteligente. Geração automática de conteúdo otimizado para WordPress.',
}

const features = [
  {
    icon: Sparkles,
    title: 'Geração Automática de Conteúdo',
    description: 'IA avançada cria matérias de blog completas, otimizadas para seu nicho e público-alvo.',
  },
  {
    icon: Target,
    title: 'SEO Otimizado',
    description: 'Conteúdo estratégico baseado em análise de concorrentes e palavras-chave relevantes.',
  },
  {
    icon: Zap,
    title: 'Publicação Direta no WordPress',
    description: 'Integração nativa que publica suas matérias automaticamente no seu blog.',
  },
]

export default function LandingPage() {
  return (
    <div className="min-h-screen bg-[var(--color-secondary-cream)]">
      <Header />

      {/* Hero Section */}
      <section className="container mx-auto px-4 pt-20 pb-16 md:pt-32 md:pb-24">
        <div className="max-w-4xl mx-auto text-center space-y-8">
          {/* Tagline Badge */}
          <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-white border border-[var(--color-primary-teal)]/20 shadow-sm">
            <span className="text-sm font-semibold font-all-round text-[var(--color-primary-purple)]">
              ✨ Naturalmente Inteligente
            </span>
          </div>

          {/* Main Heading */}
          <h1 className="text-4xl md:text-5xl lg:text-6xl font-bold font-all-round text-[var(--color-primary-dark)] leading-tight">
            Aumente seu tráfego orgânico{' '}
            <span className="text-[var(--color-primary-purple)]">com IA</span>
          </h1>

          {/* Subtitle */}
          <p className="text-lg md:text-xl font-onest text-[var(--color-primary-teal)] max-w-2xl mx-auto">
            Matérias de blog que geram autoridade e SEO automaticamente
          </p>

          {/* CTA Button */}
          <div className="pt-4">
            <Link href="/login">
              <Button
                variant="secondary"
                size="lg"
                className="text-base md:text-lg px-8 md:px-12 h-12 md:h-14"
              >
                Criar minha conta grátis
              </Button>
            </Link>
          </div>

          {/* Social Proof */}
          <p className="text-sm font-onest text-[var(--color-primary-dark)]/60">
            Junte-se a centenas de empresas que já aumentaram seu tráfego orgânico
          </p>
        </div>
      </section>

      {/* Features Section */}
      <section id="features" className="container mx-auto px-4 py-16 md:py-24">
        <div className="max-w-6xl mx-auto">
          {/* Section Title */}
          <div className="text-center mb-12 md:mb-16">
            <h2 className="text-3xl md:text-4xl font-bold font-all-round text-[var(--color-primary-dark)] mb-4">
              Como funciona
            </h2>
            <p className="text-lg font-onest text-[var(--color-primary-dark)]/70 max-w-2xl mx-auto">
              Três recursos poderosos que transformam sua estratégia de conteúdo
            </p>
          </div>

          {/* Features Grid */}
          <div className="grid md:grid-cols-3 gap-6 md:gap-8">
            {features.map((feature, index) => {
              const Icon = feature.icon

              return (
                <Card
                  key={index}
                  className="group hover:shadow-lg hover:scale-[1.02] transition-all duration-200 border-[var(--color-primary-teal)]/20"
                >
                  <CardHeader>
                    {/* Icon */}
                    <div className="mb-4 inline-flex items-center justify-center h-12 w-12 rounded-full bg-[var(--color-primary-purple)]/10 text-[var(--color-primary-purple)] group-hover:bg-[var(--color-primary-purple)] group-hover:text-white transition-colors duration-200">
                      <Icon className="h-6 w-6" />
                    </div>

                    {/* Title */}
                    <CardTitle className="text-xl">
                      {feature.title}
                    </CardTitle>
                  </CardHeader>

                  <CardContent>
                    {/* Description */}
                    <CardDescription className="text-base">
                      {feature.description}
                    </CardDescription>
                  </CardContent>
                </Card>
              )
            })}
          </div>
        </div>
      </section>

      {/* How It Works Section */}
      <section id="how-it-works" className="container mx-auto px-4 py-16 md:py-24">
        <div className="max-w-4xl mx-auto">
          <div className="bg-white rounded-[var(--radius-lg)] shadow-md p-8 md:p-12">
            <h2 className="text-3xl md:text-4xl font-bold font-all-round text-[var(--color-primary-dark)] mb-8 text-center">
              Simples e rápido
            </h2>

            <div className="space-y-8">
              {/* Step 1 */}
              <div className="flex gap-4 md:gap-6">
                <div className="flex-shrink-0 flex items-center justify-center h-10 w-10 rounded-full bg-[var(--color-primary-purple)] text-white font-bold font-all-round">
                  1
                </div>
                <div>
                  <h3 className="text-xl font-semibold font-all-round text-[var(--color-primary-dark)] mb-2">
                    Conecte seu WordPress
                  </h3>
                  <p className="font-onest text-[var(--color-primary-dark)]/70">
                    Integração segura em poucos cliques. Suporte para qualquer site WordPress.
                  </p>
                </div>
              </div>

              {/* Step 2 */}
              <div className="flex gap-4 md:gap-6">
                <div className="flex-shrink-0 flex items-center justify-center h-10 w-10 rounded-full bg-[var(--color-primary-purple)] text-white font-bold font-all-round">
                  2
                </div>
                <div>
                  <h3 className="text-xl font-semibold font-all-round text-[var(--color-primary-dark)] mb-2">
                    Configure seus objetivos
                  </h3>
                  <p className="font-onest text-[var(--color-primary-dark)]/70">
                    Defina seu nicho, público-alvo e concorrentes. Nossa IA analisa tudo para criar a melhor estratégia.
                  </p>
                </div>
              </div>

              {/* Step 3 */}
              <div className="flex gap-4 md:gap-6">
                <div className="flex-shrink-0 flex items-center justify-center h-10 w-10 rounded-full bg-[var(--color-primary-purple)] text-white font-bold font-all-round">
                  3
                </div>
                <div>
                  <h3 className="text-xl font-semibold font-all-round text-[var(--color-primary-dark)] mb-2">
                    Aprove e publique
                  </h3>
                  <p className="font-onest text-[var(--color-primary-dark)]/70">
                    Revise as ideias geradas, adicione seu toque pessoal e publique automaticamente no seu blog.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="container mx-auto px-4 py-16 md:py-24">
        <div className="max-w-4xl mx-auto text-center bg-gradient-to-br from-[var(--color-primary-purple)] to-[var(--color-primary-teal)] rounded-[var(--radius-lg)] shadow-xl p-8 md:p-12">
          <h2 className="text-3xl md:text-4xl font-bold font-all-round text-white mb-4">
            Pronto para aumentar seu tráfego?
          </h2>
          <p className="text-lg font-onest text-white/90 mb-8 max-w-2xl mx-auto">
            Comece gratuitamente e veja os resultados em poucos dias
          </p>
          <Link href="/login">
            <Button
              variant="secondary"
              size="lg"
              className="text-base md:text-lg px-8 md:px-12 h-12 md:h-14"
            >
              Criar minha conta grátis
            </Button>
          </Link>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-[var(--color-border)] bg-white">
        <div className="container mx-auto px-4 py-8">
          <div className="flex flex-col md:flex-row items-center justify-between gap-4">
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-bold font-all-round text-[var(--color-primary-purple)]">
                organiQ
              </h1>
              <span className="text-sm font-onest text-[var(--color-primary-dark)]/60">
                © 2024 Todos os direitos reservados
              </span>
            </div>
            <div className="flex items-center gap-6">
              <a
                href="#"
                className="text-sm font-onest text-[var(--color-primary-dark)]/70 hover:text-[var(--color-primary-dark)] transition-colors"
              >
                Termos de Uso
              </a>
              <a
                href="#"
                className="text-sm font-onest text-[var(--color-primary-dark)]/70 hover:text-[var(--color-primary-dark)] transition-colors"
              >
                Política de Privacidade
              </a>
              <a
                href="#"
                className="text-sm font-onest text-[var(--color-primary-dark)]/70 hover:text-[var(--color-primary-dark)] transition-colors"
              >
                Contato
              </a>
            </div>
          </div>
        </div>
      </footer>
    </div>
  )
}