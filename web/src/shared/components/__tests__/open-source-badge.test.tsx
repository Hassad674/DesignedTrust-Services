/**
 * OpenSourceBadge test.
 *
 * Asserts:
 *   1. The badge renders the translated "Open source" label (no raw
 *      i18n key leaks to the DOM).
 *   2. It is a link pointing to the public repo URL, opening in a new
 *      tab safely (`target="_blank"` + `rel="noopener noreferrer"`).
 *   3. It exposes a descriptive `aria-label` (a11y / WCAG AA).
 *   4. Both `inline` and `prominent` variants render the same link.
 */

import { describe, it, expect } from "vitest"
import { render, screen } from "@testing-library/react"
import { NextIntlClientProvider } from "next-intl"
import frMessages from "@/../messages/fr.json"
import enMessages from "@/../messages/en.json"
import {
  OpenSourceBadge,
  OPEN_SOURCE_REPO_URL,
} from "../open-source-badge"

function renderBadge(
  variant: "inline" | "prominent" = "inline",
  messages: Record<string, unknown> = frMessages,
  locale = "fr",
) {
  return render(
    <NextIntlClientProvider locale={locale} messages={messages}>
      <OpenSourceBadge variant={variant} />
    </NextIntlClientProvider>,
  )
}

describe("OpenSourceBadge", () => {
  it("renders the translated 'Open source' label", () => {
    renderBadge()
    expect(screen.getByText("Open source")).toBeInTheDocument()
  })

  it("links to the public GitHub repo and opens it safely in a new tab", () => {
    renderBadge()
    const link = screen.getByRole("link")
    expect(link).toHaveAttribute("href", OPEN_SOURCE_REPO_URL)
    expect(OPEN_SOURCE_REPO_URL).toBe(
      "https://github.com/Hassad674/serviceMarketplaceGo",
    )
    expect(link).toHaveAttribute("target", "_blank")
    expect(link).toHaveAttribute("rel", "noopener noreferrer")
  })

  it("exposes a descriptive aria-label for screen readers", () => {
    renderBadge()
    const link = screen.getByRole("link")
    const ariaLabel = link.getAttribute("aria-label")
    expect(ariaLabel).toBeTruthy()
    // Must not leak a raw i18n key.
    expect(ariaLabel).not.toContain("openSource.")
  })

  it("renders the same repo link in the prominent variant", () => {
    renderBadge("prominent")
    const link = screen.getByRole("link")
    expect(link).toHaveAttribute("href", OPEN_SOURCE_REPO_URL)
    expect(screen.getByText("Open source")).toBeInTheDocument()
  })

  it("resolves the English label on the en locale", () => {
    renderBadge("inline", enMessages, "en")
    expect(screen.getByText("Open source")).toBeInTheDocument()
    const link = screen.getByRole("link")
    expect(link.getAttribute("aria-label")).toContain("GitHub")
  })
})
