/**
 * LegalFooter test.
 *
 * Asserts:
 *   1. All canonical legal routes are linked (privacy long, cookies,
 *      legal index, CGU, CGV, code of conduct, sub-processors,
 *      automated decisions, registre). The legacy short `/privacy`
 *      page was merged into `/legal/politique-confidentialite` per
 *      CNIL's "single privacy policy" requirement.
 *   2. The DPO email (NEXT_PUBLIC_DPO_EMAIL or fallback) is rendered
 *      as a mailto: link so visitors have a one-click RGPD contact.
 *   3. The persistent "Manage cookies" entry-point is present
 *      (CNIL Recommendation 2020 point 6.3 — withdrawal must be as
 *      easy as the initial consent).
 *   4. No raw i18n key (string starting with "legal.footer.") leaks
 *      to the DOM — every label resolves through next-intl.
 */

import { describe, it, expect, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import { createElement } from "react"
import { NextIntlClientProvider } from "next-intl"
import frMessages from "@/../messages/fr.json"
import { LegalFooter } from "../legal-footer"

vi.mock("@i18n/navigation", () => ({
  Link: ({
    children,
    href,
    ...rest
  }: React.ComponentProps<"a"> & { href: string }) =>
    createElement(
      "a",
      { ...rest, href: typeof href === "string" ? href : "/" },
      children,
    ),
}))

function renderFooter() {
  return render(
    <NextIntlClientProvider locale="fr" messages={frMessages}>
      <LegalFooter />
    </NextIntlClientProvider>,
  )
}

describe("LegalFooter", () => {
  it("links to all legal routes including /decisions-automatisees (RGPD art. 22) and /legal/registre (D4)", () => {
    renderFooter()
    const expected = [
      // CNIL — long privacy policy is the single source of truth.
      // The legacy short /privacy URL was merged in (May 2026).
      "/legal/politique-confidentialite",
      "/cookies",
      "/legal",
      "/legal/cgu",
      "/legal/cgv",
      "/legal/code-de-conduite",
      "/sous-processeurs",
      "/decisions-automatisees",
      // D4 (GDPR Phase C) — pointer to the legal-documents section.
      "/legal/registre",
    ]
    for (const href of expected) {
      const link = document.querySelector(`a[href="${href}"]`)
      expect(link, `expected anchor with href=${href}`).not.toBeNull()
    }
  })

  it("does NOT expose the legacy short /privacy URL", () => {
    renderFooter()
    // After the May 2026 merge, the only privacy entry point is the
    // long /legal/politique-confidentialite. A bare /privacy link
    // would resurrect the CNIL-flagged duality.
    const stale = document.querySelector('a[href="/privacy"]')
    expect(stale).toBeNull()
  })

  it("exposes the DPO contact via a mailto: link", () => {
    renderFooter()
    const mailto = document.querySelector('a[href^="mailto:"]')
    expect(mailto).not.toBeNull()
    expect((mailto as HTMLAnchorElement).href.startsWith("mailto:")).toBe(true)
  })

  it("surfaces a persistent 'Manage cookies' affordance (CNIL 2020 §6.3)", () => {
    renderFooter()
    const manage = document.querySelector('a[aria-label]')
    // The CookieConsentManageButton sets aria-label from i18n; the
    // exact label is asserted in the manage-button test. We only
    // require its presence in the footer here.
    const manageButtons = document.querySelectorAll('[aria-label]')
    const found = Array.from(manageButtons).some((el) =>
      (el.getAttribute("aria-label") ?? "").toLowerCase().includes("cookie"),
    )
    expect(manage, "footer should expose at least one aria-labeled element").not.toBeNull()
    expect(found, "expected a cookie-management affordance in the footer").toBe(true)
  })

  it("renders no raw legal.footer.* i18n keys", () => {
    renderFooter()
    const text = screen.getByRole("contentinfo").textContent ?? ""
    expect(text).not.toMatch(/legal\.footer\./)
  })
})
