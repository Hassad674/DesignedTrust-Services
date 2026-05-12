"use client"

import { useState } from "react"
import { FileText } from "lucide-react"
import { useTranslations } from "next-intl"
import { ProjectHistoryCard, type ProjectHistoryCardEntry } from "./project-history-card"
import { useProjectHistory } from "@/shared/hooks/profile/use-project-history"
import { ReviewModal } from "@/shared/components/review/review-modal"
import type { ReviewSide } from "@/shared/types/review"

interface ProjectHistorySectionProps {
  orgId: string | undefined
  readOnly?: boolean
  // When true the section skips the query and always renders the
  // empty state. Used by the referrer profile where the freelance
  // project history would otherwise leak across personas (both share
  // the same organization_id).
  forceEmpty?: boolean
  emptyOverride?: {
    title: string
    description: string
  }
  /**
   * Side of the double-blind review the CURRENT VIEWER would submit
   * when clicking a row. When undefined, rows are non-interactive.
   * The parent profile page is responsible for computing this from
   * the relationship between the viewer's org and the profile owner
   * (client → provider profile means the viewer reviews provider,
   * provider → client profile means the viewer reviews client).
   */
  reviewSide?: ReviewSide
}

// ProjectHistorySection is the generic "completed projects" block
// used on every profile persona. Features pass `emptyOverride` when
// they need a persona-specific empty state copy — the referrer page
// uses this to say "no referral deals yet" rather than the default
// freelance phrasing. Keeps data fetching + rendering in one place
// so features don't re-implement the skeleton/empty/error trio.
export function ProjectHistorySection(props: ProjectHistorySectionProps) {
  const { orgId, readOnly = false, forceEmpty = false, emptyOverride, reviewSide } = props
  const t = useTranslations("profile")
  const query = useProjectHistory(forceEmpty ? undefined : orgId)
  const data = forceEmpty ? undefined : query.data
  const isLoading = forceEmpty ? false : query.isLoading
  const isError = forceEmpty ? false : query.isError

  const entries = forceEmpty ? [] : (data?.data ?? [])
  const count = entries.length

  const [reviewTarget, setReviewTarget] = useState<ProjectHistoryCardEntry | null>(
    null,
  )

  // Public viewers normally want the section hidden when there is
  // nothing to show — except when the consumer ships a persona-
  // specific empty state copy (e.g. the referrer "no deals yet" line),
  // in which case we keep the card visible so visitors see the
  // persona's value prop even with zero projects.
  const hasEmptyOverride = Boolean(emptyOverride)
  if (
    readOnly &&
    !isLoading &&
    !hasEmptyOverride &&
    (isError || count === 0)
  ) {
    return null
  }

  // Only wire the open-review callback when a side has been resolved
  // by the parent (the viewer is a counterparty of these projects).
  const handleOpenReview = reviewSide
    ? (entry: ProjectHistoryCardEntry) => setReviewTarget(entry)
    : undefined

  return (
    <section className="bg-card border border-border rounded-2xl p-5 shadow-[var(--shadow-card)] sm:p-7">
      <HistoryHeader count={count} />

      {isLoading ? (
        <HistorySkeleton />
      ) : isError ? (
        <div className="rounded-xl border border-destructive/20 bg-destructive/5 p-4 text-sm text-destructive">
          {t("historyLoadError")}
        </div>
      ) : count === 0 ? (
        <EmptyState override={emptyOverride} />
      ) : (
        <ul className="space-y-4">
          {entries.map((entry) => (
            <li key={entry.proposal_id}>
              <ProjectHistoryCard
                entry={entry}
                onOpenReview={handleOpenReview}
              />
            </li>
          ))}
        </ul>
      )}

      {/* Review modal — mounted at the section level so a single
          ReviewModal instance serves every row. Persona-neutral: the
          modal's `side` prop selects the right form variant. */}
      {reviewTarget && reviewSide && (
        <ReviewModal
          isOpen
          proposalId={reviewTarget.proposal_id}
          proposalTitle={reviewTarget.title}
          side={reviewSide}
          onClose={() => setReviewTarget(null)}
        />
      )}
    </section>
  )
}

interface HistoryHeaderProps {
  count: number
}

function HistoryHeader({ count }: HistoryHeaderProps) {
  const t = useTranslations("profile")
  return (
    <div className="flex items-center gap-3 mb-4">
      <h2 className="font-serif text-lg font-medium tracking-[-0.005em] text-foreground sm:text-xl">
        {t("projectHistory")}
      </h2>
      {count > 0 ? (
        <span className="rounded-full bg-muted text-muted-foreground px-2.5 py-0.5 text-xs font-medium">
          {t("completedCount", { count })}
        </span>
      ) : null}
    </div>
  )
}

interface EmptyStateProps {
  override: ProjectHistorySectionProps["emptyOverride"]
}

function EmptyState({ override }: EmptyStateProps) {
  const t = useTranslations("profile")
  const title = override?.title ?? t("noProjects")
  const description = override?.description ?? t("projectsAppearHere")
  return (
    <div className="flex flex-col items-center justify-center py-10 text-center">
      <div className="w-12 h-12 rounded-full bg-muted flex items-center justify-center mb-3">
        <FileText
          className="w-6 h-6 text-muted-foreground"
          aria-hidden="true"
        />
      </div>
      <p className="text-base font-medium text-foreground mb-1">{title}</p>
      <p className="text-sm text-muted-foreground italic">{description}</p>
    </div>
  )
}

function HistorySkeleton() {
  return (
    <div className="space-y-3" role="status" aria-live="polite">
      {[0, 1, 2].map((i) => (
        <div
          key={i}
          className="h-28 rounded-2xl border border-border bg-muted/40 animate-shimmer"
        />
      ))}
    </div>
  )
}

