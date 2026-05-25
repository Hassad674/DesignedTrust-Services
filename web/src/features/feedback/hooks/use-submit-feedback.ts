"use client"

import { useMutation } from "@tanstack/react-query"
import { submitFeedback } from "../api/feedback-api"
import type { SubmitFeedbackRequest, SubmitFeedbackResponse } from "../types"

/**
 * useSubmitFeedback wraps the POST /api/v1/feedback mutation. No cache
 * invalidation is needed — the web app has no feedback list to refresh
 * (triage lives in the admin app). The component reads `isPending`,
 * `isError`, and `isSuccess` to drive the modal's submit / error /
 * confirmation states.
 */
export function useSubmitFeedback() {
  return useMutation<SubmitFeedbackResponse, Error, SubmitFeedbackRequest>({
    mutationFn: (body) => submitFeedback(body),
  })
}
