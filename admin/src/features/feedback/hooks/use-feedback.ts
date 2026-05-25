import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  addFeedbackNote,
  getFeedback,
  listFeedback,
  updateFeedback,
} from "../api/feedback-api"
import type { FeedbackFilters, UpdateFeedbackPayload } from "../types"

export function feedbackQueryKey(filters: FeedbackFilters) {
  return ["admin", "feedback", filters] as const
}

export function feedbackDetailQueryKey(id: string) {
  return ["admin", "feedback", id] as const
}

export function useFeedbackList(filters: FeedbackFilters) {
  return useQuery({
    queryKey: feedbackQueryKey(filters),
    queryFn: () => listFeedback(filters),
    staleTime: 30 * 1000,
  })
}

export function useFeedbackDetail(id: string) {
  return useQuery({
    queryKey: feedbackDetailQueryKey(id),
    queryFn: () => getFeedback(id),
    enabled: !!id,
    staleTime: 30 * 1000,
  })
}

// useInvalidateFeedback refreshes both the detail of the affected report
// and every cached list page so a status/severity change or new note is
// reflected immediately wherever the report is shown.
function useInvalidateFeedback(id: string) {
  const queryClient = useQueryClient()
  return () => {
    queryClient.invalidateQueries({ queryKey: feedbackDetailQueryKey(id) })
    queryClient.invalidateQueries({ queryKey: ["admin", "feedback"] })
  }
}

export function useUpdateFeedback(id: string) {
  const invalidate = useInvalidateFeedback(id)
  return useMutation({
    mutationFn: (payload: UpdateFeedbackPayload) => updateFeedback(id, payload),
    onSuccess: invalidate,
  })
}

export function useAddFeedbackNote(id: string) {
  const invalidate = useInvalidateFeedback(id)
  return useMutation({
    mutationFn: (body: string) => addFeedbackNote(id, body),
    onSuccess: invalidate,
  })
}
