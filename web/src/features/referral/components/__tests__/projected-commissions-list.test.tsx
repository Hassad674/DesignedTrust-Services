import { describe, it, expect, vi } from "vitest"
import { render, screen } from "@testing-library/react"

import { ProjectedCommissionsList } from "../projected-commissions-list"
import type { ReferralCommission } from "../../types"

vi.mock("next-intl", () => ({
  useTranslations:
    (namespace?: string) =>
    (key: string, params?: Record<string, string | number>) => {
      const full = namespace ? `${namespace}.${key}` : key
      return params ? `${full}(${JSON.stringify(params)})` : full
    },
}))

const baseRow: ReferralCommission = {
  id: "c-base",
  attribution_id: "att-1",
  milestone_id: "ms-1",
  gross_amount_cents: 100_000,
  commission_cents: 10_000,
  currency: "EUR",
  status: "paid",
  created_at: "2026-05-01T00:00:00Z",
}

function row(
  overrides: Partial<ReferralCommission>,
): ReferralCommission {
  return { ...baseRow, ...overrides, id: overrides.id ?? `c-${Math.random()}` }
}

describe("ProjectedCommissionsList (Run C)", () => {
  it("renders the empty-state when no commissions and no escrow", () => {
    render(<ProjectedCommissionsList commissions={[]} escrowCents={0} />)
    expect(
      screen.getByText("referralProjection.empty"),
    ).toBeInTheDocument()
  })

  it("renders an escrow synthetic line when escrowCents > 0 and no commissions", () => {
    render(
      <ProjectedCommissionsList commissions={[]} escrowCents={50_000} />,
    )
    const escrowLine = screen.getByTestId(
      "projected-commission-escrow-line",
    )
    expect(escrowLine).toBeInTheDocument()
    expect(escrowLine.getAttribute("data-tone")).toBe("escrowed")
  })

  it("renders a paid row with the success tone", () => {
    const c = row({ status: "paid", commission_cents: 20_000 })
    render(
      <ProjectedCommissionsList commissions={[c]} escrowCents={0} />,
    )
    const li = screen.getByTestId(`projected-commission-row-${c.id}`)
    expect(li.getAttribute("data-tone")).toBe("paid")
  })

  it("renders a pending row with the pending tone (both pending and pending_kyc)", () => {
    const c1 = row({ status: "pending", commission_cents: 1_000, id: "c-p1" })
    const c2 = row({
      status: "pending_kyc",
      commission_cents: 2_000,
      id: "c-p2",
    })
    render(
      <ProjectedCommissionsList commissions={[c1, c2]} escrowCents={0} />,
    )
    expect(
      screen
        .getByTestId(`projected-commission-row-${c1.id}`)
        .getAttribute("data-tone"),
    ).toBe("pending")
    expect(
      screen
        .getByTestId(`projected-commission-row-${c2.id}`)
        .getAttribute("data-tone"),
    ).toBe("pending")
  })

  it("renders a failed row with the destructive tone", () => {
    const c = row({ status: "failed", commission_cents: 5_000 })
    render(
      <ProjectedCommissionsList commissions={[c]} escrowCents={0} />,
    )
    expect(
      screen
        .getByTestId(`projected-commission-row-${c.id}`)
        .getAttribute("data-tone"),
    ).toBe("failed")
  })

  it("SKIPS cancelled and clawed_back rows (per Run C brief)", () => {
    const skipped = row({ status: "cancelled", commission_cents: 1, id: "c-cx" })
    const skipped2 = row({
      status: "clawed_back",
      commission_cents: 2,
      id: "c-cb",
    })
    const visible = row({ status: "paid", commission_cents: 100, id: "c-ok" })
    render(
      <ProjectedCommissionsList
        commissions={[skipped, skipped2, visible]}
        escrowCents={0}
      />,
    )
    expect(
      screen.queryByTestId(`projected-commission-row-${skipped.id}`),
    ).not.toBeInTheDocument()
    expect(
      screen.queryByTestId(`projected-commission-row-${skipped2.id}`),
    ).not.toBeInTheDocument()
    expect(
      screen.getByTestId(`projected-commission-row-${visible.id}`),
    ).toBeInTheDocument()
  })

  it("stacks escrow + commission lines together when both have values", () => {
    const paid = row({ status: "paid", commission_cents: 10_000, id: "c-pa" })
    render(
      <ProjectedCommissionsList
        commissions={[paid]}
        escrowCents={30_000}
      />,
    )
    expect(
      screen.getByTestId("projected-commission-escrow-line"),
    ).toBeInTheDocument()
    expect(
      screen.getByTestId(`projected-commission-row-${paid.id}`),
    ).toBeInTheDocument()
  })
})
