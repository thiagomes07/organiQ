// app/(public)/login/page.tsx
import { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Login',
  description: 'Faça login na sua conta organiQ',
  robots: {
    index: false, // Não indexar página de login
    follow: false,
  }
}

export default function LoginPage() {
  return <div>Login</div>
}