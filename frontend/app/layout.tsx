import type { Metadata, Viewport } from 'next'
import localFont from 'next/font/local'
import './globals.css'
import { Toaster } from 'sonner'
import { Providers } from './providers'

// Carregar fontes locais
const onest = localFont({
  src: './fonts/Onest.woff2',
  variable: '--font-onest-var',
  display: 'swap',
})

const allRoundGothic = localFont({
  src: './fonts/AllRoundGothic.woff2',
  variable: '--font-all-round-gothic',
  display: 'swap',
})

// Configurar Viewport (antes era no metadata)
export const viewport: Viewport = {
  themeColor: '#001d47',
  width: 'device-width',
  initialScale: 1,
  maximumScale: 5,
}

// Metadata principal
export const metadata: Metadata = {
  metadataBase: new URL(process.env.NEXT_PUBLIC_APP_URL || 'https://organiq.com.br'),
  
  title: {
    default: 'organiQ - Aumente seu tráfego orgânico com IA',
    template: '%s | organiQ'
  },
  
  description: 'Matérias de blog que geram autoridade e SEO. Naturalmente Inteligente. Geração automática de conteúdo otimizado para WordPress.',
  
  keywords: ['SEO', 'blog', 'IA', 'conteúdo automático', 'WordPress', 'marketing digital', 'tráfego orgânico'],
  
  authors: [{ name: 'organiQ' }],
  creator: 'organiQ',
  publisher: 'organiQ',
  
  formatDetection: {
    email: false,
    address: false,
    telephone: false,
  },
  
  openGraph: {
    type: 'website',
    locale: 'pt_BR',
    url: 'https://organiq.com.br',
    title: 'organiQ - Aumente seu tráfego orgânico com IA',
    description: 'Matérias de blog que geram autoridade e SEO. Naturalmente Inteligente.',
    siteName: 'organiQ',
    images: [
      {
        url: '/images/og-image.jpg',
        width: 1200,
        height: 630,
        alt: 'organiQ - Geração Automática de Conteúdo',
      },
    ],
  },
  
  twitter: {
    card: 'summary_large_image',
    title: 'organiQ - Aumente seu tráfego orgânico com IA',
    description: 'Matérias de blog que geram autoridade e SEO. Naturalmente Inteligente.',
    images: ['/images/twitter-image.jpg'],
    creator: '@organiq',
  },
  
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      'max-video-preview': -1,
      'max-image-preview': 'large',
      'max-snippet': -1,
    },
  },
  
  icons: {
    icon: [
      { url: '/favicon-16x16.png', sizes: '16x16', type: 'image/png' },
      { url: '/favicon-32x32.png', sizes: '32x32', type: 'image/png' },
    ],
    apple: [
      { url: '/apple-touch-icon.png', sizes: '180x180', type: 'image/png' },
    ],
  },
  
  manifest: '/manifest.json',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="pt-BR" className={`${onest.variable} ${allRoundGothic.variable}`}>
      <head>
        {/* Tags extras que não tem na Metadata API */}
        <link rel="preconnect" href={process.env.NEXT_PUBLIC_API_URL} />
        <link rel="dns-prefetch" href={process.env.NEXT_PUBLIC_API_URL} />
        
        {/* Google Analytics (se tiver) */}
        {process.env.NEXT_PUBLIC_GA_ID && (
          <>
            <script async src={`https://www.googletagmanager.com/gtag/js?id=${process.env.NEXT_PUBLIC_GA_ID}`} />
            <script
              dangerouslySetInnerHTML={{
                __html: `
                  window.dataLayer = window.dataLayer || [];
                  function gtag(){dataLayer.push(arguments);}
                  gtag('js', new Date());
                  gtag('config', '${process.env.NEXT_PUBLIC_GA_ID}');
                `,
              }}
            />
          </>
        )}
      </head>
      <body>
        <Providers>
          {children}
          <Toaster position="top-right" richColors />
        </Providers>
      </body>
    </html>
  )
}