import { describe, it, expect, vi, beforeEach } from "vitest"
import { render, screen, fireEvent } from "@testing-library/react"
import { ReportModal } from "../report-modal"

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}))

// Shared Modal primitive — render a minimal dialog shell so we can probe
// the title + body without the portal/focus-trap machinery.
vi.mock("@/shared/components/ui/modal", () => ({
  Modal: ({
    open,
    title,
    children,
  }: {
    open: boolean
    title: string
    children: React.ReactNode
  }) =>
    open ? (
      <div role="dialog" aria-label={title}>
        {children}
      </div>
    ) : null,
}))

// ReportForm exposes a button that fires its onSuccess so we can assert
// the modal swaps to the confirmation state.
vi.mock("../report-form", () => ({
  ReportForm: ({ onSuccess }: { onSuccess: () => void }) => (
    <button type="button" onClick={onSuccess}>
      trigger-success
    </button>
  ),
}))

vi.mock("../report-success", () => ({
  ReportSuccess: ({ onClose }: { onClose: () => void }) => (
    <div>
      <span>success-state</span>
      <button type="button" onClick={onClose}>
        success-close
      </button>
    </div>
  ),
}))

beforeEach(() => {
  vi.clearAllMocks()
})

describe("ReportModal", () => {
  it("renders the form (with intro) while open and not submitted", () => {
    render(<ReportModal open onClose={vi.fn()} />)
    expect(screen.getByRole("dialog", { name: "modal_title" })).toBeDefined()
    expect(screen.getByText("modal_intro")).toBeDefined()
    expect(screen.getByText("trigger-success")).toBeDefined()
    expect(screen.queryByText("success-state")).toBeNull()
  })

  it("swaps to the success state after the form reports success", () => {
    render(<ReportModal open onClose={vi.fn()} />)
    fireEvent.click(screen.getByText("trigger-success"))
    expect(screen.getByText("success-state")).toBeDefined()
    expect(screen.queryByText("trigger-success")).toBeNull()
  })

  it("invokes onClose from the success state's close action", () => {
    const onClose = vi.fn()
    render(<ReportModal open onClose={onClose} />)
    fireEvent.click(screen.getByText("trigger-success"))
    fireEvent.click(screen.getByText("success-close"))
    expect(onClose).toHaveBeenCalledTimes(1)
  })

  it("renders nothing when closed", () => {
    render(<ReportModal open={false} onClose={vi.fn()} />)
    expect(screen.queryByRole("dialog")).toBeNull()
  })
})
