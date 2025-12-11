import * as React from 'react'
import { cn } from '@/lib/utils'

export interface TextareaProps
  extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  error?: string
  showCount?: boolean
  maxLength?: number
}

const Textarea = React.forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ className, error, showCount, maxLength, value, ...props }, ref) => {
    const currentLength = typeof value === 'string' ? value.length : 0

    return (
      <div className="w-full">
        <textarea
          className={cn(
            'flex min-h-[80px] w-full rounded-[var(--radius-sm)] border border-input bg-white px-3 py-2 text-sm font-onest',
            'transition-colors duration-200',
            'placeholder:text-[var(--color-primary-dark)]/40',
            'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--color-primary-purple)] focus-visible:border-transparent',
            'disabled:cursor-not-allowed disabled:opacity-50',
            'resize-y',
            error && 'border-[var(--color-error)] focus-visible:ring-[var(--color-error)]',
            className
          )}
          ref={ref}
          maxLength={maxLength}
          value={value}
          {...props}
        />
        <div className="mt-1 flex items-center justify-between">
          {error ? (
            <p className="text-xs text-[var(--color-error)] font-onest">
              {error}
            </p>
          ) : (
            <span />
          )}
          {showCount && maxLength && (
            <p className="text-xs text-[var(--color-primary-dark)]/60 font-onest">
              {currentLength} / {maxLength}
            </p>
          )}
        </div>
      </div>
    )
  }
)
Textarea.displayName = 'Textarea'

export { Textarea }