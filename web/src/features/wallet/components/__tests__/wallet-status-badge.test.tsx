import { describe, it, expect } from "vitest"
import { render } from "@testing-library/react"

import {
  WalletStatusBadge,
  resolveWalletStatusTone,
} from "../wallet-status-badge"

describe("resolveWalletStatusTone", () => {
  const cases: Array<[string, "paid" | "pending" | "escrowed" | "failed"]> = [
    ["paid", "paid"],
    ["transferred", "paid"],
    ["transferred_pending_bank", "paid"],
    ["completed", "paid"],
    ["pending", "pending"],
    ["pending_kyc", "pending"],
    ["active", "pending"],
    ["completion_requested", "pending"],
    ["accepted", "pending"],
    ["escrowed", "escrowed"],
    ["held", "escrowed"],
    ["failed", "failed"],
    ["cancelled", "failed"],
    ["clawed_back", "failed"],
    ["totally_unknown_status", "escrowed"],
    ["", "escrowed"],
  ]
  it.each(cases)(
    "maps status %s → tone %s",
    (status, tone) => {
      expect(resolveWalletStatusTone(status)).toBe(tone)
    },
  )

  it("treats uppercase input the same as lowercase (defensive)", () => {
    expect(resolveWalletStatusTone("PAID")).toBe("paid")
    expect(resolveWalletStatusTone("Pending_KYC")).toBe("pending")
  })
})

describe("WalletStatusBadge", () => {
  it("renders the label inside the pill", () => {
    const { getByText } = render(
      <WalletStatusBadge status="paid" label="Reçu" />,
    )
    expect(getByText("Reçu")).toBeDefined()
  })

  it("attaches data-tone for downstream snapshot/visual tests", () => {
    const { container } = render(
      <WalletStatusBadge status="failed" label="Échoué" />,
    )
    const span = container.querySelector("span")
    expect(span?.getAttribute("data-tone")).toBe("failed")
    expect(span?.getAttribute("data-status")).toBe("failed")
  })

  it("falls back to the escrowed tone for unknown statuses", () => {
    const { container } = render(
      <WalletStatusBadge status="brand_new_status" label="Inconnu" />,
    )
    const span = container.querySelector("span")
    expect(span?.getAttribute("data-tone")).toBe("escrowed")
  })
})
