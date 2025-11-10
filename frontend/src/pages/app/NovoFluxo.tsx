import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Checkbox } from "@/components/ui/checkbox";
import { Slider } from "@/components/ui/slider";
import { Plus, X } from "lucide-react";
import { useToast } from "@/hooks/use-toast";

const businessSchema = z.object({
  description: z.string().min(10, "Descrição deve ter no mínimo 10 caracteres"),
  objective: z.string().min(1, "Selecione um objetivo"),
  siteUrl: z.string().url("URL inválida").optional().or(z.literal("")),
  hasBlog: z.boolean(),
  blogUrls: z.array(z.string().url("URL inválida")),
  articleCount: z.number().min(1).max(50),
});

type BusinessData = z.infer<typeof businessSchema>;

const NovoFluxo = () => {
  const { toast } = useToast();
  const [currentStep, setCurrentStep] = useState(1);
  const [blogUrls, setBlogUrls] = useState<string[]>([""]);
  const [competitorUrls, setCompetitorUrls] = useState<string[]>([""]);
  const [hasBlog, setHasBlog] = useState(false);
  const [articleCount, setArticleCount] = useState([5]);

  const form = useForm<BusinessData>({
    resolver: zodResolver(businessSchema),
    defaultValues: {
      description: "",
      objective: "",
      siteUrl: "",
      hasBlog: false,
      blogUrls: [],
      articleCount: 5,
    },
  });

  const addBlogUrl = () => {
    setBlogUrls([...blogUrls, ""]);
  };

  const removeBlogUrl = (index: number) => {
    const newUrls = blogUrls.filter((_, i) => i !== index);
    setBlogUrls(newUrls.length === 0 ? [""] : newUrls);
  };

  const updateBlogUrl = (index: number, value: string) => {
    const newUrls = [...blogUrls];
    newUrls[index] = value;
    setBlogUrls(newUrls);
  };

  const addCompetitorUrl = () => {
    setCompetitorUrls([...competitorUrls, ""]);
  };

  const removeCompetitorUrl = (index: number) => {
    const newUrls = competitorUrls.filter((_, i) => i !== index);
    setCompetitorUrls(newUrls.length === 0 ? [""] : newUrls);
  };

  const updateCompetitorUrl = (index: number, value: string) => {
    const newUrls = [...competitorUrls];
    newUrls[index] = value;
    setCompetitorUrls(newUrls);
  };

  const onSubmitStep1 = async (data: BusinessData) => {
    try {
      console.log("Business data:", data);
      toast({
        title: "Dados salvos!",
        description: "Prosseguindo para análise de concorrentes...",
      });
      setCurrentStep(2);
    } catch (error) {
      toast({
        title: "Erro ao salvar",
        description: "Tente novamente mais tarde",
        variant: "destructive",
      });
    }
  };

  return (
    <div className="max-w-4xl mx-auto">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Criar Novo Fluxo</h1>
        <p className="text-muted-foreground">
          Vamos começar entendendo o seu negócio
        </p>
      </div>

      {/* Stepper */}
      <div className="flex items-center justify-between mb-12">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <div className={`w-10 h-10 rounded-full flex items-center justify-center font-semibold shadow-soft ${
              currentStep >= 1 ? "bg-primary text-primary-foreground" : "bg-muted text-muted-foreground"
            }`}>
              1
            </div>
            <span className={currentStep >= 1 ? "font-medium text-primary" : "text-muted-foreground"}>Negócio</span>
          </div>
          <div className={`w-16 h-0.5 ${currentStep >= 2 ? "bg-primary" : "bg-border"}`} />
          <div className="flex items-center gap-2">
            <div className={`w-10 h-10 rounded-full flex items-center justify-center font-semibold ${
              currentStep >= 2 ? "bg-primary text-primary-foreground shadow-soft" : "bg-muted text-muted-foreground"
            }`}>
              2
            </div>
            <span className={currentStep >= 2 ? "font-medium text-primary" : "text-muted-foreground"}>Concorrentes</span>
          </div>
          <div className={`w-16 h-0.5 ${currentStep >= 3 ? "bg-primary" : "bg-border"}`} />
          <div className="flex items-center gap-2">
            <div className={`w-10 h-10 rounded-full flex items-center justify-center font-semibold ${
              currentStep >= 3 ? "bg-primary text-primary-foreground shadow-soft" : "bg-muted text-muted-foreground"
            }`}>
              3
            </div>
            <span className={currentStep >= 3 ? "font-medium text-primary" : "text-muted-foreground"}>Integração</span>
          </div>
          <div className={`w-16 h-0.5 ${currentStep >= 4 ? "bg-primary" : "bg-border"}`} />
          <div className="flex items-center gap-2">
            <div className={`w-10 h-10 rounded-full flex items-center justify-center font-semibold ${
              currentStep >= 4 ? "bg-primary text-primary-foreground shadow-soft" : "bg-muted text-muted-foreground"
            }`}>
              4
            </div>
            <span className={currentStep >= 4 ? "font-medium text-primary" : "text-muted-foreground"}>Aprovação</span>
          </div>
        </div>
      </div>

      {currentStep === 1 && (
        <Card className="shadow-card border-border/50">
          <CardHeader>
            <CardTitle>Informações do Negócio</CardTitle>
            <CardDescription>
              Conte-nos sobre sua empresa e objetivos
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={form.handleSubmit(onSubmitStep1)} className="space-y-6">
            <div className="space-y-2">
              <Label htmlFor="description">Descreva seu negócio em uma frase *</Label>
              <Textarea
                id="description"
                placeholder="Ex: Somos uma agência de marketing digital especializada em empresas de tecnologia"
                className="min-h-[100px] resize-none"
                {...form.register("description")}
              />
              {form.formState.errors.description && (
                <p className="text-sm text-destructive">
                  {form.formState.errors.description.message}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="objective">Qual é o seu principal objetivo? *</Label>
              <Select onValueChange={(value) => form.setValue("objective", value)}>
                <SelectTrigger>
                  <SelectValue placeholder="Selecione um objetivo" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="leads">Gerar mais leads</SelectItem>
                  <SelectItem value="vendas">Vender mais online</SelectItem>
                  <SelectItem value="marca">Aumentar reconhecimento da marca</SelectItem>
                </SelectContent>
              </Select>
              {form.formState.errors.objective && (
                <p className="text-sm text-destructive">
                  {form.formState.errors.objective.message}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="siteUrl">URL do seu site principal</Label>
              <Input
                id="siteUrl"
                type="url"
                placeholder="https://seusite.com.br"
                {...form.register("siteUrl")}
              />
              {form.formState.errors.siteUrl && (
                <p className="text-sm text-destructive">
                  {form.formState.errors.siteUrl.message}
                </p>
              )}
            </div>

            <div className="space-y-4">
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="hasBlog"
                  checked={hasBlog}
                  onCheckedChange={(checked) => {
                    setHasBlog(checked as boolean);
                    form.setValue("hasBlog", checked as boolean);
                  }}
                />
                <Label htmlFor="hasBlog" className="cursor-pointer">
                  Você já tem um blog?
                </Label>
              </div>

              {hasBlog && (
                <div className="space-y-3 pl-6">
                  {blogUrls.map((url, index) => (
                    <div key={index} className="flex gap-2">
                      <Input
                        type="url"
                        placeholder="https://seublog.com.br"
                        value={url}
                        onChange={(e) => updateBlogUrl(index, e.target.value)}
                      />
                      {blogUrls.length > 1 && (
                        <Button
                          type="button"
                          variant="outline"
                          size="icon"
                          onClick={() => removeBlogUrl(index)}
                        >
                          <X className="h-4 w-4" />
                        </Button>
                      )}
                    </div>
                  ))}
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={addBlogUrl}
                    className="w-full"
                  >
                    <Plus className="h-4 w-4 mr-2" />
                    Adicionar link do blog
                  </Button>
                </div>
              )}
            </div>

            <div className="space-y-4">
              <Label>Quantidade de matérias desejadas</Label>
              <div className="space-y-2">
                <Slider
                  value={articleCount}
                  onValueChange={(value) => {
                    setArticleCount(value);
                    form.setValue("articleCount", value[0]);
                  }}
                  min={1}
                  max={50}
                  step={1}
                  className="w-full"
                />
                <div className="flex justify-between text-sm text-muted-foreground">
                  <span>1 matéria</span>
                  <span className="font-semibold text-primary">{articleCount[0]} matérias</span>
                  <span>50 matérias</span>
                </div>
              </div>
            </div>

            <div className="flex justify-end pt-6">
              <Button
                type="submit"
                size="lg"
                className="bg-secondary text-secondary-foreground hover:bg-secondary/90 hover:scale-105"
              >
                Próximo
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
      )}

      {currentStep === 2 && (
        <Card className="shadow-card border-border/50">
          <CardHeader>
            <CardTitle>Análise de Concorrentes</CardTitle>
            <CardDescription>
              Adicione URLs de concorrentes para melhorar sua estratégia de SEO
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <p className="text-sm text-muted-foreground">
                Adicionar concorrentes ajuda a melhorar a estratégia de SEO
              </p>
              <div className="space-y-3">
                {competitorUrls.map((url, index) => (
                  <div key={index} className="flex gap-2">
                    <Input
                      type="url"
                      placeholder={`https://concorrente${index + 1}.com.br`}
                      value={url}
                      onChange={(e) => updateCompetitorUrl(index, e.target.value)}
                    />
                    {competitorUrls.length > 1 && (
                      <Button
                        type="button"
                        variant="outline"
                        size="icon"
                        onClick={() => removeCompetitorUrl(index)}
                      >
                        <X className="h-4 w-4" />
                      </Button>
                    )}
                  </div>
                ))}
              </div>
              <Button
                type="button"
                variant="outline"
                size="sm"
                className="w-full"
                onClick={addCompetitorUrl}
              >
                <Plus className="h-4 w-4 mr-2" />
                Adicionar concorrente
              </Button>
              <div className="flex justify-between pt-6">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setCurrentStep(1)}
                >
                  Voltar
                </Button>
                <Button
                  type="button"
                  className="bg-secondary text-secondary-foreground hover:bg-secondary/90 hover:scale-105"
                  onClick={() => setCurrentStep(3)}
                >
                  Próximo
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {currentStep === 3 && (
        <Card className="shadow-card border-border/50">
          <CardHeader>
            <CardTitle>Integração WordPress</CardTitle>
            <CardDescription>
              Configure as credenciais do seu WordPress
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <p className="text-sm text-muted-foreground mb-4">
                Em desenvolvimento - Próxima etapa será implementada em breve
              </p>
              <div className="flex justify-between pt-6">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setCurrentStep(2)}
                >
                  Voltar
                </Button>
                <Button
                  type="button"
                  className="bg-secondary text-secondary-foreground hover:bg-secondary/90 hover:scale-105"
                  onClick={() => setCurrentStep(4)}
                >
                  Próximo
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {currentStep === 4 && (
        <Card className="shadow-card border-border/50">
          <CardHeader>
            <CardTitle>Aprovação das Matérias</CardTitle>
            <CardDescription>
              Revise e aprove as matérias geradas
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <p className="text-sm text-muted-foreground mb-4">
                Em desenvolvimento - Esta etapa será implementada em breve
              </p>
              <div className="flex justify-between pt-6">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setCurrentStep(3)}
                >
                  Voltar
                </Button>
                <Button
                  type="button"
                  className="bg-secondary text-secondary-foreground hover:bg-secondary/90 hover:scale-105"
                >
                  Publicar
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
};

export default NovoFluxo;
