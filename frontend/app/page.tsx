import { Metadata } from "next";
import Link from "next/link";
import {
  Sparkles,
  Target,
  Zap,
  Check,
  BarChart3,
  Globe2,
  ArrowRight,
  LayoutTemplate,
  Bot,
  Search,
} from "lucide-react";
import { Header } from "@/components/layouts/Header";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export const metadata: Metadata = {
  title: "organiQ - Aumente seu tráfego orgânico com IA",
  description:
    "O robô escritor que analisa concorrentes, cria estratégias de SEO local e publica no WordPress automaticamente.",
};

// --- Sub-components para Organização ---

const AbstractDashboard = () => (
  <div className="relative w-full max-w-[600px] mx-auto perspective-1000">
    {/* Background Glow */}
    <div className="absolute -inset-4 bg-[var(--color-primary-purple)]/20 blur-3xl rounded-full opacity-50 animate-pulse" />

    {/* Main Window Interface */}
    <div className="relative bg-white/80 backdrop-blur-md border border-white/50 rounded-xl shadow-2xl overflow-hidden transform rotate-y-6 hover:rotate-0 transition-transform duration-700">
      {/* Fake Browser Header */}
      <div className="bg-[var(--color-primary-dark)]/5 p-3 flex items-center gap-2 border-b border-[var(--color-primary-dark)]/5">
        <div className="w-3 h-3 rounded-full bg-red-400" />
        <div className="w-3 h-3 rounded-full bg-yellow-400" />
        <div className="w-3 h-3 rounded-full bg-green-400" />
        <div className="ml-4 h-2 w-32 bg-[var(--color-primary-dark)]/10 rounded-full" />
      </div>

      {/* Interface Content */}
      <div className="p-6 space-y-4">
        {/* Status Bar */}
        <div className="flex justify-between items-center">
          <div className="h-4 w-24 bg-[var(--color-primary-purple)]/20 rounded animate-pulse" />
          <div className="flex gap-2">
            <span className="px-2 py-1 rounded text-[10px] font-bold bg-green-100 text-green-700">
              SEO: 98/100
            </span>
          </div>
        </div>

        {/* Article Preview */}
        <div className="space-y-3">
          <div className="h-6 w-3/4 bg-[var(--color-primary-dark)]/80 rounded" />
          <div className="space-y-2">
            <div className="h-3 w-full bg-[var(--color-primary-dark)]/10 rounded" />
            <div className="h-3 w-full bg-[var(--color-primary-dark)]/10 rounded" />
            <div className="h-3 w-5/6 bg-[var(--color-primary-dark)]/10 rounded" />
          </div>
        </div>

        {/* Floating Card: Competitor Analysis */}
        <div className="absolute top-12 right-4 w-48 bg-white border border-[var(--color-primary-purple)]/20 rounded-lg shadow-lg p-3 transform translate-x-4">
          <div className="flex items-center gap-2 mb-2">
            <Search className="w-4 h-4 text-[var(--color-primary-purple)]" />
            <span className="text-xs font-bold text-[var(--color-primary-dark)]">
              Gap Encontrado
            </span>
          </div>
          <p className="text-[10px] text-gray-500 leading-tight">
            Seu concorrente não fala sobre &ldquo;Implantes em 1 dia&rdquo;. Vamos escrever
            sobre isso.
          </p>
        </div>

        {/* Floating Card: WordPress */}
        <div className="absolute bottom-4 left-4 flex items-center gap-3 bg-[var(--color-primary-teal)] text-white px-4 py-2 rounded-lg shadow-lg transform -translate-x-2">
          <div className="w-2 h-2 rounded-full bg-green-400 animate-ping" />
          <span className="text-xs font-bold">Publicado no WordPress</span>
        </div>
      </div>
    </div>
  </div>
);

const CheckListItem = ({ children }: { children: React.ReactNode }) => (
  <li className="flex items-start gap-3">
    <div className="mt-1 bg-[var(--color-success)]/10 p-1 rounded-full">
      <Check className="w-4 h-4 text-[var(--color-success)]" />
    </div>
    <span className="text-[var(--color-primary-dark)]/80 font-onest text-sm md:text-base">
      {children}
    </span>
  </li>
);

