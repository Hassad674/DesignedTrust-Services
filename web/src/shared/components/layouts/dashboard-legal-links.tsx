"use client"

import { useLocale, useTranslations } from "next-intl"

import { legalHref } from "@i18n/routing"

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
 * `/subprocessors`, …) without enumerating every other app route in
 * `defineRouting({ pathnames })`. The Next.js rewrites in
 * `next.config.ts` then map those EN URLs to the FR on-disk routes.
 */
const LINKS = [
  { canonical: "/legal/cgu", labelKey: "cgu" },
  { canonical: "/legal/cgv", labelKey: "cgv" },
  { canonical: "/legal/politique-confidentialite", labelKey: "privacy" },
  { canonical: "/cookies", labelKey: "cookies" },
  { canonical: "/legal", labelKey: "notices" },
  { canonical: "/sous-processeurs", labelKey: "subprocessors" },
] as const

export function DashboardLegalLinks() {
  const t = useTranslations("dashboardLegalLinks")
  const locale = useLocale()

  return (
    <nav
      aria-label={t("ariaLabel")}
      className="mt-12 flex flex-wrap items-center justify-center gap-x-4 gap-y-2 border-t border-border pt-6 text-xs text-muted-foreground"
    >
      {LINKS.map((link, idx) => (
        <span key={link.canonical} className="flex items-center gap-x-4">
          <a
            href={legalHref(link.canonical, locale)}
            className="hover:text-foreground hover:underline underline-offset-4 transition-colors"
          >
            {t(link.labelKey)}
          </a>
          {idx < LINKS.length - 1 && (
            <span aria-hidden="true" className="text-border">
              ·
            </span>
          )}
        </span>
      ))}
    </nav>
  )
}
