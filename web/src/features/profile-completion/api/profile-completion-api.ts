import { apiClient } from "@/shared/lib/api-client"

// ProfileCompletionSection mirrors the backend's
// app/profilecompletion.Section type. The `key` is a stable machine
// identifier — never translate on the wire — and `label_key` is the
// i18n bucket the frontend looks up for the user-facing label.
// `completion_path` is the in-app URL the missing-list modal
// navigates to when the user clicks the section.
export type ProfileCompletionSection = {
  key: string
  filled: boolean
  label_key: string
  completion_path: string
}

// ProfileCompletionReport is the response payload of
// GET /api/v1/me/profile/completion. Sections are ordered by domain
// precedence (identity -> presentation -> offer -> compliance) so
// the frontend renders the missing-list in the same intuitive order
// without a client-side sort.
export type ProfileCompletionReport = {
  role: string
  persona: string
  percent: number
  total_sections: number
  filled_sections: number
  sections: ProfileCompletionSection[]
}

// getMyProfileCompletion fetches the completion report for the
// authenticated user's current organization. Backend resolves both
// user_id and organization_id from the session — no params needed.
export async function getMyProfileCompletion(): Promise<ProfileCompletionReport> {
  return apiClient<ProfileCompletionReport>("/api/v1/me/profile/completion")
}
