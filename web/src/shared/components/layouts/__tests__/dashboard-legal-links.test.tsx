/**
 * DashboardLegalLinks — locale-aware href regression.
 *
 * The dashboard legal bar surfaces 6 legal links beneath every
 * authenticated page. On the FR locale they render under the canonical
 * FR slugs (`/fr/legal/cgu`, `/fr/sous-processeurs`, …). On the EN
 * locale they switch to the English-named segments (`/legal/terms`,
 * `/subprocessors`, …) so the URL bar stays consistent with the
 * rendered language. The component must drive both behaviours from
 * `legalHref()` — anything else regresses GDPR link surfacing for
 * the non-default locale.
 */
import { describe, expect, it } from "vitest"
import { render, screen } from "@testing-library/react"
import { NextIntlClientProvider } from "next-intl"
import frMessages from "@/../messages/fr.json"
import enMessages from "@/../messages/en.json"

import { DashboardLegalLinks } from "../dashboard-legal-links"

function renderBar(locale: "fr" | "en") {
  const messages = locale === "fr" ? frMessages : enMessages
  return render(
    <NextIntlClientProvider locale={locale} messages={messages}>
      <DashboardLegalLinks />
    </NextIntlClientProvider>,
  )
}

describe("DashboardLegalLinks", () => {
  it("renders FR-prefixed canonical hrefs on the FR locale", () => {
    renderBar("fr")
    const nav = screen.getByRole("navigation")
    for (const href of [
      "/fr/legal/cgu",
      "/fr/legal/cgv",
      "/fr/legal/politique-confidentialite",
      "/fr/cookies",
      "/fr/legal",
      "/fr/sous-processeurs",
    ]) {
      const link = nav.querySelector(`a[href="${href}"]`)
      expect(link, `expected anchor with href=${href}`).not.toBeNull()
    }
  })

  it("rewrites legal slugs to their EN segments on the EN locale (no /en prefix, default + as-needed)", () => {
    renderBar("en")
    const nav = screen.getByRole("navigation")
    for (const href of [
      "/legal/terms",
      "/legal/sales-terms",
      "/legal/privacy",
      "/cookies",
      "/legal",
      "/subprocessors",
    ]) {
      const link = nav.querySelector(`a[href="${href}"]`)
      expect(link, `expected anchor with href=${href} on EN`).not.toBeNull()
    }
  })

  it("never emits the legacy FR slug on the EN locale", () => {
    renderBar("en")
    const nav = screen.getByRole("navigation")
    // Regression: when locale is EN we must NOT render the bare FR
    // canonical segments. Otherwise a visitor on EN would land on a
    // FR-named URL after click → bad UX + sitemap drift.
    expect(nav.querySelector('a[href="/legal/cgu"]')).toBeNull()
    expect(nav.querySelector('a[href="/legal/cgv"]')).toBeNull()
    expect(
      nav.querySelector('a[href="/legal/politique-confidentialite"]'),
    ).toBeNull()
    expect(nav.querySelector('a[href="/sous-processeurs"]')).toBeNull()
  })
})
