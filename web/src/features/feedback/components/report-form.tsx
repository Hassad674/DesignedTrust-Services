"use client"

import { useState } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { useLocale, useTranslations } from "next-intl"
import { Button } from "@/shared/components/ui/button"
import { useUser } from "@/shared/hooks/use-user"
import { ReportTypeToggle } from "./report-type-toggle"
import { ReportAttachments } from "./report-attachments"
import { ReportTextFields } from "./report-text-fields"
import { ReportAttachmentsHint } from "./report-attachments-hint"
import { useFeedbackAttachments } from "../hooks/use-feedback-attachments"
import { useSubmitFeedback } from "../hooks/use-submit-feedback"
import { captureFeedbackContext, currentPageUrl } from "../lib/capture-context"
import type { FeedbackType } from "../types"

// Field length bounds mirror the backend domain validation
// (Title 3-200, Description 10-5000). Validating here keeps the submit
// button honest and surfaces a localised message before the round-trip.
// The server remains authoritative (it strips HTML before re-checking).
const reportSchema = z.object({
  title: z.string().trim().min(3).max(200),
  description: z.string().trim().min(10).max(5000),
  reporterEmail: z.string().trim().email().or(z.literal("")),
  // Honeypot: a human leaves this empty. zod does not gate it (the
  // server silently drops a filled value); it is registered only so RHF
  // forwards it into the submit body.
  hp: z.string(),
})

type ReportValues = z.infer<typeof reportSchema>

interface ReportFormProps {
  onSuccess: () => void
}

// ReportForm — the body of the report modal: type toggle, text fields,
// attachments (logged-in) or a sign-in hint (anonymous), and submit.
// Auto-captures page/locale/viewport/UA context; the honeypot ships
// hidden and empty.
export function ReportForm({ onSuccess }: ReportFormProps) {
  const t = useTranslations("feedback")
  const locale = useLocale()
  const { data: user } = useUser()
  const [type, setType] = useState<FeedbackType>("bug")
  const attachments = useFeedbackAttachments()
  const submit = useSubmitFeedback()

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ReportValues>({
    resolver: zodResolver(reportSchema),
    defaultValues: { title: "", description: "", reporterEmail: "", hp: "" },
  })

  function onValid(values: ReportValues) {
    if (values.hp !== "") return // honeypot tripped — never reaches the API
    submit.mutate(
      {
        type,
        title: values.title,
        description: values.description,
        page_url: currentPageUrl(),
        context: captureFeedbackContext({ locale, role: user?.role }),
        reporter_email: user ? "" : values.reporterEmail,
        attachment_keys: user ? attachments.uploadedRefs : [],
        hp: "",
      },
      { onSuccess },
    )
  }

  return (
    <form onSubmit={handleSubmit(onValid)} className="flex flex-col gap-4">
      <ReportTypeToggle value={type} onChange={setType} />

      <ReportTextFields
        register={register}
        errors={errors}
        showEmailField={!user}
      />

      {user ? (
        <ReportAttachments
          attachments={attachments.attachments}
          onAddFiles={attachments.addFiles}
          onRemove={attachments.remove}
          disabled={submit.isPending}
        />
      ) : (
        <ReportAttachmentsHint />
      )}

      {submit.isError && (
        <p role="alert" className="text-sm text-primary-deep">
          {t("submit_error")}
        </p>
      )}

      <Button
        type="submit"
        variant="primary"
        size="md"
        disabled={submit.isPending || attachments.isUploading}
        className="w-full rounded-full"
      >
        {submit.isPending ? t("submit_pending") : t("submit")}
      </Button>
    </form>
  )
}
