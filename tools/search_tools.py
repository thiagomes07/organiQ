import os
import json
import requests
from google.adk.tools import BaseTool

class SerperDevTool(BaseTool):
    def __init__(self):
        super().__init__(
            name="search_internet",
            description="Useful to search the internet for a given query. Returns the top results.",
        )
        self.api_key = os.getenv("SERPER_API_KEY")

    def run(self, query: str) -> str:
        """Searches the internet for the given query."""
        url = "https://google.serper.dev/search"
        payload = json.dumps({"q": query})
        headers = {
            'X-API-KEY': self.api_key,
            'Content-Type': 'application/json'
        }

        try:
            response = requests.request("POST", url, headers=headers, data=payload)
            response.raise_for_status()
            results = response.json()

            # Process and return relevant snippets
            organic = results.get("organic", [])
            output = []
            for result in organic[:5]:
                output.append(f"Title: {result.get('title')}\nLink: {result.get('link')}\nSnippet: {result.get('snippet')}\n")

            return "\n".join(output) if output else "No results found."

        except Exception as e:
            return f"Error performing search: {str(e)}"
