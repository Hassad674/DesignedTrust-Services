import { Loader2 } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/components/ui/card"
import { Select } from "@/shared/components/ui/select"
import { useUpdateFeedback } from "../hooks/use-feedback"
import type { FeedbackSeverity, FeedbackStatus } from "../types"
import { SEVERITY_OPTIONS, STATUS_OPTIONS } from "./feedback-labels"

type FeedbackTriageCardProps = {
  reportId: string
  status: string
  severity: string
}

// FeedbackTriageCard exposes the two triage controls — status and
// severity. Each select PATCHes the single field it owns on change; the
// hook invalidates the detail + list queries so the badges everywhere
// reflect the new value. Mutation state is announced inline (aria-live).
export function FeedbackTriageCard({
  reportId,
  status,
  severity,
}: FeedbackTriageCardProps) {
  const mutation = useUpdateFeedback(reportId)

  function handleStatusChange(value: string) {
    mutation.mutate({ status: value as FeedbackStatus })
  }

  function handleSeverityChange(value: string) {
    mutation.mutate({ severity: value as FeedbackSeverity })
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Triage</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <Select
          label="Statut"
          options={STATUS_OPTIONS}
          value={status}
          onChange={(e) => handleStatusChange(e.target.value)}
          disabled={mutation.isPending}
        />
        <Select
          label="Gravité"
          options={SEVERITY_OPTIONS}
          value={severity}
          onChange={(e) => handleSeverityChange(e.target.value)}
          disabled={mutation.isPending}
        />
        <p className="text-sm" role="status" aria-live="polite">
          {mutation.isPending && (
            <span className="inline-flex items-center gap-1 text-muted-foreground">
              <Loader2 className="h-3.5 w-3.5 animate-spin" aria-hidden="true" />
              Enregistrement…
            </span>
          )}
          {mutation.isSuccess && !mutation.isPending && (
            <span className="text-success">Modifications enregistrées.</span>
          )}
          {mutation.isError && (
            <span className="text-destructive">
              Erreur lors de l’enregistrement.
            </span>
          )}
        </p>
      </CardContent>
    </Card>
  )
}
