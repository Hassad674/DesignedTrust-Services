import { describe, it, expect, vi, beforeEach } from "vitest"
import { render, screen, fireEvent, waitFor } from "@testing-library/react"
import { ReportForm } from "../report-form"

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
  useLocale: () => "fr",
}))

// Session hook — tests flip `data` between anonymous and logged-in.
const mockUser: { current: { id: string; role: string } | undefined } = {
  current: undefined,
}
vi.mock("@/shared/hooks/use-user", () => ({
  useUser: () => ({ data: mockUser.current }),
}))

// Submit mutation — capture the body the form builds.
const mockMutate = vi.fn()
const mockSubmitState = {
  current: { mutate: mockMutate, isPending: false, isError: false },
}
vi.mock("../../hooks/use-submit-feedback", () => ({
  useSubmitFeedback: () => mockSubmitState.current,
}))

// Attachments hook — a static logged-in attachment ref to assert it is
// only forwarded when logged-in.
const sampleRef = {
  kind: "image",
  object_key: "feedback/abc.png",
  content_type: "image/png",
  size_bytes: 4096,
}
vi.mock("../../hooks/use-feedback-attachments", () => ({
  useFeedbackAttachments: () => ({
    attachments: [],
    addFiles: vi.fn(),
    remove: vi.fn(),
    reset: vi.fn(),
    uploadedRefs: [sampleRef],
    isUploading: false,
  }),
}))

// Context capture — deterministic so the payload assertion is stable.
vi.mock("../../lib/capture-context", () => ({
  captureFeedbackContext: ({ locale, role }: { locale: string; role?: string }) => ({
    role: role ?? "",
    locale,
    platform: "web",
    app_version: "",
    viewport: "800x600",
    user_agent: "UA",
  }),
  currentPageUrl: () => "https://app.test/projects",
}))

vi.mock("lucide-react", () => ({
  Bug: (p: Record<string, unknown>) => <span {...p} />,
  ShieldAlert: (p: Record<string, unknown>) => <span {...p} />,
  Paperclip: (p: Record<string, unknown>) => <span {...p} />,
  Trash2: (p: Record<string, unknown>) => <span {...p} />,
  Video: (p: Record<string, unknown>) => <span {...p} />,
  Loader2: (p: Record<string, unknown>) => <span {...p} />,
}))

beforeEach(() => {
  vi.clearAllMocks()
  mockUser.current = undefined
  mockSubmitState.current = { mutate: mockMutate, isPending: false, isError: false }
})

function fillValid() {
  fireEvent.change(screen.getByLabelText("title_label"), {
    target: { value: "Le bouton ne marche pas" },
  })
  fireEvent.change(screen.getByLabelText("description_label"), {
    target: { value: "Rien ne se passe au clic sur Envoyer." },
  })
}

describe("ReportForm — attachment gating", () => {
  it("hides the upload zone and shows the sign-in hint when anonymous", () => {
    mockUser.current = undefined
    render(<ReportForm onSuccess={vi.fn()} />)
    expect(screen.getByText("attachments_anonymous_hint")).toBeDefined()
    expect(screen.queryByText("attachments_add")).toBeNull()
    // Anonymous reporters get the optional contact-email field.
    expect(screen.getByLabelText("email_label")).toBeDefined()
  })

  it("shows the upload zone and no email field when logged-in", () => {
    mockUser.current = { id: "u-1", role: "agency" }
    render(<ReportForm onSuccess={vi.fn()} />)
    expect(screen.getByText("attachments_add")).toBeDefined()
    expect(screen.queryByText("attachments_anonymous_hint")).toBeNull()
    expect(screen.queryByLabelText("email_label")).toBeNull()
  })
})

describe("ReportForm — validation", () => {
  it("does not submit when title and description are empty", async () => {
    render(<ReportForm onSuccess={vi.fn()} />)
    fireEvent.click(screen.getByRole("button", { name: "submit" }))
    await waitFor(() => {
      expect(screen.getByText("title_error")).toBeDefined()
      expect(screen.getByText("description_error")).toBeDefined()
    })
    expect(mockMutate).not.toHaveBeenCalled()
  })
})

