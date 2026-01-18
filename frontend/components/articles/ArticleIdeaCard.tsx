"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { Check, MessageSquare } from "lucide-react";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";
import type { ArticleIdea } from "@/types";

interface ArticleIdeaCardProps {
  idea: ArticleIdea;
  onUpdate: (id: string, updates: Partial<ArticleIdea>) => void;
}

// Função para truncar texto com JS
function truncateText(text: string, maxLength: number): string {
  if (text.length <= maxLength) return text;
  return text.slice(0, maxLength).trim() + "...";
}

export function ArticleIdeaCard({ idea, onUpdate }: ArticleIdeaCardProps) {
  // Estado local para approved (com sincronização)
  const [isApproved, setIsApproved] = useState(idea.approved);
  const [localFeedback, setLocalFeedback] = useState(idea.feedback || "");
  const [feedbackError, setFeedbackError] = useState("");
  const timeoutRef = useRef<NodeJS.Timeout | null>(null);
  const lastSavedRef = useRef(idea.feedback || "");

  // Sincronizar isApproved quando idea.approved mudar externamente
  useEffect(() => {
    setIsApproved(idea.approved);
  }, [idea.approved]);

  // Sincronizar localFeedback quando idea.feedback mudar externamente
  useEffect(() => {
    if (idea.feedback !== localFeedback && idea.feedback !== lastSavedRef.current) {
      setLocalFeedback(idea.feedback || "");
      lastSavedRef.current = idea.feedback || "";
    }
  }, [idea.feedback, localFeedback]);

  // Validação de feedback (useMemo para memoizar a função)
  const validateFeedback = useCallback((value: string): boolean => {
    const trimmed = value.trim();

    if (value.length > 0 && trimmed.length === 0) {
      setFeedbackError("Feedback não pode conter apenas espaços em branco");
      return false;
    }

    if (trimmed.length > 0 && trimmed.length < 10) {
      setFeedbackError("Feedback deve ter pelo menos 10 caracteres");
      return false;
    }

    setFeedbackError("");
    return true;
  }, []);

  // Memoize onUpdate to prevent infinite loops
  const handleFeedbackUpdate = useCallback(
    (feedback: string) => {
      // Validar feedback antes de salvar
      const trimmedFeedback = feedback.trim();

      // Se o feedback é vazio ou só espaços, salvar como string vazia
      const validFeedback = trimmedFeedback.length > 0 ? trimmedFeedback : "";

      if (validFeedback !== lastSavedRef.current) {
        lastSavedRef.current = validFeedback;
        onUpdate(idea.id, { feedback: validFeedback });
      }
    },
    [idea.id, onUpdate]
  );

  // Debounced update para feedback
  useEffect(() => {
    // Clear existing timeout
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }

    // Set new timeout
    timeoutRef.current = setTimeout(() => {
      if (validateFeedback(localFeedback)) {
        handleFeedbackUpdate(localFeedback);
      }
    }, 1000);

    // Cleanup
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, [localFeedback, handleFeedbackUpdate, validateFeedback]);

  const handleToggleApprove = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();

    const newApprovedState = !isApproved;

    // Atualizar estado local imediatamente para feedback visual
    setIsApproved(newApprovedState);

    // Atualizar estado global
    onUpdate(idea.id, { approved: newApprovedState });
  };

  const truncatedTitle = truncateText(idea.title, 80);
  const truncatedSummary = truncateText(idea.summary, 200);

  return (
    <Card
      className={cn(
        "transition-all duration-200 h-full flex flex-col",
        isApproved && "border-l-4 border-l-[var(--color-success)]",
        !isApproved && "opacity-60"
      )}
    >
      <CardHeader className="flex-shrink-0">
        <div className="flex items-start justify-between gap-3">
          <h3
            className="text-lg font-semibold font-all-round text-[var(--color-primary-dark)] flex-1"
            title={idea.title !== truncatedTitle ? idea.title : undefined}
          >
            {truncatedTitle}
          </h3>
          {isApproved && (
            <div className="flex items-center gap-1 px-2 py-1 rounded-full bg-[var(--color-success)]/10 text-[var(--color-success)] text-xs font-medium shrink-0">
              <Check className="h-3.5 w-3.5" />
              Aprovado
            </div>
          )}
        </div>
      </CardHeader>

      <CardContent className="space-y-4 flex-1 flex flex-col">
        {/* Summary */}
        <p
          className="text-sm font-onest text-[var(--color-primary-dark)]/80 flex-shrink-0"
          title={idea.summary !== truncatedSummary ? idea.summary : undefined}
        >
          {truncatedSummary}
        </p>

        {/* Botão de Aprovar/Desaprovar */}
        <Button
          variant={isApproved ? "success" : "outline"}
          size="sm"
          onClick={handleToggleApprove}
          className={cn(
            "w-full flex-shrink-0",
            isApproved &&
            "bg-[var(--color-success)] text-white hover:bg-[var(--color-success)]/90"
          )}
        >
          <Check className="h-4 w-4 mr-2" />
          {isApproved ? "Aprovado" : "Aprovar"}
        </Button>

        {/* Feedback Field */}
        <div className="space-y-2 pt-2 border-t border-[var(--color-border)] flex-1 flex flex-col">
          <div className="flex items-center gap-2 flex-shrink-0">
            <MessageSquare className="h-4 w-4 text-[var(--color-primary-teal)]" />
            <Label htmlFor={`feedback-${idea.id}`} className="text-xs">
              Sugestões ou direcionamentos (opcional)
            </Label>
          </div>
          <Textarea
            id={`feedback-${idea.id}`}
            value={localFeedback}
            onChange={(e) => setLocalFeedback(e.target.value)}
            placeholder="Ex: Foque em pequenas empresas, adicione exemplos práticos..."
            className={cn(
              "min-h-[60px] resize-none text-sm bg-[var(--color-secondary-cream)]/50 border-[var(--color-primary-teal)]/30 focus:border-[var(--color-primary-purple)] flex-1",
              feedbackError && "border-[var(--color-error)] focus:border-[var(--color-error)]"
            )}
            maxLength={500}
            showCount
          />
          {feedbackError && (
            <p className="text-xs text-[var(--color-error)] font-onest">
              {feedbackError}
            </p>
          )}
        </div>

        {/* Badges */}
        <div className="flex items-center gap-2 flex-wrap flex-shrink-0">
          {isApproved && localFeedback.trim() && !feedbackError && (
            <div className="px-2 py-1 rounded-full bg-[var(--color-primary-purple)]/10 text-[var(--color-primary-purple)] text-xs font-medium">
              Com direcionamento
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
