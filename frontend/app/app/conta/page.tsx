"use client";

import { useRouter } from "next/navigation";
import { useState, useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { HelpCircle, Check, Calendar, Eye, EyeOff } from "lucide-react";
import {
  profileUpdateSchema,
  passwordUpdateSchema,
  type ProfileUpdateInput,
  type PasswordUpdateInput,
} from "@/lib/validations";
import { usePlans } from "@/hooks/usePlans";
import { useUser, useAuthStore } from "@/store/authStore";
import { formatDate, formatCurrency } from "@/lib/utils";
import api, { getErrorMessage } from "@/lib/axios";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { toast } from "sonner";
import { cn } from "@/lib/utils";

export default function ContaPage() {
  const router = useRouter();
  const user = useUser();
  const { updateUser } = useAuthStore();
  const { currentPlan, openPortal, isOpeningPortal } = usePlans();
  const [isSavingProfile, setIsSavingProfile] = useState(false);
  const [isSavingPassword, setIsSavingPassword] = useState(false);
  const [showCurrentPassword, setShowCurrentPassword] = useState(false);
  const [showNewPassword, setShowNewPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  // Profile Form
  const profileForm = useForm<ProfileUpdateInput>({
    resolver: zodResolver(profileUpdateSchema),
    defaultValues: {
      name: user?.name || "",
    },
  });

  // Update form when user changes
  useEffect(() => {
    if (user?.name) {
      profileForm.setValue("name", user.name);
    }
  }, [user?.name, profileForm]);

  const handleUpdateProfile = async (data: ProfileUpdateInput) => {
    setIsSavingProfile(true);
    try {
      await api.patch("/account/profile", data);
      updateUser({ name: data.name });
      toast.success("Perfil atualizado com sucesso!");
    } catch (error) {
      const message = getErrorMessage(error);
      toast.error(message || "Erro ao atualizar perfil");
    } finally {
      setIsSavingProfile(false);
    }
  };

  // Password Form
  const passwordForm = useForm<PasswordUpdateInput>({
    resolver: zodResolver(passwordUpdateSchema),
    defaultValues: {
      currentPassword: "",
      newPassword: "",
      confirmPassword: "",
    },
  });

  const handleUpdatePassword = async (data: PasswordUpdateInput) => {
    setIsSavingPassword(true);
    try {
      await api.patch("/account/password", data);
      toast.success("Senha atualizada com sucesso!");
      passwordForm.reset();
    } catch (error) {
      const message = getErrorMessage(error);
      toast.error(message || "Erro ao atualizar senha");
    } finally {
      setIsSavingPassword(false);
    }
  };



  const usagePercentage = user
    ? (user.articlesUsed / user.maxArticles) * 100
    : 0;

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
                  {...profileForm.register("name")}
                />
              </div>

              {/* Email (disabled) */}
              <div className="space-y-2">
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  type="email"
                  value={user?.email || ""}
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

        {/* Card 3: Alterar Senha */}
        <Card>
          <CardHeader>
            <CardTitle>Alterar Senha</CardTitle>
            <CardDescription>Atualize sua senha de acesso</CardDescription>
          </CardHeader>

          <form onSubmit={passwordForm.handleSubmit(handleUpdatePassword)}>
            <CardContent className="space-y-4">
              {/* Senha Atual */}
              <div className="space-y-2">
                <Label htmlFor="currentPassword" required>
                  Senha Atual
                </Label>
                <div className="relative">
                  <Input
                    id="currentPassword"
                    type={showCurrentPassword ? "text" : "password"}
                    placeholder="••••••••"
                    error={passwordForm.formState.errors.currentPassword?.message}
                    {...passwordForm.register("currentPassword")}
                  />
                  <button
                    type="button"
                    onClick={() => setShowCurrentPassword(!showCurrentPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-[var(--color-primary-dark)]/40 hover:text-[var(--color-primary-dark)]/70 transition-colors"
                  >
                    {showCurrentPassword ? (
                      <EyeOff className="h-4 w-4" />
                    ) : (
                      <Eye className="h-4 w-4" />
                    )}
                  </button>
                </div>
              </div>

              {/* Nova Senha */}
              <div className="space-y-2">
                <Label htmlFor="newPassword" required>
                  Nova Senha
                </Label>
                <div className="relative">
                  <Input
                    id="newPassword"
                    type={showNewPassword ? "text" : "password"}
                    placeholder="••••••••"
                    error={passwordForm.formState.errors.newPassword?.message}
                    {...passwordForm.register("newPassword")}
                  />
                  <button
                    type="button"
                    onClick={() => setShowNewPassword(!showNewPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-[var(--color-primary-dark)]/40 hover:text-[var(--color-primary-dark)]/70 transition-colors"
                  >
                    {showNewPassword ? (
                      <EyeOff className="h-4 w-4" />
                    ) : (
                      <Eye className="h-4 w-4" />
                    )}
                  </button>
                </div>
              </div>

              {/* Confirmar Senha */}
              <div className="space-y-2">
                <Label htmlFor="confirmPassword" required>
                  Confirmar Nova Senha
                </Label>
                <div className="relative">
                  <Input
                    id="confirmPassword"
                    type={showConfirmPassword ? "text" : "password"}
                    placeholder="••••••••"
                    error={passwordForm.formState.errors.confirmPassword?.message}
                    {...passwordForm.register("confirmPassword")}
                  />
                  <button
                    type="button"
                    onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-[var(--color-primary-dark)]/40 hover:text-[var(--color-primary-dark)]/70 transition-colors"
                  >
                    {showConfirmPassword ? (
                      <EyeOff className="h-4 w-4" />
                    ) : (
                      <Eye className="h-4 w-4" />
                    )}
                  </button>
                </div>
              </div>
            </CardContent>

            <CardFooter>
              <Button
                type="submit"
                variant="outline"
                className="w-full sm:w-auto"
                isLoading={isSavingPassword}
                disabled={isSavingPassword}
              >
                Atualizar Senha
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
                {user && user.maxArticles - user.articlesUsed} matérias
                restantes este mês
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
            <Button
              variant="primary"
              onClick={() => router.push('/app/planos')}
            >
              Fazer Upgrade
            </Button>
          </CardFooter>
        </Card>
      </div>
    </div>
  );
}
