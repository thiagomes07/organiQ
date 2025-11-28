'use client'

import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { loginSchema, type LoginInput } from '@/lib/validations'
import { useAuth } from '@/hooks/useAuth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { PasswordInput } from '@/components/ui/password-input'

export function LoginForm() {
  const { login, isLoggingIn } = useAuth()

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginInput>({
    resolver: zodResolver(loginSchema),
  })

  const onSubmit = (data: LoginInput) => {
    login(data)
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      {/* Email */}
      <div className="space-y-2">
        <Label htmlFor="login-email" required>
          Email
        </Label>
        <Input
          id="login-email"
          type="email"
          placeholder="seu@email.com"
          error={errors.email?.message}
          {...register('email')}
        />
      </div>

      {/* Senha */}
      <div className="space-y-2">
        <Label htmlFor="login-password" required>
          Senha
        </Label>
        <PasswordInput
          id="login-password"
          placeholder="••••••••"
          error={errors.password?.message}
          {...register('password')}
        />
      </div>

      {/* Link Esqueci Senha */}
      <div className="flex justify-end">
        <button
          type="button"
          disabled
          className="text-sm text-[var(--color-primary-teal)] opacity-50 cursor-not-allowed"
        >
          Esqueci minha senha
        </button>
      </div>

      {/* Submit Button */}
      <Button
        type="submit"
        variant="primary"
        className="w-full"
        isLoading={isLoggingIn}
        disabled={isLoggingIn}
      >
        {isLoggingIn ? 'Entrando...' : 'Entrar'}
      </Button>
    </form>
  )
}