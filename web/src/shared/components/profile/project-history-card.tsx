"use client"

import { Clock, Euro, PenLine } from "lucide-react"
import { useFormatter, useTranslations } from "next-intl"
import { ReviewCard } from "@/shared/components/ui/review-card"
import { useCanReview } from "@/shared/hooks/review/use-reviews"
import type { Review } from "@/shared/types/review"

// CardEntry is the minimal counterparty-agnostic shape the card
// needs. Both the provider-side project history (proposals where the
// org delivered) and the client-side project history (proposals the
// org paid for) flatten into this same shape — the only difference
// lives upstream (which endpoint populates the list), so the card is
// deliberately persona-neutral.
export type ProjectHistoryCardEntry = {
  proposal_id: string
  title: string
  amount: number
  currency?: string
  completed_at: string
  review: Review | null
}

interface ProjectHistoryCardProps {
  entry: ProjectHistoryCardEntry
  /**
   * When supplied, the card is rendered for an authenticated viewer
   * who may be eligible to leave a review on this proposal. It calls
   * useCanReview(proposal_id) and — when can_review is true AND the
   * counterparty review is not yet visible — renders the "awaiting
   * review" placeholder as a button firing onOpenReview(entry).
   * Anonymous / public visitors do not pass a callback and never
   * trigger the can-review probe.
   */
  onOpenReview?: (entry: ProjectHistoryCardEntry) => void
}

// ProjectHistoryCard renders one completed project row: amount pill
// + completion date in the header, title when set, and either the
// embedded review or the "awaiting review" placeholder below. Every
// profile surface (provider, freelance, referrer, client) uses this
// same card so a review that shows up on one page looks identical on
// every other page where the same proposal appears.
export function ProjectHistoryCard({
  entry,
  onOpenReview,
}: ProjectHistoryCardProps) {
  const t = useTranslations("profile")
  const format = useFormatter()
  const amount = entry.amount / 100
  const completedDate = new Date(entry.completed_at)
  const showTitle = entry.title.trim().length > 0

  // Only probe the can-review endpoint when a parent has expressed
  // interest by passing onOpenReview — anonymous visitors never fire
  // an authenticated call.
  const canReviewProbe = useCanReview(
    onOpenReview ? entry.proposal_id : undefined,
  )
  const canReview =
    !!onOpenReview && canReviewProbe.data?.data.can_review === true

  return (
    <article className="rounded-2xl border border-border bg-card p-5 shadow-sm transition-colors hover:border-primary/30">
      <header className="flex flex-wrap items-center justify-between gap-3">
        <div className="inline-flex items-center gap-1.5 rounded-full bg-gradient-to-r from-primary-soft to-primary-soft/60 px-3 py-1.5 text-sm font-semibold text-primary-deep">
          <Euro className="h-3.5 w-3.5" strokeWidth={2.5} />
          {format.number(amount, {
            style: "currency",
            currency: entry.currency || "EUR",
            maximumFractionDigits: 0,
          })}
        </div>
        <div className="inline-flex items-center gap-1.5 text-xs text-muted-foreground">
          <Clock className="h-3.5 w-3.5" />
          <time dateTime={entry.completed_at}>
            {t("completedOn", {
              date: format.dateTime(completedDate, {
                year: "numeric",
                month: "short",
                day: "numeric",
              }),
            })}
          </time>
        </div>
      </header>

      {showTitle ? (
        <h3 className="mt-3 text-base font-semibold text-foreground">
          {entry.title}
        </h3>
      ) : null}

      <div className="mt-4">
        {entry.review ? (
          <ReviewCard review={entry.review} />
        ) : canReview ? (
          <button
            type="button"
            onClick={() => onOpenReview!(entry)}
            className="group flex w-full items-center gap-2 rounded-xl border border-dashed border-primary/40 bg-primary-soft/40 p-4 text-left text-sm font-medium text-primary-deep transition-colors hover:border-primary hover:bg-primary-soft"
            aria-label={t("leaveYourReview")}
          >
            <PenLine className="h-4 w-4 shrink-0" />
            <span>{t("leaveYourReview")}</span>
          </button>
        ) : (
          <div className="flex items-center gap-2 rounded-xl border border-dashed border-border bg-muted/30 p-4 text-sm text-muted-foreground">
            <Clock className="h-4 w-4 shrink-0" />
            <span>{t("awaitingReview")}</span>
          </div>
        )}
      </div>
    </article>
  )
}
