import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useToast } from "@/hooks/use-toast";

const profileSchema = z.object({
  name: z.string().min(2, "Nome deve ter no mínimo 2 caracteres"),
});

const wordpressSchema = z.object({
  siteUrl: z.string().url("URL inválida"),
  username: z.string().min(1, "Nome de usuário é obrigatório"),
  appPassword: z.string().min(1, "Senha de aplicativo é obrigatória"),
});

type ProfileData = z.infer<typeof profileSchema>;
type WordPressData = z.infer<typeof wordpressSchema>;

const Conta = () => {
  const { toast } = useToast();
  const [isLoadingProfile, setIsLoadingProfile] = useState(false);
  const [isLoadingWordPress, setIsLoadingWordPress] = useState(false);

  const profileForm = useForm<ProfileData>({
    resolver: zodResolver(profileSchema),
    defaultValues: {
      name: "",
    },
  });

  const wordpressForm = useForm<WordPressData>({
    resolver: zodResolver(wordpressSchema),
    defaultValues: {
      siteUrl: "",
      username: "",
      appPassword: "",
    },
  });

  const onUpdateProfile = async (data: ProfileData) => {
    setIsLoadingProfile(true);
    try {
      // TODO: Implement API call
      console.log("Update profile:", data);
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      toast({
        title: "Perfil atualizado!",
        description: "Suas informações foram salvas com sucesso",
      });
    } catch (error) {
      toast({
        title: "Erro ao atualizar",
        description: "Tente novamente mais tarde",
        variant: "destructive",
      });
    } finally {
      setIsLoadingProfile(false);
    }
  };

  const onUpdateWordPress = async (data: WordPressData) => {
    setIsLoadingWordPress(true);
    try {
      // TODO: Implement API call
      console.log("Update WordPress:", data);
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      toast({
        title: "Credenciais atualizadas!",
        description: "Integração com WordPress configurada",
      });
    } catch (error) {
      toast({
        title: "Erro ao atualizar",
        description: "Verifique as credenciais e tente novamente",
        variant: "destructive",
      });
    } finally {
      setIsLoadingWordPress(false);
    }
  };

  return (
    <div className="max-w-4xl mx-auto">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Minha Conta</h1>
        <p className="text-muted-foreground">
          Gerencie suas informações e integrações
        </p>
      </div>

      <div className="space-y-6">
        {/* Profile Card */}
        <Card className="shadow-card border-border/50">
          <CardHeader>
            <CardTitle>Perfil</CardTitle>
            <CardDescription>Atualize suas informações pessoais</CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={profileForm.handleSubmit(onUpdateProfile)} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="name">Nome</Label>
                <Input
                  id="name"
                  type="text"
                  placeholder="Seu nome"
                  {...profileForm.register("name")}
                />
                {profileForm.formState.errors.name && (
                  <p className="text-sm text-destructive">
                    {profileForm.formState.errors.name.message}
                  </p>
                )}
              </div>

              <Button
                type="submit"
                disabled={isLoadingProfile}
                className="bg-primary text-primary-foreground hover:bg-primary/90"
              >
                {isLoadingProfile ? "Salvando..." : "Salvar Alterações"}
              </Button>
            </form>
          </CardContent>
        </Card>

        {/* WordPress Integration Card */}
        <Card className="shadow-card border-border/50">
          <CardHeader>
            <CardTitle>Integração WordPress</CardTitle>
            <CardDescription>Configure suas credenciais do WordPress</CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={wordpressForm.handleSubmit(onUpdateWordPress)} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="siteUrl">URL do site</Label>
                <Input
                  id="siteUrl"
                  type="url"
                  placeholder="https://seusite.com.br"
                  {...wordpressForm.register("siteUrl")}
                />
                {wordpressForm.formState.errors.siteUrl && (
                  <p className="text-sm text-destructive">
                    {wordpressForm.formState.errors.siteUrl.message}
                  </p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="username">Nome de usuário</Label>
                <Input
                  id="username"
                  type="text"
                  placeholder="admin"
                  {...wordpressForm.register("username")}
                />
                {wordpressForm.formState.errors.username && (
                  <p className="text-sm text-destructive">
                    {wordpressForm.formState.errors.username.message}
                  </p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="appPassword">Senha de Aplicativo</Label>
                <Input
                  id="appPassword"
                  type="password"
                  placeholder="••••••••••••••••"
                  {...wordpressForm.register("appPassword")}
                />
                {wordpressForm.formState.errors.appPassword && (
                  <p className="text-sm text-destructive">
                    {wordpressForm.formState.errors.appPassword.message}
                  </p>
                )}
                <p className="text-sm text-muted-foreground">
                  Gere uma senha de aplicativo no painel do WordPress em Usuários → Perfil
                </p>
              </div>

              <Button
                type="submit"
                disabled={isLoadingWordPress}
                className="bg-primary text-primary-foreground hover:bg-primary/90"
              >
                {isLoadingWordPress ? "Atualizando..." : "Atualizar Credenciais"}
              </Button>
            </form>
          </CardContent>
        </Card>

        {/* Plan Card (Disabled) */}
        <Card className="shadow-card border-border/50 opacity-50">
          <CardHeader>
            <CardTitle>Meu Plano</CardTitle>
            <CardDescription>Informações sobre sua assinatura</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <p className="font-semibold text-lg">Plano Atual: Starter</p>
              <p className="text-muted-foreground">Matérias mensais: 5 de 10</p>
            </div>
            <Button disabled className="cursor-not-allowed">
              Gerenciar Assinatura
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

export default Conta;
