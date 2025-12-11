import { MetadataRoute } from 'next'

/**
 * Sitemap Generator
 * 
 * Gera sitemap.xml automaticamente para SEO
 * Next.js 14+ suporta geração dinâmica de sitemap
 */
export default function sitemap(): MetadataRoute.Sitemap {
  const baseUrl = process.env.NEXT_PUBLIC_APP_URL || 'https://organiq.com.br'
  
  // Data atual para lastModified
  const now = new Date()

  return [
    {
      url: baseUrl,
      lastModified: now,
      changeFrequency: 'weekly',
      priority: 1,
    },
    {
      url: `${baseUrl}/login`,
      lastModified: now,
      changeFrequency: 'monthly',
      priority: 0.8,
    },
    // Rotas protegidas não devem estar no sitemap público
    // pois requerem autenticação
  ]
}