import os
import re
import asyncio
from dotenv import load_dotenv
from google.adk.apps import App
from google.adk import Agent
from google.adk.models import Gemini
from google.adk.tools import AgentTool
from google.adk.runners import InMemoryRunner
from google.genai.types import Content
from agents.analysis_agents import CompetitorIdentifier, CompetitorScraper, GapIdentifier, Writer
from agents.gso_agents import OrchestratorGSO, AEOOptimizer, SEOOptimizer, GEOOptimizer
from config import get_model

load_dotenv()

# Definição dos agentes
competitor_identifier = CompetitorIdentifier()
competitor_scraper = CompetitorScraper()
gap_identifier = GapIdentifier()
writer = Writer()
orchestrator_gso = OrchestratorGSO()
aeo_optimizer = AEOOptimizer()
seo_optimizer = SEOOptimizer()
geo_optimizer = GEOOptimizer()

# Definição dos agentes como tools
tools = [
    AgentTool(agent=competitor_identifier),
    AgentTool(agent=competitor_scraper),
    AgentTool(agent=gap_identifier),
    AgentTool(agent=writer),
    AgentTool(agent=orchestrator_gso),
    AgentTool(agent=aeo_optimizer),
    AgentTool(agent=seo_optimizer),
    AgentTool(agent=geo_optimizer)
]

# Definição do agente root
root_agent = Agent(
    name="root_agent",
    instruction="""
    Você é um agente roteador. Seu objetivo é encaminhar as informações fornecidas pelo usuário para o agente especialista apropriado.

    Agentes Disponíveis:
    - identificador_concorrentes: Identifica concorrentes com base no nicho de mercado do website fornecido.
    - agente_scraping_concorrentes: Faz o scraping (coleta) de estratégias de concorrentes.
    - identificador_lacunas: Identifica lacunas de conteúdo com base nas estratégias coletadas.
    - escritor: Escreve conteúdo de um blog com base nas lacunas identificadas.
    - orq_gso: Orquestra a otimização GSO.
    - agente_aeo: Otimiza os conteúdos gerados para AEO.
    - agente_seo: Otimiza os conteúdos gerados para SEO.
    - agente_geo: Otimiza os conteúdos gerados para GEO.

    PROCEDIMENTO OPERACIONAL PADRÃO (POP) - ANÁLISE COMPLETA:
    Quando o usuário fornecer uma URL (ex: "[https://example.com](https://example.com)"), você **DEVE** executar os seguintes passos em ordem:
    1.  **Identificar Concorrentes**: Use o `identificador_concorrentes` para encontrar concorrentes para a URL fornecida.
    2.  **Coletar Estratégias**: Use o `agente_scraping_concorrentes` para analisar as estratégias dos concorrentes identificados.
    3.  **Identificar Lacunas**: Use o `identificador_lacunas` para encontrar lacunas de conteúdo com base nas estratégias coletadas.
    4.  **Escrever Rascunho**: Use o `escritor` para escrever um blog com conteúdo profundo sobre os temas das lacunas identificadas.
    5.  **Otimizar Conteúdo**: Use o `orq_gso` para otimizar os textos de blogs para AEO, SEO e GEO.

    Retorne os **TEXTOS FINAIS OTIMIZADOS** da etapa 5 ao usuário.

    **IMPORTANTE**: Separe cada um dos 3 blogs com a string exata: "---BLOG_SEPARATOR---".
    Não coloque nada antes do primeiro blog.
    Exemplo de output:
    [Conteúdo do Blog 1]
    ---BLOG_SEPARATOR---
    [Conteúdo do Blog 2]
    ---BLOG_SEPARATOR---
    [Conteúdo do Blog 3]
    """,
    tools=tools,
    model=get_model()
)

app = App(
    name="agents",
    root_agent=root_agent
)

async def main():
    print("Inicializando Agente Runner...")
    runner = InMemoryRunner(app=app)

    async with runner:
        session = await runner.session_service.create_session(
            app_name=app.name,
            user_id="user"
        )
        print(f"Sessão criada: {session.id}")
        print("Insira uma URL para iniciar a análise completa GSO (ou escreve 'exit' para sair).")

        while True:
            try:
                user_input = input("\nURL para análise: ")
                if user_input.lower() in ['exit', 'quit']:
                    break

                user_input = user_input.strip()
                main_url = user_input

                competitor_urls = []
                if input("Deseja inserir URLs de concorrentes? (s/n): ").lower() == 's':
                    for i in range(3):
                        comp_url = input(f"URL do concorrente {i+1} (deixe em branco para parar): ").strip()
                        if comp_url:
                            competitor_urls.append(comp_url)
                        else:
                            break

                preferred_blogs = []
                if input("Deseja inserir URLs de blogs de sua preferência? (s/n): ").lower() == 's':
                    for i in range(3):
                        blog_url = input(f"URL do blog preferido {i+1} (deixe em branco para parar): ").strip()
                        if blog_url:
                            preferred_blogs.append(blog_url)
                        else:
                            break

                prompt_parts = [f"\nRealize uma análise completa e otimização para esta URL: {main_url}"]
                if competitor_urls:
                    prompt_parts.append(f"\nConsidere também os seguintes concorrentes: {', '.join(competitor_urls)}")
                if preferred_blogs:
                    prompt_parts.append(f"\nE inspire-se nestes blogs de referência: {', '.join(preferred_blogs)}")

                prompt = " ".join(prompt_parts)

                # Create a directory for the URL
                sanitized_name = re.sub(r'[^a-zA-Z0-9]', '_', user_input)
                output_dir = os.path.abspath(os.path.join("output", sanitized_name))
                os.makedirs(output_dir, exist_ok=True)
                print(f"Diretório de saída: {output_dir}")

                full_response = ""

                async for event in runner.run_async(
                    user_id="user",
                    session_id=session.id,
                    new_message=Content(parts=[{"text": prompt}])
                ):
                    if event.content and event.content.parts:
                         for part in event.content.parts:
                             if part.text:
                                 print(f"{part.text}", end="", flush=True)
                                 full_response += part.text
                print()

                if full_response:
                    blogs = full_response.split("---BLOG_SEPARATOR---")

                    for i, blog_content in enumerate(blogs):
                        if not blog_content.strip():
                            continue

                        filename = os.path.join(output_dir, f"blog_{i+1}.md")
                        with open(filename, "w", encoding="utf-8") as f:
                            f.write(blog_content.strip())
                        print(f"Blog {i+1} salvo em: {filename}")

                    print(f"\nTodos os outputs salvos em: {output_dir}")

            except Exception as e:
                print(f"Erro: {e}")

if __name__ == "__main__":
    asyncio.run(main())
