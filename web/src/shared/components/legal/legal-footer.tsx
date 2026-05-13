import { useTranslations } from "next-intl"
import { Link } from "@i18n/navigation"
import { getDpoEmail } from "@/shared/lib/dpo"
import { CookieConsentManageButton } from "@/shared/components/analytics/cookie-consent-manage-button"

// LegalFooter — minimalist legal-link footer rendered under public
// pages. Surfaces the 7 legal routes plus the DPO email + a persistent
// "Manage cookies" entry point so every visitor has a one-click path
// to exercise their RGPD rights (CNIL Recommendation 2020 point 6.3).
//
// Note: `/privacy` has been merged into `/legal/politique-confidentialite`
// (the long canonical version) — CNIL requires a single policy, not
// two parallel documents. The "privacy" footer key now resolves to the
// long URL.
//
// Server-renderable (no client interaction). The cookie manage button
// is the only client-island in here. Uses Soleil v2 tokens.
export function LegalFooter() {
  const t = useTranslations("legal.footer")
  const dpoEmail = getDpoEmail()
  const year = new Date().getFullYear()

  const links: ReadonlyArray<{ href: string; key: string }> = [
    { href: "/legal/politique-confidentialite", key: "privacy" },
    { href: "/cookies", key: "cookies" },
    { href: "/legal", key: "legal" },
    { href: "/legal/cgu", key: "cgu" },
    { href: "/legal/cgv", key: "cgv" },
    { href: "/legal/code-de-conduite", key: "codeOfConduct" },
    { href: "/sous-processeurs", key: "subprocessors" },
    { href: "/decisions-automatisees", key: "automatedDecisions" },
    // D4 (GDPR Phase C) — link to /legal/registre as the canonical
    // entry point to the documents section. The /legal index page
    // surfaces the full list (registre, AIPD, DPA, politique, CGU, CGV).
    { href: "/legal/registre", key: "documents" },
  ]

  return (
    <footer
      role="contentinfo"
      aria-label={t("ariaLabel")}
      className="border-t border-border bg-background"
    >
      <div className="mx-auto flex max-w-7xl flex-col gap-3 px-6 py-6 text-xs text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
        <nav aria-label={t("ariaLabel")} className="flex flex-wrap items-center gap-x-4 gap-y-2">
          {links.map(({ href, key }) => (
            <Link
              key={href}
              href={href}
              className="hover:text-foreground hover:underline"
            >
              {t(key)}
            </Link>
          ))}
          <a
            href={`mailto:${dpoEmail}`}
            className="hover:text-foreground hover:underline"
          >
            {t("dpoContact")}
          </a>
          <CookieConsentManageButton variant="inline" />
        </nav>
        <p className="text-muted-foreground/80">{t("copyright", { year })}</p>
      </div>
    </footer>
  )
}
