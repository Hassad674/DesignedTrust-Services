import { apiClient } from "@/shared/lib/api-client"
import type { Get, Post } from "@/shared/lib/api-paths"
import type { Review, AverageRating } from "@/shared/types/review"

// Re-export shared types for backward compatibility.
export type { Review, AverageRating }

export type ReviewListResponse = {
  data: Review[]
  next_cursor: string
  has_more: boolean
}

export type CanReviewResponse = {
  can_review: boolean
}

export type CreateReviewPayload = {
  proposal_id: string
  global_rating: number
  timeliness?: number
  communication?: number
  quality?: number
  comment?: string
  video_url?: string
  title_visible?: boolean
}

export async function fetchReviewsByUser(userId: string, cursor?: string) {
  const params = new URLSearchParams()
  if (cursor) params.set("cursor", cursor)
  const query = params.toString()
  const url = `/api/v1/reviews/user/${userId}${query ? `?${query}` : ""}`
  return apiClient<ReviewListResponse>(url)
}

export async function fetchAverageRating(userId: string) {
  return apiClient<Get<"/api/v1/reviews/average/{orgId}"> & { data: AverageRating }>(`/api/v1/reviews/average/${userId}`)
}

export async function fetchCanReview(proposalId: string) {
  return apiClient<Get<"/api/v1/reviews/can-review/{proposalId}"> & { data: CanReviewResponse }>(`/api/v1/reviews/can-review/${proposalId}`)
}

export async function createReview(payload: CreateReviewPayload) {
  return apiClient<Post<"/api/v1/reviews"> & { data: Review }>("/api/v1/reviews", {
    method: "POST",
    body: payload,
  })
}

import { uploadVideoDirect } from "@/shared/lib/upload/direct-video-upload"

// Review video — DIRECT-to-R2 presigned upload (presign + complete) so
// videos bypass the Vercel proxy ~4.5 MB body cap that 413'd large
// videos in production. /review-video/complete returns the public URL
// (carried into the review create payload as video_url) and fires the
// SAME moderation pipeline as the legacy multipart POST.
export async function uploadReviewVideo(file: File): Promise<string> {
  const { url } = await uploadVideoDirect<{ url: string }>(
    {
      presign: "/api/v1/upload/review-video/presign",
      complete: "/api/v1/upload/review-video/complete",
    },
    file,
  )
  return url
}
