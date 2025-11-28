import { Check, X } from "lucide-react";
import { cn } from "@/lib/utils";

interface PasswordStrengthIndicatorProps {
  password: string;
  showRules?: boolean;
}

interface PasswordRule {
  id: string;
  label: string;
  test: (password: string) => boolean;
}

const passwordRules: PasswordRule[] = [
  {
    id: "minLength",
    label: "Mínimo de 8 caracteres",
    test: (pwd) => pwd.length >= 8,
  },
  {
    id: "hasUpperCase",
    label: "Pelo menos uma letra maiúscula",
    test: (pwd) => /[A-Z]/.test(pwd),
  },
  {
    id: "hasLowerCase",
    label: "Pelo menos uma letra minúscula",
    test: (pwd) => /[a-z]/.test(pwd),
  },
  {
    id: "hasNumber",
    label: "Pelo menos um número",
    test: (pwd) => /[0-9]/.test(pwd),
  },
];

export function getPasswordStrength(password: string): {
  strength: "weak" | "medium" | "strong";
  score: number;
  color: string;
} {
  if (!password) {
    return { strength: "weak", score: 0, color: "bg-gray-300" };
  }

  const score = passwordRules.filter((rule) => rule.test(password)).length;

  if (score <= 1) {
    return { strength: "weak", score, color: "bg-[var(--color-error)]" };
  } else if (score <= 3) {
    return { strength: "medium", score, color: "bg-[var(--color-warning)]" };
  } else {
    return { strength: "strong", score, color: "bg-[var(--color-success)]" };
  }
}

export function PasswordStrengthIndicator({
  password,
  showRules = true,
}: PasswordStrengthIndicatorProps) {
  const { strength, score, color } = getPasswordStrength(password);

  if (!password && !showRules) return null;

  return (
    <div className="space-y-3">
      {/* Barra de Força */}
      {password && (
        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium font-onest text-[var(--color-primary-dark)]/70">
              Força da senha
            </span>
            <span
              className={cn(
                "text-xs font-semibold font-onest capitalize",
                strength === "weak" && "text-[var(--color-error)]",
                strength === "medium" && "text-[var(--color-warning)]",
                strength === "strong" && "text-[var(--color-success)]"
              )}
            >
              {strength === "weak" && "Fraca"}
              {strength === "medium" && "Média"}
              {strength === "strong" && "Forte"}
            </span>
          </div>
          <div className="flex gap-1">
            {[1, 2, 3, 4].map((level) => (
              <div
                key={level}
                className={cn(
                  "h-1.5 flex-1 rounded-full transition-all duration-300",
                  level <= score ? color : "bg-gray-200"
                )}
              />
            ))}
          </div>
        </div>
      )}

      {/* Regras de Validação */}
      {showRules && (
        <div className="space-y-1.5">
          {passwordRules.map((rule) => {
            const isValid = rule.test(password);
            return (
              <div
                key={rule.id}
                className="flex items-center gap-2 text-xs font-onest"
              >
                <div
                  className={cn(
                    "flex items-center justify-center h-4 w-4 rounded-full transition-colors",
                    isValid
                      ? "bg-[var(--color-success)] text-white"
                      : "bg-gray-200 text-gray-400"
                  )}
                >
                  {isValid ? (
                    <Check className="h-3 w-3" />
                  ) : (
                    <X className="h-3 w-3" />
                  )}
                </div>
                <span
                  className={cn(
                    "transition-colors",
                    isValid
                      ? "text-[var(--color-primary-dark)]"
                      : "text-[var(--color-primary-dark)]/60"
                  )}
                >
                  {rule.label}
                </span>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
