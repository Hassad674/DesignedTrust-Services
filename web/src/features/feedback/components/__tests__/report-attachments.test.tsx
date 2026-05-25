import { describe, it, expect, vi, beforeEach } from "vitest"
import { render, screen, fireEvent } from "@testing-library/react"
import { ReportAttachments } from "../report-attachments"
import type { FeedbackAttachment } from "../../hooks/use-feedback-attachments"

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}))

// next/image → plain img so the preview renders in jsdom.
vi.mock("next/image", () => ({
  default: (props: Record<string, unknown>) => {
    // eslint-disable-next-line @next/next/no-img-element, jsx-a11y/alt-text
    return <img {...props} />
  },
}))

vi.mock("lucide-react", () => ({
  Paperclip: (p: Record<string, unknown>) => <span {...p} />,
  Trash2: (p: Record<string, unknown>) => <span data-testid="trash" {...p} />,
  Video: (p: Record<string, unknown>) => <span data-testid="video-icon" {...p} />,
  Loader2: (p: Record<string, unknown>) => <span {...p} />,
}))

function imageAttachment(over: Partial<FeedbackAttachment> = {}): FeedbackAttachment {
  return {
    id: "a1",
    file: new File([new Uint8Array(1)], "shot.png", { type: "image/png" }),
    kind: "image",
    previewUrl: "blob:preview-1",
    status: "uploaded",
    ...over,
  }
}

beforeEach(() => {
  vi.clearAllMocks()
})

describe("ReportAttachments", () => {
  it("renders the labelled add-files control and accepted formats", () => {
    render(
      <ReportAttachments
        attachments={[]}
        onAddFiles={vi.fn()}
        onRemove={vi.fn()}
        disabled={false}
      />,
    )
    expect(screen.getByText("attachments_add")).toBeDefined()
    expect(screen.getByText("attachments_formats")).toBeDefined()
  })

  it("disables the add control when disabled", () => {
    render(
      <ReportAttachments
        attachments={[]}
        onAddFiles={vi.fn()}
        onRemove={vi.fn()}
        disabled={true}
      />,
    )
    const button = screen.getByRole("button", { name: "attachments_add" })
    expect((button as HTMLButtonElement).disabled).toBe(true)
  })

  it("renders an image preview for an uploaded image", () => {
    render(
      <ReportAttachments
        attachments={[imageAttachment()]}
        onAddFiles={vi.fn()}
        onRemove={vi.fn()}
        disabled={false}
      />,
    )
    expect(screen.getByText("shot.png")).toBeDefined()
    expect(screen.getByText("attachments_uploaded")).toBeDefined()
    expect(
      (document.querySelector("img") as HTMLImageElement).getAttribute("src"),
    ).toBe("blob:preview-1")
  })

  it("shows a video icon (not a preview) for a video attachment", () => {
    render(
      <ReportAttachments
        attachments={[
          imageAttachment({ id: "v1", kind: "video", status: "uploaded" }),
        ]}
        onAddFiles={vi.fn()}
        onRemove={vi.fn()}
        disabled={false}
      />,
    )
    expect(screen.getByTestId("video-icon")).toBeDefined()
  })

  it("announces an upload error via role=alert", () => {
    render(
      <ReportAttachments
        attachments={[
          imageAttachment({ status: "error", rejection: "too_large" }),
        ]}
        onAddFiles={vi.fn()}
        onRemove={vi.fn()}
        disabled={false}
      />,
    )
    const alert = screen.getByRole("alert")
    expect(alert.textContent).toBe("attachments_error_too_large")
  })

  it("calls onRemove with the attachment id", () => {
    const onRemove = vi.fn()
    render(
      <ReportAttachments
        attachments={[imageAttachment()]}
        onAddFiles={vi.fn()}
        onRemove={onRemove}
        disabled={false}
      />,
    )
    fireEvent.click(screen.getByRole("button", { name: "attachments_remove" }))
    expect(onRemove).toHaveBeenCalledWith("a1")
  })

  it("forwards selected files to onAddFiles", () => {
    const onAddFiles = vi.fn()
    const { container } = render(
      <ReportAttachments
        attachments={[]}
        onAddFiles={onAddFiles}
        onRemove={vi.fn()}
        disabled={false}
      />,
    )
    const input = container.querySelector('input[type="file"]') as HTMLInputElement
    const file = new File([new Uint8Array(1)], "x.png", { type: "image/png" })
    fireEvent.change(input, { target: { files: [file] } })
    expect(onAddFiles).toHaveBeenCalledTimes(1)
  })
})
