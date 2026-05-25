import { ChevronLeft, ChevronRight } from "lucide-react"
import { cn } from "@/shared/lib/utils"

type FeedbackPaginationProps = {
  count: number
  hasMore: boolean
  hasPrevious: boolean
  onNext: () => void
  onPrevious: () => void
}

// FeedbackPagination is a cursor-based prev/next pager (no page numbers,
// no offset). The shared DataTablePagination is offset/page based, so
// the feedback feature ships its own cursor pager — mirroring the
// invoices page's inline pager but extracted into a focused component.
export function FeedbackPagination({
  count,
  hasMore,
  hasPrevious,
  onNext,
  onPrevious,
}: FeedbackPaginationProps) {
  return (
    <div className="flex items-center justify-between px-1">
      <div className="text-sm text-muted-foreground">
        {count} résultat{count > 1 ? "s" : ""}
        {hasMore ? " (plus de résultats disponibles)" : ""}
      </div>
      <div className="flex items-center gap-2">
        <PagerButton
          label="Précédent"
          icon="left"
          disabled={!hasPrevious}
          onClick={onPrevious}
        />
        <PagerButton
          label="Suivant"
          icon="right"
          disabled={!hasMore}
          onClick={onNext}
        />
      </div>
    </div>
  )
}

function PagerButton({
  label,
  icon,
  disabled,
  onClick,
}: {
  label: string
  icon: "left" | "right"
  disabled: boolean
  onClick: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      className={cn(
        "inline-flex items-center gap-1 rounded-lg px-3 py-1.5 text-sm font-medium transition-all duration-200 ease-out",
        disabled
          ? "cursor-not-allowed text-muted-foreground/50"
          : "text-muted-foreground hover:bg-muted",
      )}
    >
      {icon === "left" && <ChevronLeft className="h-4 w-4" aria-hidden="true" />}
      {label}
      {icon === "right" && (
        <ChevronRight className="h-4 w-4" aria-hidden="true" />
      )}
    </button>
  )
}
