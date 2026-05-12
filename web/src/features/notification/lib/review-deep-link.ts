// Review deep-link helper.
//
// Some notification types ("Mission terminée — laissez un avis",
// "review_received") must navigate to the messaging surface AND ask
// it to auto-open the review modal for the right proposal. The
// backend (dispute & proposal flows) ships the navigation hints in
// `notification.data`:
//
//   {
//     "proposal_id":     "<uuid>",
//     "conversation_id": "<uuid>",
//     "proposal_title":  "..."
//   }
//
// `buildReviewDeepLink` reads that payload and produces the canonical
// URL the messaging-page understands:
//
//   /messages?id=<convId>&openReview=1&reviewProposalId=<proposalId>
//
// Returns `null` when the notification kind is not review-related OR
// when the payload is missing/malformed — the caller then falls back
// to the legacy "mark as read only" behaviour.

import type { Notification, NotificationType } from "../types"

const REVIEW_KINDS: ReadonlySet<NotificationType> = new Set<NotificationType>([
  // Both completion flows (proposal + dispute) fire this kind with
  // the "leave a review" call-to-action body. Distinct from the
  // "review_received" kind which fires when the counterparty has
  // already submitted theirs.
  "proposal_completed",
  "review_received",
])

function readString(data: Record<string, unknown>, key: string): string | null {
  const value = data[key]
  return typeof value === "string" && value.length > 0 ? value : null
}

export function buildReviewDeepLink(notification: Notification): string | null {
  if (!REVIEW_KINDS.has(notification.type)) return null

  const data = (notification.data ?? {}) as Record<string, unknown>
  const conversationId = readString(data, "conversation_id")
  const proposalId = readString(data, "proposal_id")
  if (!conversationId || !proposalId) return null

  const params = new URLSearchParams({
    id: conversationId,
    openReview: "1",
    reviewProposalId: proposalId,
  })
  return `/messages?${params.toString()}`
}
