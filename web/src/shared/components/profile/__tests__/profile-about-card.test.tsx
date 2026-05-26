import { describe, expect, it, vi } from "vitest"
import { fireEvent, render, screen } from "@testing-library/react"
import { NextIntlClientProvider } from "next-intl"
import messages from "@/../messages/en.json"
import { ProfileAboutCard } from "../profile-about-card"

function renderCard(props: Parameters<typeof ProfileAboutCard>[0]) {
  return render(
    <NextIntlClientProvider locale="en" messages={messages}>
      <ProfileAboutCard {...props} />
    </NextIntlClientProvider>,
  )
}

describe("ProfileAboutCard", () => {
  it("renders the heading and content when provided", () => {
    renderCard({
      content: "Hello world",
      label: "About",
      placeholder: "Tell us more",
      onSave: vi.fn(),
    })

    expect(screen.getByRole("heading", { name: "About" })).toBeInTheDocument()
    expect(screen.getByText("Hello world")).toBeInTheDocument()
  })

  // Regression: hooks must be called in the same order across renders.
  // readOnly + empty content returns null AFTER the hooks, so toggling
  // these props between renders must not throw a hook order violation.
  it("returns null when readOnly + no content, then renders cleanly when content arrives", () => {
    const props = {
      content: "",
      label: "About",
      placeholder: "Tell us more",
      readOnly: true,
    }
    const { container, rerender } = render(
      <NextIntlClientProvider locale="en" messages={messages}>
        <ProfileAboutCard {...props} />
      </NextIntlClientProvider>,
    )
    expect(container.firstChild).toBeNull()

    rerender(
      <NextIntlClientProvider locale="en" messages={messages}>
        <ProfileAboutCard {...props} content="Now visible" />
      </NextIntlClientProvider>,
    )
    expect(screen.getByText("Now visible")).toBeInTheDocument()
  })

  // Regression: the About field cap was raised from 1000 to 3000 chars.
  // The counter must read "/ 3000" and the textarea must clamp input
  // beyond 3000 characters (never silently allow more, never cap at 1000).
  it("caps the About editor at 3000 characters and clamps overflow", () => {
    renderCard({
      content: "",
      label: "About",
      placeholder: "Tell us more",
      onSave: vi.fn(),
    })

    fireEvent.click(screen.getByRole("button", { name: "Edit about" }))

    expect(screen.getByText("0 / 3000 characters")).toBeInTheDocument()

    const textarea = screen.getByRole("textbox", { name: "About" })
    fireEvent.change(textarea, { target: { value: "x".repeat(3500) } })

    expect((textarea as HTMLTextAreaElement).value).toHaveLength(3000)
    expect(screen.getByText("3000 / 3000 characters")).toBeInTheDocument()
  })
})
