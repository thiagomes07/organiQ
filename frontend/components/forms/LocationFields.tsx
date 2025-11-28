"use client";

import { useFieldArray, Control } from "react-hook-form";
import { Plus, X, MapPin, GripVertical } from "lucide-react";
import { useEffect, useRef } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { cn } from "@/lib/utils";
import type { BusinessInput } from "@/lib/validations";

/**
 * ESTRUTURA DE DADOS:
 * 
 * const BRAZIL_STATES: string[] = ["AC", "AL", "AP", ...]
 * 
 * const cidadesPorEstado: Record<string, string[]> = {
 *   "AC": ["Acrelândia", "Assis Brasil", ...],
 *   "AL": ["Anadia", "Arapiraca", ...],
 *   ...
 * }
 */
import {
  BRAZIL_STATES,
  getCitiesByState,
  getStateName,
} from "@/lib/brazil-locations";

interface LocationFieldsProps {
  control: Control<BusinessInput>;
  watch: (name: string) => any;
  setValue: (name: string, value: any) => void;
  errors: any;
  register: any;
}

export function LocationFields({
  control,
  watch,
  setValue,
  errors,
  register,
}: LocationFieldsProps) {
  const hasMultipleUnits = watch("location.hasMultipleUnits");
  const country = watch("location.country");
  const state = watch("location.state");

  // Refs para detectar mudanças de estado
  const statePreviousValue = useRef<string>("");

  const { fields, append, remove } = useFieldArray({
    control,
    name: "location.units",
  });

  // Limpar cidade quando estado muda (single location)
  useEffect(() => {
    if (!hasMultipleUnits && state) {
      if (statePreviousValue.current === "") {
        statePreviousValue.current = state;
        return;
      }
      if (statePreviousValue.current !== state) {
        setValue("location.city", "");
        statePreviousValue.current = state;
      }
    }
  }, [state, hasMultipleUnits, setValue]);

  const handleAddUnit = () => {
    const newUnit = {
      id: crypto.randomUUID(),
      name: "",
      country: "Brasil",
      state: "",
      city: "",
    };
    append(newUnit as any);
  };

  const handleStateChange = (unitIndex: number, newState: string) => {
    setValue(`location.units.${unitIndex}.state`, newState);
    setValue(`location.units.${unitIndex}.city`, "");
  };

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-start gap-3">
        <div className="flex items-center justify-center h-10 w-10 rounded-full bg-[var(--color-primary-teal)]/10 shrink-0 mt-1">
          <MapPin className="h-5 w-5 text-[var(--color-primary-teal)]" />
        </div>
        <div className="flex-1">
          <Label required className="text-base">
            Localização do Negócio
          </Label>
          <p className="text-sm font-onest text-[var(--color-primary-dark)]/60 mt-1">
            Informe onde seu negócio atua para otimizar o SEO local
          </p>
        </div>
      </div>

      {/* Card Container */}
      <div className="border-2 border-[var(--color-primary-teal)]/20 rounded-[var(--radius-md)] p-4 bg-[var(--color-secondary-cream)]/30 space-y-4">
        {/* Info: Foco no Brasil */}
        <div className="bg-blue-50 border border-blue-200 rounded-[var(--radius-sm)] p-3">
          <p className="text-xs font-onest text-blue-900">
            🇧🇷 <strong>Foco no Brasil:</strong> Por enquanto, o organiQ está
            otimizado para empresas brasileiras. Selecione o estado e cidade da
            sua operação.
          </p>
        </div>

        {/* Checkbox: Múltiplas Unidades */}
        <div className="flex items-center gap-2">
          <input
            type="checkbox"
            id="hasMultipleUnits"
            className="h-4 w-4 rounded border-[var(--color-border)] text-[var(--color-primary-purple)] focus:ring-[var(--color-primary-purple)]"
            checked={hasMultipleUnits}
            onChange={(e) =>
              setValue("location.hasMultipleUnits", e.target.checked)
            }
          />
          <Label
            htmlFor="hasMultipleUnits"
            className="cursor-pointer font-medium"
          >
            Meu negócio tem mais de uma unidade
          </Label>
        </div>
        <p className="text-xs text-[var(--color-primary-dark)]/60 font-onest pl-6">
          Marque se você deseja especificar localizações diferentes no Brasil
        </p>

        {/* Single Location */}
        {!hasMultipleUnits && (
          <div className="space-y-3 pt-2">
            {/* País (fixo em Brasil) */}
            <input
              type="hidden"
              {...register("location.country")}
              value="Brasil"
            />

            <div className="bg-gray-50 border border-gray-200 rounded-[var(--radius-sm)] p-3">
              <p className="text-sm font-onest text-gray-700">
                <strong>País:</strong> Brasil 🇧🇷
              </p>
            </div>

            {/* Estado */}
            <div className="space-y-2">
              <Label htmlFor="location.state" required>
                Estado
              </Label>
              <Select
                value={state || ""}
                onValueChange={(value) => setValue("location.state", value)}
              >
                <SelectTrigger
                  id="location.state"
                  error={errors.location?.state?.message}
                >
                  <SelectValue placeholder="Selecione o estado" />
                </SelectTrigger>
                <SelectContent>
                  {BRAZIL_STATES.map((uf) => (
                    <SelectItem key={uf} value={uf}>
                      {uf} - {getStateName(uf)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Cidade */}
            {state && (
              <div className="space-y-2">
                <Label htmlFor="location.city" required>
                  Cidade
                </Label>
                <Select
                  value={watch("location.city") || ""}
                  onValueChange={(value) => setValue("location.city", value)}
                >
                  <SelectTrigger
                    id="location.city"
                    error={errors.location?.city?.message}
                  >
                    <SelectValue placeholder="Selecione a cidade" />
                  </SelectTrigger>
                  <SelectContent className="max-h-[200px] overflow-y-auto">
                    {getCitiesByState(state).map((city) => (
                      <SelectItem key={city} value={city}>
                        {city}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}
          </div>
        )}

        {/* Multiple Units */}
        {hasMultipleUnits && (
          <div className="space-y-4 pt-2">
            {/* Empty State */}
            {fields.length === 0 && (
              <div className="text-center py-6 px-4 border-2 border-dashed border-[var(--color-border)] rounded-[var(--radius-md)] bg-white">
                <p className="text-sm font-onest text-[var(--color-primary-dark)]/60 mb-3">
                  Nenhuma unidade adicionada ainda
                </p>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={handleAddUnit}
                >
                  <Plus className="h-4 w-4 mr-2" />
                  Adicionar primeira unidade
                </Button>
              </div>
            )}

            {/* Units List */}
            {fields.map((field, index) => {
              const unitState = watch(`location.units.${index}.state`);

              return (
                <div
                  key={field.id}
                  className={cn(
                    "border-l-4 border-[var(--color-primary-purple)] rounded-[var(--radius-md)] bg-white p-4 space-y-3",
                    "shadow-sm hover:shadow-md transition-shadow"
                  )}
                >
                  {/* Header */}
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <GripVertical className="h-4 w-4 text-[var(--color-primary-dark)]/40" />
                      <span className="text-sm font-semibold font-all-round text-[var(--color-primary-dark)]">
                        Unidade {index + 1}
                      </span>
                    </div>
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      onClick={() => {
                        if (fields.length === 1) {
                          if (
                            confirm(
                              "Remover a última unidade? Isso desativará múltiplas unidades."
                            )
                          ) {
                            remove(index);
                            setValue("location.hasMultipleUnits", false);
                          }
                        } else {
                          remove(index);
                        }
                      }}
                    >
                      <X className="h-4 w-4" />
                    </Button>
                  </div>

                  {/* Nome da Unidade (opcional) */}
                  <div className="space-y-2">
                    <Label htmlFor={`location.units.${index}.name`}>
                      Nome da unidade (opcional)
                    </Label>
                    <Input
                      type="text"
                      placeholder="Ex: Filial Centro, Matriz..."
                      {...register(`location.units.${index}.name`)}
                    />
                  </div>

                  {/* País (fixo) */}
                  <input
                    type="hidden"
                    {...register(`location.units.${index}.country`)}
                    value="Brasil"
                  />
                  <div className="bg-gray-50 border border-gray-200 rounded-[var(--radius-sm)] p-2">
                    <p className="text-xs font-onest text-gray-700">
                      <strong>País:</strong> Brasil 🇧🇷
                    </p>
                  </div>

                  {/* Estado */}
                  <div className="space-y-2">
                    <Label htmlFor={`location.units.${index}.state`} required>
                      Estado
                    </Label>
                    <Select
                      value={unitState || ""}
                      onValueChange={(value) => handleStateChange(index, value)}
                    >
                      <SelectTrigger
                        error={errors.location?.units?.[index]?.state?.message}
                      >
                        <SelectValue placeholder="Selecione o estado" />
                      </SelectTrigger>
                      <SelectContent>
                        {BRAZIL_STATES.map((uf) => (
                          <SelectItem key={uf} value={uf}>
                            {uf} - {getStateName(uf)}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  {/* Cidade */}
                  {unitState && (
                    <div className="space-y-2">
                      <Label htmlFor={`location.units.${index}.city`} required>
                        Cidade
                      </Label>
                      <Select
                        value={watch(`location.units.${index}.city`) || ""}
                        onValueChange={(value) =>
                          setValue(`location.units.${index}.city`, value)
                        }
                      >
                        <SelectTrigger
                          error={errors.location?.units?.[index]?.city?.message}
                        >
                          <SelectValue placeholder="Selecione a cidade" />
                        </SelectTrigger>
                        <SelectContent className="max-h-[200px] overflow-y-auto">
                          {getCitiesByState(unitState).map((city) => (
                            <SelectItem key={city} value={city}>
                              {city}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  )}
                </div>
              );
            })}

            {/* Add More Button */}
            {fields.length > 0 && fields.length < 10 && (
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={handleAddUnit}
                className="w-full"
              >
                <Plus className="h-4 w-4 mr-2" />
                Adicionar unidade ({fields.length}/10)
              </Button>
            )}

            {/* Badge: Total de Unidades */}
            {fields.length > 0 && (
              <div className="flex items-center gap-2 text-sm font-onest text-[var(--color-primary-purple)]">
                <MapPin className="h-4 w-4" />
                <span>
                  {fields.length}{" "}
                  {fields.length === 1
                    ? "unidade cadastrada"
                    : "unidades cadastradas"}
                </span>
              </div>
            )}

            {/* Max Limit Warning */}
            {fields.length >= 10 && (
              <p className="text-xs text-[var(--color-warning)] font-onest text-center">
                Limite de 10 unidades atingido
              </p>
            )}
          </div>
        )}

        {/* Error Message */}
        {errors.location?.units && (
          <p className="text-xs text-[var(--color-error)] font-onest">
            {errors.location.units.message}
          </p>
        )}
      </div>

      {/* Info Box */}
      <div className="bg-[var(--color-primary-purple)]/5 border border-[var(--color-primary-purple)]/20 rounded-[var(--radius-md)] p-3">
        <p className="text-xs font-onest text-[var(--color-primary-dark)]/80">
          💡 <strong>Dica:</strong> Informações de localização ajudam a criar
          conteúdo otimizado para SEO local, aumentando sua visibilidade em
          buscas geográficas específicas do Brasil.
        </p>
      </div>
    </div>
  );
}
