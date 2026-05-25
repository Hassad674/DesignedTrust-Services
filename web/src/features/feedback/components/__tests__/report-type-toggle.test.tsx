import { describe, it, expect, vi, beforeEach } from "vitest"
import { render, screen, fireEvent } from "@testing-library/react"
import { ReportTypeToggle } from "../report-type-toggle"
import type { FeedbackType } from "../../types"

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}))

vi.mock("lucide-react", () => ({
  Bug: (props: Record<string, unknown>) => <span data-testid="bug" {...props} />,
  ShieldAlert: (props: Record<string, unknown>) => (
    <span data-testid="shield" {...props} />
  ),
}))

beforeEach(() => {
  vi.clearAllMocks()
})

describe("ReportTypeToggle", () => {
  it("renders a radiogroup with bug + security options", () => {
    render(<ReportTypeToggle value="bug" onChange={vi.fn()} />)
    expect(screen.getByRole("radiogroup", { name: "type_legend" })).toBeDefined()
    expect(screen.getByRole("radio", { name: "type_bug" })).toBeDefined()
    expect(screen.getByRole("radio", { name: "type_security" })).toBeDefined()
  })

  it("marks the active option with aria-checked", () => {
    render(<ReportTypeToggle value="security" onChange={vi.fn()} />)
    expect(
      screen.getByRole("radio", { name: "type_security" }).getAttribute("aria-checked"),
    ).toBe("true")
    expect(
      screen.getByRole("radio", { name: "type_bug" }).getAttribute("aria-checked"),
    ).toBe("false")
  })

  it("calls onChange with the chosen type", () => {
    const onChange = vi.fn<(value: FeedbackType) => void>()
    render(<ReportTypeToggle value="bug" onChange={onChange} />)
    fireEvent.click(screen.getByRole("radio", { name: "type_security" }))
    expect(onChange).toHaveBeenCalledWith("security")
  })
})
