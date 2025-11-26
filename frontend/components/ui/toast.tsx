/**
 * Toast Configuration for Sonner
 * 
 * Este arquivo configura o Sonner com as cores do projeto organiQ.
 * O Toaster já está incluído no layout principal (app/layout.tsx).
 * 
 * Para usar em qualquer componente:
 * 
 * import { toast } from 'sonner'
 * 
 * toast.success('Matéria publicada!')
 * toast.error('Erro ao salvar')
 * toast.warning('Atenção: limite atingido')
 * toast.info('Nova atualização disponível')
 * 
 * Com loading:
 * const toastId = toast.loading('Salvando...')
 * // ... operação async
 * toast.success('Salvo!', { id: toastId })
 * 
 * Com ação:
 * toast.success('Matéria criada!', {
 *   action: {
 *     label: 'Ver',
 *     onClick: () => router.push('/app/materias')
 *   }
 * })
 */

import { toast as sonnerToast } from 'sonner'

// Configurações padrão do toast
export const toastConfig = {
  position: 'top-right' as const,
  duration: 5000,
  richColors: true,
  closeButton: true,
  
  // Estilos customizados
  style: {
    fontFamily: 'var(--font-onest)',
  },
  
  // Classes para cada tipo
  classNames: {
    toast: 'font-onest',
    title: 'font-semibold',
    description: 'text-sm opacity-90',
    actionButton: 'bg-[var(--color-primary-purple)] text-white hover:opacity-90',
    cancelButton: 'bg-[var(--color-primary-dark)]/10 hover:bg-[var(--color-primary-dark)]/20',
    closeButton: 'bg-white hover:bg-[var(--color-primary-dark)]/5',
  }
}

// Re-exportar o toast do sonner com tipagem
export const toast = {
  success: (message: string, data?: Parameters<typeof sonnerToast.success>[1]) =>
    sonnerToast.success(message, data),
  
  error: (message: string, data?: Parameters<typeof sonnerToast.error>[1]) =>
    sonnerToast.error(message, data),
  
  warning: (message: string, data?: Parameters<typeof sonnerToast.warning>[1]) =>
    sonnerToast.warning(message, data),
  
  info: (message: string, data?: Parameters<typeof sonnerToast.info>[1]) =>
    sonnerToast.info(message, data),
  
  loading: (message: string, data?: Parameters<typeof sonnerToast.loading>[1]) =>
    sonnerToast.loading(message, data),
  
  promise: sonnerToast.promise,
  dismiss: sonnerToast.dismiss,
  custom: sonnerToast.custom,
}

// Helper para toast com promise
export const toastPromise = <T,>(
  promise: Promise<T>,
  messages: {
    loading: string
    success: string | ((data: T) => string)
    error: string | ((error: unknown) => string)
  }
) => {
  return sonnerToast.promise(promise, messages)
}

// Exemplos de uso:
export const toastExamples = {
  // Básico
  basic: () => {
    toast.success('Operação realizada com sucesso!')
  },
  
  // Com descrição
  withDescription: () => {
    toast.success('Matéria publicada!', {
      description: 'A matéria está disponível no seu blog WordPress'
    })
  },
  
  // Com ação
  withAction: () => {
    toast.success('Matéria criada!', {
      action: {
        label: 'Visualizar',
        onClick: () => console.log('Navegando...')
      }
    })
  },
  
  // Loading com atualização
  loadingUpdate: async () => {
    const toastId = toast.loading('Salvando alterações...')
    
    // Simular operação
    await new Promise(resolve => setTimeout(resolve, 2000))
    
    toast.success('Alterações salvas!', { id: toastId })
  },
  
  // Promise
  withPromise: async () => {
    const promise = new Promise((resolve, reject) => {
      setTimeout(() => Math.random() > 0.5 ? resolve('OK') : reject('Erro'), 2000)
    })
    
    toastPromise(promise, {
      loading: 'Processando...',
      success: 'Sucesso!',
      error: 'Falha ao processar'
    })
  }
}

export default toast