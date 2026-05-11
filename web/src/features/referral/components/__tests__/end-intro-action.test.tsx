import { describe, it, expect, vi, beforeEach } from "vitest"
import { render, screen, fireEvent, waitFor } from "@testing-library/react"

import { ApiError } from "@/shared/lib/api-client"

vi.mock("next-intl", () => ({
  useTranslations:
    (namespace?: string) =>
    (key: string, params?: Record<string, string | number>) => {
      const full = namespace ? `${namespace}.${key}` : key
      return params ? `${full}(${JSON.stringify(params)})` : full
    },
}))

// Toast stub — assert calls + the .error sub-method.
vi.mock("sonner", () => {
  const fn = vi.fn() as unknown as ((msg: string) => void) & {
    error: (msg: string) => void
  }
  ;(fn as { error: (msg: string) => void }).error = vi.fn()
  return { toast: fn }
})
import { toast as mockToast } from "sonner"

// useEndIntroAttribution mock — full control of mutation lifecycle.
const mockMutate = vi.fn()
let mockIsPending = false
vi.mock("../../hooks/use-referrals", () => ({
  useEndIntroAttribution: () => ({
    mutate: mockMutate,
    isPending: mockIsPending,
  }),
}))

import { EndIntroAction } from "../end-intro-action"

beforeEach(() => {
  vi.clearAllMocks()
  mockIsPending = false
})

describe("EndIntroAction (Run C)", () => {
  it("renders the destructive trigger button by default", () => {
    render(<EndIntroAction attributionId="att-1" />)
    const btn = screen.getByTestId("end-intro-trigger")
    expect(btn).toBeInTheDocument()
    expect(btn.textContent).toContain("referralEndIntro.ctaLabel")
  })

  it("renders the badge directly when initialEndedAt is supplied", () => {
    render(
      <EndIntroAction
        attributionId="att-2"
        initialEndedAt="2026-05-11T10:00:00Z"
      />,
    )
    const badge = screen.getByTestId("end-intro-badge")
    expect(badge).toBeInTheDocument()
    // Date formatted dd/mm/yyyy.
    expect(badge.textContent).toContain("11/05/2026")
    // The trigger button is replaced.
    expect(
      screen.queryByTestId("end-intro-trigger"),
    ).not.toBeInTheDocument()
  })

  it("opens the modal on trigger click, then closes on Annuler without firing the mutation", () => {
    render(<EndIntroAction attributionId="att-3" />)
    fireEvent.click(screen.getByTestId("end-intro-trigger"))
    expect(screen.getByTestId("end-intro-confirm")).toBeInTheDocument()
    fireEvent.click(screen.getByTestId("end-intro-cancel"))
    expect(mockMutate).not.toHaveBeenCalled()
  })

  it("fires the mutation on Confirm and swaps to the badge on success", async () => {
    mockMutate.mockImplementation((id, opts) => {
      opts?.onSuccess?.({
        id,
        referral_id: "ref-x",
        proposal_id: "prop-x",
        ended_at: "2026-05-11T12:34:00Z",
      })
    })

    render(<EndIntroAction attributionId="att-4" />)
    fireEvent.click(screen.getByTestId("end-intro-trigger"))
    fireEvent.click(screen.getByTestId("end-intro-confirm"))

    await waitFor(() =>
      expect(screen.getByTestId("end-intro-badge")).toBeInTheDocument(),
    )
    expect(mockMutate).toHaveBeenCalledWith(
      "att-4",
      expect.any(Object),
    )
    // Trigger button is gone — replaced by the badge.
    expect(
      screen.queryByTestId("end-intro-trigger"),
    ).not.toBeInTheDocument()
  })

  it("surfaces a 403 ApiError via the forbidden toast", async () => {
    mockMutate.mockImplementation((_id, opts) => {
      const err = new ApiError(403, "forbidden", "forbidden")
      opts?.onError?.(err)
    })
    render(<EndIntroAction attributionId="att-5" />)
    fireEvent.click(screen.getByTestId("end-intro-trigger"))
    fireEvent.click(screen.getByTestId("end-intro-confirm"))
    await waitFor(() =>
      expect(
        (mockToast as unknown as { error: ReturnType<typeof vi.fn> }).error,
      ).toHaveBeenCalledWith("referralEndIntro.error.forbidden"),
    )
    // Trigger button stays — user can retry after fixing perms.
    expect(screen.getByTestId("end-intro-trigger")).toBeInTheDocument()
  })

  it("surfaces a 404 ApiError via the notFound toast", async () => {
    mockMutate.mockImplementation((_id, opts) => {
      const err = new ApiError(404, "not_found", "not found")
      opts?.onError?.(err)
    })
    render(<EndIntroAction attributionId="att-6" />)
    fireEvent.click(screen.getByTestId("end-intro-trigger"))
    fireEvent.click(screen.getByTestId("end-intro-confirm"))
    await waitFor(() =>
      expect(
        (mockToast as unknown as { error: ReturnType<typeof vi.fn> }).error,
      ).toHaveBeenCalledWith("referralEndIntro.error.notFound"),
    )
  })

  it("surfaces a network error via the generic toast", async () => {
    mockMutate.mockImplementation((_id, opts) => {
      opts?.onError?.(new Error("offline"))
    })
    render(<EndIntroAction attributionId="att-7" />)
    fireEvent.click(screen.getByTestId("end-intro-trigger"))
    fireEvent.click(screen.getByTestId("end-intro-confirm"))
    await waitFor(() =>
      expect(
        (mockToast as unknown as { error: ReturnType<typeof vi.fn> }).error,
      ).toHaveBeenCalledWith("referralEndIntro.error.generic"),
    )
  })
})
