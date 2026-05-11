import { cn } from "@/shared/lib/utils"

/**
 * WalletStatusBadge — single source of truth for mapping the
 * backend's free-form `status` string on a unified transaction row
 * to one of four Soleil v2 tones. Lives in its own file so the
 * mapping is testable in isolation and shared by the header /
 * history list without duplication.
 *
 * The four tones cover every status the backend currently emits on
 * the wallet/summary timeline:
 *
 *   - paid              → green (success)
 *   - pending           → amber (in flight)
 *   - escrowed / held   → muted (held funds, not yet released)
 *   - failed            → red (action required)
 *
 * Unknown statuses fall back to the muted tone — graceful
 * degradation for any new backend status that ships before the UI
 * adds it.
 */
export type WalletStatusTone = "paid" | "pending" | "escrowed" | "failed"

const STATUS_MAP: Record<string, WalletStatusTone> = {
  paid: "paid",
  transferred: "paid",
  transferred_pending_bank: "paid",
  completed: "paid",
  pending: "pending",
  pending_kyc: "pending",
  active: "pending",
  in_progress: "pending",
  completion_requested: "pending",
  accepted: "pending",
  escrowed: "escrowed",
  held: "escrowed",
  failed: "failed",
  cancelled: "failed",
  clawed_back: "failed",
}

/**
 * Maps any backend status string to one of the four tones. Exported
 * so tests can drive the matrix without instantiating the badge.
 */
export function resolveWalletStatusTone(status: string): WalletStatusTone {
  return STATUS_MAP[status.toLowerCase()] ?? "escrowed"
}

type WalletStatusBadgeProps = {
  status: string
  /** Translated label to display — keeps the badge i18n-agnostic. */
  label: string
}

export function WalletStatusBadge({ status, label }: WalletStatusBadgeProps) {
  const tone = resolveWalletStatusTone(status)
  return (
    <span
      className={cn(
        "inline-flex shrink-0 items-center rounded-full px-2 py-0.5",
        "text-[10px] font-semibold uppercase tracking-wide",
        toneClass(tone),
      )}
      data-tone={tone}
      data-status={status}
    >
      {label}
    </span>
  )
}

function toneClass(tone: WalletStatusTone): string {
  switch (tone) {
    case "paid":
      return "bg-success-soft text-success"
    case "pending":
      return "bg-amber-soft text-warning"
    case "failed":
      return "bg-destructive/10 text-destructive"
    case "escrowed":
    default:
      return "bg-muted text-muted-foreground"
  }
}
