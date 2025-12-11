"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { registerSchema, type RegisterInput } from "@/lib/validations";
import { useAuth } from "@/hooks/useAuth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { PasswordInput } from "@/components/ui/password-input";
import { PasswordStrengthIndicator } from "@/components/ui/password-strength-indicator";

export function RegisterForm() {
  const { register: registerUser, isRegistering } = useAuth();

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<RegisterInput>({
    resolver: zodResolver(registerSchema),
  });

  const password = watch("password") || "";

  const onSubmit = (data: RegisterInput) => {
    registerUser(data);
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      {/* Nome */}
      <div className="space-y-2">
        <Label htmlFor="register-name" required>
          Nome completo
        </Label>
        <Input
          id="register-name"
          type="text"
          placeholder="Seu nome"
          error={errors.name?.message}
          {...register("name")}
        />
      </div>

      {/* Email */}
      <div className="space-y-2">
        <Label htmlFor="register-email" required>
          Email
        </Label>
        <Input
          id="register-email"
          type="email"
          placeholder="seu@email.com"
          error={errors.email?.message}
          {...register("email")}
        />
      </div>

      {/* Senha */}
      <div className="space-y-2">
        <Label htmlFor="register-password" required>
          Senha
        </Label>
        <PasswordInput
          id="register-password"
          placeholder="••••••••"
          error={errors.password?.message}
          {...register("password")}
        />

        {/* Indicador de Força da Senha */}
        <PasswordStrengthIndicator password={password} />
      </div>

      {/* Confirmação de Senha */}
      <div className="space-y-2">
        <Label htmlFor="register-confirmPassword" required>
          Confirmar senha
        </Label>
        <PasswordInput
          id="register-confirmPassword"
          placeholder="••••••••"
          error={errors.confirmPassword?.message}
          {...register("confirmPassword")}
        />
      </div>

      {/* Disclaimer de Segurança */}
      <div className="bg-[var(--color-primary-purple)]/5 border border-[var(--color-primary-purple)]/20 rounded-[var(--radius-sm)] p-3">
        <p className="text-xs font-onest text-[var(--color-primary-dark)]/70">
          🔒 Sua senha é criptografada e nunca será compartilhada. Use uma senha
          forte e única.
        </p>
      </div>

      {/* Submit Button */}
      <Button
        type="submit"
        variant="secondary"
        className="w-full"
        isLoading={isRegistering}
        disabled={isRegistering}
      >
        {isRegistering ? "Criando conta..." : "Criar conta"}
      </Button>
    </form>
  );
}
