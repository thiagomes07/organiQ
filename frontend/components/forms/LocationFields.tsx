"use client";

import {
  useFieldArray,
  type Control,
  type FieldErrors,
  type UseFormRegister,
  type UseFormSetValue,
  type UseFormWatch,
} from "react-hook-form";
import { Plus, Trash2, MapPin, GripVertical } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
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
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
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

const COUNTRIES = [
  { value: "Brasil", label: "Brasil" },
  { value: "Portugal", label: "Portugal" },
  { value: "Estados Unidos", label: "Estados Unidos" },
  { value: "Espanha", label: "Espanha" },
  { value: "Argentina", label: "Argentina" },
];

function createUUID(): string {
  const cryptoObj: Crypto | undefined =
    typeof globalThis !== "undefined" ? globalThis.crypto : undefined;

  if (cryptoObj?.randomUUID) {
    return cryptoObj.randomUUID();
  }

  // Fallback simples (RFC4122-ish) para ambientes sem randomUUID
  const bytes = new Uint8Array(16);
  if (cryptoObj?.getRandomValues) {
    cryptoObj.getRandomValues(bytes);
  } else {
    for (let i = 0; i < bytes.length; i++)
      bytes[i] = Math.floor(Math.random() * 256);
  }

  bytes[6] = (bytes[6] & 0x0f) | 0x40;
  bytes[8] = (bytes[8] & 0x3f) | 0x80;
  const hex = Array.from(bytes).map((b) => b.toString(16).padStart(2, "0"));
  return `${hex.slice(0, 4).join("")}-${hex.slice(4, 6).join("")}-${hex
    .slice(6, 8)
    .join("")}-${hex.slice(8, 10).join("")}-${hex.slice(10, 16).join("")}`;
}

interface LocationFieldsProps {
  control: Control<BusinessInput>;
  watch: UseFormWatch<BusinessInput>;
  setValue: UseFormSetValue<BusinessInput>;
  errors: FieldErrors<BusinessInput>;
  register: UseFormRegister<BusinessInput>;
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

  const [removeIndex, setRemoveIndex] = useState<number | null>(null);

  const { fields, append, remove, replace } = useFieldArray<
    BusinessInput,
    "location.units"
  >({
    control,
    name: "location.units",
  });

  const isBrazil = country === "Brasil";
  const canShowState = Boolean(country);
  const canShowCity = Boolean(state);

