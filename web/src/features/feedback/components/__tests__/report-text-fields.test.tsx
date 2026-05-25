import { describe, it, expect, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import { useForm } from "react-hook-form"
import { ReportTextFields } from "../report-text-fields"

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}))

type Values = {
  title: string
  description: string
  reporterEmail: string
  hp: string
}

// Harness wires a real RHF register so the fields behave like in the form.
function Harness({ showEmailField }: { showEmailField: boolean }) {
  const { register, formState } = useForm<Values>({
    defaultValues: { title: "", description: "", reporterEmail: "", hp: "" },
  })
  return (
    <ReportTextFields
      register={register}
      errors={formState.errors}
      showEmailField={showEmailField}
    />
  )
}

describe("ReportTextFields", () => {
  it("renders labelled title and description controls", () => {
    render(<Harness showEmailField={false} />)
    expect(screen.getByLabelText("title_label")).toBeDefined()
    expect(screen.getByLabelText("description_label")).toBeDefined()
  })

  it("shows the contact-email field only when requested (anonymous)", () => {
    const { rerender } = render(<Harness showEmailField={false} />)
    expect(screen.queryByLabelText("email_label")).toBeNull()
    rerender(<Harness showEmailField={true} />)
    expect(screen.getByLabelText("email_label")).toBeDefined()
  })

  it("renders a hidden, aria-hidden, non-focusable honeypot input", () => {
    const { container } = render(<Harness showEmailField={false} />)
    const honeypot = container.querySelector('input[name="hp"]')
    expect(honeypot).not.toBeNull()
    expect(honeypot?.getAttribute("aria-hidden")).toBe("true")
    expect(honeypot?.getAttribute("tabindex")).toBe("-1")
    expect(honeypot?.className).toContain("sr-only")
  })
})
