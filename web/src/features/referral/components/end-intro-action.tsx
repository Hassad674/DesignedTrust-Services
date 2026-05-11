"use client"

import { CheckCircle2 } from "lucide-react"
import { useTranslations } from "next-intl"
import { useState } from "react"
import { toast } from "sonner"

import { ApiError } from "@/shared/lib/api-client"
import { Button } from "@/shared/components/ui/button"
import { cn } from "@/shared/lib/utils"

import { useEndIntroAttribution } from "../hooks/use-referrals"
import { EndIntroConfirmationModal } from "./end-intro-confirmation-modal"

/**
 * EndIntroAction — the button → modal → badge state machine for
 * the apporteur's "Terminer l'intro" action on a single attribution.
 *
 * - When `initialEndedAt` is set, render the badge directly (the
 *   page may have been reloaded after a prior end).
 * - When the badge is null, render the destructive ghost button.
 * - Click → open EndIntroConfirmationModal.
 * - Confirm → fire useEndIntroAttribution(attributionId). On
 *   success, swap to the badge with the returned `ended_at`. The
 *   query invalidations already happen inside the hook
 *   (referralKeys.all + ["wallet"]).
 * - Cancel → close the modal without firing the mutation.
 * - Error → toast with the appropriate i18n message + keep the
 *   button so the user can retry.
 *
 * Strictly per-attribution (Run C scope) — does NOT terminate the
 * whole referral. The existing "Terminer l'intro" in ReferralActions
 * (PER-REFERRAL) is left untouched.
 */
export type EndIntroActionProps = {
  attributionId: string
  providerName?: string
  clientName?: string
  /** When set, the badge is rendered immediately (idempotent reload path). */
  initialEndedAt?: string
}

export function EndIntroAction({
  attributionId,
  providerName,
  clientName,
  initialEndedAt,
}: EndIntroActionProps) {
  const t = useTranslations("referralEndIntro")
  const tErr = useTranslations("referralEndIntro.error")

  const [open, setOpen] = useState(false)
  const [localEndedAt, setLocalEndedAt] = useState<string | undefined>(
    initialEndedAt,
  )
  const mutation = useEndIntroAttribution()

  function handleConfirm() {
    mutation.mutate(attributionId, {
      onSuccess: (result) => {
        setOpen(false)
        if (result.ended_at) {
          setLocalEndedAt(result.ended_at)
        }
      },
      onError: (err) => {
        setOpen(false)
        if (err instanceof ApiError) {
          if (err.status === 403) {
            toast.error(tErr("forbidden"))
            return
          }
          if (err.status === 404) {
            toast.error(tErr("notFound"))
            return
          }
        }
        toast.error(tErr("generic"))
      },
    })
  }

  if (localEndedAt) {
    return (
      <span
        data-testid="end-intro-badge"
        className={cn(
          "inline-flex items-center gap-1.5 rounded-full px-3 py-1",
          "bg-success-soft text-success",
          "text-[12px] font-semibold",
        )}
      >
        <CheckCircle2 className="h-3.5 w-3.5" aria-hidden="true" />
        {t("badge", { date: formatFrenchDate(localEndedAt) })}
      </span>
    )
  }

  return (
    <>
      <Button
        variant="outline"
        size="sm"
        type="button"
        onClick={() => setOpen(true)}
        disabled={mutation.isPending}
        data-testid="end-intro-trigger"
        className={cn(
          "rounded-full border-destructive/40 text-destructive",
          "hover:bg-destructive/10 hover:text-destructive",
        )}
      >
        {t("ctaLabel")}
      </Button>
      <EndIntroConfirmationModal
        open={open}
        onClose={() => setOpen(false)}
        onConfirm={handleConfirm}
        providerName={providerName}
        clientName={clientName}
        pending={mutation.isPending}
      />
    </>
  )
}

function formatFrenchDate(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return new Intl.DateTimeFormat("fr-FR", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  }).format(d)
}
