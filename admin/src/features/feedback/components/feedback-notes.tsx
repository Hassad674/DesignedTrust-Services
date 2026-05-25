import { useState, type FormEvent } from "react"
import { Loader2, MessageSquare } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/components/ui/card"
import { Button } from "@/shared/components/ui/button"
import { Textarea } from "@/shared/components/ui/textarea"
import { formatRelativeDate } from "@/shared/lib/utils"
import { useAddFeedbackNote } from "../hooks/use-feedback"
import type { FeedbackNote } from "../types"

type FeedbackNotesProps = {
  reportId: string
  notes: FeedbackNote[]
}

// FeedbackNotes renders the internal admin note thread (newest first)
// plus the add-note form. On submit it POSTs the note; the hook
// invalidates the report detail so the thread refreshes with the
// persisted note. Success / error are announced inline (the admin app
// has no toast layer — it surfaces mutation state via aria-live text,
// mirroring the disputes detail page).
export function FeedbackNotes({ reportId, notes }: FeedbackNotesProps) {
  const [body, setBody] = useState("")
  const mutation = useAddFeedbackNote(reportId)

  function handleSubmit(e: FormEvent) {
    e.preventDefault()
    const trimmed = body.trim()
    if (!trimmed || mutation.isPending) return
    mutation.mutate(trimmed, {
      onSuccess: () => setBody(""),
    })
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <MessageSquare className="h-4 w-4 text-muted-foreground" aria-hidden="true" />
          Notes internes ({notes.length})
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {notes.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            Aucune note pour le moment.
          </p>
        ) : (
          <ul className="space-y-3">
            {notes.map((note) => (
              <NoteItem key={note.id} note={note} />
            ))}
          </ul>
        )}

        <form onSubmit={handleSubmit} className="space-y-2">
          <Textarea
            label="Ajouter une note"
            value={body}
            onChange={(e) => setBody(e.target.value)}
            rows={3}
            placeholder="Note de triage interne…"
          />
          <div className="flex items-center gap-3">
            <Button
              type="submit"
              variant="primary"
              size="sm"
              disabled={mutation.isPending || body.trim() === ""}
            >
              {mutation.isPending && (
                <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" />
              )}
              Ajouter
            </Button>
            <span className="text-sm" role="status" aria-live="polite">
              {mutation.isSuccess && (
                <span className="text-success">Note ajoutée.</span>
              )}
              {mutation.isError && (
                <span className="text-destructive">
                  Erreur lors de l’ajout de la note.
                </span>
              )}
            </span>
          </div>
        </form>
      </CardContent>
    </Card>
  )
}

function NoteItem({ note }: { note: FeedbackNote }) {
  return (
    <li className="rounded-lg border border-border bg-muted/30 p-3">
      <p className="whitespace-pre-wrap text-sm text-foreground">{note.body}</p>
      <p className="mt-1 text-xs text-muted-foreground">
        {formatRelativeDate(note.created_at)}
      </p>
    </li>
  )
}
