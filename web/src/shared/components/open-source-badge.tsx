import { Github } from "lucide-react"
import { useTranslations } from "next-intl"
import { cn } from "@/shared/lib/utils"

// Public repository — the project is open source. Hardcoded on purpose:
// it is a fixed, non-tenant, non-env-driven URL (same as the legal-page
// `sourceHref` links that already point at this repo). This is an
// `<a href>` navigation to github.com (NOT a fetch / script src), so no
// Content-Security-Policy change is required.
export const OPEN_SOURCE_REPO_URL =
  "https://github.com/Hassad674/serviceMarketplaceGo"

type OpenSourceBadgeVariant = "inline" | "prominent"

interface OpenSourceBadgeProps {
  // `inline` (default) — discreet text-link sized to sit in a legal-link
  //   bar next to muted-foreground siblings (public footer, dashboard
  //   legal links).
  // `prominent` — a soft pill with a subtle border, for the landing
  //   footer where open-source is a tasteful selling point.
  variant?: OpenSourceBadgeVariant
  className?: string
}

// OpenSourceBadge — a discreet "Open source" badge (GitHub icon + label)
// that links to the public repository in a new tab. Mounted into the
// EXISTING global chrome (public footer + dashboard legal-links bar +
// landing footer) so it appears on every page — public or authenticated
// — without adding a floating element that could collide with the
// feedback FAB (bottom-left) or the chat widget (bottom-right).
//
// a11y: real `<a>` with a descriptive `aria-label`, visible focus ring
// (project-standard `focus-visible:ring-4 ring-primary/20`), and
// `target="_blank" rel="noopener noreferrer"` for safe new-tab opening.
export function OpenSourceBadge({
  variant = "inline",
  className,
}: OpenSourceBadgeProps) {
  const t = useTranslations("openSource")

  return (
    <a
      href={OPEN_SOURCE_REPO_URL}
      target="_blank"
      rel="noopener noreferrer"
      aria-label={t("ariaLabel")}
      title={t("tooltip")}
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full transition-colors",
        "focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-primary/20",
        variant === "inline" &&
          "text-muted-foreground hover:text-foreground hover:underline underline-offset-4",
        variant === "prominent" &&
          "border border-border bg-card px-3 py-1.5 text-foreground hover:border-primary/40 hover:text-primary-deep",
        className,
      )}
    >
      <Github
        className={cn(variant === "prominent" ? "size-4" : "size-3.5")}
        strokeWidth={1.8}
        aria-hidden="true"
      />
      <span className={cn(variant === "prominent" && "text-[13px] font-medium")}>
        {t("label")}
      </span>
    </a>
  )
}
