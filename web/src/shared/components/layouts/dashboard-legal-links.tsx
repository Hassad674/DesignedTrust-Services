"use client"

import { useTranslations } from "next-intl"
import { Link } from "@i18n/navigation"

/**
 * Minimalist legal links bar shown at the bottom of every authenticated
 * dashboard page (mirror of Malt's footer pattern). Discrete styling
 * — small muted-foreground text, comma-separated horizontal layout —
 * so it never competes with the page's primary content.
 *
 * Rendered inside the scrollable <main>, NOT position-fixed: it
 * scrolls with the content so the chrome stays single (DashboardShell
 * top nav + sidebar) rather than introducing a third persistent bar.
 */
const LINKS = [
  { href: "/legal/cgu", labelKey: "cgu" },
  { href: "/legal/cgv", labelKey: "cgv" },
  { href: "/legal/code-de-conduite", labelKey: "codeOfConduct" },
  { href: "/legal/politique-confidentialite", labelKey: "privacy" },
  { href: "/cookies", labelKey: "cookies" },
  { href: "/legal", labelKey: "notices" },
  { href: "/sous-processeurs", labelKey: "subprocessors" },
  { href: "/decisions-automatisees", labelKey: "automatedDecisions" },
] as const

export function DashboardLegalLinks() {
  const t = useTranslations("dashboardLegalLinks")

  return (
    <nav
      aria-label={t("ariaLabel")}
      className="mt-12 flex flex-wrap items-center justify-center gap-x-4 gap-y-2 border-t border-border pt-6 text-xs text-muted-foreground"
    >
      {LINKS.map((link, idx) => (
        <span key={link.href} className="flex items-center gap-x-4">
          <Link
            href={link.href}
            className="hover:text-foreground hover:underline underline-offset-4 transition-colors"
          >
            {t(link.labelKey)}
          </Link>
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
