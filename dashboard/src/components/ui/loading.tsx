import { motion } from "framer-motion";

export function LoadingSkeleton({ className = "" }: { className?: string }) {
  return (
    <div className={`animate-pulse rounded-lg bg-card-border/50 ${className}`} />
  );
}

export function CardSkeleton() {
  return (
    <div className="rounded-xl border border-card-border bg-card p-5 space-y-4 animate-pulse">
      <div className="flex items-center justify-between">
        <div className="space-y-2">
          <div className="h-4 w-24 rounded bg-card-border/50" />
          <div className="h-3 w-16 rounded bg-card-border/50" />
        </div>
        <div className="h-10 w-10 rounded-full bg-card-border/50" />
      </div>
      <div className="space-y-2">
        <div className="h-3 w-full rounded bg-card-border/50" />
        <div className="h-3 w-3/4 rounded bg-card-border/50" />
        <div className="h-3 w-1/2 rounded bg-card-border/50" />
      </div>
    </div>
  );
}

export function StatsSkeleton() {
  return (
    <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
      {Array.from({ length: 4 }).map((_, i) => (
        <div key={i} className="rounded-xl border border-card-border bg-card p-4 animate-pulse">
          <div className="h-3 w-16 rounded bg-card-border/50 mb-2" />
          <div className="h-6 w-12 rounded bg-card-border/50" />
        </div>
      ))}
    </div>
  );
}

export function ListSkeleton({ count = 3 }: { count?: number }) {
  return (
    <div className="space-y-3">
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} className="rounded-lg border border-card-border bg-card p-4 animate-pulse">
          <div className="flex items-center gap-3">
            <div className="h-3 w-3 rounded-full bg-card-border/50" />
            <div className="flex-1 space-y-2">
              <div className="h-4 w-3/4 rounded bg-card-border/50" />
              <div className="h-3 w-1/2 rounded bg-card-border/50" />
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}

export function PageLoader() {
  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      className="flex items-center justify-center h-64"
    >
      <div className="flex flex-col items-center gap-3">
        <div className="h-8 w-8 border-2 border-accent border-t-transparent rounded-full animate-spin" />
        <span className="text-xs text-muted">Loading...</span>
      </div>
    </motion.div>
  );
}

export function EmptyState({ message, icon }: { message: string; icon?: React.ReactNode }) {
  return (
    <div className="flex flex-col items-center justify-center h-48 text-muted">
      {icon && <div className="mb-3 opacity-50">{icon}</div>}
      <p className="text-sm">{message}</p>
    </div>
  );
}

export function QueryError({ error, onRetry }: { error: Error | null; onRetry?: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center p-6 rounded-xl border border-danger/20 bg-danger/5">
      <p className="text-sm text-danger font-medium mb-1">Failed to load data</p>
      <p className="text-xs text-muted mb-3">{error?.message || "Unknown error"}</p>
      {onRetry && (
        <button
          onClick={onRetry}
          className="px-3 py-1.5 rounded-lg text-xs font-medium border border-card-border hover:border-accent transition-colors"
        >
          Retry
        </button>
      )}
    </div>
  );
}
