"use client";

import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { businessSchema, competitorsSchema, integrationsUpdateSchema, type BusinessInput, type CompetitorsInput, type IntegrationsUpdateInput } from "@/lib/validations";
import { Plus, X, Upload, Building2, Users, Loader2, Save, ArrowLeft, Eye, EyeOff, Trash2, FileText, Globe, Check, HelpCircle } from "lucide-react";
import { useRouter } from "next/navigation";
import * as Accordion from "@radix-ui/react-accordion";
import { cn } from "@/lib/utils";
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
import { LocationFields } from "@/components/forms/LocationFields";
import {
    Card,
    CardContent,
    CardDescription,
    CardFooter,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { toast } from "sonner";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import api, { getErrorMessage } from "@/lib/axios";

interface WizardData {
    onboardingStep: number;
    business?: {
        description: string;
        primaryObjective: string;
        secondaryObjective?: string;
        location?: {
            country: string;
            state: string;
            city: string;
            hasMultipleUnits: boolean;
            units: Array<{
                id: string;
                name: string;
                country: string;
                state: string;
                city: string;
                isPrimary?: boolean;
            }>;
        };
        siteUrl?: string;
        hasBlog: boolean;
        blogUrls?: string[];
        brandFileUrl?: string;
    };
    competitors?: string[];
}

export default function ConfiguracoesPage() {
    const router = useRouter();
    const user = useUser();
    const [isLoading, setIsLoading] = useState(true);
    const [isSavingBusiness, setIsSavingBusiness] = useState(false);
    const [isSavingCompetitors, setIsSavingCompetitors] = useState(false);
    const [wizardData, setWizardData] = useState<WizardData | null>(null);
    const [selectedFile, setSelectedFile] = useState<File | null>(null);
    const [activeTab, setActiveTab] = useState("business");
    const [removeBrandFile, setRemoveBrandFile] = useState(false);
    const [isSavingIntegrations, setIsSavingIntegrations] = useState(false);
    const [showPassword, setShowPassword] = useState(false);

    // Business Form
    const businessForm = useForm({
        resolver: zodResolver(businessSchema),
        defaultValues: {
            description: "",
            primaryObjective: "leads" as const,
            hasBlog: false,
            blogUrls: [] as string[],
            articleCount: 1,
            location: {
                country: "",
                state: "",
                city: "",
                hasMultipleUnits: false,
                units: [] as Array<{
                    id: string;
                    name: string;
                    country: string;
                    state: string;
                    city: string;
                    isPrimary?: boolean;
                }>,
            },
            siteUrl: "",
        },
    });

    // Competitors Form
    const competitorsForm = useForm<CompetitorsInput>({
        resolver: zodResolver(competitorsSchema),
        defaultValues: {
            competitorUrls: [],
        },
    });

    // Integrations Form
    const integrationsForm = useForm<IntegrationsUpdateInput>({
        resolver: zodResolver(integrationsUpdateSchema),
        defaultValues: {
            wordpress: {
                siteUrl: "",
                username: "",
                appPassword: "",
            },
            searchConsole: {
                enabled: false,
            },
            analytics: {
                enabled: false,
            },
        },
    });

    const watchSearchConsoleEnabled = integrationsForm.watch("searchConsole.enabled");
    const watchAnalyticsEnabled = integrationsForm.watch("analytics.enabled");

    // Fetch wizard data on mount
    useEffect(() => {
        const fetchData = async () => {
            try {
                const response = await api.get<WizardData>("/wizard/data");
                setWizardData(response.data);

                // Populate business form
                if (response.data.business) {
                    const biz = response.data.business;
                    businessForm.reset({
                        description: biz.description || "",
                        primaryObjective: biz.primaryObjective as "leads" | "sales" | "branding",
                        secondaryObjective: biz.secondaryObjective as "leads" | "sales" | "branding" | undefined,
                        hasBlog: biz.hasBlog || false,
                        blogUrls: biz.blogUrls || [],
                        articleCount: 1, // This field is not needed for settings
                        location: biz.location ? {
                            country: biz.location.country || "",
                            state: biz.location.state || "",
                            city: biz.location.city || "",
                            hasMultipleUnits: biz.location.hasMultipleUnits || false,
                            units: biz.location.units?.map(unit => ({
                                id: unit.id,
                                name: unit.name || "",
                                country: unit.country || "",
                                state: unit.state || "",
                                city: unit.city || "",
                                isPrimary: unit.isPrimary || false,
                            })) || [],
                        } : {
                            country: "",
                            state: "",
                            city: "",
                            hasMultipleUnits: false,
                            units: [],
                        },
                        siteUrl: biz.siteUrl || "",
                    });
                }

                // Populate competitors form
                if (response.data.competitors && response.data.competitors.length > 0) {
                    competitorsForm.reset({
                        competitorUrls: response.data.competitors,
                    });
                }
            } catch (error) {
                console.error("Error fetching wizard data:", error);
                toast.error("Erro ao carregar configurações");
            } finally {
                setIsLoading(false);
            }
        };

        fetchData();
    }, []);

    const watchPrimaryObjective = businessForm.watch("primaryObjective");
    const watchDescription = businessForm.watch("description");
    const watchHasBlog = businessForm.watch("hasBlog");
    const blogUrls = (businessForm.watch("blogUrls") ?? []) as string[];
    const competitorUrls = competitorsForm.watch("competitorUrls") ?? [];

    const availableSecondaryObjectives = OBJECTIVES.filter(
        (obj) => obj.value !== watchPrimaryObjective
    );

    const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (file) {
            setSelectedFile(file);
            businessForm.setValue("brandFile", file);
        }
    };

    const handleSaveBusiness = async (data: BusinessInput) => {
        setIsSavingBusiness(true);

        try {
            // Build FormData for multipart
            const formData = new FormData();
            formData.append("description", data.description);
            formData.append("primaryObjective", data.primaryObjective);
            if (data.secondaryObjective) {
                formData.append("secondaryObjective", data.secondaryObjective);
            }
            formData.append("location", JSON.stringify(data.location));
            if (data.siteUrl) {
                formData.append("siteUrl", data.siteUrl);
            }
            formData.append("hasBlog", String(data.hasBlog));
            if (data.blogUrls && data.blogUrls.length > 0) {
                data.blogUrls.forEach((url) => formData.append("blogUrls", url));
            }
            if (selectedFile) {
                formData.append("brandFile", selectedFile);
            }
            // Enviar flag de remoção do arquivo de marca
            if (removeBrandFile && !selectedFile) {
                formData.append("removeBrandFile", "true");
            }

            await api.post("/wizard/business", formData, {
                headers: { "Content-Type": "multipart/form-data" },
            });

            // Atualizar estado local após sucesso
            if (removeBrandFile && !selectedFile) {
                setWizardData(prev => prev ? {
                    ...prev,
                    business: prev.business ? { ...prev.business, brandFileUrl: undefined } : undefined
                } : null);
            } else if (selectedFile) {
                // Se enviou novo arquivo, atualizar o estado (simplificado - idealmente pegaria a URL do response)
                setWizardData(prev => prev ? {
                    ...prev,
                    business: prev.business ? { ...prev.business, brandFileUrl: "uploaded" } : undefined
                } : null);
            }

            setRemoveBrandFile(false);
            setSelectedFile(null);
            toast.success("Informações do negócio atualizadas!");
        } catch (error) {
            const message = getErrorMessage(error);
            toast.error(message || "Erro ao atualizar informações do negócio");
        } finally {
            setIsSavingBusiness(false);
        }
    };

    const handleSaveCompetitors = async (data: CompetitorsInput) => {
        setIsSavingCompetitors(true);

        try {
            await api.post("/wizard/competitors", {
                competitorUrls: data.competitorUrls.filter((url) => url.trim() !== ""),
            });

            toast.success("Concorrentes atualizados!");
        } catch (error) {
            const message = getErrorMessage(error);
            toast.error(message || "Erro ao atualizar concorrentes");
        } finally {
            setIsSavingCompetitors(false);
        }
    };

    const handleSaveIntegrations = async (data: IntegrationsUpdateInput) => {
        setIsSavingIntegrations(true);
        try {
            await api.patch("/account/integrations", data);
            toast.success("Integrações atualizadas com sucesso!");
        } catch (error) {
            const message = getErrorMessage(error);
            toast.error(message || "Erro ao atualizar integrações");
        } finally {
            setIsSavingIntegrations(false);
        }
    };

    // Redirect if not authenticated or not completed onboarding
    if (!user?.hasCompletedOnboarding) {
        return (
            <div className="flex flex-col items-center justify-center min-h-[400px] space-y-4">
                <div className="text-center space-y-2">
                    <h2 className="text-xl font-semibold font-all-round text-[var(--color-primary-dark)]">
                        Complete o onboarding primeiro
                    </h2>
                    <p className="text-sm font-onest text-[var(--color-primary-dark)]/70">
                        Você precisa completar o onboarding antes de acessar as configurações
                    </p>
                </div>
                <Button variant="primary" onClick={() => router.push("/app/onboarding")}>
                    Ir para Onboarding
                </Button>
            </div>
        );
    }

    if (isLoading) {
        return (
            <div className="flex items-center justify-center min-h-[400px]">
                <div className="flex flex-col items-center gap-4">
                    <Loader2 className="h-8 w-8 animate-spin text-[var(--color-primary-purple)]" />
                    <p className="text-sm font-onest text-[var(--color-primary-dark)]/70">
                        Carregando configurações...
                    </p>
                </div>
            </div>
        );
    }

    return (
        <div className="space-y-6">
            {/* Header */}
            <div className="flex items-center gap-4">
                <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => router.back()}
                >
                    <ArrowLeft className="h-5 w-5" />
                </Button>
                <div>
                    <h1 className="text-3xl font-bold font-all-round text-[var(--color-primary-dark)]">
                        Configurações do Negócio
                    </h1>
                    <p className="text-sm font-onest text-[var(--color-primary-dark)]/70 mt-1">
                        Edite as informações do seu negócio e concorrentes
                    </p>
                </div>
            </div>

            {/* Tabs */}
            <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
                <TabsList className="grid w-full grid-cols-3 max-w-md">
                    <TabsTrigger value="business" className="flex items-center gap-2">
                        <Building2 className="h-4 w-4" />
                        Meu Negócio
                    </TabsTrigger>
                    <TabsTrigger value="competitors" className="flex items-center gap-2">
                        <Users className="h-4 w-4" />
                        Concorrentes
                    </TabsTrigger>
                    <TabsTrigger value="integrations" className="flex items-center gap-2">
                        <Globe className="h-4 w-4" />
                        Integrações
                    </TabsTrigger>
                </TabsList>

                {/* Business Tab */}
                <TabsContent value="business" className="mt-6">
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <Building2 className="h-5 w-5 text-[var(--color-primary-purple)]" />
                                Informações do Negócio
                            </CardTitle>
                            <CardDescription>
                                Atualize as informações do seu negócio para melhorar a precisão do conteúdo gerado
                            </CardDescription>
                        </CardHeader>

                        <form onSubmit={businessForm.handleSubmit(handleSaveBusiness)}>
                            <CardContent className="space-y-6">
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
                                        error={businessForm.formState.errors.description?.message as string}
                                        {...businessForm.register("description")}
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
                                                businessForm.setValue("primaryObjective", value as "leads" | "sales" | "branding")
                                            }
                                        >
                                            <SelectTrigger error={businessForm.formState.errors.primaryObjective?.message as string}>
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
                                                value={businessForm.watch("secondaryObjective") || ""}
                                                onValueChange={(value) =>
                                                    businessForm.setValue("secondaryObjective", value as "leads" | "sales" | "branding" | undefined)
                                                }
                                            >
                                                <SelectTrigger error={businessForm.formState.errors.secondaryObjective?.message as string}>
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

                                {/* Location Fields */}
                                <LocationFields
                                    control={businessForm.control as any}
                                    watch={businessForm.watch as any}
                                    setValue={businessForm.setValue as any}
                                    errors={businessForm.formState.errors as any}
                                    register={businessForm.register as any}
                                />

                                {/* URL do Site */}
                                <div className="space-y-2">
                                    <Label htmlFor="siteUrl">URL do seu site (opcional)</Label>
                                    <Input
                                        id="siteUrl"
                                        type="text"
                                        placeholder="seusite.com.br"
                                        error={businessForm.formState.errors.siteUrl?.message as string}
                                        {...businessForm.register("siteUrl")}
                                    />
                                </div>

                                {/* Tem Blog? */}
                                <div className="space-y-2">
                                    <div className="flex items-center gap-2">
                                        <input
                                            type="checkbox"
                                            id="hasBlog"
                                            className="h-4 w-4 rounded border-[var(--color-border)] text-[var(--color-primary-purple)] focus:ring-[var(--color-primary-purple)]"
                                            {...businessForm.register("hasBlog")}
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
                                                        error={(businessForm.formState.errors.blogUrls as any)?.[index]?.message as string}
                                                        {...businessForm.register(`blogUrls.${index}` as const)}
                                                    />
                                                    <Button
                                                        type="button"
                                                        variant="ghost"
                                                        size="icon"
                                                        onClick={() =>
                                                            businessForm.setValue(
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
                                                onClick={() => businessForm.setValue("blogUrls", [...blogUrls, ""])}
                                            >
                                                <Plus className="h-4 w-4 mr-2" />
                                                Adicionar URL
                                            </Button>
                                        </div>
                                    </div>
                                )}

                                {/* Upload Manual da Marca */}
                                <div className="space-y-3">
                                    <Label htmlFor="brandFile">Manual da marca (opcional)</Label>

                                    {/* Arquivo existente */}
                                    {wizardData?.business?.brandFileUrl && !removeBrandFile && !selectedFile && (
                                        <div className="flex items-center gap-3 p-3 bg-[var(--color-primary-teal)]/5 border border-[var(--color-primary-teal)]/20 rounded-[var(--radius-sm)]">
                                            <div className="flex items-center justify-center h-10 w-10 rounded-full bg-[var(--color-primary-teal)]/10">
                                                <FileText className="h-5 w-5 text-[var(--color-primary-teal)]" />
                                            </div>
                                            <div className="flex-1 min-w-0">
                                                <p className="text-sm font-medium font-onest text-[var(--color-primary-dark)] truncate">
                                                    Arquivo enviado
                                                </p>
                                                <p className="text-xs text-[var(--color-primary-dark)]/60 font-onest">
                                                    Manual da marca atual
                                                </p>
                                            </div>
                                            <div className="flex items-center gap-2">
                                                <Button
                                                    type="button"
                                                    variant="ghost"
                                                    size="sm"
                                                    onClick={() => window.open(wizardData.business!.brandFileUrl!, '_blank')}
                                                    title="Visualizar arquivo"
                                                >
                                                    <Eye className="h-4 w-4" />
                                                </Button>
                                                <Button
                                                    type="button"
                                                    variant="ghost"
                                                    size="sm"
                                                    onClick={() => setRemoveBrandFile(true)}
                                                    title="Remover arquivo"
                                                    className="text-[var(--color-error)] hover:text-[var(--color-error)] hover:bg-[var(--color-error)]/10"
                                                >
                                                    <Trash2 className="h-4 w-4" />
                                                </Button>
                                            </div>
                                        </div>
                                    )}

                                    {/* Confirmação de remoção */}
                                    {removeBrandFile && !selectedFile && (
                                        <div className="flex items-center gap-3 p-3 bg-[var(--color-error)]/5 border border-[var(--color-error)]/20 rounded-[var(--radius-sm)]">
                                            <p className="flex-1 text-sm text-[var(--color-error)] font-onest">
                                                O arquivo será removido ao salvar as alterações.
                                            </p>
                                            <Button
                                                type="button"
                                                variant="ghost"
                                                size="sm"
                                                onClick={() => setRemoveBrandFile(false)}
                                            >
                                                Desfazer
                                            </Button>
                                        </div>
                                    )}

                                    {/* Novo arquivo selecionado */}
                                    {selectedFile && (
                                        <div className="flex items-center gap-3 p-3 bg-[var(--color-primary-purple)]/5 border border-[var(--color-primary-purple)]/20 rounded-[var(--radius-sm)]">
                                            <div className="flex items-center justify-center h-10 w-10 rounded-full bg-[var(--color-primary-purple)]/10">
                                                <Upload className="h-5 w-5 text-[var(--color-primary-purple)]" />
                                            </div>
                                            <div className="flex-1 min-w-0">
                                                <p className="text-sm font-medium font-onest text-[var(--color-primary-dark)] truncate">
                                                    {selectedFile.name}
                                                </p>
                                                <p className="text-xs text-[var(--color-primary-dark)]/60 font-onest">
                                                    {(selectedFile.size / 1024).toFixed(1)} KB - Novo arquivo
                                                </p>
                                            </div>
                                            <Button
                                                type="button"
                                                variant="ghost"
                                                size="sm"
                                                onClick={() => {
                                                    setSelectedFile(null);
                                                    businessForm.setValue("brandFile", undefined);
                                                }}
                                                title="Remover seleção"
                                                className="text-[var(--color-error)] hover:text-[var(--color-error)] hover:bg-[var(--color-error)]/10"
                                            >
                                                <X className="h-4 w-4" />
                                            </Button>
                                        </div>
                                    )}

                                    {/* Botão de upload */}
                                    <div className="flex items-center gap-4">
                                        <label
                                            htmlFor="brandFile"
                                            className="flex items-center gap-2 px-4 py-2 rounded-[var(--radius-sm)] border-2 border-dashed border-[var(--color-border)] hover:border-[var(--color-primary-purple)] transition-colors cursor-pointer"
                                        >
                                            <Upload className="h-4 w-4" />
                                            <span className="text-sm font-onest">
                                                {wizardData?.business?.brandFileUrl && !removeBrandFile && !selectedFile
                                                    ? "Substituir arquivo"
                                                    : selectedFile
                                                        ? "Escolher outro"
                                                        : "Escolher arquivo"}
                                            </span>
                                        </label>
                                        <input
                                            id="brandFile"
                                            type="file"
                                            accept=".pdf,.jpg,.jpeg,.png"
                                            className="hidden"
                                            onChange={(e) => {
                                                handleFileChange(e);
                                                setRemoveBrandFile(false);
                                            }}
                                        />
                                    </div>
                                    <p className="text-xs text-[var(--color-primary-dark)]/60 font-onest">
                                        PDF, JPG ou PNG (máx. 5MB)
                                    </p>
                                </div>
                            </CardContent>

                            <CardFooter>
                                <Button
                                    type="submit"
                                    variant="primary"
                                    isLoading={isSavingBusiness}
                                    disabled={isSavingBusiness}
                                    className="flex items-center gap-2"
                                >
                                    <Save className="h-4 w-4" />
                                    Salvar Alterações
                                </Button>
                            </CardFooter>
                        </form>
                    </Card>
                </TabsContent>

                {/* Competitors Tab */}
                <TabsContent value="competitors" className="mt-6">
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <Users className="h-5 w-5 text-[var(--color-primary-teal)]" />
                                Concorrentes
                            </CardTitle>
                            <CardDescription>
                                Adicione sites ou nomes de concorrentes para criar uma estratégia de SEO mais
                                competitiva
                            </CardDescription>
                        </CardHeader>

                        <form onSubmit={competitorsForm.handleSubmit(handleSaveCompetitors)}>
                            <CardContent className="space-y-4">
                                {/* Lista de URLs */}
                                <div className="space-y-3">
                                    {competitorUrls.length === 0 ? (
                                        <div className="text-center py-8 px-4 border-2 border-dashed border-[var(--color-border)] rounded-[var(--radius-md)]">
                                            <p className="text-sm font-onest text-[var(--color-primary-dark)]/60 mb-4">
                                                Nenhum concorrente adicionado ainda
                                            </p>
                                            <Button
                                                type="button"
                                                variant="outline"
                                                size="sm"
                                                onClick={() => competitorsForm.setValue("competitorUrls", [""])}
                                            >
                                                <Plus className="h-4 w-4 mr-2" />
                                                Adicionar primeiro concorrente
                                            </Button>
                                        </div>
                                    ) : (
                                        <>
                                            {competitorUrls.map((_, index) => (
                                                <div key={index} className="space-y-2">
                                                    <div className="flex items-start gap-2">
                                                        <div className="flex-1">
                                                            <Label htmlFor={`competitor-${index}`}>
                                                                Concorrente {index + 1}
                                                            </Label>
                                                            <div className="mt-1 flex gap-2">
                                                                <Input
                                                                    id={`competitor-${index}`}
                                                                    type="text"
                                                                    placeholder="URL ou nome do concorrente (ex: Coca Cola)"
                                                                    error={competitorsForm.formState.errors.competitorUrls?.[index]?.message}
                                                                    {...competitorsForm.register(`competitorUrls.${index}` as const)}
                                                                />
                                                                <Button
                                                                    type="button"
                                                                    variant="ghost"
                                                                    size="icon"
                                                                    onClick={() =>
                                                                        competitorsForm.setValue(
                                                                            "competitorUrls",
                                                                            competitorUrls.filter((__, i) => i !== index)
                                                                        )
                                                                    }
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
                                            {competitorUrls.length < 20 && (
                                                <Button
                                                    type="button"
                                                    variant="outline"
                                                    size="sm"
                                                    onClick={() => competitorsForm.setValue("competitorUrls", [...competitorUrls, ""])}
                                                    className="w-full"
                                                >
                                                    <Plus className="h-4 w-4 mr-2" />
                                                    Adicionar concorrente ({competitorUrls.length}/20)
                                                </Button>
                                            )}

                                            {competitorUrls.length >= 20 && (
                                                <p className="text-xs text-[var(--color-warning)] font-onest text-center">
                                                    Limite de 20 concorrentes atingido
                                                </p>
                                            )}
                                        </>
                                    )}
                                </div>

                                {/* Info Box */}
                                <div className="bg-[var(--color-primary-purple)]/5 border border-[var(--color-primary-purple)]/20 rounded-[var(--radius-md)] p-4">
                                    <p className="text-sm font-onest text-[var(--color-primary-dark)]/80">
                                        💡 <strong>Dica:</strong> A preferência é por <strong>Links (URLs)</strong>, pois permitem uma análise mais profunda. Mas você também pode adicionar apenas o nome do concorrente.
                                    </p>
                                </div>
                            </CardContent>

                            <CardFooter>
                                <Button
                                    type="submit"
                                    variant="primary"
                                    isLoading={isSavingCompetitors}
                                    disabled={isSavingCompetitors}
                                    className="flex items-center gap-2"
                                >
                                    <Save className="h-4 w-4" />
                                    Salvar Concorrentes
                                </Button>
                            </CardFooter>
                        </form>
                    </Card>
                </TabsContent>

                {/* Integrations Tab */}
                <TabsContent value="integrations" className="mt-6">
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <Globe className="h-5 w-5 text-[var(--color-primary-purple)]" />
                                Integrações
                            </CardTitle>
                            <CardDescription>
                                Configure suas conexões com WordPress e Google para automatizar seu fluxo
                            </CardDescription>
                        </CardHeader>

                        <form onSubmit={integrationsForm.handleSubmit(handleSaveIntegrations)}>
                            <CardContent>
                                <Accordion.Root type="multiple" className="space-y-4">
                                    {/* WordPress */}
                                    <Accordion.Item value="wordpress">
                                        <div className="border-2 border-[var(--color-primary-purple)] rounded-[var(--radius-md)] overflow-hidden">
                                            <Accordion.Header>
                                                <Accordion.Trigger className="flex items-center justify-between w-full p-4 hover:bg-[var(--color-primary-purple)]/5 transition-colors">
                                                    <div className="flex items-center gap-3">
                                                        <div className="flex items-center justify-center h-10 w-10 rounded-full bg-[var(--color-primary-purple)]/10">
                                                            <span className="text-xl">🔌</span>
                                                        </div>
                                                        <div className="text-left">
                                                            <h4 className="text-base font-semibold font-all-round text-[var(--color-primary-dark)]">
                                                                WordPress
                                                            </h4>
                                                            <p className="text-xs font-onest text-[var(--color-primary-dark)]/60">
                                                                Publicação automática de matérias
                                                            </p>
                                                        </div>
                                                    </div>
                                                    <Check className="h-5 w-5 text-[var(--color-success)]" />
                                                </Accordion.Trigger>
                                            </Accordion.Header>

                                            <Accordion.Content className="p-4 pt-0 space-y-4">
                                                <div className="space-y-2">
                                                    <Label htmlFor="wp-siteUrl">URL do site</Label>
                                                    <Input
                                                        id="wp-siteUrl"
                                                        type="url"
                                                        placeholder="https://seusite.com.br"
                                                        {...integrationsForm.register("wordpress.siteUrl")}
                                                    />
                                                </div>

                                                <div className="space-y-2">
                                                    <Label htmlFor="wp-username">Nome de usuário</Label>
                                                    <Input
                                                        id="wp-username"
                                                        type="text"
                                                        placeholder="seu_usuario"
                                                        {...integrationsForm.register("wordpress.username")}
                                                    />
                                                </div>

                                                <div className="space-y-2">
                                                    <div className="flex items-center justify-between">
                                                        <Label htmlFor="wp-appPassword">
                                                            Senha de aplicativo
                                                        </Label>
                                                        <Dialog>
                                                            <DialogTrigger asChild>
                                                                <button
                                                                    type="button"
                                                                    className="text-[var(--color-primary-teal)] hover:text-[var(--color-primary-purple)]"
                                                                >
                                                                    <HelpCircle className="h-4 w-4" />
                                                                </button>
                                                            </DialogTrigger>
                                                            <DialogContent>
                                                                <DialogHeader>
                                                                    <DialogTitle>Como obter a senha?</DialogTitle>
                                                                    <DialogDescription className="space-y-2 text-left">
                                                                        <p>1. WordPress → Usuários → Perfil</p>
                                                                        <p>
                                                                            2. Role até &ldquo;Senhas de aplicativo&rdquo;
                                                                        </p>
                                                                        <p>3. Adicione uma nova senha</p>
                                                                        <p>4. Copie e cole aqui</p>
                                                                    </DialogDescription>
                                                                </DialogHeader>
                                                            </DialogContent>
                                                        </Dialog>
                                                    </div>
                                                    <div className="relative">
                                                        <Input
                                                            id="wp-appPassword"
                                                            type={showPassword ? "text" : "password"}
                                                            placeholder="••••••••••••"
                                                            autoComplete="new-password"
                                                            {...integrationsForm.register("wordpress.appPassword")}
                                                        />
                                                        <button
                                                            type="button"
                                                            onClick={() => setShowPassword(!showPassword)}
                                                            className="absolute right-3 top-1/2 -translate-y-1/2 text-[var(--color-primary-dark)]/40 hover:text-[var(--color-primary-dark)]/70 transition-colors"
                                                        >
                                                            {showPassword ? (
                                                                <EyeOff className="h-4 w-4" />
                                                            ) : (
                                                                <Eye className="h-4 w-4" />
                                                            )}
                                                        </button>
                                                    </div>
                                                </div>
                                            </Accordion.Content>
                                        </div>
                                    </Accordion.Item>

                                    {/* Google Search Console */}
                                    <Accordion.Item value="searchConsole">
                                        <div
                                            className={cn(
                                                "border-2 rounded-[var(--radius-md)] overflow-hidden",
                                                watchSearchConsoleEnabled
                                                    ? "border-[var(--color-primary-teal)]"
                                                    : "border-[var(--color-border)]"
                                            )}
                                        >
                                            <Accordion.Header>
                                                <Accordion.Trigger className="flex items-center justify-between w-full p-4 hover:bg-[var(--color-primary-teal)]/5">
                                                    <div className="flex items-center gap-3">
                                                        <div className="h-10 w-10 rounded-full bg-[var(--color-primary-teal)]/10 flex items-center justify-center">
                                                            <span className="text-xl">📊</span>
                                                        </div>
                                                        <div className="text-left">
                                                            <h4 className="text-base font-semibold font-all-round text-[var(--color-primary-dark)]">
                                                                Google Search Console
                                                            </h4>
                                                            <p className="text-xs font-onest text-[var(--color-primary-dark)]/60">
                                                                Análise de palavras-chave
                                                            </p>
                                                        </div>
                                                    </div>
                                                    <div className="flex items-center gap-2">
                                                        {watchSearchConsoleEnabled && (
                                                            <Check className="h-5 w-5 text-[var(--color-success)]" />
                                                        )}
                                                        <input
                                                            type="checkbox"
                                                            {...integrationsForm.register("searchConsole.enabled")}
                                                            onClick={(e) => e.stopPropagation()}
                                                            className="h-4 w-4 rounded"
                                                        />
                                                    </div>
                                                </Accordion.Trigger>
                                            </Accordion.Header>

                                            {watchSearchConsoleEnabled && (
                                                <Accordion.Content className="p-4 pt-0">
                                                    <Input
                                                        type="url"
                                                        placeholder="https://seusite.com.br"
                                                        {...integrationsForm.register("searchConsole.propertyUrl")}
                                                    />
                                                </Accordion.Content>
                                            )}
                                        </div>
                                    </Accordion.Item>

                                    {/* Google Analytics */}
                                    <Accordion.Item value="analytics">
                                        <div
                                            className={cn(
                                                "border-2 rounded-[var(--radius-md)] overflow-hidden",
                                                watchAnalyticsEnabled
                                                    ? "border-[var(--color-primary-teal)]"
                                                    : "border-[var(--color-border)]"
                                            )}
                                        >
                                            <Accordion.Header>
                                                <Accordion.Trigger className="flex items-center justify-between w-full p-4 hover:bg-[var(--color-primary-teal)]/5">
                                                    <div className="flex items-center gap-3">
                                                        <div className="h-10 w-10 rounded-full bg-[var(--color-primary-teal)]/10 flex items-center justify-center">
                                                            <span className="text-xl">📈</span>
                                                        </div>
                                                        <div className="text-left">
                                                            <h4 className="text-base font-semibold font-all-round text-[var(--color-primary-dark)]">
                                                                Google Analytics
                                                            </h4>
                                                            <p className="text-xs font-onest text-[var(--color-primary-dark)]/60">
                                                                Análise de tráfego
                                                            </p>
                                                        </div>
                                                    </div>
                                                    <div className="flex items-center gap-2">
                                                        {watchAnalyticsEnabled && (
                                                            <Check className="h-5 w-5 text-[var(--color-success)]" />
                                                        )}
                                                        <input
                                                            type="checkbox"
                                                            {...integrationsForm.register("analytics.enabled")}
                                                            onClick={(e) => e.stopPropagation()}
                                                            className="h-4 w-4 rounded"
                                                        />
                                                    </div>
                                                </Accordion.Trigger>
                                            </Accordion.Header>

                                            {watchAnalyticsEnabled && (
                                                <Accordion.Content className="p-4 pt-0">
                                                    <Input
                                                        type="text"
                                                        placeholder="G-XXXXXXXXXX"
                                                        {...integrationsForm.register("analytics.measurementId")}
                                                    />
                                                </Accordion.Content>
                                            )}
                                        </div>
                                    </Accordion.Item>
                                </Accordion.Root>
                            </CardContent>

                            <CardFooter>
                                <Button
                                    type="submit"
                                    variant="primary"
                                    isLoading={isSavingIntegrations}
                                    disabled={isSavingIntegrations}
                                    className="flex items-center gap-2"
                                >
                                    <Save className="h-4 w-4" />
                                    Atualizar Integrações
                                </Button>
                            </CardFooter>
                        </form>
                    </Card>
                </TabsContent>
            </Tabs>
        </div>
    );
}
