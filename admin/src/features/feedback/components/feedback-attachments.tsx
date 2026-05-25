import { useState } from "react"
import { Paperclip } from "lucide-react"
import { Dialog } from "@/shared/components/ui/dialog"
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/components/ui/card"
import type { FeedbackAttachment } from "../types"
import { formatBytes } from "./feedback-format"

type FeedbackAttachmentsProps = {
  attachments: FeedbackAttachment[]
}

// FeedbackAttachments renders the report's media. Images are clickable
// thumbnails that open a full-size lightbox (shared Dialog); videos use
// the native <video> player with controls. Both read the presigned GET
// `url` the backend stamps on each attachment.
export function FeedbackAttachments({ attachments }: FeedbackAttachmentsProps) {
  const [lightbox, setLightbox] = useState<FeedbackAttachment | null>(null)

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Paperclip className="h-4 w-4 text-muted-foreground" aria-hidden="true" />
          Pièces jointes ({attachments.length})
        </CardTitle>
      </CardHeader>
      <CardContent>
        {attachments.length === 0 ? (
          <p className="text-sm text-muted-foreground">Aucune pièce jointe.</p>
        ) : (
          <ul className="grid grid-cols-2 gap-4 sm:grid-cols-3">
            {attachments.map((attachment) => (
              <li key={attachment.id}>
                <AttachmentTile
                  attachment={attachment}
                  onOpenImage={() => setLightbox(attachment)}
                />
              </li>
            ))}
          </ul>
        )}
      </CardContent>

      <Dialog
        open={lightbox !== null}
        onClose={() => setLightbox(null)}
        className="max-w-3xl"
      >
        {lightbox && (
          <img
            src={lightbox.url}
            alt="Pièce jointe en taille réelle"
            className="max-h-[75vh] w-full rounded-lg object-contain"
          />
        )}
      </Dialog>
    </Card>
  )
}

type AttachmentTileProps = {
  attachment: FeedbackAttachment
  onOpenImage: () => void
}

function AttachmentTile({ attachment, onOpenImage }: AttachmentTileProps) {
  const caption = formatBytes(attachment.size_bytes)

  if (attachment.kind === "video") {
    return (
      <figure className="space-y-1">
        <video
          src={attachment.url}
          controls
          className="h-32 w-full rounded-lg bg-black object-contain"
        >
          <track kind="captions" />
        </video>
        <figcaption className="truncate text-xs text-muted-foreground">
          Vidéo · {caption}
        </figcaption>
      </figure>
    )
  }

  return (
    <figure className="space-y-1">
      <button
        type="button"
        onClick={onOpenImage}
        className="block w-full overflow-hidden rounded-lg border border-border transition-all duration-200 ease-out hover:border-primary/40 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/50"
        aria-label="Agrandir la pièce jointe"
      >
        <img
          src={attachment.url}
          alt="Aperçu de la pièce jointe"
          loading="lazy"
          className="h-32 w-full object-cover"
        />
      </button>
      <figcaption className="truncate text-xs text-muted-foreground">
        Image · {caption}
      </figcaption>
    </figure>
  )
}
