"use client"

import { useSearchParams } from "next/navigation"
import { useTranslations } from "next-intl"
import { useRouter } from "@i18n/navigation"
import { OpportunityList } from "@/features/job/components/opportunity-list"
import { ApplicationList } from "@/features/job/components/application-list"
import { Button } from "@/shared/components/ui/button"
import { cn } from "@/shared/lib/utils"

/**
 * URL-bookmarkable tab switcher for the Opportunités page. Mirrors the
 * `?tab=` pattern established by `invoices-tabs.tsx` so the two surfaces
 * feel consistent — single URL, no breaking SEO, easy share.
 *
 * Lives in the route directory rather than inside any feature folder
 * because composition is the page's job (per `web/CLAUDE.md` —
 * features never import other features). Importing
 * `features/job` twice (for `OpportunityList` and `ApplicationList`)
 * is the legitimate composition point.
 *
 * Lazy-load contract: the inactive tab is NEVER mounted. `ApplicationList`
 * (the moved-in "Mes candidatures" surface) only fires its TanStack Query
 * once the user clicks the second tab; the active tab keeps its query
 * cache via TanStack's defaults so toggling back and forth is free after
 * the first activation.
 */
const VALID_TABS = ["all", "applications"] as const
type OpportunitiesTab = (typeof VALID_TABS)[number]
const DEFAULT_TAB: OpportunitiesTab = "all"

export function OpportunitiesTabs() {
  const t = useTranslations("opportunitiesTabs")
  const searchParams = useSearchParams()
  const router = useRouter()

  const rawTab = searchParams.get("tab") || DEFAULT_TAB
  const tab: OpportunitiesTab = (VALID_TABS as readonly string[]).includes(
    rawTab,
  )
    ? (rawTab as OpportunitiesTab)
    : DEFAULT_TAB

  function handleTabChange(next: OpportunitiesTab) {
    if (next === DEFAULT_TAB) {
      router.replace("/opportunities")
      return
    }
    router.replace(`/opportunities?tab=${next}`)
  }

  return (
    <div className="space-y-6">
      <nav
        role="tablist"
        aria-label={t("navLabel")}
        className="inline-flex items-center gap-1 rounded-full border border-border bg-card p-1 shadow-[var(--shadow-card)]"
      >
        <TabButton
          isActive={tab === "all"}
          onClick={() => handleTabChange("all")}
          controls="panel-opportunities-all"
          id="tab-opportunities-all"
          label={t("allTab")}
        />
        <TabButton
          isActive={tab === "applications"}
          onClick={() => handleTabChange("applications")}
          controls="panel-opportunities-applications"
          id="tab-opportunities-applications"
          label={t("applicationsTab")}
        />
      </nav>

      <div
        role="tabpanel"
        id="panel-opportunities-all"
        aria-labelledby="tab-opportunities-all"
        hidden={tab !== "all"}
      >
        {tab === "all" ? <OpportunityList /> : null}
      </div>
      <div
        role="tabpanel"
        id="panel-opportunities-applications"
        aria-labelledby="tab-opportunities-applications"
        hidden={tab !== "applications"}
      >
        {tab === "applications" ? <ApplicationList /> : null}
      </div>
    </div>
  )
}

type TabButtonProps = {
  isActive: boolean
  onClick: () => void
  controls: string
  id: string
  label: string
}

function TabButton({ isActive, onClick, controls, id, label }: TabButtonProps) {
  return (
    <Button
      variant="ghost"
      size="auto"
      type="button"
      role="tab"
      id={id}
      aria-controls={controls}
      aria-selected={isActive}
      onClick={onClick}
      className={cn(
        "inline-flex items-center rounded-full px-4 py-2 text-[13px] font-semibold transition-colors",
        isActive
          ? "bg-primary-soft text-[var(--primary-deep)] hover:bg-primary-soft"
          : "text-muted-foreground hover:text-foreground",
      )}
    >
      {label}
    </Button>
  )
}
