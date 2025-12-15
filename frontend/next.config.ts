import type { NextConfig } from 'next'

const nextConfig: NextConfig = {
  // --- OUTPUT STANDALONE ---
  // Otimiza o build para containers Docker (~150MB vs ~1GB)
  output: 'standalone',
  
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