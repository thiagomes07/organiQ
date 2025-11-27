import type { Metadata, Viewport } from 'next'
import localFont from 'next/font/local'
import './globals.css'
import { Toaster } from 'sonner'
import { Providers } from './providers'

// Configuração da Fonte Secundária (Corpo/Texto) - Onest
// Mapeando todos os arquivos específicos que você baixou para seus respectivos pesos css
const onest = localFont({
  src: [
    {
      path: './fonts/Onest-Thin.woff2',
      weight: '100',
      style: 'normal',
    },
    {
      path: './fonts/Onest-ExtraLight.woff2',
      weight: '200',
      style: 'normal',
    },
    {
      path: './fonts/Onest-Light.woff2',
      weight: '300',
      style: 'normal',
    },
    {
      path: './fonts/Onest-Regular.woff2',
      weight: '400',
      style: 'normal',
    },
    {
      path: './fonts/Onest-Medium.woff2',
      weight: '500',
      style: 'normal',
    },
    {
      path: './fonts/Onest-SemiBold.woff2',
      weight: '600',
      style: 'normal',
    },
    {
      path: './fonts/Onest-Bold.woff2',
      weight: '700',
      style: 'normal',
    },
    {
      path: './fonts/Onest-ExtraBold.woff2',
      weight: '800',
      style: 'normal',
    },
    {
      path: './fonts/Onest-Black.woff2',
      weight: '900',
      style: 'normal',
    },
  ],
  variable: '--font-onest-var',
  display: 'swap',
})

// Configuração da Fonte Primária (Títulos) - All Round Gothic
// Como temos apenas o arquivo Medium, definimos ele como a fonte padrão para esta família
const allRoundGothic = localFont({
  src: [
    {
      path: './fonts/AllRoundGothic-Medium.woff2',
      weight: '500', // Peso nativo da fonte (Medium)
      style: 'normal',
    },
  ],
  variable: '--font-all-round-gothic',
  display: 'swap',
})

// Configurar Viewport
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