"use client";

import { useForm, useFieldArray } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Plus, X } from "lucide-react";
import { competitorsSchema, type CompetitorsInput } from "@/lib/validations";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

interface CompetitorsFormProps {
  onSubmit: (data: CompetitorsInput) => void;
  onBack: () => void;
  isLoading?: boolean;
  defaultValues?: Partial<CompetitorsInput>;
}

export function CompetitorsForm({
  onSubmit,
  onBack,
  isLoading,
  defaultValues,
}: CompetitorsFormProps) {
  const {
    register,
    handleSubmit,
    control,
    formState: { errors },
  } = useForm<CompetitorsInput>({
    resolver: zodResolver(competitorsSchema),
    defaultValues: {
      competitorUrls: defaultValues?.competitorUrls || [],
    },
  });

  const { fields, append, remove } = useFieldArray({
    control,
    name: "competitorUrls",
  });

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {/* Header */}
      <div className="space-y-2">
        <h3 className="text-xl font-semibold font-all-round text-[var(--color-primary-dark)]">
          Concorrentes (Opcional)
        </h3>
        <p className="text-sm font-onest text-[var(--color-primary-dark)]/70">
          Adicione URLs de concorrentes para criar uma estratégia de SEO mais
          competitiva. Esta etapa é opcional, mas recomendada.
        </p>
      </div>

      {/* Lista de URLs */}
      <div className="space-y-3">
        {fields.length === 0 ? (
          <div className="text-center py-8 px-4 border-2 border-dashed border-[var(--color-border)] rounded-[var(--radius-md)]">
            <p className="text-sm font-onest text-[var(--color-primary-dark)]/60 mb-4">
              Nenhum concorrente adicionado ainda
            </p>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => append("")}
            >
              <Plus className="h-4 w-4 mr-2" />
              Adicionar primeiro concorrente
            </Button>
          </div>
        ) : (
          <>
            {fields.map((field, index) => (
              <div key={field.id} className="space-y-2">
                <div className="flex items-start gap-2">
                  <div className="flex-1">
                    <Label htmlFor={`competitor-${index}`}>
                      Concorrente {index + 1}
                    </Label>
                    <div className="mt-1 flex gap-2">
                      <Input
                        id={`competitor-${index}`}
                        type="url"
                        placeholder="https://concorrente.com.br"
                        error={errors.competitorUrls?.[index]?.message}
                        {...register(`competitorUrls.${index}` as const)}
                      />
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        onClick={() => remove(index)}
                        className="flex-shrink-0"
                      >
                        <X className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                </div>
              </div>
            ))}

            {/* Add More Button */}
            {fields.length < 10 && (
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => append("")}
                className="w-full"
              >
                <Plus className="h-4 w-4 mr-2" />
                Adicionar concorrente ({fields.length}/10)
              </Button>
            )}

            {fields.length >= 10 && (
              <p className="text-xs text-[var(--color-warning)] font-onest text-center">
                Limite de 10 concorrentes atingido
              </p>
            )}
          </>
        )}
      </div>

      {/* Info Box */}
      <div className="bg-[var(--color-primary-purple)]/5 border border-[var(--color-primary-purple)]/20 rounded-[var(--radius-md)] p-4">
        <p className="text-sm font-onest text-[var(--color-primary-dark)]/80">
          💡 <strong>Dica:</strong> Adicione sites que produzem conteúdo similar
          ao seu. Nossa IA analisará suas estratégias de SEO para criar matérias
          ainda melhores.
        </p>
      </div>

      {/* Action Buttons */}
      <div className="flex items-center justify-between pt-4">
        <Button
          type="button"
          variant="outline"
          onClick={onBack}
          disabled={isLoading}
        >
          Voltar
        </Button>

        <Button
          type="submit"
          variant="secondary"
          size="lg"
          isLoading={isLoading}
          disabled={isLoading}
        >
          {fields.length === 0 ? "Pular esta etapa" : "Próximo"}
        </Button>
      </div>
    </form>
  );
}
