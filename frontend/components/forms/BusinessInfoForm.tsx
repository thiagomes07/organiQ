"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Plus, X, Upload } from "lucide-react";
import { businessSchema, type BusinessInput } from "@/lib/validations";
import { OBJECTIVES } from "@/lib/constants";
import { useUser } from "@/store/authStore";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Slider } from "@/components/ui/slider";
import { LocationFields } from "./LocationFields";
import { useState } from "react";

interface BusinessInfoFormProps {
  onSubmit: (data: BusinessInput) => void;
  isLoading?: boolean;
  defaultValues?: Partial<BusinessInput>;
}

export function BusinessInfoForm({
  onSubmit,
  isLoading,
  defaultValues,
}: BusinessInfoFormProps) {
  const user = useUser();
  const [selectedFile, setSelectedFile] = useState<File | null>(null);

  // Plan Limits Logic
  const maxArticles = user?.maxArticles || 1; // Default to 1 if missing
  const articlesUsed = user?.articlesUsed || 0;
  const articlesRemaining = Math.max(0, maxArticles - articlesUsed);
  const isLimitReached = articlesRemaining === 0;

  const {
    register,
    handleSubmit,
    watch,
    setValue,
    control,
    setError,
    formState: { errors },
  } = useForm({
    resolver: zodResolver(businessSchema),
    defaultValues: {
      description: "",
      primaryObjective: "leads",
      hasBlog: false,
      blogUrls: [],
      articleCount: isLimitReached ? 0 : 1, // Start with 1 valid if possible
      location: {
        country: "",
        state: "",
        city: "",
        hasMultipleUnits: false,
        units: [],
      },
      siteUrl: "",
      ...defaultValues,
    } as any,
  });

  const watchPrimaryObjective = watch("primaryObjective");
  const watchDescription = watch("description");
  const watchHasBlog = watch("hasBlog");
  const watchArticleCount = watch("articleCount");
  const blogUrls = watch("blogUrls") ?? [];

  const availableSecondaryObjectives = OBJECTIVES.filter(
    (obj) => obj.value !== watchPrimaryObjective
  );

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setSelectedFile(file);
      setValue("brandFile", file);
    }
  };

  const onFormError = (errors: any) => {
    console.log('[BusinessInfoForm] Validation errors:', errors);
  };

  const handleInternalSubmit = (data: BusinessInput) => {
    // Validação de limite de plano no frontend
    if (data.articleCount > articlesRemaining) {
      setError("articleCount", {
        type: "manual",
        message: `Seu plano permite apenas ${articlesRemaining} novas matérias. Faça um upgrade para gerar mais.`
      });
      return;
    }
    onSubmit(data);
  };

  return (
    <form onSubmit={handleSubmit(handleInternalSubmit, onFormError)} className="space-y-6">
      {/* Descrição do Negócio */}
      <div className="space-y-2">
        <Label htmlFor="description" required>
          Descreva seu negócio
        </Label>
        <Textarea
          id="description"
          placeholder="Ex: Somos uma agência de marketing digital especializada em pequenas empresas..."
          maxLength={500}
          showCount
          value={watchDescription}
          error={errors.description?.message as string}
          {...register("description")}
        />
        <p className="text-xs text-[var(--color-primary-dark)]/60 font-onest">
          Quanto mais detalhes, melhor será o conteúdo gerado
        </p>
      </div>

      {/* Objetivos */}
      <div className="space-y-4">
        <Label required>Quais são seus objetivos?</Label>

        {/* Objetivo Principal */}
        <div className="space-y-2">
          <Label htmlFor="primaryObjective">Objetivo Principal</Label>
          <Select
            value={watchPrimaryObjective}
            onValueChange={(value) =>
              setValue("primaryObjective", value as "leads" | "sales" | "branding")
            }
          >
            <SelectTrigger error={errors.primaryObjective?.message as string}>
              <SelectValue placeholder="Selecione seu objetivo principal" />
            </SelectTrigger>
            <SelectContent>
              {OBJECTIVES.map((obj) => (
                <SelectItem key={obj.value} value={obj.value}>
                  {obj.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* Objetivo Secundário */}
        {watchPrimaryObjective && (
          <div className="space-y-2">
            <Label htmlFor="secondaryObjective">
              Objetivo Secundário (opcional)
            </Label>
            <Select
              value={watch("secondaryObjective") || ""}
              onValueChange={(value) =>
                setValue("secondaryObjective", value as "leads" | "sales" | "branding" | undefined)
              }
            >
              <SelectTrigger error={errors.secondaryObjective?.message as string}>
                <SelectValue placeholder="Selecione um objetivo secundário (opcional)" />
              </SelectTrigger>
              <SelectContent>
                {availableSecondaryObjectives.map((obj) => (
                  <SelectItem key={obj.value} value={obj.value}>
                    {obj.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <p className="text-xs text-[var(--color-primary-dark)]/60 font-onest">
              Um objetivo secundário ajuda a criar conteúdo mais diversificado
            </p>
          </div>
        )}
      </div>

      {/* ========================================== */}
      {/* LOCATION FIELDS - NOVO */}
      {/* ========================================== */}
      <LocationFields
        control={control}
        watch={watch}
        setValue={setValue}
        errors={errors}
        register={register}
      />

      {/* URL do Site */}
      <div className="space-y-2">
        <Label htmlFor="siteUrl">URL do seu site (opcional)</Label>
        <Input
          id="siteUrl"
          type="text"
          placeholder="seusite.com.br"
          error={errors.siteUrl?.message as string}
          {...register("siteUrl")}
        />
      </div>

      {/* Tem Blog? */}
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <input
            type="checkbox"
            id="hasBlog"
            className="h-4 w-4 rounded border-[var(--color-border)] text-[var(--color-primary-purple)] focus:ring-[var(--color-primary-purple)]"
            {...register("hasBlog")}
          />
          <Label htmlFor="hasBlog" className="cursor-pointer">
            Meu site já tem um blog
          </Label>
        </div>
      </div>

      {/* URLs do Blog */}
      {watchHasBlog && (
        <div className="space-y-2">
          <Label>URLs do blog</Label>
          <div className="space-y-2">
            {blogUrls.map((_: string, index: number) => (
              <div key={index} className="flex gap-2">
                <Input
                  type="text"
                  placeholder="https://seusite.com.br/blog"
                  error={(errors.blogUrls as any)?.[index]?.message as string}
                  {...register(`blogUrls.${index}` as const)}
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  onClick={() =>
                    setValue(
                      "blogUrls",
                      blogUrls.filter((__: string, i: number) => i !== index)
                    )
                  }
                >
                  <X className="h-4 w-4" />
                </Button>
              </div>
            ))}
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => setValue("blogUrls", [...blogUrls, ""])}
            >
              <Plus className="h-4 w-4 mr-2" />
              Adicionar URL
            </Button>
          </div>
        </div>
      )}

      {/* Quantidade de Matérias */}
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <Label required>Quantas matérias deseja criar?</Label>
          <div className="text-right">
            <span className={isLimitReached ? "text-[var(--color-error)] text-xs font-bold" : "text-[var(--color-primary-purple)] text-xs font-bold"}>
              {articlesRemaining} créditos disponíveis
            </span>
            <br />
            <a href="/app/conta" className="text-[10px] underline text-[var(--color-primary-dark)]/60 hover:text-[var(--color-primary-purple)]">
              Aumentar limite
            </a>
          </div>
        </div>

        {isLimitReached ? (
          <div className="p-3 bg-[var(--color-error)]/10 border border-[var(--color-error)]/20 rounded-[var(--radius-sm)]">
            <p className="text-xs text-[var(--color-error)] font-onest">
              Você atingiu o limite de matérias do seu plano. <br />
              <a href="/app/conta" className="underline font-bold">Faça upgrade agora</a> para continuar gerando conteúdo.
            </p>
          </div>
        ) : (
          <Slider
            min={1}
            max={articlesRemaining} // Limita ao restante
            step={1}
            value={[watchArticleCount || 1]}
            onValueChange={(value) => setValue("articleCount", value[0])}
            showValue
            formatValue={(value) =>
              `${value} ${value === 1 ? "ma    téria" : "matérias"}`
            }
          />
        )}

        {errors.articleCount && (
          <p className="text-xs text-[var(--color-error)] font-onest">
            {errors.articleCount.message as string}
          </p>
        )}
      </div>

      {/* Upload Manual da Marca */}
      <div className="space-y-2">
        <Label htmlFor="brandFile">Manual da marca (opcional)</Label>
        <div className="flex items-center gap-4">
          <label
            htmlFor="brandFile"
            className="flex items-center gap-2 px-4 py-2 rounded-[var(--radius-sm)] border-2 border-dashed border-[var(--color-border)] hover:border-[var(--color-primary-purple)] transition-colors cursor-pointer"
          >
            <Upload className="h-4 w-4" />
            <span className="text-sm font-onest">
              {selectedFile ? selectedFile.name : "Escolher arquivo"}
            </span>
          </label>
          <input
            id="brandFile"
            type="file"
            accept=".pdf,.jpg,.jpeg,.png"
            className="hidden"
            onChange={handleFileChange}
          />
        </div>
        <p className="text-xs text-[var(--color-primary-dark)]/60 font-onest">
          PDF, JPG ou PNG (máx. 5MB)
        </p>
        {errors.brandFile && (
          <p className="text-xs text-[var(--color-error)] font-onest">
            {errors.brandFile.message as string}
          </p>
        )}
      </div>

      {/* Submit Button */}
      <div className="flex justify-end pt-4">
        <Button
          type="submit"
          variant="secondary"
          size="lg"
          isLoading={isLoading}
          disabled={isLoading}
        >
          Próximo
        </Button>
      </div>
    </form>
  );
}