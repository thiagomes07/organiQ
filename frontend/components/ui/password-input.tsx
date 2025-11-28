import * as React from "react";
import { Eye, EyeOff } from "lucide-react";
import { cn } from "@/lib/utils";

export interface PasswordInputProps
  extends Omit<React.InputHTMLAttributes<HTMLInputElement>, "type"> {
  error?: string;
}

const PasswordInput = React.forwardRef<HTMLInputElement, PasswordInputProps>(
  ({ className, error, ...props }, ref) => {
    const [showPassword, setShowPassword] = React.useState(false);

    return (
      <div className="w-full">
        <div className="relative">
          <input
            type={showPassword ? "text" : "password"}
            className={cn(
              "flex h-10 w-full rounded-[var(--radius-sm)] border border-input bg-white px-3 py-2 pr-10 text-sm font-onest",
              "transition-colors duration-200",
              "placeholder:text-[var(--color-primary-dark)]/40",
              "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--color-primary-purple)] focus-visible:border-transparent",
              "disabled:cursor-not-allowed disabled:opacity-50",
              error &&
                "border-[var(--color-error)] focus-visible:ring-[var(--color-error)]",
              className
            )}
            ref={ref}
            {...props}
          />
          <button
            type="button"
            onClick={() => setShowPassword(!showPassword)}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-[var(--color-primary-dark)]/60 hover:text-[var(--color-primary-dark)] transition-colors"
            tabIndex={-1}
            aria-label={showPassword ? "Ocultar senha" : "Mostrar senha"}
          >
            {showPassword ? (
              <EyeOff className="h-4 w-4" />
            ) : (
              <Eye className="h-4 w-4" />
            )}
          </button>
        </div>
        {error && (
          <p className="mt-1 text-xs text-[var(--color-error)] font-onest">
            {error}
          </p>
        )}
      </div>
    );
  }
);
PasswordInput.displayName = "PasswordInput";

export { PasswordInput };
