import * as React from 'react'
import * as SliderPrimitive from '@radix-ui/react-slider'
import { cn } from '@/lib/utils'

interface SliderProps
  extends React.ComponentPropsWithoutRef<typeof SliderPrimitive.Root> {
  showValue?: boolean
  formatValue?: (value: number) => string
}

const Slider = React.forwardRef<
  React.ElementRef<typeof SliderPrimitive.Root>,
  SliderProps
>(({ className, showValue, formatValue, value, ...props }, ref) => {
  const currentValue = Array.isArray(value) ? value[0] : value || 0
  const displayValue = formatValue ? formatValue(currentValue) : currentValue

  return (
    <div className="w-full space-y-2">
      {showValue && (
        <div className="flex items-center justify-between">
          <span className="text-sm font-medium font-onest text-[var(--color-primary-dark)]">
            Quantidade
          </span>
          <span className="text-sm font-semibold font-all-round text-[var(--color-primary-purple)]">
            {displayValue}
          </span>
        </div>
      )}
      <SliderPrimitive.Root
        ref={ref}
        className={cn(
          'relative flex w-full touch-none select-none items-center',
          className
        )}
        value={value}
        {...props}
      >
        <SliderPrimitive.Track className="relative h-2 w-full grow overflow-hidden rounded-full bg-[var(--color-primary-dark)]/10">
          <SliderPrimitive.Range className="absolute h-full bg-[var(--color-primary-purple)]" />
        </SliderPrimitive.Track>
        <SliderPrimitive.Thumb className="block h-5 w-5 rounded-full border-2 border-[var(--color-primary-purple)] bg-white ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 cursor-pointer hover:scale-110" />
      </SliderPrimitive.Root>
    </div>
  )
})
Slider.displayName = SliderPrimitive.Root.displayName

export { Slider }