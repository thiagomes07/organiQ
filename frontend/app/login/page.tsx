"use client";

import { useState } from "react";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { LoginForm } from "@/components/forms/LoginForm";
import { RegisterForm } from "@/components/forms/RegisterForm";

export default function LoginPage() {
  const [activeTab, setActiveTab] = useState("login");

  // Lógica de cores baseada na aba ativa
  const isRegister = activeTab === "register";

  // Título: Roxo no login, Amarelo no registro
  // Adicionei drop-shadow no amarelo para garantir leitura no fundo claro
  const titleColorClass = isRegister
    ? "text-[var(--color-secondary-yellow)] drop-shadow"
    : "text-[var(--color-primary-purple)]";

  // Borda: Roxa no login, Amarela no registro
  const borderColorClass = isRegister
    ? "border-t-[var(--color-secondary-yellow)]"
    : "border-t-[var(--color-primary-purple)]";

  return (
    <div className="min-h-screen flex flex-col items-center justify-center p-8 overflow-hidden bg-gray-50/50">
      {/* Back to Home */}
      <div className="fixed top-8 left-4 md:left-8 z-10">
        <Link
          href="/"
          className="inline-flex items-center gap-2 text-sm font-medium font-onest text-[var(--color-primary-dark)]/70 hover:text-[var(--color-primary-dark)] transition-colors"
        >
          <ArrowLeft className="h-4 w-4" />
          Voltar para o início
        </Link>
      </div>

      {/* Main Container - Âncora Relativa */}
      <div className="relative w-full max-w-md mt-12 md:mt-0 mr-8 md:mr-0">
        {/* Logo - Posicionado acima e à direita */}
        <div className="absolute bottom-full right-0 mb-[-11px] z-10">
          <h1
            className={`text-6xl md:text-7xl font-bold font-all-round leading-none select-none transition-all duration-500 ease-in-out ${titleColorClass}`}
          >
            organiQ
          </h1>
        </div>

        {/* Slogan Conceitual */}
        <div className="absolute left-full top-1 h-full ml-3 md:ml-5">
          <p className="origin-top-left rotate-90 whitespace-nowrap text-xs md:text-sm font-light font-onest text-[var(--color-primary-teal)] tracking-[0.4em] uppercase opacity-80 select-none">
            Naturalmente Inteligente
          </p>
        </div>

        {/* Login/Register Card */}
        {/* A borda superior (border-t-4) muda de cor suavemente */}
        <Card
          className={`w-full shadow-xl border-t-4 transition-colors duration-500 ease-in-out ${borderColorClass}`}
        >
          <CardHeader className="space-y-1 pb-4 pt-8">
            <CardTitle className="text-2xl text-center font-onest">
              Bem-vindo
            </CardTitle>
            <CardDescription className="text-center font-onest">
              {isRegister
                ? "Crie sua conta para começar sua jornada"
                : "Entre na sua conta ou crie uma nova para começar"}
            </CardDescription>
          </CardHeader>

          <CardContent>
            {/* O onValueChange atualiza o estado e dispara a troca de cores */}
            <Tabs
              defaultValue="login"
              value={activeTab}
              onValueChange={setActiveTab}
              className="w-full"
            >
              <TabsList className="grid w-full grid-cols-2 mb-6">
                <TabsTrigger value="login" className="font-onest">
                  Entrar
                </TabsTrigger>
                <TabsTrigger value="register" className="font-onest">
                  Cadastrar
                </TabsTrigger>
              </TabsList>

              <TabsContent value="login">
                <LoginForm />
              </TabsContent>

              <TabsContent value="register">
                <RegisterForm />
              </TabsContent>
            </Tabs>
          </CardContent>
        </Card>
      </div>

      {/* Footer Note */}
      <div className="mt-12 text-center relative z-10">
        <p className="text-xs font-onest text-[var(--color-primary-dark)]/60">
          Ao continuar, você concorda com nossos{" "}
          <a
            href="#"
            className="text-[var(--color-primary-purple)] hover:underline font-medium"
          >
            Termos de Uso
          </a>
          {" e "}
          <a
            href="#"
            className="text-[var(--color-primary-purple)] hover:underline font-medium"
          >
            Política de Privacidade
          </a>
        </p>
      </div>
    </div>
  );
}