  // Limpar dependências quando country muda (single)
  useEffect(() => {
    if (!hasMultipleUnits) {
      setValue("location.state", "", {
        shouldDirty: true,
        shouldValidate: true,
      });
      setValue("location.city", "", {
        shouldDirty: true,
        shouldValidate: true,
      });
    }

    // Sempre limpa state/city de topo quando trocar de país
    // (como são opcionais e dependentes)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [country]);

  // Limpar city quando state muda (single)
  useEffect(() => {
    if (!hasMultipleUnits) {
      setValue("location.city", "", {
        shouldDirty: true,
        shouldValidate: true,
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [state]);

  const unitsCountLabel = useMemo(() => {
    if (fields.length === 0) return "Nenhuma unidade adicionada ainda";
    return `${fields.length} ${
      fields.length === 1 ? "unidade cadastrada" : "unidades cadastradas"
    }`;
  }, [fields.length]);

  const handleAddUnit = () => {
    const baseCountry = country || "";
    append({
      id: createUUID(),
      name: "",
      country: baseCountry,
      state: "",
      city: "",
    });
  };

  const setUnitCountry = (index: number, value: string) => {
    setValue(`location.units.${index}.country` as const, value, {
      shouldDirty: true,
      shouldValidate: true,
    });
  };

  const setUnitState = (index: number, value: string) => {
    setValue(`location.units.${index}.state` as const, value, {
      shouldDirty: true,
      shouldValidate: true,
    });
  };

  const setUnitCity = (index: number, value: string) => {
    setValue(`location.units.${index}.city` as const, value, {
      shouldDirty: true,
      shouldValidate: true,
    });
  };

  const clearUnitStateCity = (index: number) => {
    setUnitState(index, "");
    setUnitCity(index, "");
  };

  const handleToggleMultipleUnits = (checked: boolean) => {
    setValue("location.hasMultipleUnits", checked, {
      shouldDirty: true,
      shouldValidate: true,
    });
    if (!checked) {
      replace([]);
    } else {
      // No modo múltiplas unidades, topo não usa state/city
      setValue("location.state", "", {
        shouldDirty: true,
        shouldValidate: true,
      });
      setValue("location.city", "", {
        shouldDirty: true,
        shouldValidate: true,
      });
    }
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
        {/* País (Obrigatório) */}
        <div className="space-y-2">
          <Label htmlFor="location.country" required>
            País
          </Label>
          <Select
            value={country || ""}
            onValueChange={(value) => {
              setValue("location.country", value, {
                shouldDirty: true,
                shouldValidate: true,
              });
              // Reinicia dependências do topo
              setValue("location.state", "", {
                shouldDirty: true,
                shouldValidate: true,
              });
              setValue("location.city", "", {
                shouldDirty: true,
                shouldValidate: true,
              });
              // Atualiza defaults das unidades existentes (se houver)
              if (hasMultipleUnits && fields.length > 0) {
                fields.forEach((_, idx) => {
                  setUnitCountry(idx, value);
                  clearUnitStateCity(idx);
                });
              }
            }}
          >
            <SelectTrigger
              id="location.country"
              error={errors.location?.country?.message}
            >
              <SelectValue placeholder="Selecione o país" />
            </SelectTrigger>
            <SelectContent>
              {COUNTRIES.map((c) => (
                <SelectItem key={c.value} value={c.value}>
                  {c.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* Checkbox: Múltiplas Unidades */}
        <div className="flex items-center gap-2">
          <input
            type="checkbox"
            id="hasMultipleUnits"
            className="h-4 w-4 rounded border-[var(--color-border)] text-[var(--color-primary-purple)] focus:ring-[var(--color-primary-purple)]"
            checked={hasMultipleUnits}
            onChange={(e) => handleToggleMultipleUnits(e.target.checked)}
          />
          <Label
            htmlFor="hasMultipleUnits"
            className="cursor-pointer font-medium"
          >
            Meu negócio tem mais de uma unidade
          </Label>
        </div>
        <p className="text-xs text-[var(--color-primary-dark)]/60 font-onest pl-6">
          Marque se você deseja especificar localizações diferentes
        </p>

        {/* Single Location */}
        {!hasMultipleUnits && (
          <div className="space-y-3 pt-2">
            {/* Estado */}
            {canShowState && (
              <div className="space-y-2">
                <Label htmlFor="location.state">Estado (opcional)</Label>
                {isBrazil ? (
                  <Select
                    value={state || ""}
                    onValueChange={(value) =>
                      setValue("location.state", value, {
                        shouldDirty: true,
                        shouldValidate: true,
                      })
                    }
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
                ) : (
                  <Input
                    id="location.state"
                    type="text"
                    placeholder="Ex: Lisboa"
                    error={errors.location?.state?.message}
                    {...register("location.state")}
                  />
                )}
              </div>
            )}

            {/* Cidade */}
            {canShowCity && (
              <div className="space-y-2">
                <Label htmlFor="location.city">Cidade (opcional)</Label>
                {isBrazil ? (
                  <Select
                    value={watch("location.city") || ""}
                    onValueChange={(value) =>
                      setValue("location.city", value, {
                        shouldDirty: true,
                        shouldValidate: true,
                      })
                    }
                  >
                    <SelectTrigger
                      id="location.city"
                      error={errors.location?.city?.message}
                    >
                      <SelectValue placeholder="Selecione a cidade" />
                    </SelectTrigger>
                    <SelectContent className="max-h-[200px] overflow-y-auto">
                      {getCitiesByState(state ?? "").map((city) => (
                        <SelectItem key={city} value={city}>
                          {city}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                ) : (
                  <Input
                    id="location.city"
                    type="text"
                    placeholder="Ex: Porto"
                    error={errors.location?.city?.message}
                    {...register("location.city")}
                  />
                )}
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
              const unitCountry = watch(
                `location.units.${index}.country` as const
              );
              const unitIsBrazil = unitCountry === "Brasil";
              const unitState = watch(`location.units.${index}.state` as const);
              const unitCanShowState = Boolean(unitCountry);
              const unitCanShowCity = Boolean(unitState);

              return (
                <div
                  key={field.id}
                  className={cn(
                    "border-l-4 border-[var(--color-primary-purple)] rounded-[var(--radius-md)] bg-white p-4 space-y-3",
                    "shadow-sm hover:shadow-md transition-shadow animate-in slide-in-from-bottom-2 duration-200"
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
                      onClick={() => setRemoveIndex(index)}
                    >
                      <Trash2 className="h-4 w-4" />
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
                      {...register(`location.units.${index}.name` as const)}
                    />
                  </div>

                  {/* País (obrigatório) */}
                  <div className="space-y-2">
                    <Label required>País</Label>
                    <Select
                      value={unitCountry || ""}
                      onValueChange={(value) => {
                        setUnitCountry(index, value);
                        clearUnitStateCity(index);
                      }}
                    >
                      <SelectTrigger
                        error={
                          errors.location?.units?.[index]?.country?.message
                        }
                      >
                        <SelectValue placeholder="Selecione o país" />
                      </SelectTrigger>
                      <SelectContent>
                        {COUNTRIES.map((c) => (
                          <SelectItem key={c.value} value={c.value}>
                            {c.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  {/* Estado */}
                  {unitCanShowState && (
                    <div className="space-y-2">
                      <Label>Estado (opcional)</Label>
                      {unitIsBrazil ? (
                        <Select
                          value={unitState || ""}
                          onValueChange={(value) => {
                            setUnitState(index, value);
                            setUnitCity(index, "");
                          }}
                        >
                          <SelectTrigger
                            error={
                              errors.location?.units?.[index]?.state?.message
                            }
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
                      ) : (
                        <Input
                          type="text"
                          placeholder="Ex: Lisboa"
                          error={
                            errors.location?.units?.[index]?.state?.message
                          }
                          {...register(
                            `location.units.${index}.state` as const
                          )}
                        />
                      )}
                    </div>
                  )}

                  {/* Cidade */}
                  {unitCanShowCity && (
                    <div className="space-y-2">
                      <Label>Cidade (opcional)</Label>
                      {unitIsBrazil ? (
                        <Select
                          value={
                            watch(`location.units.${index}.city` as const) || ""
                          }
                          onValueChange={(value) => setUnitCity(index, value)}
                        >
                          <SelectTrigger
                            error={
                              errors.location?.units?.[index]?.city?.message
                            }
                          >
                            <SelectValue placeholder="Selecione a cidade" />
                          </SelectTrigger>
                          <SelectContent className="max-h-[200px] overflow-y-auto">
                            {getCitiesByState(unitState ?? "").map((city) => (
                              <SelectItem key={city} value={city}>
                                {city}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      ) : (
                        <Input
                          type="text"
                          placeholder="Ex: Porto"
                          error={errors.location?.units?.[index]?.city?.message}
                          {...register(`location.units.${index}.city` as const)}
                        />
                      )}
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
                <span>{unitsCountLabel}</span>
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

      {/* Confirmação de Remoção */}
      <Dialog
        open={removeIndex !== null}
        onOpenChange={(open) => !open && setRemoveIndex(null)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Remover unidade?</DialogTitle>
            <DialogDescription>
              Esta ação remove a unidade selecionada. Se for a última unidade, o
              modo de múltiplas unidades será desativado.
            </DialogDescription>
          </DialogHeader>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => setRemoveIndex(null)}
            >
              Cancelar
            </Button>
            <Button
              type="button"
              variant="danger"
              onClick={() => {
                if (removeIndex === null) return;
                const isLast = fields.length === 1;
                remove(removeIndex);
                setRemoveIndex(null);
                if (isLast) {
                  handleToggleMultipleUnits(false);
                }
              }}
            >
              Remover
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
