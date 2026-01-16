import type { Metadata } from 'next'
import { Toaster } from 'sonner'
import { Providers } from './providers'
import './globals.css'

// Fonts
import localFont from 'next/font/local'

const allRoundGothic = localFont({
  src: './fonts/AllRoundGothic-Medium.woff2',
  variable: '--font-all-round-gothic',
  weight: '500',
  display: 'swap',
})

const onest = localFont({
  src: [
    { path: './fonts/Onest-Thin.woff2', weight: '100', style: 'normal' },
    { path: './fonts/Onest-ExtraLight.woff2', weight: '200', style: 'normal' },
    { path: './fonts/Onest-Light.woff2', weight: '300', style: 'normal' },
    { path: './fonts/Onest-Regular.woff2', weight: '400', style: 'normal' },
    { path: './fonts/Onest-Medium.woff2', weight: '500', style: 'normal' },
    { path: './fonts/Onest-SemiBold.woff2', weight: '600', style: 'normal' },
    { path: './fonts/Onest-Bold.woff2', weight: '700', style: 'normal' },
    { path: './fonts/Onest-ExtraBold.woff2', weight: '800', style: 'normal' },
    { path: './fonts/Onest-Black.woff2', weight: '900', style: 'normal' },
  ],
  variable: '--font-onest',
  display: 'swap',
})

export const metadata: Metadata = {
  title: {
    default: 'organiQ - Aumente seu tráfego orgânico com IA',
    template: '%s | organiQ',
  },
  description: 'Matérias de blog que geram autoridade e SEO. Naturalmente Inteligente.',
  keywords: ['SEO', 'Marketing de Conteúdo', 'IA', 'WordPress', 'Blog', 'Tráfego Orgânico'],
  authors: [{ name: 'organiQ' }],
  creator: 'organiQ',
  publisher: 'organiQ',
  metadataBase: new URL(process.env.NEXT_PUBLIC_APP_URL || 'https://organiq.com.br'),
  
  openGraph: {
    type: 'website',
    locale: 'pt_BR',
    url: '/',
    siteName: 'organiQ',
    title: 'organiQ - Aumente seu tráfego orgânico com IA',
    description: 'Matérias de blog que geram autoridade e SEO. Naturalmente Inteligente.',
    images: [
      {
        url: '/og-image.jpg',
        width: 1200,
        height: 630,
        alt: 'organiQ - Naturalmente Inteligente',
      },
    ],
  },
  
  twitter: {
    card: 'summary_large_image',
    title: 'organiQ - Aumente seu tráfego orgânico com IA',
    description: 'Matérias de blog que geram autoridade e SEO. Naturalmente Inteligente.',
    images: ['/twitter-image.jpg'],
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
  
  manifest: '/manifest.json',
  
  icons: {
    icon: [
      { url: '/favicon-16x16.png', sizes: '16x16', type: 'image/png' },
      { url: '/favicon-32x32.png', sizes: '32x32', type: 'image/png' },
    ],
    apple: [
      { url: '/apple-touch-icon.png', sizes: '180x180', type: 'image/png' },
    ],
  },
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="pt-BR" className={`${allRoundGothic.variable} ${onest.variable}`}>
      <body className="antialiased" suppressHydrationWarning>
        <Providers>
          {children}
          <Toaster position="top-right" richColors closeButton />
        </Providers>
      </body>
    </html>
  )
}