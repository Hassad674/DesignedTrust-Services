import { useParams, useNavigate } from "react-router-dom"
import { ArrowLeft } from "lucide-react"
import { PageHeader } from "@/shared/components/layouts/page-header"
import { Button } from "@/shared/components/ui/button"
import { Skeleton } from "@/shared/components/ui/skeleton"
import { useFeedbackDetail } from "../hooks/use-feedback"
import { FeedbackInfoCard } from "./feedback-info-card"
import { FeedbackTriageCard } from "./feedback-triage-card"
import { FeedbackAttachments } from "./feedback-attachments"
import { FeedbackNotes } from "./feedback-notes"

// FeedbackDetailPage is the triage surface for a single report. It
// composes the read-only info + attachments, the triage controls
// (status/severity), and the internal note thread. All mutations live
// in their own child components; this page only loads the detail and
// lays out the cards.
export function FeedbackDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data: report, isLoading, isError } = useFeedbackDetail(id ?? "")

  if (isLoading) return <FeedbackDetailSkeleton />

  if (isError || !report) {
    return (
      <div className="rounded-xl border border-destructive/20 bg-destructive/5 p-6 text-center text-sm text-destructive">
        Erreur lors du chargement du signalement
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Détail du signalement"
        actions={
          <Button variant="ghost" size="sm" onClick={() => navigate(-1)}>
            <ArrowLeft className="h-4 w-4" aria-hidden="true" />
            Retour
          </Button>
        }
      />

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="space-y-6 lg:col-span-2">
          <FeedbackInfoCard report={report} />
          <FeedbackAttachments attachments={report.attachments} />
          <FeedbackNotes reportId={report.id} notes={report.notes} />
        </div>
        <div className="space-y-6">
          <FeedbackTriageCard
            reportId={report.id}
            status={report.status}
            severity={report.severity}
          />
        </div>
      </div>
    </div>
  )
}

function FeedbackDetailSkeleton() {
  return (
    <div className="space-y-6">
      <Skeleton className="h-8 w-48" />
      <div className="grid gap-6 lg:grid-cols-3">
        <div className="space-y-6 lg:col-span-2">
          <Skeleton className="h-48" />
          <Skeleton className="h-40" />
        </div>
        <Skeleton className="h-56" />
      </div>
    </div>
  )
}
