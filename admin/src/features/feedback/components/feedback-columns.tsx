import type { ColumnDef } from "@tanstack/react-table"
import { Paperclip, MessageSquare } from "lucide-react"
import { Badge } from "@/shared/components/ui/badge"
import { formatRelativeDate } from "@/shared/lib/utils"
import type { FeedbackReportRow } from "../types"
import {
  severityLabel,
  severityVariant,
  statusLabel,
  statusVariant,
  typeLabel,
  typeVariant,
} from "./feedback-labels"

export const feedbackColumns: ColumnDef<FeedbackReportRow, unknown>[] = [
  {
    id: "type",
    header: "Type",
    cell: ({ row }) => (
      <Badge variant={typeVariant(row.original.type)}>
        {typeLabel(row.original.type)}
      </Badge>
    ),
  },
  {
    id: "title",
    header: "Titre",
    cell: ({ row }) => {
      const { title } = row.original
      const truncated = title.length > 60 ? `${title.slice(0, 60)}…` : title
      return <span className="font-medium text-foreground">{truncated}</span>
    },
  },
  {
    id: "status",
    header: "Statut",
    cell: ({ row }) => (
      <Badge variant={statusVariant(row.original.status)}>
        {statusLabel(row.original.status)}
      </Badge>
    ),
  },
  {
    id: "severity",
    header: "Gravité",
    cell: ({ row }) => (
      <Badge variant={severityVariant(row.original.severity)}>
        {severityLabel(row.original.severity)}
      </Badge>
    ),
  },
  {
    id: "reporter",
    header: "Rapporteur",
    cell: ({ row }) => {
      const email = row.original.reporter_email
      if (!email) {
        return <span className="text-sm text-muted-foreground">Anonyme</span>
      }
      return <span className="text-sm text-foreground">{email}</span>
    },
  },
  {
    id: "counts",
    header: "Pièces / Notes",
    cell: ({ row }) => (
      <div className="flex items-center gap-3 text-xs text-muted-foreground">
        <span className="inline-flex items-center gap-1">
          <Paperclip className="h-3.5 w-3.5" aria-hidden="true" />
          {row.original.attachment_count}
        </span>
        <span className="inline-flex items-center gap-1">
          <MessageSquare className="h-3.5 w-3.5" aria-hidden="true" />
          {row.original.note_count}
        </span>
      </div>
    ),
  },
  {
    accessorKey: "created_at",
    header: "Reçu",
    cell: ({ row }) => (
      <span className="text-sm text-muted-foreground">
        {formatRelativeDate(row.original.created_at)}
      </span>
    ),
  },
]
