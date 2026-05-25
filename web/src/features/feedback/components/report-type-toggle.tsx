"use client"

import { Bug, ShieldAlert } from "lucide-react"
import { useTranslations } from "next-intl"
import { cn } from "@/shared/lib/utils"
import type { FeedbackType } from "../types"

interface ReportTypeToggleProps {
  value: FeedbackType
  onChange: (value: FeedbackType) => void
}

const OPTIONS: { value: FeedbackType; labelKey: string; Icon: typeof Bug }[] = [
  { value: "bug", labelKey: "type_bug", Icon: Bug },
  { value: "security", labelKey: "type_security", Icon: ShieldAlert },
]

// ReportTypeToggle — a two-option segmented control implemented as an
// ARIA radiogroup so arrow keys move between Bug / Faille de sécurité
// and the active option is announced. The corail-soft active state pairs
// the colour with an icon + label so meaning never rides on colour alone.
export function ReportTypeToggle({ value, onChange }: ReportTypeToggleProps) {
  const t = useTranslations("feedback")
  return (
    <div
      role="radiogroup"
      aria-label={t("type_legend")}
      className="grid grid-cols-2 gap-2"
    >
      {OPTIONS.map(({ value: optionValue, labelKey, Icon }) => {
        const active = value === optionValue
        return (
          <button
            key={optionValue}
            type="button"
            role="radio"
            aria-checked={active}
            onClick={() => onChange(optionValue)}
            className={cn(
              "flex items-center justify-center gap-2 rounded-xl border px-3 py-2.5 text-sm font-medium",
              "transition-colors duration-150 ease-out",
              "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/30 focus-visible:ring-offset-2 focus-visible:ring-offset-background",
              active
                ? "border-primary bg-primary-soft text-primary-deep"
                : "border-border bg-card text-muted-foreground hover:border-border-strong hover:text-foreground",
            )}
          >
            <Icon className="h-4 w-4" strokeWidth={1.6} aria-hidden="true" />
            {t(labelKey)}
          </button>
        )
      })}
    </div>
  )
}
