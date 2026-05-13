/**
 * `legalHref()` unit tests.
 *
 * Why a dedicated suite for a 6-line helper: this is the single source
 * of truth for the locale-aware legal URL on every surface that cannot
 * use next-intl's typed `<Link>` (cookie consent banner, server-rendered
 * `<a>` markup in MDX, raw href interpolation). A regression in the
 * mapping would silently 404 every legal link on the EN locale and
 * break GDPR compliance — over-test it.
 *
 * Pinned cases:
 *   - canonical FR slug stays FR-prefixed on FR locale (non-default
 *     locale + `localePrefix: as-needed` → `/fr/<slug>`);
 *   - canonical FR slug rewrites to the EN-named slug on EN locale,
 *     with NO `/en` prefix (EN is default + as-needed);
 *   - a path NOT in the legal table falls through to the canonical
 *     argument (no crash, no implicit prefix injection beyond the
 *     usual locale prefix);
 *   - the table contains every legal route in scope as of 2026-05-13
 *     so a silent typo doesn't strip a link from the EN locale.
 */
import { describe, expect, it } from "vitest"

import { legalHref, legalPathnames, routing } from "@i18n/routing"

describe("legalHref", () => {
  it("returns the FR canonical slug under /fr on FR locale", () => {
    expect(legalHref("/legal/cgu", "fr")).toBe("/fr/legal/cgu")
    expect(legalHref("/sous-processeurs", "fr")).toBe(
      "/fr/sous-processeurs",
    )
    expect(legalHref("/legal/politique-confidentialite", "fr")).toBe(
      "/fr/legal/politique-confidentialite",
    )
  })

  it("returns the EN-named slug WITHOUT prefix on EN (default) locale", () => {
    expect(legalHref("/legal/cgu", "en")).toBe("/legal/terms")
    expect(legalHref("/legal/cgv", "en")).toBe("/legal/sales-terms")
    expect(legalHref("/legal/politique-confidentialite", "en")).toBe(
      "/legal/privacy",
    )
    expect(legalHref("/sous-processeurs", "en")).toBe("/subprocessors")
    expect(legalHref("/legal/registre", "en")).toBe(
      "/legal/processing-register",
    )
    expect(legalHref("/legal/aipd", "en")).toBe("/legal/dpia")
    expect(legalHref("/legal/code-de-conduite", "en")).toBe(
      "/legal/code-of-conduct",
    )
    expect(legalHref("/decisions-automatisees", "en")).toBe(
      "/automated-decisions",
    )
  })

  it("returns the canonical path unchanged for unmapped routes", () => {
    // /cookies + /legal share the same segment on both locales, so
    // they intentionally live outside `legalPathnames`. They must
    // still resolve correctly through `legalHref`.
    expect(legalHref("/cookies", "en")).toBe("/cookies")
    expect(legalHref("/cookies", "fr")).toBe("/fr/cookies")
    expect(legalHref("/legal", "en")).toBe("/legal")
    expect(legalHref("/legal", "fr")).toBe("/fr/legal")
    // Random un-known path: pass through, but still respect the prefix
    expect(legalHref("/totally-unknown", "en")).toBe("/totally-unknown")
    expect(legalHref("/totally-unknown", "fr")).toBe("/fr/totally-unknown")
  })

  it("aligns with the active routing config (default locale + locales)", () => {
    // If someone bumps the default locale in routing.ts the helper's
    // prefix logic must follow. This test fails loudly when the
    // contract drifts.
    expect(routing.defaultLocale).toBe("en")
    expect(routing.locales).toEqual(["en", "fr"])
  })

  it("`legalPathnames` declares every route in the GDPR corpus", () => {
    // Pin the canonical keys so a future edit cannot silently drop
    // one. Update this list IF AND ONLY IF you also update
    // `next.config.ts` rewrites and the cookie consent + legal footer
    // link tables — the four must stay in lockstep.
    expect(Object.keys(legalPathnames).sort()).toEqual(
      [
        "/legal/cgu",
        "/legal/cgv",
        "/legal/politique-confidentialite",
        "/sous-processeurs",
        "/legal/registre",
        "/legal/aipd",
        "/legal/code-de-conduite",
        "/decisions-automatisees",
      ].sort(),
    )
  })
})
