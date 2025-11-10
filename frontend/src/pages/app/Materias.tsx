import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";
import { useNavigate } from "react-router-dom";

const Materias = () => {
  const navigate = useNavigate();

  return (
    <div className="max-w-6xl mx-auto">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold mb-2">Minhas Matérias</h1>
          <p className="text-muted-foreground">
            Gerencie todo o conteúdo publicado
          </p>
        </div>
      </div>

      {/* Empty State */}
      <Card className="shadow-card border-border/50">
        <CardContent className="flex flex-col items-center justify-center py-16">
          <div className="w-20 h-20 rounded-full bg-gradient-accent flex items-center justify-center mb-6">
            <Plus className="h-10 w-10 text-primary-foreground" />
          </div>
          <CardTitle className="mb-2 text-center">Nenhuma matéria publicada ainda</CardTitle>
          <CardDescription className="text-center mb-6 max-w-md">
            Comece criando seu primeiro fluxo de conteúdo para gerar artigos otimizados para SEO
          </CardDescription>
          <Button
            onClick={() => navigate("/app/novo")}
            className="bg-secondary text-secondary-foreground hover:bg-secondary/90"
          >
            <Plus className="h-4 w-4 mr-2" />
            Criar primeira matéria
          </Button>
        </CardContent>
      </Card>
    </div>
  );
};

export default Materias;
