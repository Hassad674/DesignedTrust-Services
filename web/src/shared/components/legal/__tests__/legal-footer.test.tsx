/**
 * Phase A.4 + A.5 — LegalFooter test.
 *
 * Asserts:
 *   1. All 6 legal placeholder routes are linked (privacy, cookies,
 *      legal, cgu, cgv, sous-processeurs).
 *   2. The DPO email (NEXT_PUBLIC_DPO_EMAIL or fallback) is rendered
 *      as a mailto: link so visitors have a one-click RGPD contact.
 *   3. No raw i18n key (string starting with "legal.footer.") leaks to
 *      the DOM — every label resolves through next-intl.
 *   4. Legal-corpus hrefs are locale-aware: `/legal/cgu` stays
 *      canonical on FR but becomes `/legal/terms` on EN; same shape
 *      for /sous-processeurs → /subprocessors etc.
 */

import { describe, it, expect } from "vitest"
import { render, screen } from "@testing-library/react"
import { NextIntlClientProvider } from "next-intl"
import enMessages from "@/../messages/en.json"
import frMessages from "@/../messages/fr.json"
import { LegalFooter } from "../legal-footer"

function renderFooter(locale: "fr" | "en" = "fr") {
  const messages = locale === "fr" ? frMessages : enMessages
  return render(
    <NextIntlClientProvider locale={locale} messages={messages}>
      <LegalFooter />
    </NextIntlClientProvider>,
  )
}

describe("LegalFooter", () => {
  it("links to all legal routes including /decisions-automatisees (RGPD art. 22) and /legal/registre (D4) on FR", () => {
    renderFooter("fr")
    // FR locale + as-needed + non-default → /fr/<canonical-or-mapped>
    const expected = [
      "/fr/privacy",
      "/fr/cookies",
      "/fr/legal",
      "/fr/cgu",
      "/fr/cgv",
      "/fr/sous-processeurs",
      "/fr/decisions-automatisees",
      // D4 (GDPR Phase C) — pointer to the legal-documents section.
      "/fr/legal/registre",
    ]
    for (const href of expected) {
      const link = document.querySelector(`a[href="${href}"]`)
      expect(link, `expected anchor with href=${href}`).not.toBeNull()
    }
  })

  it("rewrites the FR legal slugs to their EN counterparts on the EN locale", () => {
    renderFooter("en")
    // EN locale = default + as-needed → no /en prefix.
    // Legal segments are remapped to their English names.
    const expected = [
      "/privacy", // canonical /privacy is identical FR/EN (no mapping)
      "/cookies",
      "/legal",
      "/cgu", // /cgu page (not /legal/cgu) is identical
      "/cgv",
      "/subprocessors", // /sous-processeurs → /subprocessors
      "/automated-decisions", // /decisions-automatisees → /automated-decisions
      "/legal/processing-register", // /legal/registre → /legal/processing-register
    ]
    for (const href of expected) {
      const link = document.querySelector(`a[href="${href}"]`)
      expect(link, `expected anchor with href=${href} on EN`).not.toBeNull()
    }
  })

  it("exposes the DPO contact via a mailto: link", () => {
    renderFooter()
    const mailto = document.querySelector('a[href^="mailto:"]')
    expect(mailto).not.toBeNull()
    expect((mailto as HTMLAnchorElement).href.startsWith("mailto:")).toBe(true)
  })

  it("renders no raw legal.footer.* i18n keys", () => {
    renderFooter()
    const text = screen.getByRole("contentinfo").textContent ?? ""
    expect(text).not.toMatch(/legal\.footer\./)
  })
})
