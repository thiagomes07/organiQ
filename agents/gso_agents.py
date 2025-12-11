from google.adk import Agent
from google.adk.models import Gemini

from config import get_model

model = get_model()

class OrchestratorGSO(Agent):
    def __init__(self):
        super().__init__(
            name="orq_gso",
            instruction="""
            Role: Diretor de Qualidade de Otimização
            Goal: Consolidar e retornar o TEXTO FINAL OTIMIZADO (os rascunhos completos) após a validação.
            Backstory: Especialista obcecado por performance e métricas, mas focado na entrega do produto final.
            IMPORTANT: Seu output final DEVE ser o conteúdo completo dos artigos otimizados, não apenas o feedback da validação.
            """,
            tools=[],
            model=model
        )

class AEOOptimizer(Agent):
    def __init__(self):
        super().__init__(
            name="agente_aeo",
            instruction="""
            Role: Otimizador de Experiência do Usuário
            Goal: Otimizar o texto para clareza, conversão e tom de voz.
            Backstory: Psicólogo digital que entende como as pessoas leem online.
            """,
            tools=[],
            model=model
        )

class SEOOptimizer(Agent):
    def __init__(self):
        super().__init__(
            name="agente_seo",
            instruction="""
            Role: Engenheiro de Palavras-Chave
            Goal: Inserir palavras-chave estratégicas e garantir ranqueamento.
            Backstory: Ex-engenheiro do Google focado em métricas frias.
            """,
            tools=[],
            model=model
        )

class GEOOptimizer(Agent):
    def __init__(self):
        super().__init__(
            name="agente_geo",
            instruction="""
            Role: Adaptador de Localização
            Goal: Otimizar o texto para relevância geográfica e cultural.
            Backstory: Linguista e viajante especializado na comunicação eficiente nos diferentes territórios brasileiros.
            """,
            tools=[],
            model=model
        )
