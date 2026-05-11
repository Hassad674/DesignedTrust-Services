import { describe, it, expect, vi, beforeEach } from "vitest"
import { render, screen, fireEvent, waitFor } from "@testing-library/react"

// next-intl echo stub.
vi.mock("next-intl", () => ({
  useTranslations:
    (namespace?: string) =>
    (key: string, params?: Record<string, string | number>) => {
      const full = namespace ? `${namespace}.${key}` : key
      return params ? `${full}(${JSON.stringify(params)})` : full
    },
}))

// useWalletSummary returns whatever the test prescribes per render.
const mockSummary = vi.fn()
vi.mock("../../hooks/use-wallet", () => ({
  useWalletSummary: (cursor?: string) => mockSummary(cursor),
}))

import { WalletUnifiedHistory } from "../wallet-unified-history"

const emptyLeg = {
  total_cents: 0,
  available_cents: 0,
  escrowed_cents: 0,
  transmitted_cents: 0,
}

function summaryPage(
  rows: Array<{
    type: "mission" | "commission"
    amount_cents: number
    status: string
    occurred_at: string
    reference_id: string
    mission_title?: string
  }>,
  nextCursor?: string,
) {
  return {
    data: {
      currency: "EUR",
      total_cents: 0,
      available_cents: 0,
      escrowed_cents: 0,
      transmitted_cents: 0,
      breakdown: { missions: emptyLeg, commissions: emptyLeg },
      recent_transactions: rows.map((r) => ({ currency: "EUR", ...r })),
      next_cursor: nextCursor,
    },
    isLoading: false,
    isError: false,
  }
}

beforeEach(() => {
  vi.clearAllMocks()
})

describe("WalletUnifiedHistory (Run C)", () => {
  it("renders the mission + commission rows with correct labels and amounts", () => {
    mockSummary.mockReturnValue(
      summaryPage([
        {
          type: "mission",
          amount_cents: 50_000,
          status: "paid",
          occurred_at: "2026-05-10T12:00:00Z",
          reference_id: "rec-1",
          mission_title: "Logo design",
        },
        {
          type: "commission",
          amount_cents: 12_000,
          status: "pending_kyc",
          occurred_at: "2026-05-08T12:00:00Z",
          reference_id: "comm-1",
          mission_title: "Dev React",
        },
      ]),
    )

    render(<WalletUnifiedHistory />)
    expect(screen.getByTestId("wallet-history-row-rec-1")).toBeInTheDocument()
    expect(
      screen.getByTestId("wallet-history-row-comm-1"),
    ).toBeInTheDocument()
    expect(screen.getByText("Logo design")).toBeInTheDocument()
    expect(screen.getByText("Dev React")).toBeInTheDocument()
  })

  it("renders the empty state when the page has no rows", () => {
    mockSummary.mockReturnValue(summaryPage([]))
    render(<WalletUnifiedHistory />)
    expect(
      screen.getByText("walletUnified.history.empty"),
    ).toBeInTheDocument()
  })

  it("hides the Charger plus button when no next_cursor", () => {
    mockSummary.mockReturnValue(
      summaryPage([
        {
          type: "mission",
          amount_cents: 1_000,
          status: "paid",
          occurred_at: "2026-05-10T12:00:00Z",
          reference_id: "rec-x",
        },
      ]),
    )
    render(<WalletUnifiedHistory />)
    expect(
      screen.queryByTestId("wallet-history-load-more"),
    ).not.toBeInTheDocument()
  })

  it("shows the Charger plus button when next_cursor is set, and advances the cursor on click", async () => {
    // The hook is called on every render with the current `cursor`
    // state. Drive page selection from the cursor argument so the
    // sequence is deterministic regardless of how many renders the
    // component does.
    mockSummary.mockImplementation((cursor?: string) => {
      if (cursor === "NEXT_CUR") {
        return summaryPage(
          [
            {
              type: "commission",
              amount_cents: 500,
              status: "paid",
              occurred_at: "2026-05-09T12:00:00Z",
              reference_id: "comm-2",
            },
          ],
          undefined,
        )
      }
      return summaryPage(
        [
          {
            type: "mission",
            amount_cents: 1_000,
            status: "paid",
            occurred_at: "2026-05-10T12:00:00Z",
            reference_id: "rec-1",
          },
        ],
        "NEXT_CUR",
      )
    })

    render(<WalletUnifiedHistory />)
    const loadMore = screen.getByTestId("wallet-history-load-more")
    fireEvent.click(loadMore)

    await waitFor(() => {
      // After the cursor advance, the hook was called with NEXT_CUR.
      expect(mockSummary).toHaveBeenCalledWith("NEXT_CUR")
    })
  })

  it("applies the right status tone to each row through WalletStatusBadge", () => {
    mockSummary.mockReturnValue(
      summaryPage([
        {
          type: "mission",
          amount_cents: 100,
          status: "paid",
          occurred_at: "2026-05-10T12:00:00Z",
          reference_id: "rec-paid",
        },
        {
          type: "commission",
          amount_cents: 200,
          status: "pending_kyc",
          occurred_at: "2026-05-09T12:00:00Z",
          reference_id: "comm-pending",
        },
        {
          type: "commission",
          amount_cents: 300,
          status: "failed",
          occurred_at: "2026-05-08T12:00:00Z",
          reference_id: "comm-failed",
        },
        {
          type: "mission",
          amount_cents: 400,
          status: "escrowed",
          occurred_at: "2026-05-07T12:00:00Z",
          reference_id: "rec-escrow",
        },
      ]),
    )

    const { container } = render(<WalletUnifiedHistory />)
    const badges = container.querySelectorAll("[data-tone]")
    // 4 rows × 1 badge each.
    expect(badges.length).toBe(4)
    const tones = Array.from(badges).map((b) => b.getAttribute("data-tone"))
    expect(tones).toEqual(["paid", "pending", "failed", "escrowed"])
  })

  it("returns null on isError (defensive degradation)", () => {
    mockSummary.mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
    })
    const { container } = render(<WalletUnifiedHistory />)
    expect(container.firstChild).toBeNull()
  })
})
