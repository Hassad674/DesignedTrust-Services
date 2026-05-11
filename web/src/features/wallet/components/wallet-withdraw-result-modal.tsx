"use client"

import { AlertTriangle, CheckCircle2 } from "lucide-react"
import { useTranslations } from "next-intl"

import { Modal } from "@/shared/components/ui/modal"
import { Button } from "@/shared/components/ui/button"
import { cn } from "@/shared/lib/utils"

import type { WithdrawLegError } from "../api/wallet-api"

/**
 * WalletWithdrawResultModal — surfaces a 207 Multi-Status response
 * to the user. The host opens it whenever `errors[]` is non-empty
 * AND `drained_cents > 0` so the user understands that part of the
 * money moved while another leg failed.
 *
 * The body has three sub-sections:
 *   1. Success line — total drained + per-leg breakdown.
 *   2. Per-leg error list — one entry per failed source.
 *   3. Close button (ghost outline).
 *
 * Strictly presentational — opens via `open` prop, closes via
 * `onClose`. No state of its own.
 */
export type WalletWithdrawResultModalProps = {
  open: boolean
  onClose: () => void
  drainedCents: number
  missionsCents: number
  commissionsCents: number
  errors: WithdrawLegError[]
}

export function WalletWithdrawResultModal({
  open,
  onClose,
  drainedCents,
  missionsCents,
  commissionsCents,
  errors,
}: WalletWithdrawResultModalProps) {
  const t = useTranslations("walletUnified.result")
  const tErr = useTranslations("walletUnified.result")

  return (
    <Modal open={open} onClose={onClose} title={t("title")}>
      <div className="space-y-4">
        <SuccessSummary
          drainedLabel={t("drained", {
            amount: formatEurCents(drainedCents),
          })}
          missionsLine={t("missionsLine", {
            amount: formatEurCents(missionsCents),
          })}
          commissionsLine={t("commissionsLine", {
            amount: formatEurCents(commissionsCents),
          })}
        />
        {errors.length > 0 && (
          <ErrorsList
            heading={t("errorsHeading")}
            labels={{
              missions: tErr("errorMissions"),
              commissions: tErr("errorCommissions"),
            }}
            errors={errors}
          />
        )}
        <div className="flex justify-end pt-1">
          <Button
            variant="outline"
            size="md"
            type="button"
            onClick={onClose}
            data-testid="withdraw-result-close"
          >
            {t("close")}
          </Button>
        </div>
      </div>
    </Modal>
  )
}

function SuccessSummary({
  drainedLabel,
  missionsLine,
  commissionsLine,
}: {
  drainedLabel: string
  missionsLine: string
  commissionsLine: string
}) {
  return (
    <div
      className={cn(
        "flex items-start gap-3 rounded-xl bg-success-soft p-3 text-[13.5px]",
      )}
    >
      <CheckCircle2
        className="mt-0.5 h-4 w-4 shrink-0 text-success"
        aria-hidden="true"
      />
      <div className="space-y-1">
        <p className="font-semibold text-foreground">{drainedLabel}</p>
        <p className="text-muted-foreground">{missionsLine}</p>
        <p className="text-muted-foreground">{commissionsLine}</p>
      </div>
    </div>
  )
}

function ErrorsList({
  heading,
  labels,
  errors,
}: {
  heading: string
  labels: { missions: string; commissions: string }
  errors: WithdrawLegError[]
}) {
  return (
    <div className="space-y-2" data-testid="withdraw-errors-list">
      <p className="text-[12px] font-semibold uppercase tracking-wide text-muted-foreground">
        {heading}
      </p>
      <ul className="space-y-2">
        {errors.map((err, idx) => (
          <li
            key={`${err.source}-${idx}`}
            className={cn(
              "flex items-start gap-2 rounded-xl border border-destructive/30 bg-destructive/5 p-3",
              "text-[13px] text-foreground",
            )}
          >
            <AlertTriangle
              className="mt-0.5 h-4 w-4 shrink-0 text-destructive"
              aria-hidden="true"
            />
            <div>
              <p className="font-semibold">
                {err.source === "missions" ? labels.missions : labels.commissions}
              </p>
              <p className="mt-0.5 text-muted-foreground">{err.message}</p>
            </div>
          </li>
        ))}
      </ul>
    </div>
  )
}

function formatEurCents(cents: number): string {
  return new Intl.NumberFormat("fr-FR", {
    style: "currency",
    currency: "EUR",
    maximumFractionDigits: 0,
  }).format(cents / 100)
}
