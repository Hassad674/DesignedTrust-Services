"use client"

import { useState } from "react"
import { Check, ChevronRight, X } from "lucide-react"
import { useTranslations } from "next-intl"
import { Link } from "@i18n/navigation"

import { useProfileCompletion } from "../hooks/use-profile-completion"
import type { ProfileCompletionSection } from "../api/profile-completion-api"
import { Button } from "@/shared/components/ui/button"
import { cn } from "@/shared/lib/utils"

// ProfileCompletionBarProps controls how the bar lays out — sidebar
// vs page header — without forking the component. Keeps the prop
// surface under the 4-cap.
type ProfileCompletionBarProps = {
  // variant changes the visual density: "sidebar" is compact (used in
  // the sidebar user card); "page" is the larger card used at the top
  // of the profile page. Default "page".
  variant?: "sidebar" | "page"
  // collapsed is forwarded by the sidebar so the bar can shrink to a
  // single percent label when the sidebar is collapsed. Has no effect
  // for the page variant.
  collapsed?: boolean
  // hideWhenComplete suppresses the bar at 100%. Default false — the
  // page variant keeps the bar visible to celebrate the milestone.
  hideWhenComplete?: boolean
}

// ProfileCompletionBar renders "Profil rempli à X%" with a Soleil-
// corail progress fill. Clicking the bar opens a modal listing the
// missing sections with click-through links to the editor pages.
export function ProfileCompletionBar(props: ProfileCompletionBarProps) {
  const { variant = "page", collapsed = false, hideWhenComplete = false } = props
  const t = useTranslations("profileCompletion")
  const [modalOpen, setModalOpen] = useState(false)
  const { data, isLoading } = useProfileCompletion()

  if (isLoading || !data) return null
  if (hideWhenComplete && data.percent >= 100) return null

  const missingCount = data.total_sections - data.filled_sections
  const isComplete = data.percent >= 100
  const a11yLabel = t("a11yLabel", { percent: data.percent })

  if (variant === "sidebar" && collapsed) {
    return (
      <div
        className="mx-auto flex h-9 w-9 items-center justify-center rounded-full bg-primary-soft text-[11px] font-semibold text-primary-deep"
        title={a11yLabel}
        aria-label={a11yLabel}
      >
        {data.percent}%
      </div>
    )
  }

  return (
    <>
      <button
        type="button"
        onClick={() => setModalOpen(true)}
        className={cn(
          "group block w-full rounded-xl text-left transition-colors",
          "focus-visible:outline-2 focus-visible:outline-ring focus-visible:outline-offset-2",
          variant === "page"
            ? "bg-card border border-border p-5 shadow-[var(--shadow-card)] hover:bg-primary-soft/40"
            : "bg-background hover:bg-primary-soft/40 p-3",
        )}
        aria-label={a11yLabel}
      >
        <div className="flex items-center justify-between gap-3">
          <div className="min-w-0 flex-1">
            <p
              className={cn(
                "truncate font-medium text-foreground",
                variant === "page" ? "font-serif text-lg" : "text-sm",
              )}
            >
              {t("title", { percent: data.percent })}
            </p>
            <p className="mt-0.5 text-xs text-muted-foreground">
              {isComplete
                ? t("subtitleComplete")
                : t("subtitle", {
                    filled: data.filled_sections,
                    total: data.total_sections,
                  })}
            </p>
          </div>
          {!isComplete && (
            <span
              className="inline-flex shrink-0 items-center gap-1 rounded-full bg-primary-soft px-2.5 py-1 text-[11px] font-semibold uppercase tracking-wider text-primary-deep"
            >
              {missingCount}
              <ChevronRight className="h-3 w-3" aria-hidden="true" />
            </span>
          )}
        </div>
        <div
          className="mt-3 h-2 w-full overflow-hidden rounded-full bg-muted"
          role="progressbar"
          aria-valuenow={data.percent}
          aria-valuemin={0}
          aria-valuemax={100}
        >
          <div
            className={cn(
              "h-full rounded-full transition-[width] duration-500 ease-out",
              isComplete ? "bg-success" : "bg-primary",
            )}
            style={{ width: `${data.percent}%` }}
          />
        </div>
      </button>

      {modalOpen && (
        <ProfileCompletionModal
          sections={data.sections}
          percent={data.percent}
          onClose={() => setModalOpen(false)}
        />
      )}
    </>
  )
}

type ProfileCompletionModalProps = {
  sections: ProfileCompletionSection[]
  percent: number
  onClose: () => void
}

function ProfileCompletionModal(props: ProfileCompletionModalProps) {
  const { sections, percent, onClose } = props
  const t = useTranslations("profileCompletion")

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-foreground/30 p-4 backdrop-blur-sm"
      role="dialog"
      aria-modal="true"
      aria-labelledby="profile-completion-modal-title"
      onClick={onClose}
    >
      <div
        className="w-full max-w-md rounded-2xl bg-card p-6 shadow-[var(--shadow-card-strong)]"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-start justify-between gap-3">
          <div>
            <h2
              id="profile-completion-modal-title"
              className="font-serif text-xl font-medium text-foreground"
            >
              {t("modalTitle", { percent })}
            </h2>
            <p className="mt-1 text-sm text-muted-foreground">
              {t("modalSubtitle")}
            </p>
          </div>
          <Button
            variant="ghost"
            size="auto"
            onClick={onClose}
            className="rounded-md p-2 text-muted-foreground hover:text-foreground"
            aria-label={t("modalCloseLabel")}
          >
            <X className="h-4 w-4" aria-hidden="true" />
          </Button>
        </div>
        <ul className="mt-5 space-y-2" data-testid="completion-section-list">
          {sections.map((section) => (
            <SectionRow key={section.key} section={section} onNavigate={onClose} />
          ))}
        </ul>
      </div>
    </div>
  )
}

type SectionRowProps = {
  section: ProfileCompletionSection
  onNavigate: () => void
}

function SectionRow({ section, onNavigate }: SectionRowProps) {
  const t = useTranslations()
  // The label_key is a fully-qualified i18n path the backend ships
  // (e.g. "profile.completion.section.title"). next-intl resolves it
  // at the root scope so we always pass the full path to t().
  const label = t(section.label_key)

  if (section.filled) {
    return (
      <li className="flex items-center gap-3 rounded-lg bg-primary-soft/40 px-3 py-2 text-sm">
        <Check className="h-4 w-4 shrink-0 text-success" aria-hidden="true" />
        <span className="truncate text-muted-foreground line-through">
          {label}
        </span>
      </li>
    )
  }

  return (
    <li>
      <Link
        href={section.completion_path}
        onClick={onNavigate}
        className="flex items-center justify-between gap-3 rounded-lg border border-border px-3 py-2 text-sm text-foreground transition-colors hover:bg-primary-soft/60 focus-visible:outline-2 focus-visible:outline-ring focus-visible:outline-offset-2"
      >
        <span className="truncate">{label}</span>
        <ChevronRight className="h-4 w-4 shrink-0 text-muted-foreground" aria-hidden="true" />
      </Link>
    </li>
  )
}
