"use client"

import { Paperclip } from "lucide-react"
import { useTranslations } from "next-intl"

// ReportAttachmentsHint — shown to anonymous reporters in place of the
// upload zone. Attachments are logged-in-only (the presign endpoint
// 401s anonymous callers), so we gently explain that text-only is fine
// and signing in unlocks media — without blocking the submission.
export function ReportAttachmentsHint() {
  const t = useTranslations("feedback")
  return (
    <p className="flex items-center gap-2 rounded-xl border border-dashed border-border bg-background px-4 py-3 text-xs text-muted-foreground">
      <Paperclip className="h-4 w-4 shrink-0" strokeWidth={1.6} aria-hidden="true" />
      {t("attachments_anonymous_hint")}
    </p>
  )
}
