import { Metadata } from 'next'

export const metadata: Metadata = {
  robots: {
    index: true,
    follow: true,
  },
}

export default function PublicLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <div className="min-h-screen bg-[var(--color-secondary-cream)]">
      {children}
    </div>
  )
}