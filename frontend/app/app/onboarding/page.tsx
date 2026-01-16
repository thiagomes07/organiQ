import { Metadata } from 'next'
import { OnboardingWizard } from '@/components/wizards/OnboardingWizard'

export const metadata: Metadata = {
  title: 'Configuração Inicial',
  description: 'Configure sua conta organiQ',
  robots: {
    index: false,
    follow: false,
  },
}

export default function OnboardingPage() {
  return (
    <div className="min-h-screen p-4 md:p-8">
      <div className="max-w-4xl mx-auto px-6">
        <OnboardingWizard />
      </div>
    </div>
  )
}