describe("ReportForm — submit payload", () => {
  it("forwards attachment refs + empty email when logged-in", async () => {
    mockUser.current = { id: "u-1", role: "agency" }
    render(<ReportForm onSuccess={vi.fn()} />)
    fillValid()
    fireEvent.click(screen.getByRole("button", { name: "submit" }))

    await waitFor(() => expect(mockMutate).toHaveBeenCalledTimes(1))
    const [body] = mockMutate.mock.calls[0]
    expect(body).toMatchObject({
      type: "bug",
      title: "Le bouton ne marche pas",
      description: "Rien ne se passe au clic sur Envoyer.",
      page_url: "https://app.test/projects",
      reporter_email: "",
      attachment_keys: [sampleRef],
      hp: "",
    })
    expect(body.context.role).toBe("agency")
  })

  it("sends no attachments + carries the contact email when anonymous", async () => {
    mockUser.current = undefined
    render(<ReportForm onSuccess={vi.fn()} />)
    fillValid()
    fireEvent.change(screen.getByLabelText("email_label"), {
      target: { value: "visitor@example.com" },
    })
    fireEvent.click(screen.getByRole("button", { name: "submit" }))

    await waitFor(() => expect(mockMutate).toHaveBeenCalledTimes(1))
    const [body] = mockMutate.mock.calls[0]
    expect(body.attachment_keys).toEqual([])
    expect(body.reporter_email).toBe("visitor@example.com")
    expect(body.context.role).toBe("")
  })

  it("switches the report type before submitting", async () => {
    render(<ReportForm onSuccess={vi.fn()} />)
    fireEvent.click(screen.getByRole("radio", { name: "type_security" }))
    fillValid()
    fireEvent.click(screen.getByRole("button", { name: "submit" }))

    await waitFor(() => expect(mockMutate).toHaveBeenCalledTimes(1))
    expect(mockMutate.mock.calls[0][0].type).toBe("security")
  })
})

describe("ReportForm — honeypot", () => {
  it("renders a hidden, aria-hidden, non-focusable honeypot field", () => {
    const { container } = render(<ReportForm onSuccess={vi.fn()} />)
    const honeypot = container.querySelector('input[name="hp"]')
    expect(honeypot).not.toBeNull()
    expect(honeypot?.getAttribute("aria-hidden")).toBe("true")
    expect(honeypot?.getAttribute("tabindex")).toBe("-1")
    expect(honeypot?.className).toContain("sr-only")
  })

  it("never calls the API when the honeypot is filled (bot path)", async () => {
    const { container } = render(<ReportForm onSuccess={vi.fn()} />)
    fillValid()
    const honeypot = container.querySelector('input[name="hp"]') as HTMLInputElement
    fireEvent.change(honeypot, { target: { value: "i-am-a-bot" } })
    fireEvent.click(screen.getByRole("button", { name: "submit" }))

    // Give the async submit handler a tick; mutate must stay untouched.
    await new Promise((r) => setTimeout(r, 10))
    expect(mockMutate).not.toHaveBeenCalled()
  })
})

describe("ReportForm — submit states", () => {
  it("disables the button and shows the pending label while submitting", () => {
    mockSubmitState.current = { mutate: mockMutate, isPending: true, isError: false }
    render(<ReportForm onSuccess={vi.fn()} />)
    const button = screen.getByRole("button", { name: "submit_pending" })
    expect((button as HTMLButtonElement).disabled).toBe(true)
  })

  it("shows an inline error when the submission fails", () => {
    mockSubmitState.current = { mutate: mockMutate, isPending: false, isError: true }
    render(<ReportForm onSuccess={vi.fn()} />)
    expect(screen.getByText("submit_error")).toBeDefined()
  })
})
