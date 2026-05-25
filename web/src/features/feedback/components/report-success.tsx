"use client"

import { CheckCircle2 } from "lucide-react"
import { useTranslations } from "next-intl"
import { Button } from "@/shared/components/ui/button"

interface ReportSuccessProps {
  onClose: () => void
}

// ReportSuccess — the post-submit confirmation state. Shown in place of
// the form once the report is accepted; a single action closes (and
// resets) the modal.
export function ReportSuccess({ onClose }: ReportSuccessProps) {
  const t = useTranslations("feedback")
  return (
    <div className="flex flex-col items-center gap-4 py-4 text-center">
      <span className="flex h-14 w-14 items-center justify-center rounded-full bg-success-soft text-success">
        <CheckCircle2 className="h-7 w-7" strokeWidth={1.6} aria-hidden="true" />
      </span>
      <div className="flex flex-col gap-1">
        <h3 className="font-serif text-xl font-medium text-foreground">
          {t("success_title")}
        </h3>
        <p className="text-sm text-muted-foreground">{t("success_body")}</p>
      </div>
      <Button
        type="button"
        variant="primary"
        size="md"
        onClick={onClose}
        className="rounded-full px-6"
      >
        {t("success_close")}
      </Button>
    </div>
  )
}
