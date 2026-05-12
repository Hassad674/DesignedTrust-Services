/**
 * Notification item navigation tests.
 *
 * The notification row is a button: on click it (1) marks the item as
 * read and (2) — for review-related kinds — navigates to the messaging
 * deep-link that auto-opens the review modal. Tests pin both
 * behaviours via mocked navigation + mocked mark-as-read mutation.
 */
import { describe, it, expect, vi, beforeEach } from "vitest"
import { render, screen, fireEvent } from "@testing-library/react"
import { NextIntlClientProvider } from "next-intl"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { createElement } from "react"

import { NotificationItem } from "../notification-item"
import type { Notification, NotificationType } from "../../types"

const mockPush = vi.fn()
vi.mock("@i18n/navigation", () => ({
  useRouter: () => ({ push: mockPush, replace: vi.fn() }),
  Link: ({ children }: { children: React.ReactNode }) => children,
}))

const mockMutate = vi.fn()
vi.mock("../../hooks/use-notification-actions", () => ({
  useMarkAsRead: () => ({ mutate: mockMutate }),
}))

const messages = {
  notifications: {
    timeJustNow: "à l'instant",
    timeMinutes: "il y a {n} min",
    timeHours: "il y a {n} h",
    timeDays: "il y a {n} j",
  },
}

function wrapper({ children }: { children: React.ReactNode }) {
  const qc = new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0 },
      mutations: { retry: false },
    },
  })
  return createElement(
    QueryClientProvider,
    { client: qc },
    createElement(
      NextIntlClientProvider,
      { locale: "fr", messages },
      children,
    ),
  )
}

function makeNotification(
  type: NotificationType,
  data: Record<string, unknown> = {},
  opts: Partial<Notification> = {},
): Notification {
  return {
    id: "n-1",
    user_id: "u-1",
    type,
    title: "Some title",
    body: "Some body",
    data,
    read_at: null,
    created_at: new Date().toISOString(),
    ...opts,
  }
}

beforeEach(() => {
  mockPush.mockReset()
  mockMutate.mockReset()
})

describe("NotificationItem — click handling", () => {
  it("review notification → marks-as-read AND deep-links to messages with openReview=1", () => {
    const notif = makeNotification("proposal_completed", {
      proposal_id: "p-1",
      conversation_id: "c-1",
      proposal_title: "Site web Acme",
    })
    render(<NotificationItem notification={notif} />, { wrapper })

    fireEvent.click(screen.getByRole("button"))

    expect(mockMutate).toHaveBeenCalledWith("n-1")
    expect(mockPush).toHaveBeenCalledTimes(1)
    const url = mockPush.mock.calls[0][0] as string
    expect(url).toContain("/messages")
    const parsed = new URL(`http://x${url}`)
    expect(parsed.searchParams.get("id")).toBe("c-1")
    expect(parsed.searchParams.get("openReview")).toBe("1")
    expect(parsed.searchParams.get("reviewProposalId")).toBe("p-1")
  })

  it("review_received notification → also deep-links with openReview=1", () => {
    const notif = makeNotification("review_received", {
      proposal_id: "p-2",
      conversation_id: "c-2",
    })
    render(<NotificationItem notification={notif} />, { wrapper })

    fireEvent.click(screen.getByRole("button"))
    expect(mockPush).toHaveBeenCalledTimes(1)
  })

  it("non-review notification (proposal_accepted) → only marks-as-read, no navigation", () => {
    const notif = makeNotification("proposal_accepted", {
      proposal_id: "p-3",
      conversation_id: "c-3",
    })
    render(<NotificationItem notification={notif} />, { wrapper })

    fireEvent.click(screen.getByRole("button"))
    expect(mockMutate).toHaveBeenCalledWith("n-1")
    expect(mockPush).not.toHaveBeenCalled()
  })

  it("review notification with stale payload (only dispute_id) → marks-as-read, no navigation", () => {
    // Regression: the legacy dispute notification payload would only
    // carry dispute_id. A stale notification arriving post-fix must
    // not crash the click handler.
    const notif = makeNotification("proposal_completed", {
      dispute_id: "d-1",
    })
    render(<NotificationItem notification={notif} />, { wrapper })

    fireEvent.click(screen.getByRole("button"))
    expect(mockMutate).toHaveBeenCalledWith("n-1")
    expect(mockPush).not.toHaveBeenCalled()
  })

  it("already-read notification → still navigates if eligible but does NOT call markAsRead", () => {
    const notif = makeNotification(
      "proposal_completed",
      { proposal_id: "p-1", conversation_id: "c-1" },
      { read_at: new Date().toISOString() },
    )
    render(<NotificationItem notification={notif} />, { wrapper })

    fireEvent.click(screen.getByRole("button"))
    expect(mockMutate).not.toHaveBeenCalled()
    expect(mockPush).toHaveBeenCalledTimes(1)
  })
})
