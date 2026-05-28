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
  // score is the 0-100 WEIGHTED completion score — the exact number the
  // Typesense search gate filters on (`profile_completion_score:>=50`).
  // Distinct from `percent` (a section-count ratio for the checklist
  // UX). The UI must surface THIS number for the search-visibility
  // message so what the user sees matches what gates them — never
  // recompute a percentage locally.
  score: number
  // listed_in_search is true when the profile currently clears the
  // search-visibility threshold (score >= 50 AND published). Drives the
  // "visible / complete to 50%" message — read it straight off the
  // backend, do not derive it from `score` client-side.
  listed_in_search: boolean
}

// CompletionPersona is the optional override the caller passes when a
// provider_personal user wants the apporteur checklist (rendered on
// /referral) instead of the freelance one (rendered on /profile).
// `undefined` means "auto-select from the org type".
export type CompletionPersona = "freelance" | "referrer" | undefined

// getMyProfileCompletion fetches the completion report for the
// authenticated user's current organization. Backend resolves both
// user_id and organization_id from the session — only the optional
// persona override is forwarded as a query string. Unsupported
// personas (e.g. referrer for an enterprise org) silently fall back
// to the default persona on the server.
export async function getMyProfileCompletion(
  persona?: CompletionPersona,
): Promise<ProfileCompletionReport> {
  const path = persona
    ? `/api/v1/me/profile/completion?persona=${encodeURIComponent(persona)}`
    : "/api/v1/me/profile/completion"
  return apiClient<ProfileCompletionReport>(path)
}
