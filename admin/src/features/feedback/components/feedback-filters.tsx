import { DataTableToolbar } from "@/shared/components/data-table/data-table-toolbar"
import { Select } from "@/shared/components/ui/select"
import type { FeedbackFilters } from "../types"
import { SEVERITY_OPTIONS, STATUS_OPTIONS, TYPE_OPTIONS } from "./feedback-labels"

type FeedbackFiltersBarProps = {
  filters: FeedbackFilters
  onChange: (filters: FeedbackFilters) => void
}

// FeedbackFiltersBar wires the type / status / severity selects and the
// title search input. Every change resets the cursor to "" so the
// caller restarts pagination from the first page (cursor pagination has
// no concept of "go to page N" after a filter change).
export function FeedbackFiltersBar({
  filters,
  onChange,
}: FeedbackFiltersBarProps) {
  return (
    <DataTableToolbar
      searchValue={filters.search}
      onSearchChange={(search) => onChange({ ...filters, search, cursor: "" })}
      searchPlaceholder="Rechercher par titre..."
    >
      <Select
        aria-label="Filtrer par type"
        options={TYPE_OPTIONS}
        placeholder="Tous les types"
        value={filters.type}
        onChange={(e) => onChange({ ...filters, type: e.target.value, cursor: "" })}
        className="w-40"
      />
      <Select
        aria-label="Filtrer par statut"
        options={STATUS_OPTIONS}
        placeholder="Tous les statuts"
        value={filters.status}
        onChange={(e) =>
          onChange({ ...filters, status: e.target.value, cursor: "" })
        }
        className="w-40"
      />
      <Select
        aria-label="Filtrer par gravité"
        options={SEVERITY_OPTIONS}
        placeholder="Toutes les gravités"
        value={filters.severity}
        onChange={(e) =>
          onChange({ ...filters, severity: e.target.value, cursor: "" })
        }
        className="w-44"
      />
    </DataTableToolbar>
  )
}
