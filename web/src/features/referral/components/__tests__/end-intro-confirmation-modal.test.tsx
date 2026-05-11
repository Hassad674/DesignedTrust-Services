import { describe, it, expect, vi } from "vitest"
import { render, screen, fireEvent } from "@testing-library/react"

import { EndIntroConfirmationModal } from "../end-intro-confirmation-modal"

vi.mock("next-intl", () => ({
  useTranslations:
    (namespace?: string) =>
    (key: string, params?: Record<string, string | number>) => {
      const full = namespace ? `${namespace}.${key}` : key
      return params ? `${full}(${JSON.stringify(params)})` : full
    },
}))

function build(
  overrides: Partial<Parameters<typeof EndIntroConfirmationModal>[0]> = {},
) {
  const onClose = vi.fn()
  const onConfirm = vi.fn()
  return {
    onClose,
    onConfirm,
    props: {
      open: true,
      onClose,
      onConfirm,
      providerName: "Provider Inc",
      clientName: "Client Co",
      ...overrides,
    },
  }
}

describe("EndIntroConfirmationModal (Run C)", () => {
  it("does not render its body when open=false", () => {
    const { props } = build({ open: false })
    const { container } = render(
      <EndIntroConfirmationModal {...props} />,
    )
    expect(container.firstChild).toBeNull()
  })

  it("renders the title + body with the provider / client names interpolated", () => {
    const { props } = build()
    render(<EndIntroConfirmationModal {...props} />)
    expect(
      screen.getByText("referralEndIntro.modal.title"),
    ).toBeInTheDocument()
    const body = screen.getByText(/referralEndIntro\.modal\.body/)
    expect(body.textContent).toContain("Provider Inc")
    expect(body.textContent).toContain("Client Co")
  })

  it("falls back to generic nouns when names are absent", () => {
    const { props } = build({
      providerName: undefined,
      clientName: undefined,
    })
    render(<EndIntroConfirmationModal {...props} />)
    const body = screen.getByText(/referralEndIntro\.modal\.body/)
    // The fallback keys are echoed verbatim by the mock.
    expect(body.textContent).toContain(
      "referralEndIntro.modal.fallbackProvider",
    )
    expect(body.textContent).toContain(
      "referralEndIntro.modal.fallbackClient",
    )
  })

  it("fires onClose when Annuler is clicked", () => {
    const { props, onClose } = build()
    render(<EndIntroConfirmationModal {...props} />)
    fireEvent.click(screen.getByTestId("end-intro-cancel"))
    expect(onClose).toHaveBeenCalledOnce()
  })

  it("fires onConfirm when Terminer définitivement is clicked", () => {
    const { props, onConfirm, onClose } = build()
    render(<EndIntroConfirmationModal {...props} />)
    fireEvent.click(screen.getByTestId("end-intro-confirm"))
    expect(onConfirm).toHaveBeenCalledOnce()
    // Clicking the confirm button does NOT close the modal — the
    // parent owns the close-on-success behaviour so the loading
    // state can render while the mutation is in flight.
    expect(onClose).not.toHaveBeenCalled()
  })

  it("disables both buttons while pending", () => {
    const { props } = build({ pending: true })
    render(<EndIntroConfirmationModal {...props} />)
    const cancel = screen.getByTestId("end-intro-cancel") as HTMLButtonElement
    const confirm = screen.getByTestId(
      "end-intro-confirm",
    ) as HTMLButtonElement
    expect(cancel.disabled).toBe(true)
    expect(confirm.disabled).toBe(true)
  })
})
