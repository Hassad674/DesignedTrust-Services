"use client"

import { useTranslations } from "next-intl"
import { Cookie } from "lucide-react"
import * as CookieConsent from "vanilla-cookieconsent"

import { Link } from "@i18n/navigation"

/**
 * Persistent "Manage cookies" affordance — CNIL Recommendation 2020
 * point 6.3: the consent withdrawal must be exposed as easily as the
 * consent itself. The original CMP banner only fires on first visit;
 * after that the user has no obvious way back into the preferences
 * modal. This button is the missing link.
 *
 * Two render modes:
 * - "inline"  — discreet link sitting inside footers (landing + legal
 *               pages). Visually neutral, no fixed positioning.
 * - "floating"— floating pill at the bottom-left corner of public
 *               pages. Reserved for surfaces where the inline footer
 *               link is hard to reach (e.g. long marketing pages).
 *
 * Both render an anchor first (so the link is meaningful in HTML even
 * when JS is disabled, e.g. `/cookies` page) and override the click
 * to call `CookieConsent.showPreferences()` so the user can adjust
 * their choices on the same page without a full navigation.
 */
export type CookieConsentManageButtonVariant = "inline" | "floating"

interface CookieConsentManageButtonProps {
  variant?: CookieConsentManageButtonVariant
  className?: string
}

const FLOATING_CLASSES =
  "fixed bottom-4 left-4 z-40 flex items-center gap-2 rounded-full border border-border bg-card px-3 py-2 text-sm text-foreground shadow-md transition-colors hover:bg-muted focus:outline-none focus-visible:ring-2 focus-visible:ring-accent"

const INLINE_CLASSES =
  "inline-flex items-center gap-1.5 text-sm text-muted-foreground underline-offset-4 transition-colors hover:text-foreground hover:underline focus:outline-none focus-visible:ring-2 focus-visible:ring-accent rounded-sm"

export function CookieConsentManageButton({
  variant = "inline",
  className,
}: CookieConsentManageButtonProps) {
  const t = useTranslations("cookieConsent.banner")

  const onClick = (event: React.MouseEvent<HTMLAnchorElement>) => {
    // Best-effort: open the CMP preferences modal. If the CMP hasn't
    // booted (e.g. SSR mismatch, third-party script blocked), fall
    // back to the natural `/cookies` navigation handled by the
    // wrapping <Link>.
    try {
      CookieConsent.showPreferences()
      event.preventDefault()
    } catch {
      // navigation continues to /cookies
    }
  }

  const base = variant === "floating" ? FLOATING_CLASSES : INLINE_CLASSES

  return (
    <Link
      href="/cookies"
      onClick={onClick}
      aria-label={t("manageAria")}
      className={className ? `${base} ${className}` : base}
    >
      <Cookie className="size-4" aria-hidden />
      <span className={variant === "floating" ? "hidden sm:inline" : ""}>
        {t("manageLabel")}
      </span>
    </Link>
  )
}
