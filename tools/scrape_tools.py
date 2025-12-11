import requests
from bs4 import BeautifulSoup
from google.adk.tools import BaseTool

class ScrapeWebsiteTool(BaseTool):
    def __init__(self):
        super().__init__(
            name="scrape_website",
            description="Useful to scrape the content of a website given its URL.",
        )

    def run(self, url: str) -> str:
        """Scrapes the content of the given URL."""
        try:
            headers = {
                'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36'
            }
            response = requests.get(url, headers=headers, timeout=10)
            response.raise_for_status()

            soup = BeautifulSoup(response.content, 'html.parser')

            # Remove script and style elements
            for script in soup(["script", "style"]):
                script.decompose()

            text = soup.get_text()

            # Break into lines and remove leading/trailing space on each
            lines = (line.strip() for line in text.splitlines())
            # Break multi-headlines into a line each
            chunks = (phrase.strip() for line in lines for phrase in line.split("  "))
            # Drop blank lines
            text = '\n'.join(chunk for chunk in chunks if chunk)

            # Limit text length to avoid context window issues (approx 8000 chars)
            return text[:8000]

        except Exception as e:
            return f"Error scraping website: {str(e)}"
