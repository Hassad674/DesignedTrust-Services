import { describe, it, expect, vi } from "vitest"
import { render, screen, fireEvent } from "@testing-library/react"

import { WalletWithdrawResultModal } from "../wallet-withdraw-result-modal"
import type { WithdrawLegError } from "../../api/wallet-api"

// next-intl stub — echoes namespace.key (with optional params JSON-
// stringified) so tests can assert against the i18n contract.
vi.mock("next-intl", () => ({
  useTranslations:
    (namespace?: string) =>
    (key: string, params?: Record<string, string | number>) => {
      const full = namespace ? `${namespace}.${key}` : key
      return params ? `${full}(${JSON.stringify(params)})` : full
    },
}))

function build(
  overrides: Partial<Parameters<typeof WalletWithdrawResultModal>[0]> = {},
) {
  const onClose = vi.fn()
  return {
    onClose,
    props: {
      open: true,
      onClose,
      drainedCents: 500_00,
      missionsCents: 300_00,
      commissionsCents: 200_00,
      errors: [] as WithdrawLegError[],
      ...overrides,
    },
  }
}

describe("WalletWithdrawResultModal (Run C)", () => {
  it("renders the success summary lines with the formatted amounts", () => {
    const { props } = build()
    render(<WalletWithdrawResultModal {...props} />)
    expect(
      screen.getByText(/walletUnified\.result\.drained/),
    ).toBeInTheDocument()
    expect(
      screen.getByText(/walletUnified\.result\.missionsLine/),
    ).toBeInTheDocument()
    expect(
      screen.getByText(/walletUnified\.result\.commissionsLine/),
    ).toBeInTheDocument()
  })

  it("hides the errors section when errors is empty", () => {
    const { props } = build()
    render(<WalletWithdrawResultModal {...props} />)
    expect(
      screen.queryByTestId("withdraw-errors-list"),
    ).not.toBeInTheDocument()
  })

  it("renders one entry per error with the right source label", () => {
    const { props } = build({
      errors: [
        {
          source: "missions",
          code: "stripe_account_missing",
          message: "Stripe account missing",
        },
        {
          source: "commissions",
          code: "commission_drain_failed",
          message: "Stripe transfer failed",
        },
      ],
    })
    render(<WalletWithdrawResultModal {...props} />)
    const list = screen.getByTestId("withdraw-errors-list")
    expect(list.children).toBeDefined()
    expect(
      screen.getByText("walletUnified.result.errorMissions"),
    ).toBeInTheDocument()
    expect(
      screen.getByText("walletUnified.result.errorCommissions"),
    ).toBeInTheDocument()
    expect(
      screen.getByText("Stripe account missing"),
    ).toBeInTheDocument()
    expect(
      screen.getByText("Stripe transfer failed"),
    ).toBeInTheDocument()
  })

  it("calls onClose when the Fermer button is clicked", () => {
    const { props, onClose } = build()
    render(<WalletWithdrawResultModal {...props} />)
    fireEvent.click(screen.getByTestId("withdraw-result-close"))
    expect(onClose).toHaveBeenCalledOnce()
  })

  it("renders nothing when open=false", () => {
    const { props } = build({ open: false })
    const { container } = render(
      <WalletWithdrawResultModal {...props} />,
    )
    expect(container.firstChild).toBeNull()
  })
})
