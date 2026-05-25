"use client"

import { useState } from "react"
import dynamic from "next/dynamic"
import { Bug } from "lucide-react"
import { useTranslations } from "next-intl"
import { Button } from "@/shared/components/ui/button"
import { cn } from "@/shared/lib/utils"

// Lazy-load the modal so the form + upload logic stay out of the initial
// bundle on every page — the button is the only always-mounted cost.
const ReportModal = dynamic(
  () => import("./report-modal").then((m) => ({ default: m.ReportModal })),
  { ssr: false },
)

// ReportButton — the always-visible "Signaler" floating shortcut.
//
// Placement: anchored BOTTOM-LEFT (`fixed bottom-6 left-6`). The
// messaging ChatWidget owns BOTTOM-RIGHT (`fixed bottom-6 right-6`,
// desktop-only) and there is no global bottom navbar, so left/right
// anchoring guarantees the two never overlap on any breakpoint —
// including mobile, where the chat widget is hidden but the report
// button must still appear. Shares the chat widget's `z-50` so both
// float above page content but below the modal overlay (`z-[100]`).
export function ReportButton() {
  const t = useTranslations("feedback")
  const [open, setOpen] = useState(false)

  return (
    <>
      <Button
        type="button"
        variant="ghost"
        size="auto"
        onClick={() => setOpen(true)}
        title={t("button_tooltip")}
        aria-haspopup="dialog"
        aria-label={t("button_label")}
        className={cn(
          "fixed bottom-6 left-6 z-50 inline-flex items-center gap-2.5",
          "h-12 rounded-full pl-2 pr-4",
          "bg-card text-foreground",
          "border border-border",
          "shadow-[0_8px_24px_rgba(42,31,21,0.12)]",
          "transition-all duration-200 ease-out",
          "hover:shadow-[0_12px_28px_rgba(42,31,21,0.16)]",
          "hover:-translate-y-0.5",
        )}
      >
        <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-primary-soft text-primary-deep">
          <Bug className="h-[18px] w-[18px]" strokeWidth={1.6} aria-hidden="true" />
        </span>
        <span className="text-sm font-semibold">{t("button_label")}</span>
      </Button>

      {open && <ReportModal open={open} onClose={() => setOpen(false)} />}
    </>
  )
}
