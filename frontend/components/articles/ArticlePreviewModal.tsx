"use client";

import { useState } from "react";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";

import { Article } from "@/types";
import { ExternalLink, Loader2 } from "lucide-react";
import { toast } from "sonner";
import ReactMarkdown from 'react-markdown';

interface ArticlePreviewModalProps {
  article: Article | null;
  isOpen: boolean;
  onClose: () => void;
  onPublish: (id: string) => Promise<void>;
  isPublishing: boolean;
}

export function ArticlePreviewModal({
  article,
  isOpen,
  onClose,
  onPublish,
  isPublishing
}: ArticlePreviewModalProps) {
  if (!article) return null;

  const handlePublish = async () => {
    try {
      await onPublish(article.id);
      onClose();
    } catch (error) {
      console.error("Erro ao publicar:", error);
      toast.error("Erro ao iniciar publicação");
    }
  };

  const isPublished = article.status === 'published';

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-4xl h-[90vh] flex flex-col p-0 overflow-hidden">
        <DialogHeader className="px-6 py-4 border-b">
          <DialogTitle className="text-xl font-bold font-all-round text-[var(--color-primary-dark)] pr-8">
            {article.title}
          </DialogTitle>
          <DialogDescription>
            {isPublished 
              ? "Visualize o conteúdo da matéria publicada" 
              : "Revise o conteúdo gerado antes de publicar no WordPress"}
          </DialogDescription>
        </DialogHeader>

      <div className="flex-1 overflow-hidden bg-gray-50/50">
          <div className="h-full overflow-y-auto p-6 md:p-10">
              <div className="prose prose-purple max-w-none prose-headings:font-all-round prose-p:font-onest prose-p:text-gray-600 prose-headings:text-[var(--color-primary-dark)]">
                 {article.content ? (
                    <ReactMarkdown>{article.content}</ReactMarkdown>
                 ) : (
                    <div className="flex flex-col items-center justify-center py-20 text-gray-400">
                        <p>Conteúdo não disponível para visualização.</p>
                        {article.postUrl && (
                            <a 
                                href={article.postUrl} 
                                target="_blank" 
                                rel="noopener noreferrer"
                                className="text-[var(--color-primary-purple)] hover:underline mt-2 flex items-center"
                            >
                                Ver artigo original <ExternalLink className="ml-1 h-3 w-3" />
                            </a>
                        )}
                    </div>
                 )}
              </div>
          </div>
        </div>

        <DialogFooter className="px-6 py-4 border-t bg-white">
          <Button variant="outline" onClick={onClose} disabled={isPublishing}>
            Fechar
          </Button>
          
          {isPublished && article.postUrl ? (
             <a href={article.postUrl} target="_blank" rel="noopener noreferrer">
                <Button className="bg-[var(--color-primary-purple)] hover:bg-[var(--color-primary-purple)]/90 text-white">
                    Ver no WordPress
                    <ExternalLink className="ml-2 h-4 w-4" />
                </Button>
             </a>
          ) : (
            <Button 
                onClick={handlePublish} 
                disabled={isPublishing || !article.content}
                className="bg-[var(--color-primary-purple)] hover:bg-[var(--color-primary-purple)]/90 text-white min-w-[140px]"
            >
                {isPublishing ? (
                <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Publicando...
                </>
                ) : (
                <>
                    Publicar no WordPress
                    <ExternalLink className="ml-2 h-4 w-4" />
                </>
                )}
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
