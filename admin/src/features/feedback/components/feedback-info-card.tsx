import { Link } from "react-router-dom"
import { ExternalLink } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/components/ui/card"
import { Badge } from "@/shared/components/ui/badge"
import { formatDate } from "@/shared/lib/utils"
import type { FeedbackReportDetail } from "../types"
import { typeLabel, typeVariant } from "./feedback-labels"
import { formatContext } from "./feedback-format"

type FeedbackInfoCardProps = {
  report: FeedbackReportDetail
}

// FeedbackInfoCard shows the report body: type badge, description,
// reporter (email or "Anonyme" + a link to the user when present), the
// originating page URL, and the flattened context key/values.
export function FeedbackInfoCard({ report }: FeedbackInfoCardProps) {
  const context = formatContext(report.context)

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Badge variant={typeVariant(report.type)}>
            {typeLabel(report.type)}
          </Badge>
          <span>{report.title}</span>
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div>
          <p className="mb-1 text-sm text-muted-foreground">Description</p>
          <p className="whitespace-pre-wrap text-sm text-foreground">
            {report.description}
          </p>
        </div>

        <dl className="grid gap-3 sm:grid-cols-2">
          <InfoRow label="Rapporteur">
            <ReporterValue
              email={report.reporter_email}
              userId={report.reporter_user_id}
            />
          </InfoRow>
          <InfoRow label="Reçu le">
            <span className="text-foreground">
              {formatDate(report.created_at)}
            </span>
          </InfoRow>
          {report.page_url && (
            <InfoRow label="Page concernée" full>
              <span className="break-all text-foreground">{report.page_url}</span>
            </InfoRow>
          )}
        </dl>

        {context.length > 0 && (
          <div>
            <p className="mb-2 text-sm text-muted-foreground">Contexte</p>
            <dl className="grid gap-2 rounded-lg border border-border bg-muted/30 p-3 sm:grid-cols-2">
              {context.map((entry) => (
                <div key={entry.key} className="flex flex-col">
                  <dt className="text-xs text-muted-foreground">{entry.label}</dt>
                  <dd className="break-all text-sm text-foreground">
                    {entry.value}
                  </dd>
                </div>
              ))}
            </dl>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

function InfoRow({
  label,
  full,
  children,
}: {
  label: string
  full?: boolean
  children: React.ReactNode
}) {
  return (
    <div className={full ? "sm:col-span-2" : undefined}>
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd className="text-sm">{children}</dd>
    </div>
  )
}

function ReporterValue({
  email,
  userId,
}: {
  email: string
  userId?: string | null
}) {
  if (!email) {
    return <span className="text-muted-foreground">Anonyme</span>
  }
  if (userId) {
    return (
      <Link
        to={`/users/${userId}`}
        className="inline-flex items-center gap-1 text-primary hover:underline"
      >
        {email}
        <ExternalLink className="h-3 w-3" aria-hidden="true" />
      </Link>
    )
  }
  return <span className="text-foreground">{email}</span>
}
