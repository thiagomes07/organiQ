import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Sparkles, TrendingUp, Target, Zap } from "lucide-react";
import { useNavigate } from "react-router-dom";

const Landing = () => {
  const navigate = useNavigate();

  const features = [
    {
      icon: Sparkles,
      title: "Conteúdo com IA",
      description: "Artigos otimizados para SEO gerados automaticamente com inteligência artificial"
    },
    {
      icon: TrendingUp,
      title: "Aumente o Tráfego",
      description: "Atraia mais visitantes qualificados com conteúdo estratégico e relevante"
    },
    {
      icon: Target,
      title: "Autoridade Digital",
      description: "Construa credibilidade e destaque-se como referência no seu mercado"
    }
  ];

  return (
    <div className="min-h-screen gradient-subtle">
      {/* Header */}
      <header className="border-b border-border/50 bg-card/50 backdrop-blur-sm sticky top-0 z-50">
        <div className="container mx-auto px-6 h-16 flex items-center">
          <div className="flex items-center gap-2">
            <Zap className="h-6 w-6 text-primary" />
            <span className="text-2xl font-bold text-primary-dark">organiQ</span>
          </div>
        </div>
      </header>

      {/* Hero Section */}
      <section className="container mx-auto px-6 py-20 md:py-32">
        <div className="max-w-4xl mx-auto text-center">
          <div className="inline-block mb-6 px-4 py-2 bg-card rounded-full shadow-soft border border-border/50">
            <span className="text-sm font-medium text-accent">Naturalmente Inteligente</span>
          </div>
          
          <h1 className="mb-6 text-balance">
            Gere Conteúdo de Qualidade para Seu Blog em Minutos
          </h1>
          
          <p className="text-xl md:text-2xl text-accent mb-12 text-balance max-w-2xl mx-auto">
            Transforme seu site em uma máquina de gerar leads com artigos otimizados para SEO, 
            criados por inteligência artificial
          </p>

          <Button 
            size="lg"
            onClick={() => navigate('/login')}
            className="bg-secondary text-secondary-foreground hover:bg-secondary/90 hover:scale-105 hover:shadow-elevated text-lg px-8 py-6 h-auto font-semibold"
          >
            Criar minha conta
          </Button>

          <p className="mt-6 text-sm text-muted-foreground">
            Comece gratuitamente • Sem cartão de crédito
          </p>
        </div>
      </section>

      {/* Features Grid */}
      <section className="container mx-auto px-6 pb-20 md:pb-32">
        <div className="grid md:grid-cols-3 gap-8 max-w-6xl mx-auto">
          {features.map((feature, index) => {
            const Icon = feature.icon;
            return (
              <Card 
                key={index}
                className="border-border/50 shadow-card hover:shadow-elevated bg-card group hover:-translate-y-1"
              >
                <CardContent className="p-8">
                  <div className="w-12 h-12 rounded-lg bg-gradient-accent flex items-center justify-center mb-6 group-hover:scale-110">
                    <Icon className="h-6 w-6 text-primary-foreground" />
                  </div>
                  <h3 className="mb-3 text-xl font-semibold">{feature.title}</h3>
                  <p className="text-muted-foreground leading-relaxed">
                    {feature.description}
                  </p>
                </CardContent>
              </Card>
            );
          })}
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-border/50 bg-card/50 backdrop-blur-sm">
        <div className="container mx-auto px-6 py-8">
          <div className="flex items-center justify-center gap-2">
            <Zap className="h-5 w-5 text-primary" />
            <span className="text-lg font-bold text-primary-dark">organiQ</span>
            <span className="text-muted-foreground">• 2024</span>
          </div>
        </div>
      </footer>
    </div>
  );
};

export default Landing;
