"use client"

import { useQueryClient, useQuery } from "@tanstack/react-query"
import { useCurrentUserId } from "@/shared/hooks/use-current-user-id"
import {
  getMyProfileCompletion,
  type ProfileCompletionReport,
} from "../api/profile-completion-api"

// Query key convention follows the rest of the user-scoped surfaces
// — root namespace is ["user", uid, …] so a global "user logged out"
// invalidation can fan out without naming every feature explicitly.
export function profileCompletionQueryKey(uid: string | undefined) {
  return ["user", uid, "profile-completion"] as const
}

// useProfileCompletion reads the authenticated user's completion
// report. staleTime is 30 seconds — matches the backend's
// `Cache-Control: private, max-age=30`. After a write that affects
// completion (e.g. saving the profile expertise), callers SHOULD
// invalidate via useInvalidateProfileCompletion so the bar updates
// instantly instead of waiting for the staleTime window. As a
// belt-and-braces refresh path, the query also re-fetches when the
// user re-focuses the tab after navigating to an editor — that
// covers cross-tab edits and any mutation hook that has not been
// wired to the invalidator yet.
export function useProfileCompletion() {
  const uid = useCurrentUserId()
  return useQuery<ProfileCompletionReport>({
    queryKey: profileCompletionQueryKey(uid),
    queryFn: getMyProfileCompletion,
    staleTime: 30 * 1000,
    enabled: Boolean(uid),
    refetchOnWindowFocus: true,
  })
}

// useInvalidateProfileCompletion returns a stable invalidator the
// per-section save flows can call after a successful mutation. The
// hook lives here (not at every call site) so the query-key shape
// stays encapsulated.
export function useInvalidateProfileCompletion() {
  const queryClient = useQueryClient()
  const uid = useCurrentUserId()
  return () => {
    if (!uid) return
    queryClient.invalidateQueries({ queryKey: profileCompletionQueryKey(uid) })
  }
}
