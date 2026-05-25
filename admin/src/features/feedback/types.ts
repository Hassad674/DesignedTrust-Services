import type { components } from "@/shared/types/api"

// The feedback feature consumes the backend's generated OpenAPI schemas
// directly so the types never drift from the contract. We re-export the
// generated shapes under feature-local aliases and add the UI-only
// filter + list-envelope types the generator does not describe (the
// admin list endpoint returns a flat `{ data, next_cursor, has_more }`
// envelope built ad-hoc in the handler, not a documented schema).

export type FeedbackReportRow =
  components["schemas"]["AdminFeedbackReportResponse"]

export type FeedbackReportDetail =
  components["schemas"]["AdminFeedbackReportDetailResponse"]

export type FeedbackAttachment = FeedbackReportDetail["attachments"][number]

export type FeedbackNote = FeedbackReportDetail["notes"][number]

// FeedbackType / FeedbackStatus / FeedbackSeverity narrow the loosely
// typed `string` fields the backend emits to the closed sets the domain
// enforces, so the UI can exhaustively map labels + badges.
export type FeedbackType = "bug" | "vulnerability"

export type FeedbackStatus =
  | "new"
  | "triaged"
  | "in_progress"
  | "resolved"
  | "rejected"

export type FeedbackSeverity = "low" | "medium" | "high" | "critical"

// FeedbackListResponse is the flat envelope the admin list handler
// returns. Cursor-only — no offset, no total. `has_more` is true when a
// `next_cursor` is present.
export type FeedbackListResponse = {
  data: FeedbackReportRow[]
  next_cursor: string
  has_more: boolean
}

// FeedbackFilters drives the list query. All fields are optional on the
// wire; an empty string means "no filter". `cursor` empty = first page.
export type FeedbackFilters = {
  type: string
  status: string
  severity: string
  search: string
  cursor: string
}

export const EMPTY_FEEDBACK_FILTERS: FeedbackFilters = {
  type: "",
  status: "",
  severity: "",
  search: "",
  cursor: "",
}

// UpdateFeedbackPayload mirrors the PATCH body. Both fields optional —
// the handler only touches the ones present.
export type UpdateFeedbackPayload = {
  status?: FeedbackStatus
  severity?: FeedbackSeverity
}
