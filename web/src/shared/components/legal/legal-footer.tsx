import { useLocale, useTranslations } from "next-intl"

import { legalHref } from "@i18n/routing"
import { getDpoEmail } from "@/shared/lib/dpo"

// LegalFooter — minimalist legal-link footer rendered under public
// pages. Surfaces the 6 legal placeholder routes plus the DPO email so
// every visitor has a one-click path to exercise their RGPD rights.
//
// Server-renderable (no client interaction). Uses Soleil v2 tokens.
//
// Uses plain `<a>` (not next-intl `<Link>`) so the rendered href can
// be the locale-aware EN-named slug (`/legal/terms`,
// `/subprocessors`, …) without enumerating every other app route in
// `defineRouting({ pathnames })`. The Next.js rewrites in
// `next.config.ts` then map those EN URLs to the FR on-disk routes.
export function LegalFooter() {
  const t = useTranslations("legal.footer")
  const locale = useLocale()
  const dpoEmail = getDpoEmail()
  const year = new Date().getFullYear()

  const links: ReadonlyArray<{ canonical: string; key: string }> = [
    { canonical: "/privacy", key: "privacy" },
    { canonical: "/cookies", key: "cookies" },
    { canonical: "/legal", key: "legal" },
    { canonical: "/cgu", key: "cgu" },
    { canonical: "/cgv", key: "cgv" },
    { canonical: "/sous-processeurs", key: "subprocessors" },
    { canonical: "/decisions-automatisees", key: "automatedDecisions" },
    // D4 (GDPR Phase C) — link to /legal/registre as the canonical
    // entry point to the documents section. The /legal index page
    // surfaces the full list (registre, AIPD, DPA, politique, CGU, CGV).
    { canonical: "/legal/registre", key: "documents" },
  ]

  return (
    <footer
      role="contentinfo"
      aria-label={t("ariaLabel")}
      className="border-t border-border bg-background"
    >
      <div className="mx-auto flex max-w-7xl flex-col gap-3 px-6 py-6 text-xs text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
        <nav aria-label={t("ariaLabel")} className="flex flex-wrap items-center gap-x-4 gap-y-2">
          {links.map(({ canonical, key }) => (
            <a
              key={canonical}
              href={legalHref(canonical, locale)}
              className="hover:text-foreground hover:underline"
            >
              {t(key)}
            </a>
          ))}
          <a
            href={`mailto:${dpoEmail}`}
            className="hover:text-foreground hover:underline"
          >
            {t("dpoContact")}
          </a>
        </nav>
        <p className="text-muted-foreground/80">{t("copyright", { year })}</p>
      </div>
    </footer>
  )
}
