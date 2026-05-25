"use client"

import { useState } from "react"
import { useTranslations } from "next-intl"
import { Modal } from "@/shared/components/ui/modal"
import { ReportForm } from "./report-form"
import { ReportSuccess } from "./report-success"

interface ReportModalProps {
  open: boolean
  onClose: () => void
}

// ReportModal — the lazy-loaded dialog behind the "Signaler" button.
// Delegates focus-trap / ESC / backdrop / aria-modal to the shared Modal
// primitive and swaps the form for a confirmation once the report is
// accepted. The button mounts this only while open, so every open is a
// fresh form (state resets for free on unmount).
export function ReportModal({ open, onClose }: ReportModalProps) {
  const t = useTranslations("feedback")
  const [submitted, setSubmitted] = useState(false)

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={t("modal_title")}
      maxWidthClassName="max-w-lg"
    >
      {submitted ? (
        <ReportSuccess onClose={onClose} />
      ) : (
        <>
          <p className="mb-4 text-sm text-muted-foreground">
            {t("modal_intro")}
          </p>
          <ReportForm onSuccess={() => setSubmitted(true)} />
        </>
      )}
    </Modal>
  )
}
