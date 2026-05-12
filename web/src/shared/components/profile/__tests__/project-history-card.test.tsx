/**
 * ProjectHistoryCard tests — pins the three states of the "awaiting
 * review" placeholder:
 *
 *  - viewer is not a counterparty (no onOpenReview prop) → static
 *    placeholder, no can-review probe fired.
 *  - viewer is a counterparty AND can_review=true → clickable button
 *    that calls onOpenReview.
 *  - viewer is a counterparty AND can_review=false → static
 *    placeholder (already reviewed or window closed).
 *
 *  AND: when the entry already has an embedded review, the placeholder
 *  never shows up regardless of can-review.
 */
import { describe, it, expect, vi, beforeEach } from "vitest"
import { render, screen, fireEvent } from "@testing-library/react"
import { NextIntlClientProvider } from "next-intl"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"

import { ProjectHistoryCard, type ProjectHistoryCardEntry } from "../project-history-card"
import type { Review } from "@/shared/types/review"

const mockCanReview = vi.fn()
vi.mock("@/shared/hooks/review/use-reviews", () => ({
  useCanReview: (proposalId: string | undefined) => mockCanReview(proposalId),
}))

vi.mock("@/shared/components/ui/review-card", () => ({
  ReviewCard: ({ review }: { review: Review }) => (
    <div data-testid="review-card">{review.id}</div>
  ),
}))

const messages = {
  profile: {
    awaitingReview: "En attente d'avis",
    leaveYourReview: "Laisser ton avis sur ce projet",
    completedOn: "Terminée le {date}",
  },
}

function wrapper({ children }: { children: React.ReactNode }) {
  const qc = new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0 },
      mutations: { retry: false },
    },
  })
  return (
    <QueryClientProvider client={qc}>
      <NextIntlClientProvider locale="fr" messages={messages}>
        {children}
      </NextIntlClientProvider>
    </QueryClientProvider>
  )
}

const baseEntry: ProjectHistoryCardEntry = {
  proposal_id: "p-1",
  title: "Refonte site web",
  amount: 350000,
  currency: "EUR",
  completed_at: "2026-04-01T10:00:00Z",
  review: null,
}

beforeEach(() => {
  mockCanReview.mockReset()
})

describe("ProjectHistoryCard — review entry point", () => {
  it("no onOpenReview prop → static placeholder, never calls useCanReview with a real id", () => {
    mockCanReview.mockReturnValue({ data: undefined, isLoading: false })
    render(<ProjectHistoryCard entry={baseEntry} />, { wrapper })

    expect(screen.getByText("En attente d'avis")).toBeInTheDocument()
    // The probe is called with undefined so the underlying query is
    // disabled — anonymous visitors never trigger an auth probe.
    expect(mockCanReview).toHaveBeenCalledWith(undefined)
    // No clickable button.
    expect(screen.queryByRole("button")).not.toBeInTheDocument()
  })

  it("onOpenReview + can_review=true → clickable 'Laisser ton avis' button, fires callback", () => {
    mockCanReview.mockReturnValue({
      data: { data: { can_review: true } },
      isLoading: false,
    })
    const onOpenReview = vi.fn()
    render(
      <ProjectHistoryCard entry={baseEntry} onOpenReview={onOpenReview} />,
      { wrapper },
    )

    const button = screen.getByRole("button", { name: "Laisser ton avis sur ce projet" })
    expect(button).toBeInTheDocument()
    fireEvent.click(button)
    expect(onOpenReview).toHaveBeenCalledWith(baseEntry)
  })

  it("onOpenReview + can_review=false → static placeholder (already reviewed / window closed)", () => {
    mockCanReview.mockReturnValue({
      data: { data: { can_review: false } },
      isLoading: false,
    })
    const onOpenReview = vi.fn()
    render(
      <ProjectHistoryCard entry={baseEntry} onOpenReview={onOpenReview} />,
      { wrapper },
    )

    expect(screen.getByText("En attente d'avis")).toBeInTheDocument()
    expect(
      screen.queryByRole("button", { name: "Laisser ton avis sur ce projet" }),
    ).not.toBeInTheDocument()
  })

  it("onOpenReview but data still loading → falls back to static placeholder", () => {
    mockCanReview.mockReturnValue({ data: undefined, isLoading: true })
    const onOpenReview = vi.fn()
    render(
      <ProjectHistoryCard entry={baseEntry} onOpenReview={onOpenReview} />,
      { wrapper },
    )
    expect(screen.getByText("En attente d'avis")).toBeInTheDocument()
    expect(screen.queryByRole("button")).not.toBeInTheDocument()
  })

  it("entry has an embedded review → renders ReviewCard, ignores can-review entirely", () => {
    mockCanReview.mockReturnValue({
      data: { data: { can_review: true } },
      isLoading: false,
    })
    const entryWithReview: ProjectHistoryCardEntry = {
      ...baseEntry,
      review: {
        id: "r-1",
        proposal_id: "p-1",
        reviewer_id: "u-1",
        reviewed_id: "u-2",
        side: "client_to_provider",
        global_rating: 5,
        comment: "Très bon travail.",
        video_url: "",
        title_visible: true,
        created_at: "2026-04-15T10:00:00Z",
        published_at: "2026-04-15T10:00:00Z",
      } as unknown as Review,
    }
    render(
      <ProjectHistoryCard entry={entryWithReview} onOpenReview={vi.fn()} />,
      { wrapper },
    )

    expect(screen.getByTestId("review-card")).toBeInTheDocument()
    expect(screen.queryByText("En attente d'avis")).not.toBeInTheDocument()
    expect(screen.queryByRole("button")).not.toBeInTheDocument()
  })
})
