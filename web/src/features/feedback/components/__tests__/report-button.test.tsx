import { describe, it, expect, vi, beforeEach } from "vitest"
import { render, screen, fireEvent } from "@testing-library/react"
import { ReportButton } from "../report-button"

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}))

// `next/dynamic` returns the eager component synchronously in tests when
// we hand it a resolved factory; mock it to render a probe modal so we
// can assert the open/close wiring without the async chunk.
vi.mock("next/dynamic", () => ({
  default: () =>
    function MockReportModal({
      open,
      onClose,
    }: {
      open: boolean
      onClose: () => void
    }) {
      if (!open) return null
      return (
        <div data-testid="report-modal">
          <button type="button" onClick={onClose}>
            close-modal
          </button>
        </div>
      )
    },
}))

vi.mock("lucide-react", () => ({
  Bug: (props: Record<string, unknown>) => <span data-testid="bug-icon" {...props} />,
}))

beforeEach(() => {
  vi.clearAllMocks()
})

describe("ReportButton", () => {
  it("renders a labelled, keyboard-operable trigger with the bug icon", () => {
    render(<ReportButton />)
    const button = screen.getByRole("button", { name: "button_label" })
    expect(button).toBeDefined()
    expect(button.getAttribute("title")).toBe("button_tooltip")
    expect(button.getAttribute("aria-haspopup")).toBe("dialog")
    expect(screen.getByTestId("bug-icon")).toBeDefined()
  })

  it("is anchored bottom-left (never overlaps the bottom-right chat widget)", () => {
    render(<ReportButton />)
    const button = screen.getByRole("button", { name: "button_label" })
    expect(button.className).toContain("bottom-6")
    expect(button.className).toContain("left-6")
    expect(button.className).not.toContain("right-6")
  })

  it("does not mount the modal until the button is clicked", () => {
    render(<ReportButton />)
    expect(screen.queryByTestId("report-modal")).toBeNull()
  })

  it("opens the modal on click and closes it on the modal's close action", () => {
    render(<ReportButton />)
    fireEvent.click(screen.getByRole("button", { name: "button_label" }))
    expect(screen.getByTestId("report-modal")).toBeDefined()

    fireEvent.click(screen.getByText("close-modal"))
    expect(screen.queryByTestId("report-modal")).toBeNull()
  })
})
