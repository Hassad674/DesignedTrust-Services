import { describe, it, expect, vi } from "vitest"
import { render, screen, fireEvent } from "@testing-library/react"
import { ReportSuccess } from "../report-success"
import { ReportAttachmentsHint } from "../report-attachments-hint"

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}))

vi.mock("lucide-react", () => ({
  CheckCircle2: (p: Record<string, unknown>) => <span {...p} />,
  Paperclip: (p: Record<string, unknown>) => <span {...p} />,
}))

describe("ReportSuccess", () => {
  it("shows the confirmation copy and a close action", () => {
    const onClose = vi.fn()
    render(<ReportSuccess onClose={onClose} />)
    expect(screen.getByText("success_title")).toBeDefined()
    expect(screen.getByText("success_body")).toBeDefined()
    fireEvent.click(screen.getByRole("button", { name: "success_close" }))
    expect(onClose).toHaveBeenCalledTimes(1)
  })
})

describe("ReportAttachmentsHint", () => {
  it("renders the sign-in-to-attach hint for anonymous reporters", () => {
    render(<ReportAttachmentsHint />)
    expect(screen.getByText("attachments_anonymous_hint")).toBeDefined()
  })
})
