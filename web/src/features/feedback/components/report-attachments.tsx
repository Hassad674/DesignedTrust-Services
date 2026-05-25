"use client"

import { useId, useRef } from "react"
import Image from "next/image"
import { Paperclip, Trash2, Video as VideoIcon, Loader2 } from "lucide-react"
import { useTranslations } from "next-intl"
import { cn } from "@/shared/lib/utils"
import { ATTACHMENT_ACCEPT } from "../lib/attachment-constraints"
import type { FeedbackAttachment } from "../hooks/use-feedback-attachments"

interface ReportAttachmentsProps {
  attachments: FeedbackAttachment[]
  onAddFiles: (files: FileList) => void
  onRemove: (id: string) => void
  disabled: boolean
}

// ReportAttachments — the logged-in-only media zone (images + videos).
// Anonymous reporters never see this; the modal renders a gentle hint
// instead. Each file shows a preview, an upload spinner, or a localised
// error, plus a remove control.
export function ReportAttachments({
  attachments,
  onAddFiles,
  onRemove,
  disabled,
}: ReportAttachmentsProps) {
  const t = useTranslations("feedback")
  const inputRef = useRef<HTMLInputElement>(null)
  const labelId = useId()

  return (
    <section aria-labelledby={labelId} className="flex flex-col gap-2">
      <span id={labelId} className="text-sm font-medium text-foreground">
        {t("attachments_label")}
      </span>

      <button
        type="button"
        disabled={disabled}
        onClick={() => inputRef.current?.click()}
        className={cn(
          "flex items-center justify-center gap-2 rounded-xl border border-dashed border-border-strong",
          "bg-background px-4 py-3 text-sm text-muted-foreground",
          "transition-colors duration-150 ease-out hover:border-primary hover:text-foreground",
          "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/30 focus-visible:ring-offset-2 focus-visible:ring-offset-background",
          "disabled:cursor-not-allowed disabled:opacity-60",
        )}
      >
        <Paperclip className="h-4 w-4" strokeWidth={1.6} aria-hidden="true" />
        {t("attachments_add")}
      </button>
      <input
        ref={inputRef}
        type="file"
        accept={ATTACHMENT_ACCEPT}
        multiple
        className="sr-only"
        aria-label={t("attachments_add")}
        onChange={(e) => {
          if (e.target.files?.length) onAddFiles(e.target.files)
          e.target.value = ""
        }}
      />

      <p className="text-xs text-muted-foreground">{t("attachments_formats")}</p>

      {attachments.length > 0 && (
        <ul className="flex flex-col gap-2">
          {attachments.map((attachment) => (
            <AttachmentRow
              key={attachment.id}
              attachment={attachment}
              onRemove={onRemove}
            />
          ))}
        </ul>
      )}
    </section>
  )
}

function AttachmentRow({
  attachment,
  onRemove,
}: {
  attachment: FeedbackAttachment
  onRemove: (id: string) => void
}) {
  const t = useTranslations("feedback")
  return (
    <li className="flex items-center gap-3 rounded-xl border border-border bg-card p-2">
      <AttachmentThumb attachment={attachment} />
      <div className="flex min-w-0 flex-1 flex-col">
        <span className="truncate text-sm text-foreground">
          {attachment.file.name}
        </span>
        <AttachmentStatus attachment={attachment} />
      </div>
      <button
        type="button"
        onClick={() => onRemove(attachment.id)}
        aria-label={t("attachments_remove")}
        className="rounded-lg p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/30"
      >
        <Trash2 className="h-4 w-4" strokeWidth={1.6} aria-hidden="true" />
      </button>
    </li>
  )
}

function AttachmentThumb({ attachment }: { attachment: FeedbackAttachment }) {
  if (attachment.kind === "video") {
    return (
      <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-primary-soft text-primary-deep">
        <VideoIcon className="h-5 w-5" strokeWidth={1.6} aria-hidden="true" />
      </span>
    )
  }
  return (
    <Image
      src={attachment.previewUrl}
      alt=""
      width={40}
      height={40}
      unoptimized
      className="h-10 w-10 shrink-0 rounded-lg object-cover"
    />
  )
}

function AttachmentStatus({ attachment }: { attachment: FeedbackAttachment }) {
  const t = useTranslations("feedback")
  if (attachment.status === "uploading") {
    return (
      <span className="flex items-center gap-1.5 text-xs text-muted-foreground">
        <Loader2 className="h-3 w-3 animate-spin" aria-hidden="true" />
        {t("attachments_uploading")}
      </span>
    )
  }
  if (attachment.status === "error") {
    return (
      <span role="alert" className="text-xs text-primary-deep">
        {t(`attachments_error_${attachment.rejection ?? "upload_failed"}`)}
      </span>
    )
  }
  return (
    <span className="text-xs text-success">{t("attachments_uploaded")}</span>
  )
}
