import { adminApi } from "@/shared/lib/api-client"
import type {
  FeedbackFilters,
  FeedbackListResponse,
  FeedbackNote,
  FeedbackReportDetail,
  FeedbackReportRow,
  UpdateFeedbackPayload,
} from "../types"

// listFeedback fetches one page of feedback reports. Cursor pagination
// only — an empty cursor returns the first page. Filters that are empty
// strings are omitted from the query string so the URL stays minimal.
export function listFeedback(
  filters: FeedbackFilters,
): Promise<FeedbackListResponse> {
  const params = new URLSearchParams()
  if (filters.type) params.set("type", filters.type)
  if (filters.status) params.set("status", filters.status)
  if (filters.severity) params.set("severity", filters.severity)
  if (filters.search) params.set("search", filters.search)
  if (filters.cursor) params.set("cursor", filters.cursor)
  params.set("limit", "20")
  const qs = params.toString()
  return adminApi<FeedbackListResponse>(
    `/api/v1/admin/feedback${qs ? `?${qs}` : ""}`,
  )
}

// getFeedback fetches a single report with its attachments (each
// carrying a presigned GET url) and notes (newest first).
export function getFeedback(id: string): Promise<FeedbackReportDetail> {
  return adminApi<FeedbackReportDetail>(`/api/v1/admin/feedback/${id}`)
}

// updateFeedback patches the status and/or severity of a report.
export function updateFeedback(
  id: string,
  payload: UpdateFeedbackPayload,
): Promise<FeedbackReportRow> {
  return adminApi<FeedbackReportRow>(`/api/v1/admin/feedback/${id}`, {
    method: "PATCH",
    body: payload,
  })
}

// addFeedbackNote appends an internal admin note to a report and returns
// the created note.
export function addFeedbackNote(
  id: string,
  body: string,
): Promise<FeedbackNote> {
  return adminApi<FeedbackNote>(`/api/v1/admin/feedback/${id}/notes`, {
    method: "POST",
    body: { body },
  })
}
