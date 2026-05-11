"use client"

import { useTranslations } from "next-intl"

import { cn } from "@/shared/lib/utils"

import type { ReferralCommission } from "../types"

/**
 * ProjectedCommissionsList — replaces the misleading "0 €" sub-line
 * shown on the per-mission row when the apporteur has not yet been
 * paid. Renders one entry per commission row scoped to the
 * attribution + an optional "escrowed" preview line driven by the
 * attribution.escrow_commission_cents aggregate (funds held against
 * milestones not yet released).
 *
 * Status → tone matrix (Run C brief):
 *   - paid              → "+X € reçue"     — green
 *   - pending|pending_kyc → "X € en attente" — orange
 *   - failed            → "X € échouée"   — red
 *   - cancelled|clawed_back → skip (per brief)
 *   - synthetic escrow  → "≈ X € (en séquestre)" — muted italic
 */
export type ProjectedCommissionsListProps = {
  /** Commission rows scoped to this attribution. */
  commissions: ReferralCommission[]
  /** Total funds held in escrow against this attribution, cents. */
  escrowCents: number
}

export function ProjectedCommissionsList({
  commissions,
  escrowCents,
}: ProjectedCommissionsListProps) {
  const t = useTranslations("referralProjection")
  const tStatus = useTranslations("referralProjection.status")

  const renderable = commissions.filter(
    (c) => c.status !== "cancelled" && c.status !== "clawed_back",
  )

  if (renderable.length === 0 && escrowCents === 0) {
    return (
      <p className="text-[11px] italic text-muted-foreground">
        {t("empty")}
      </p>
    )
  }

  return (
    <div className="space-y-1.5" data-testid="projected-commissions-list">
      <p className="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">
        {t("perMilestoneTitle")}
      </p>
      <ul className="space-y-1">
        {escrowCents > 0 && (
          <li
            data-testid="projected-commission-escrow-line"
            data-tone="escrowed"
            className="flex items-center justify-between gap-2 text-[12px]"
          >
            <span className="italic text-muted-foreground">
              {tStatus("escrowed", { amount: formatEurCents(escrowCents) })}
            </span>
          </li>
        )}
        {renderable.map((c) => (
          <li
            key={c.id}
            data-tone={statusTone(c.status)}
            data-testid={`projected-commission-row-${c.id}`}
            className="flex items-center justify-between gap-2 text-[12px]"
          >
            <span
              className={cn(
                "rounded-full px-2 py-0.5 text-[10px] font-semibold",
                pillClass(statusTone(c.status)),
              )}
            >
              {labelFor(c.status, c.commission_cents, tStatus)}
            </span>
          </li>
        ))}
      </ul>
    </div>
  )
}

type Tone = "paid" | "pending" | "failed" | "escrowed"

function statusTone(status: ReferralCommission["status"]): Tone {
  switch (status) {
    case "paid":
      return "paid"
    case "pending":
    case "pending_kyc":
      return "pending"
    case "failed":
      return "failed"
    default:
      return "escrowed"
  }
}

function pillClass(tone: Tone): string {
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

function labelFor(
  status: ReferralCommission["status"],
  cents: number,
  tStatus: (key: string, params?: Record<string, string | number>) => string,
): string {
  const amount = formatEurCents(cents)
  switch (status) {
    case "paid":
      return tStatus("paid", { amount })
    case "pending":
    case "pending_kyc":
      return tStatus("pending", { amount })
    case "failed":
      return tStatus("failed", { amount })
    default:
      return tStatus("escrowed", { amount })
  }
}

function formatEurCents(cents: number): string {
  return new Intl.NumberFormat("fr-FR", {
    style: "currency",
    currency: "EUR",
    maximumFractionDigits: 0,
  }).format(cents / 100)
}
