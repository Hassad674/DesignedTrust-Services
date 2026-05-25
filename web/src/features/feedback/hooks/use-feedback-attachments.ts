"use client"

import { useCallback, useState } from "react"
import { uploadFeedbackAttachment } from "../api/feedback-upload"
import {
  validateAttachment,
  type AttachmentRejection,
} from "../lib/attachment-constraints"
import type { FeedbackAttachmentKind, FeedbackAttachmentRef } from "../types"

/** One file's lifecycle in the attachment zone. */
export type FeedbackAttachment = {
  /** Stable client id (not the server key) for list rendering + removal. */
  id: string
  file: File
  kind: FeedbackAttachmentKind
  /** Object URL for the preview thumbnail; revoked on removal. */
  previewUrl: string
  status: "uploading" | "uploaded" | "error"
  /** Populated once the presign + PUT succeed — fed into the submit body. */
  ref?: FeedbackAttachmentRef
  /** A localisation reason when the file is rejected or the upload fails. */
  rejection?: AttachmentRejection | "upload_failed"
}

let attachmentSeq = 0
function nextId(): string {
  attachmentSeq += 1
  return `fb-att-${attachmentSeq}`
}

/**
 * useFeedbackAttachments owns the attachment list and the per-file
 * presign→PUT lifecycle for the logged-in report flow. The hook keeps
 * the modal component thin (state + orchestration live here, not in
 * JSX) and exposes only the surface the UI needs.
 */
export function useFeedbackAttachments() {
  const [attachments, setAttachments] = useState<FeedbackAttachment[]>([])

  const update = useCallback(
    (id: string, patch: Partial<FeedbackAttachment>) => {
      setAttachments((prev) =>
        prev.map((a) => (a.id === id ? { ...a, ...patch } : a)),
      )
    },
    [],
  )

  const addFiles = useCallback(
    (files: FileList | File[]) => {
      for (const file of Array.from(files)) {
        const verdict = validateAttachment(file)
        if (!verdict.ok) {
          setAttachments((prev) => [...prev, rejectedEntry(file, verdict.reason)])
          continue
        }
        const entry = uploadingEntry(file, verdict.kind)
        setAttachments((prev) => [...prev, entry])
        void runUpload(entry, update)
      }
    },
    [update],
  )

  const remove = useCallback((id: string) => {
    setAttachments((prev) => {
      const target = prev.find((a) => a.id === id)
      if (target) URL.revokeObjectURL(target.previewUrl)
      return prev.filter((a) => a.id !== id)
    })
  }, [])

  const reset = useCallback(() => {
    setAttachments((prev) => {
      prev.forEach((a) => URL.revokeObjectURL(a.previewUrl))
      return []
    })
  }, [])

  /** Refs for every successfully-uploaded file — the submit payload. */
  const uploadedRefs: FeedbackAttachmentRef[] = attachments
    .filter((a) => a.status === "uploaded" && a.ref)
    .map((a) => a.ref as FeedbackAttachmentRef)

  /** True while at least one file is still mid-upload (blocks submit). */
  const isUploading = attachments.some((a) => a.status === "uploading")

  return { attachments, addFiles, remove, reset, uploadedRefs, isUploading }
}

function uploadingEntry(
  file: File,
  kind: FeedbackAttachmentKind,
): FeedbackAttachment {
  return {
    id: nextId(),
    file,
    kind,
    previewUrl: URL.createObjectURL(file),
    status: "uploading",
  }
}

function rejectedEntry(
  file: File,
  reason: AttachmentRejection,
): FeedbackAttachment {
  return {
    id: nextId(),
    file,
    // Best-effort kind for the icon; the entry is in error state anyway.
    kind: file.type.startsWith("video/") ? "video" : "image",
    previewUrl: URL.createObjectURL(file),
    status: "error",
    rejection: reason,
  }
}

async function runUpload(
  entry: FeedbackAttachment,
  update: (id: string, patch: Partial<FeedbackAttachment>) => void,
): Promise<void> {
  try {
    const ref = await uploadFeedbackAttachment(entry.file, entry.kind)
    update(entry.id, { status: "uploaded", ref })
  } catch {
    update(entry.id, { status: "error", rejection: "upload_failed" })
  }
}
