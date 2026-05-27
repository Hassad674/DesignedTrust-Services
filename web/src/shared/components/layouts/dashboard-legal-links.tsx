"use client"

import { useLocale, useTranslations } from "next-intl"

import { legalHref } from "@i18n/routing"
import { OpenSourceBadge } from "@/shared/components/open-source-badge"

/**
 * Minimalist legal links bar shown at the bottom of every authenticated
 * dashboard page (mirror of Malt's footer pattern). Discrete styling
 * — small muted-foreground text, comma-separated horizontal layout —
 * so it never competes with the page's primary content.
 *
 * Rendered inside the scrollable <main>, NOT position-fixed: it
 * scrolls with the content so the chrome stays single (DashboardShell
 * top nav + sidebar) rather than introducing a third persistent bar.
 *
 * Uses plain `<a>` (not next-intl `<Link>`) so the rendered href can
 * be the locale-aware EN-named slug (`/legal/terms`,
 * `/subprocessors`, …) via `legalHref(canonical, locale)`, without
 * enumerating every other app route in `defineRouting({ pathnames })`.
 * The Next.js rewrites in `next.config.ts` then map those EN URLs to
 * the FR on-disk routes. This mirrors LandingFooter / LegalFooter so
 * EN visitors never land on a FR-named URL after clicking.
 */
const LINKS = [
  { canonical: "/legal/cgu", labelKey: "cgu" },
  { canonical: "/legal/cgv", labelKey: "cgv" },
  { canonical: "/legal/code-de-conduite", labelKey: "codeOfConduct" },
  { canonical: "/legal/politique-confidentialite", labelKey: "privacy" },
  { canonical: "/cookies", labelKey: "cookies" },
  { canonical: "/legal", labelKey: "notices" },
  { canonical: "/sous-processeurs", labelKey: "subprocessors" },
  { canonical: "/decisions-automatisees", labelKey: "automatedDecisions" },
] as const

export function DashboardLegalLinks() {
  const t = useTranslations("dashboardLegalLinks")
  const locale = useLocale()

  return (
    <nav
      aria-label={t("ariaLabel")}
      className="mt-12 flex flex-wrap items-center justify-center gap-x-4 gap-y-2 border-t border-border pt-6 text-xs text-muted-foreground"
    >
      {LINKS.map((link) => (
        <span key={link.canonical} className="flex items-center gap-x-4">
          <a
            href={legalHref(link.canonical, locale)}
            className="hover:text-foreground hover:underline underline-offset-4 transition-colors"
          >
            {t(link.labelKey)}
          </a>
          <span aria-hidden="true" className="text-border">
            ·
          </span>
        </span>
      ))}
      <OpenSourceBadge />
    </nav>
  )
}
