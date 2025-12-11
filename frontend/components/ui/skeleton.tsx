import { cn } from '@/lib/utils'

function Skeleton({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn(
        'animate-pulse rounded-[var(--radius-sm)] bg-[var(--color-primary-dark)]/10',
        className
      )}
      {...props}
    />
  )
}

// Skeleton pré-definidos para casos comuns
const SkeletonCard = () => (
  <div className="space-y-3 p-4">
    <Skeleton className="h-4 w-3/4" />
    <Skeleton className="h-4 w-1/2" />
    <Skeleton className="h-20 w-full" />
  </div>
)

const SkeletonTable = ({ rows = 5 }: { rows?: number }) => (
  <div className="space-y-2">
    {Array.from({ length: rows }).map((_, i) => (
      <div key={i} className="flex items-center gap-4">
        <Skeleton className="h-12 w-full" />
      </div>
    ))}
  </div>
)

const SkeletonButton = () => <Skeleton className="h-10 w-24" />

const SkeletonAvatar = () => <Skeleton className="h-12 w-12 rounded-full" />

const SkeletonInput = () => <Skeleton className="h-10 w-full" />

const SkeletonText = ({ lines = 3 }: { lines?: number }) => (
  <div className="space-y-2">
    {Array.from({ length: lines }).map((_, i) => (
      <Skeleton
        key={i}
        className={cn(
          'h-4',
          i === lines - 1 ? 'w-2/3' : 'w-full' // Última linha mais curta
        )}
      />
    ))}
  </div>
)

export {
  Skeleton,
  SkeletonCard,
  SkeletonTable,
  SkeletonButton,
  SkeletonAvatar,
  SkeletonInput,
  SkeletonText,
}