export default function LandingPage() {
  return (
    <div className="!scroll-smooth min-h-screen bg-[var(--color-secondary-cream)] selection:bg-[var(--color-secondary-yellow)] selection:text-[var(--color-primary-dark)]">
      <Header />

      {/* --- HERO SECTION --- */}
      <section className="relative pt-28 pb-20 md:pt-40 md:pb-32 overflow-hidden">
        {/* Background Elements */}
        <div className="absolute top-0 right-0 w-[500px] h-[500px] bg-[var(--color-primary-purple)]/5 rounded-full blur-3xl -translate-y-1/2 translate-x-1/3" />
        <div className="absolute bottom-0 left-0 w-[500px] h-[500px] bg-[var(--color-secondary-yellow)]/10 rounded-full blur-3xl translate-y-1/3 -translate-x-1/3" />

        <div className="container mx-auto px-4 relative z-10">
          <div className="grid md:grid-cols-2 gap-12 items-center">
            {/* Copy */}
            <div className="space-y-6 md:space-y-8">
              <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-white border border-[var(--color-primary-purple)]/20 shadow-sm w-fit">
                <Sparkles className="w-4 h-4 text-[var(--color-primary-purple)]" />
                <span className="text-xs font-bold font-all-round tracking-wide text-[var(--color-primary-dark)] uppercase">
                  Naturalmente Inteligente
                </span>
              </div>

              <h1 className="text-4xl md:text-5xl lg:text-6xl font-bold font-all-round text-[var(--color-primary-dark)] leading-[1.1]">
                Seu Blog no Piloto Automático,{" "}
                <span className="text-transparent bg-clip-text bg-gradient-to-r from-[var(--color-primary-purple)] to-[var(--color-primary-teal)]">
                  Sem Alucinações.
                </span>
              </h1>

              <p className="text-lg md:text-xl font-onest text-[var(--color-primary-dark)]/70 max-w-lg leading-relaxed">
                O organiQ analisa seus concorrentes, encontra gaps de conteúdo e
                escreve artigos otimizados para SEO que são publicados direto no
                seu WordPress.
              </p>

              <div className="flex flex-col sm:flex-row gap-4 pt-2">
                <Link href="/login" className="w-full sm:w-auto">
                  <Button
                    size="lg"
                    className="w-full sm:w-auto h-14 px-8 text-lg font-all-round bg-[var(--color-secondary-yellow)] text-[var(--color-primary-dark)] hover:bg-[var(--color-secondary-yellow)]/90 hover:scale-105 transition-all shadow-lg hover:shadow-[var(--color-secondary-yellow)]/50"
                  >
                    Começar Grátis
                  </Button>
                </Link>
                <Link href="#how-it-works" className="w-full sm:w-auto">
                  <Button
                    variant="ghost"
                    size="lg"
                    className="w-full sm:w-auto h-14 px-8 text-lg font-onest text-[var(--color-primary-dark)] hover:bg-[var(--color-primary-purple)]/5"
                  >
                    Ver como funciona <ArrowRight className="ml-2 w-5 h-5" />
                  </Button>
                </Link>
              </div>

              {/* Mini Social Proof */}
              <div className="pt-6 border-t border-[var(--color-primary-dark)]/10">
                <p className="text-sm font-onest text-[var(--color-primary-dark)]/60 mb-3">
                  Integração nativa com:
                </p>
                <div className="flex items-center gap-6 opacity-60 grayscale hover:grayscale-0 transition-all duration-300">
                  <div className="flex items-center gap-2">
                    <Globe2 className="w-5 h-5" />
                    <span className="font-bold">WordPress</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <BarChart3 className="w-5 h-5" />
                    <span className="font-bold">Analytics</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <Search className="w-5 h-5" />
                    <span className="font-bold">Search Console</span>
                  </div>
                </div>
              </div>
            </div>

            {/* Visual */}
            <div className="relative hidden md:block">
              <AbstractDashboard />
            </div>
          </div>
        </div>
      </section>

      {/* --- BENTO GRID FEATURES --- */}
      <section id="features" className="py-20 md:py-32 bg-white relative">
        <div className="container mx-auto px-4">
          <div className="text-center max-w-3xl mx-auto mb-16">
            <h2 className="text-3xl md:text-4xl font-bold font-all-round text-[var(--color-primary-dark)] mb-4">
              Não é só um gerador de texto. <br />É um estrategista de SEO.
            </h2>
            <p className="text-lg font-onest text-[var(--color-primary-dark)]/70">
              A maioria das IAs escreve conteúdo genérico. O organiQ usa dados
              reais do seu negócio e dos seus concorrentes.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 max-w-6xl mx-auto">
            {/* Feature 1: Large (Competitor Analysis) */}
            <Card className="md:col-span-2 bg-[var(--color-primary-dark)] text-white overflow-hidden relative border-none group">
              <div className="absolute top-0 right-0 p-32 bg-[var(--color-primary-purple)]/20 blur-[100px] rounded-full" />
              <CardHeader className="relative z-10">
                <div className="w-12 h-12 bg-white/10 rounded-lg flex items-center justify-center mb-4 group-hover:bg-[var(--color-primary-purple)] transition-colors duration-300">
                  <Target className="w-6 h-6 text-white" />
                </div>
                <CardTitle className="text-2xl font-all-round">
                  Análise de Concorrência & Gaps
                </CardTitle>
                <CardDescription className="text-gray-300 text-base font-onest mt-2">
                  Você nos diz quem são seus concorrentes. Nós analisamos sobre
                  o que eles escrevem e encontramos oportunidades (gaps) que
                  ninguém está explorando no seu nicho.
                </CardDescription>
              </CardHeader>
              <CardContent className="relative z-10 pt-4">
                <div className="bg-white/5 border border-white/10 rounded-lg p-4 font-mono text-sm text-green-400">
                  {`> Analisando clinicadentes.com.br...`} <br />
                  {`> Gap detectado: "Implante carga imediata"`} <br />
                  {`> Sugestão: "Guia de Preços 2024"`}
                </div>
              </CardContent>
            </Card>

            {/* Feature 2: Tall (Local SEO) */}
            <Card className="md:row-span-2 bg-[var(--color-secondary-cream)] border-[var(--color-primary-teal)]/20 group hover:border-[var(--color-primary-purple)]/50 transition-colors">
              <CardHeader>
                <div className="w-12 h-12 bg-[var(--color-primary-teal)]/10 rounded-lg flex items-center justify-center mb-4">
                  <Globe2 className="w-6 h-6 text-[var(--color-primary-teal)]" />
                </div>
                <CardTitle className="text-xl font-all-round text-[var(--color-primary-dark)]">
                  SEO Local Multi-unidade
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="font-onest text-[var(--color-primary-dark)]/70 mb-6">
                  Tem franquias ou várias unidades? O organiQ entende a
                  hierarquia País &gt; Estado &gt; Cidade e gera conteúdo
                  específico para cada localidade.
                </p>
                <div className="space-y-2">
                  <div className="flex items-center gap-2 p-2 bg-white rounded border border-[var(--color-primary-teal)]/10 text-sm">
                    <span className="w-2 h-2 rounded-full bg-green-500" />{" "}
                    Unidade São Paulo
                  </div>
                  <div className="flex items-center gap-2 p-2 bg-white rounded border border-[var(--color-primary-teal)]/10 text-sm">
                    <span className="w-2 h-2 rounded-full bg-green-500" />{" "}
                    Unidade Rio de Janeiro
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Feature 3: Standard (Human Control) */}
            <Card className="bg-white border-[var(--color-primary-teal)]/20 hover:shadow-lg transition-all">
              <CardHeader>
                <div className="w-12 h-12 bg-[var(--color-secondary-yellow)]/20 rounded-lg flex items-center justify-center mb-4">
                  <Bot className="w-6 h-6 text-[var(--color-primary-dark)]" />
                </div>
                <CardTitle className="text-xl font-all-round text-[var(--color-primary-dark)]">
                  Controle Humano Total
                </CardTitle>
              </CardHeader>
              <CardContent>
                <CardDescription className="font-onest text-base">
                  A IA sugere, você aprova. Adicione feedbacks como &ldquo;Seja mais
                  técnico&rdquo; ou &ldquo;Mencione o produto X&rdquo; e o artigo será reescrito
                  antes de publicar.
                </CardDescription>
              </CardContent>
            </Card>

            {/* Feature 4: Standard (WordPress) */}
            <Card className="bg-gradient-to-br from-white to-[var(--color-primary-purple)]/5 border-[var(--color-primary-teal)]/20 hover:shadow-lg transition-all">
              <CardHeader>
                <div className="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center mb-4">
                  <LayoutTemplate className="w-6 h-6 text-blue-600" />
                </div>
                <CardTitle className="text-xl font-all-round text-[var(--color-primary-dark)]">
                  Integração WordPress
                </CardTitle>
              </CardHeader>
              <CardContent>
                <CardDescription className="font-onest text-base">
                  Adeus Copiar e Colar. O conteúdo vai formatado (H1, H2,
                  listas) direto para o seu blog com status de rascunho ou
                  publicado.
                </CardDescription>
              </CardContent>
            </Card>
          </div>
        </div>
      </section>

      {/* --- HOW IT WORKS (ZIG ZAG) --- */}
      <section
        id="how-it-works"
        className="py-20 md:py-32 bg-[var(--color-secondary-cream)]"
      >
        <div className="container mx-auto px-4">
          <div className="text-center mb-16">
            <span className="text-[var(--color-primary-purple)] font-bold font-all-round text-sm tracking-wider uppercase">
              Fluxo de Trabalho
            </span>
            <h2 className="text-3xl md:text-4xl font-bold font-all-round text-[var(--color-primary-dark)] mt-2">
              Do cadastro ao tráfego em 3 passos
            </h2>
          </div>

          <div className="max-w-5xl mx-auto space-y-24">
            {/* Step 1 */}
            <div className="flex flex-col md:flex-row items-center gap-12">
              <div className="md:w-1/2">
                <div className="relative p-8 bg-white rounded-2xl shadow-xl border border-[var(--color-primary-purple)]/10 transform -rotate-2 hover:rotate-0 transition-transform duration-500">
                  <div className="absolute -top-4 -left-4 w-12 h-12 bg-[var(--color-primary-purple)] text-white rounded-xl flex items-center justify-center text-xl font-bold font-all-round shadow-lg">
                    1
                  </div>
                  <div className="space-y-4">
                    <div className="h-4 bg-gray-100 rounded w-1/3"></div>
                    <div className="h-10 bg-gray-50 border border-gray-200 rounded px-3 flex items-center text-sm text-gray-500">
                      clinicadentes.com.br
                    </div>
                    <div className="h-10 bg-gray-50 border border-gray-200 rounded px-3 flex items-center text-sm text-gray-500">
                      Concorrente: odontocompany.com
                    </div>
                    <Button
                      className="w-full bg-[var(--color-primary-purple)]"
                      size="sm"
                    >
                      Analisar Negócio
                    </Button>
                  </div>
                </div>
              </div>
              <div className="md:w-1/2 space-y-4">
                <h3 className="text-2xl font-bold font-all-round text-[var(--color-primary-dark)]">
                  Ensine a IA sobre seu negócio
                </h3>
                <p className="text-lg font-onest text-[var(--color-primary-dark)]/70">
                  Preencha seu site, objetivos (vendas ou leads) e adicione seus
                  concorrentes. O sistema varre essas URLs para entender o tom
                  de voz e os assuntos do momento.
                </p>
                <ul className="space-y-2 mt-4">
                  <CheckListItem>
                    Definição de Localização (País/Estado/Cidade)
                  </CheckListItem>
                  <CheckListItem>
                    Upload de Manual de Marca (Opcional)
                  </CheckListItem>
                </ul>
              </div>
            </div>

            {/* Step 2 */}
            <div className="flex flex-col md:flex-row-reverse items-center gap-12">
              <div className="md:w-1/2">
                <div className="relative p-8 bg-white rounded-2xl shadow-xl border border-[var(--color-secondary-yellow)]/30 transform rotate-2 hover:rotate-0 transition-transform duration-500">
                  <div className="absolute -top-4 -right-4 w-12 h-12 bg-[var(--color-secondary-yellow)] text-[var(--color-primary-dark)] rounded-xl flex items-center justify-center text-xl font-bold font-all-round shadow-lg">
                    2
                  </div>
                  <div className="space-y-3">
                    <div className="p-3 border border-green-200 bg-green-50 rounded-lg flex justify-between items-center">
                      <span className="text-sm font-semibold text-green-900">
                        Implantes: Guia de Preços
                      </span>
                      <Check className="w-4 h-4 text-green-600" />
                    </div>
                    <div className="p-3 border border-gray-200 bg-gray-50 rounded-lg opacity-50 flex justify-between items-center">
                      <span className="text-sm text-gray-500">
                        História da Odontologia
                      </span>
                      <span className="text-xs text-red-400">Rejeitado</span>
                    </div>
                    <div className="text-xs text-[var(--color-primary-purple)] font-bold mt-2">
                      + Adicionar Feedback: &ldquo;Focar em pagamento facilitado&rdquo;
                    </div>
                  </div>
                </div>
              </div>
              <div className="md:w-1/2 space-y-4">
                <h3 className="text-2xl font-bold font-all-round text-[var(--color-primary-dark)]">
                  Curadoria e Feedback
                </h3>
                <p className="text-lg font-onest text-[var(--color-primary-dark)]/70">
                  A IA gera sugestões de títulos baseados em dados, não em
                  palpites. Você aprova o que gosta e pode adicionar feedbacks
                  específicos para direcionar a escrita.
                </p>
                <ul className="space-y-2 mt-4">
                  <CheckListItem>
                    Sugestões baseadas em Search Console
                  </CheckListItem>
                  <CheckListItem>Aprovação em lote</CheckListItem>
                </ul>
              </div>
            </div>

            {/* Step 3 */}
            <div className="flex flex-col md:flex-row items-center gap-12">
              <div className="md:w-1/2">
                <div className="relative p-6 bg-white rounded-2xl shadow-xl border border-[var(--color-primary-teal)]/20 flex flex-col items-center justify-center min-h-[200px] text-center">
                  <div className="absolute -top-4 -left-4 w-12 h-12 bg-[var(--color-primary-teal)] text-white rounded-xl flex items-center justify-center text-xl font-bold font-all-round shadow-lg">
                    3
                  </div>
                  <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mb-4 animate-bounce">
                    <Zap className="w-8 h-8 text-green-600" />
                  </div>
                  <h4 className="font-bold text-[var(--color-primary-dark)]">
                    Publicado com Sucesso!
                  </h4>
                  <p className="text-sm text-gray-500 mt-2">
                    Seu artigo já está indexando no Google.
                  </p>
                </div>
              </div>
              <div className="md:w-1/2 space-y-4">
                <h3 className="text-2xl font-bold font-all-round text-[var(--color-primary-dark)]">
                  Publicação Automática
                </h3>
                <p className="text-lg font-onest text-[var(--color-primary-dark)]/70">
                  Esqueça a formatação manual no WordPress. O artigo vai
                  completo, com headers, negritos e otimizações de SEO on-page,
                  pronto para gerar tráfego.
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* --- PRICING SECTION --- */}
      <section
        id="pricing"
        className="py-20 md:py-32 bg-white relative overflow-hidden"
      >
        {/* Decorative Blob */}
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[800px] bg-[var(--color-primary-purple)]/5 rounded-full blur-3xl" />

        <div className="container mx-auto px-4 relative z-10">
          <div className="text-center mb-16">
            <h2 className="text-3xl md:text-4xl font-bold font-all-round text-[var(--color-primary-dark)] mb-4">
              Investimento simples, retorno orgânico
            </h2>
            <p className="text-lg font-onest text-[var(--color-primary-dark)]/70">
              Escolha o plano ideal para a velocidade de crescimento que você
              deseja.
            </p>
          </div>

          <div className="grid md:grid-cols-3 gap-8 max-w-6xl mx-auto items-start">
            {/* Plan: Starter */}
            <Card className="border-2 border-[var(--color-primary-teal)]/20 hover:border-[var(--color-primary-teal)]/50 transition-all bg-white shadow-sm">
              <CardHeader>
                <CardTitle className="text-2xl font-all-round text-[var(--color-primary-teal)]">
                  Starter
                </CardTitle>
                <div className="mt-4 mb-2">
                  <span className="text-4xl font-bold font-all-round text-[var(--color-primary-dark)]">
                    R$ 49,90
                  </span>
                  <span className="text-gray-500 font-onest">/mês</span>
                </div>
                <CardDescription className="font-onest">
                  Para quem está começando a validar o canal orgânico.
                </CardDescription>
              </CardHeader>
              <CardContent>
                <ul className="space-y-3">
                  <CheckListItem>
                    <strong>5 matérias</strong> por mês
                  </CheckListItem>
                  <CheckListItem>SEO Básico</CheckListItem>
                  <CheckListItem>Integração WordPress</CheckListItem>
                  <CheckListItem>Suporte por Email</CheckListItem>
                </ul>
              </CardContent>
              <CardFooter>
                <Link href="/login" className="w-full">
                  <Button
                    variant="outline"
                    className="w-full border-[var(--color-primary-teal)] text-[var(--color-primary-teal)] hover:bg-[var(--color-primary-teal)]/10 font-bold"
                  >
                    Começar Starter
                  </Button>
                </Link>
              </CardFooter>
            </Card>

            {/* Plan: Pro (Highlighted) */}
            <Card className="border-2 border-[var(--color-primary-purple)] bg-white shadow-xl relative scale-105 z-10">
              <div className="absolute top-0 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-[var(--color-primary-purple)] text-white px-4 py-1 rounded-full text-sm font-bold shadow-md">
                Mais Popular
              </div>
              <CardHeader>
                <CardTitle className="text-2xl font-all-round text-[var(--color-primary-purple)]">
                  Pro
                </CardTitle>
                <div className="mt-4 mb-2">
                  <span className="text-5xl font-bold font-all-round text-[var(--color-primary-dark)]">
                    R$ 99,90
                  </span>
                  <span className="text-gray-500 font-onest">/mês</span>
                </div>
                <CardDescription className="font-onest text-[var(--color-primary-dark)]/80">
                  Ideal para pequenas empresas que precisam de consistência.
                </CardDescription>
              </CardHeader>
              <CardContent>
                <ul className="space-y-3">
                  <CheckListItem>
                    <strong>15 matérias</strong> por mês
                  </CheckListItem>
                  <CheckListItem>
                    <strong>SEO Avançado</strong> (Concorrência)
                  </CheckListItem>
                  <CheckListItem>Múltiplas Localizações</CheckListItem>
                  <CheckListItem>Suporte Prioritário</CheckListItem>
                  <CheckListItem>Análise de Gaps de Conteúdo</CheckListItem>
                </ul>
              </CardContent>
              <CardFooter>
                <Link href="/login" className="w-full">
                  <Button className="w-full bg-[var(--color-primary-purple)] hover:bg-[var(--color-primary-purple)]/90 text-white font-bold h-12 text-lg shadow-lg shadow-purple-200">
                    Escolher Plano Pro
                  </Button>
                </Link>
              </CardFooter>
            </Card>

            {/* Plan: Enterprise */}
            <Card className="border-2 border-[var(--color-primary-dark)]/10 hover:border-[var(--color-primary-dark)]/30 transition-all bg-[var(--color-primary-dark)] text-white">
              <CardHeader>
                <CardTitle className="text-2xl font-all-round text-[var(--color-secondary-yellow)]">
                  Enterprise
                </CardTitle>
                <div className="mt-4 mb-2">
                  <span className="text-4xl font-bold font-all-round text-white">
                    R$ 249,90
                  </span>
                  <span className="text-gray-400 font-onest">/mês</span>
                </div>
                <CardDescription className="font-onest text-gray-300">
                  Para agências e empresas que buscam domínio total do nicho.
                </CardDescription>
              </CardHeader>
              <CardContent className="text-gray-200">
                <ul className="space-y-3">
                  <li className="flex items-start gap-3">
                    <Check className="w-4 h-4 mt-1 text-[var(--color-secondary-yellow)]" />
                    <span className="font-bold text-white">
                      50 matérias por mês
                    </span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="w-4 h-4 mt-1 text-[var(--color-secondary-yellow)]" />
                    <span>SEO Premium + Schema Markup</span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="w-4 h-4 mt-1 text-[var(--color-secondary-yellow)]" />
                    <span>Suporte Dedicado</span>
                  </li>
                  <li className="flex items-start gap-3">
                    <Check className="w-4 h-4 mt-1 text-[var(--color-secondary-yellow)]" />
                    <span>Gestão de Múltiplas Contas</span>
                  </li>
                </ul>
              </CardContent>
              <CardFooter>
                <Link href="/login" className="w-full">
                  <Button
                    variant="secondary"
                    className="w-full bg-[var(--color-secondary-yellow)] text-[var(--color-primary-dark)] hover:bg-[var(--color-secondary-yellow)]/90 font-bold"
                  >
                    Contratar Enterprise
                  </Button>
                </Link>
              </CardFooter>
            </Card>
          </div>
        </div>
      </section>

      {/* --- CTA FINAL --- */}
      <section className="py-20 container mx-auto px-4">
        <div className="bg-[var(--color-primary-dark)] rounded-[2rem] p-8 md:p-16 text-center relative overflow-hidden shadow-2xl">
          {/* Background Accents */}
          <div className="absolute top-0 right-0 w-64 h-64 bg-[var(--color-primary-purple)] rounded-full blur-[80px] opacity-40 translate-x-1/2 -translate-y-1/2" />
          <div className="absolute bottom-0 left-0 w-64 h-64 bg-[var(--color-primary-teal)] rounded-full blur-[80px] opacity-40 -translate-x-1/2 translate-y-1/2" />

          <div className="relative z-10 max-w-3xl mx-auto space-y-8">
            <h2 className="text-3xl md:text-5xl font-bold font-all-round text-white leading-tight">
              Pare de escrever. <br /> Comece a editar.
            </h2>
            <p className="text-lg md:text-xl font-onest text-gray-300">
              Junte-se a empresas inteligentes que usam o organiQ para dominar o
              Google sem gastar horas escrevendo.
            </p>
            <div className="flex flex-col sm:flex-row justify-center gap-4">
              <Link href="/login">
                <Button
                  size="lg"
                  className="w-full sm:w-auto h-16 px-10 text-xl font-bold font-all-round bg-[var(--color-secondary-yellow)] text-[var(--color-primary-dark)] hover:bg-[var(--color-secondary-yellow)]/90 hover:scale-105 transition-all"
                >
                  Criar conta grátis agora
                </Button>
              </Link>
            </div>
            <p className="text-sm text-gray-500 font-onest">
              Não é necessário cartão de crédito para criar a conta.
            </p>
          </div>
        </div>
      </section>

      {/* --- FOOTER --- */}
      <footer className="border-t border-[var(--color-primary-dark)]/10 bg-white">
        <div className="container mx-auto px-4 py-12">
          <div className="grid md:grid-cols-4 gap-8 mb-8">
            <div className="col-span-1 md:col-span-2">
              <div className="flex items-center gap-2 mb-4">
                {/* Simple Text Logo for Footer */}
                <h1 className="text-2xl font-bold font-all-round text-[var(--color-primary-purple)]">
                  organiQ
                </h1>
              </div>
              <p className="font-onest text-[var(--color-primary-dark)]/70 max-w-xs">
                Automação de conteúdo inteligente para empresas que querem
                crescer organicamente.
              </p>
            </div>

            <div>
              <h4 className="font-bold font-all-round text-[var(--color-primary-dark)] mb-4">
                Produto
              </h4>
              <ul className="space-y-2 font-onest text-[var(--color-primary-dark)]/70">
                <li>
                  <Link
                    href="#features"
                    className="hover:text-[var(--color-primary-purple)]"
                  >
                    Funcionalidades
                  </Link>
                </li>
                <li>
                  <Link
                    href="#pricing"
                    className="hover:text-[var(--color-primary-purple)]"
                  >
                    Preços
                  </Link>
                </li>
                <li>
                  <Link
                    href="/login"
                    className="hover:text-[var(--color-primary-purple)]"
                  >
                    Login
                  </Link>
                </li>
              </ul>
            </div>

            <div>
              <h4 className="font-bold font-all-round text-[var(--color-primary-dark)] mb-4">
                Legal
              </h4>
              <ul className="space-y-2 font-onest text-[var(--color-primary-dark)]/70">
                <li>
                  <a
                    href="#"
                    className="hover:text-[var(--color-primary-purple)]"
                  >
                    Termos de Uso
                  </a>
                </li>
                <li>
                  <a
                    href="#"
                    className="hover:text-[var(--color-primary-purple)]"
                  >
                    Privacidade
                  </a>
                </li>
              </ul>
            </div>
          </div>

          <div className="pt-8 border-t border-gray-100 flex flex-col md:flex-row justify-between items-center gap-4">
            <span className="text-sm font-onest text-[var(--color-primary-dark)]/50">
              © 2024 organiQ. Feito com inteligência natural.
            </span>
            <div className="flex gap-4">
              {/* Social placeholders if needed */}
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}
