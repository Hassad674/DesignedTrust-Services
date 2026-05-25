import type { FeedbackSeverity, FeedbackStatus, FeedbackType } from "../types"

// Shared label + badge-variant lookup tables for the feedback feature.
// Extracted here (rather than duplicated in columns, filters, and the
// detail view) because the same mappings are consumed by all three —
// the rule-of-three threshold for extraction.

type BadgeVariant = "default" | "success" | "warning" | "destructive" | "outline"

export const TYPE_LABELS: Record<FeedbackType, string> = {
  bug: "Bug",
  vulnerability: "Sécurité",
}

export const TYPE_VARIANTS: Record<FeedbackType, BadgeVariant> = {
  bug: "warning",
  vulnerability: "destructive",
}

export const STATUS_LABELS: Record<FeedbackStatus, string> = {
  new: "Nouveau",
  triaged: "Trié",
  in_progress: "En cours",
  resolved: "Résolu",
  rejected: "Rejeté",
}

export const STATUS_VARIANTS: Record<FeedbackStatus, BadgeVariant> = {
  new: "default",
  triaged: "outline",
  in_progress: "warning",
  resolved: "success",
  rejected: "destructive",
}

export const SEVERITY_LABELS: Record<FeedbackSeverity, string> = {
  low: "Faible",
  medium: "Moyenne",
  high: "Élevée",
  critical: "Critique",
}

export const SEVERITY_VARIANTS: Record<FeedbackSeverity, BadgeVariant> = {
  low: "outline",
  medium: "default",
  high: "warning",
  critical: "destructive",
}

// The select option lists reuse the label tables so the dropdowns and
// the badges never drift. A leading empty option is prepended by the
// callers that need a "tous" sentinel.
export const TYPE_OPTIONS = (Object.keys(TYPE_LABELS) as FeedbackType[]).map(
  (value) => ({ value, label: TYPE_LABELS[value] }),
)

export const STATUS_OPTIONS = (
  Object.keys(STATUS_LABELS) as FeedbackStatus[]
).map((value) => ({ value, label: STATUS_LABELS[value] }))

export const SEVERITY_OPTIONS = (
  Object.keys(SEVERITY_LABELS) as FeedbackSeverity[]
).map((value) => ({ value, label: SEVERITY_LABELS[value] }))

// fallbackVariant returns a safe badge variant for an unknown enum value
// the backend might add in the future, so the UI degrades gracefully.
export function typeVariant(type: string): BadgeVariant {
  return TYPE_VARIANTS[type as FeedbackType] ?? "outline"
}

export function typeLabel(type: string): string {
  return TYPE_LABELS[type as FeedbackType] ?? type
}

export function statusVariant(status: string): BadgeVariant {
  return STATUS_VARIANTS[status as FeedbackStatus] ?? "outline"
}

export function statusLabel(status: string): string {
  return STATUS_LABELS[status as FeedbackStatus] ?? status
}

export function severityVariant(severity: string): BadgeVariant {
  return SEVERITY_VARIANTS[severity as FeedbackSeverity] ?? "outline"
}

export function severityLabel(severity: string): string {
  return SEVERITY_LABELS[severity as FeedbackSeverity] ?? severity
}
