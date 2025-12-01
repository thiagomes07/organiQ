import type { NextConfig } from 'next'

const nextConfig: NextConfig = {
  // --- REMOVIDO O OUTPUT STANDALONE ---
  // O Amplify vai gerenciar o servidor automaticamente
  
  // --- CONFIGURAÇÕES DE BUILD ---
  typescript: {
    ignoreBuildErrors: true,
  },
  
  // --- IMAGENS ---
  images: {
    remotePatterns: [
      {
        protocol: 'https',
        hostname: '**.organiq.com.br',
      },
      {
        protocol: 'https',
        hostname: '**.amplifyapp.com',
      },
    ],
  }, 

  // --- OTIMIZAÇÕES ---
  compiler: {
    removeConsole: process.env.NODE_ENV === 'production',
  },

  // --- CABEÇALHOS DE SEGURANÇA ---
  async headers() {
    return [
      {
        source: '/:path*',
        headers: [
          {
            key: 'X-DNS-Prefetch-Control',
            value: 'on'
          },
          {
            key: 'Strict-Transport-Security',
            value: 'max-age=63072000; includeSubDomains; preload'
          },
          {
            key: 'X-Frame-Options',
            value: 'SAMEORIGIN'
          },
          {
            key: 'X-Content-Type-Options',
            value: 'nosniff'
          },
          {
            key: 'X-XSS-Protection',
            value: '1; mode=block'
          },
          {
            key: 'Referrer-Policy',
            value: 'origin-when-cross-origin'
          }
        ]
      }
    ]
  }
}

export default nextConfig