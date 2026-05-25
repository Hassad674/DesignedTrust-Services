import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { MessageSquareWarning } from "lucide-react"
import { PageHeader } from "@/shared/components/layouts/page-header"
import { DataTable } from "@/shared/components/data-table/data-table"
import { EmptyState } from "@/shared/components/ui/empty-state"
import { TableSkeleton } from "@/shared/components/ui/skeleton"
import { useFeedbackList } from "../hooks/use-feedback"
import { EMPTY_FEEDBACK_FILTERS, type FeedbackFilters, type FeedbackReportRow } from "../types"
import { feedbackColumns } from "./feedback-columns"
import { FeedbackFiltersBar } from "./feedback-filters"
import { FeedbackPagination } from "./feedback-pagination"

// FeedbackPage lists bug & vulnerability reports with type/status/
// severity filters and cursor pagination. It owns the filter state and
// the cursor stack (push on next, pop on previous); the table, filters,
// and pager are dumb children.
export function FeedbackPage() {
  const navigate = useNavigate()
  const [filters, setFilters] = useState<FeedbackFilters>(EMPTY_FEEDBACK_FILTERS)
  const [cursorStack, setCursorStack] = useState<string[]>([])

  const { data, isLoading, isError } = useFeedbackList(filters)

  function handleFiltersChange(next: FeedbackFilters) {
    setFilters(next)
    setCursorStack([])
  }

  function handleNext() {
    if (!data?.next_cursor) return
    setCursorStack((stack) => [...stack, filters.cursor])
    setFilters((f) => ({ ...f, cursor: data.next_cursor }))
  }

  function handlePrevious() {
    setCursorStack((stack) => {
      if (stack.length === 0) return stack
      const previousCursor = stack[stack.length - 1] ?? ""
      setFilters((f) => ({ ...f, cursor: previousCursor }))
      return stack.slice(0, -1)
    })
  }

  function handleRowClick(report: FeedbackReportRow) {
    navigate(`/feedback/${report.id}`)
  }

  const rows = data?.data ?? []
  const hasPrevious = cursorStack.length > 0

  return (
    <div className="space-y-6">
      <PageHeader
        title="Signalements"
        description="Bugs et vulnérabilités remontés par les utilisateurs."
      />

      <FeedbackFiltersBar filters={filters} onChange={handleFiltersChange} />

      {isLoading ? (
        <TableSkeleton rows={8} cols={7} />
      ) : isError ? (
        <div className="rounded-xl border border-destructive/20 bg-destructive/5 p-6 text-center text-sm text-destructive">
          Erreur lors du chargement des signalements
        </div>
      ) : rows.length === 0 && !hasPrevious ? (
        <EmptyState
          icon={MessageSquareWarning}
          title="Aucun signalement"
          description="Aucun signalement ne correspond aux filtres sélectionnés."
        />
      ) : (
        <>
          <DataTable
            columns={feedbackColumns}
            data={rows}
            onRowClick={handleRowClick}
          />
          <FeedbackPagination
            count={rows.length}
            hasMore={data?.has_more ?? false}
            hasPrevious={hasPrevious}
            onNext={handleNext}
            onPrevious={handlePrevious}
          />
        </>
      )}
    </div>
  )
}
