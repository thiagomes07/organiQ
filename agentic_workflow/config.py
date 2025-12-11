from google.adk.models import Gemini
import os
from dotenv import load_dotenv

load_dotenv()

# Nome do modelo a ser utilizado em todo o projeto.
# Opções comuns: "gemini-1.5-flash", "gemini-1.5-pro"
MODEL_NAME = "gemini-1.5-flash"

def get_model():
    return Gemini(model_name=MODEL_NAME)
