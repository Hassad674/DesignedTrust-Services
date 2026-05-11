import { describe, it, expect, vi } from "vitest"
import { render, screen, fireEvent } from "@testing-library/react"

import { WalletUnifiedHeader } from "../wallet-unified-header"

// Mock next-intl so the tests assert against the i18n key — keeps
// the test wired to the contract (which keys must exist) rather than
// to the French copy that may evolve.
vi.mock("next-intl", () => ({
  useTranslations:
    (namespace?: string) =>
    (key: string, params?: Record<string, string | number>) => {
      const full = namespace ? `${namespace}.${key}` : key
      if (!params) return full
      return `${full}(${JSON.stringify(params)})`
    },
}))

function build(
  overrides: Partial<Parameters<typeof WalletUnifiedHeader>[0]> = {},
) {
  return {
    totalCents: 1_500_00,
    escrowedCents: 300_00,
    availableCents: 700_00,
    transmittedCents: 500_00,
    payoutPending: false,
    onWithdraw: vi.fn(),
    ...overrides,
  }
}

describe("WalletUnifiedHeader (Run C)", () => {
  it("renders the title, subtitle and total earned formatted as EUR", () => {
    render(<WalletUnifiedHeader {...build()} />)
    expect(screen.getByText("walletUnified.title")).toBeInTheDocument()
    expect(screen.getByText("walletUnified.subtitle")).toBeInTheDocument()
    // Total formatted with no decimals — 1500.00 € → "1 500 €"
    expect(screen.getByTestId("wallet-unified-total").textContent).toMatch(
      /1\s*500/,
    )
  })

  it("renders the three stat cards with the correct amounts", () => {
    render(
      <WalletUnifiedHeader
        {...build({
          escrowedCents: 123_00,
          availableCents: 456_00,
          transmittedCents: 789_00,
        })}
      />,
    )
    expect(
      screen.getByTestId("wallet-stat-escrowed").textContent,
    ).toMatch(/123/)
    expect(
      screen.getByTestId("wallet-stat-available").textContent,
    ).toMatch(/456/)
    expect(
      screen.getByTestId("wallet-stat-transmitted").textContent,
    ).toMatch(/789/)
  })

  it("fires onWithdraw when the Retirer button is clicked", () => {
    const onWithdraw = vi.fn()
    render(<WalletUnifiedHeader {...build({ onWithdraw })} />)
    const btn = screen.getByTestId("wallet-unified-withdraw")
    fireEvent.click(btn)
    expect(onWithdraw).toHaveBeenCalledOnce()
  })

  it("disables the Retirer button when available_cents === 0", () => {
    const onWithdraw = vi.fn()
    render(
      <WalletUnifiedHeader
        {...build({ availableCents: 0, onWithdraw })}
      />,
    )
    const btn = screen.getByTestId(
      "wallet-unified-withdraw",
    ) as HTMLButtonElement
    expect(btn.disabled).toBe(true)
    fireEvent.click(btn)
    // Disabled buttons silently swallow clicks — onWithdraw must
    // NOT fire even when the caller tries to bypass via JS.
    expect(onWithdraw).not.toHaveBeenCalled()
    // No-funds caption rendered as a hint.
    expect(screen.getByText("walletUnified.noFunds")).toBeInTheDocument()
  })

  it("disables the Retirer button while the mutation is in flight", () => {
    render(<WalletUnifiedHeader {...build({ payoutPending: true })} />)
    const btn = screen.getByTestId(
      "wallet-unified-withdraw",
    ) as HTMLButtonElement
    expect(btn.disabled).toBe(true)
    // The label flips to the "withdrawing" key while the spinner is
    // visible.
    expect(screen.getByText("walletUnified.withdrawing")).toBeInTheDocument()
  })
})
