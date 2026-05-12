import { describe, it, expect } from "vitest"
import { buildReviewDeepLink } from "../review-deep-link"
import type { Notification, NotificationType } from "../../types"

function makeNotification(
  type: NotificationType,
  data: Record<string, unknown>,
): Notification {
  return {
    id: "notif-1",
    user_id: "user-1",
    type,
    title: "Whatever",
    body: "Whatever",
    data,
    read_at: null,
    created_at: new Date().toISOString(),
  }
}

const PROPOSAL_ID = "11111111-1111-1111-1111-111111111111"
const CONVERSATION_ID = "22222222-2222-2222-2222-222222222222"

describe("buildReviewDeepLink", () => {
  describe("review-related notification kinds", () => {
    const reviewKinds: NotificationType[] = [
      "proposal_completed",
      "review_received",
    ]

    it.each(reviewKinds)(
      "builds /messages?id=…&openReview=1&reviewProposalId=… for kind %s",
      (kind) => {
        const url = buildReviewDeepLink(
          makeNotification(kind, {
            proposal_id: PROPOSAL_ID,
            conversation_id: CONVERSATION_ID,
            proposal_title: "Site web Acme",
          }),
        )
        expect(url).not.toBeNull()
        const parsed = new URL(`http://x${url!}`)
        expect(parsed.pathname).toBe("/messages")
        expect(parsed.searchParams.get("id")).toBe(CONVERSATION_ID)
        expect(parsed.searchParams.get("openReview")).toBe("1")
        expect(parsed.searchParams.get("reviewProposalId")).toBe(PROPOSAL_ID)
      },
    )
  })

  describe("non-review notification kinds", () => {
    const otherKinds: NotificationType[] = [
      "proposal_received",
      "proposal_accepted",
      "proposal_declined",
      "proposal_modified",
      "proposal_paid",
      "completion_requested",
      "new_message",
      "system_announcement",
    ]

    it.each(otherKinds)(
      "returns null for kind %s (no deep-link)",
      (kind) => {
        const url = buildReviewDeepLink(
          makeNotification(kind, {
            proposal_id: PROPOSAL_ID,
            conversation_id: CONVERSATION_ID,
          }),
        )
        expect(url).toBeNull()
      },
    )
  })

  it("returns null when proposal_id is missing", () => {
    const url = buildReviewDeepLink(
      makeNotification("proposal_completed", {
        conversation_id: CONVERSATION_ID,
      }),
    )
    expect(url).toBeNull()
  })

  it("returns null when conversation_id is missing", () => {
    const url = buildReviewDeepLink(
      makeNotification("proposal_completed", {
        proposal_id: PROPOSAL_ID,
      }),
    )
    expect(url).toBeNull()
  })

  it("returns null when data is empty (legacy dispute payload)", () => {
    // Regression guard: pre-fix, dispute resolution notifications
    // shipped only { dispute_id: "..." } — the new payload aligns
    // with the proposal flow, but in case an older client still
    // surfaces a stale notification we should not crash, just no-op.
    const url = buildReviewDeepLink(
      makeNotification("proposal_completed", {
        dispute_id: "00000000-0000-0000-0000-000000000000",
      }),
    )
    expect(url).toBeNull()
  })

  it("returns null when data is malformed", () => {
    const url = buildReviewDeepLink(
      makeNotification("proposal_completed", {
        proposal_id: 12345,
        conversation_id: { x: 1 },
      }),
    )
    expect(url).toBeNull()
  })
})
