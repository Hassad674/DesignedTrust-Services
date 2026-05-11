"use client"

import { useTranslations } from "next-intl"

import { Modal } from "@/shared/components/ui/modal"
import { Button } from "@/shared/components/ui/button"
import { cn } from "@/shared/lib/utils"

/**
 * EndIntroConfirmationModal — destructive confirmation dialog
 * fired by the apporteur's "Terminer l'intro" action on a referral
 * attribution row. Strict copy from the Run C brief:
 *
 *   "Tu vas mettre fin à ta mise en relation entre {provider} et
 *    {client}. Tu ne toucheras plus de commission sur les jalons
 *    approuvés à partir d'aujourd'hui. Les commissions déjà
 *    acquises restent dues. Confirmer ?"
 *
 * Two buttons: ghost "Annuler" + destructive "Terminer
 * définitivement". The mutation lives in the parent so the modal
 * stays presentational.
 */
export type EndIntroConfirmationModalProps = {
  open: boolean
  onClose: () => void
  onConfirm: () => void
  /** Provider display name — falls back to a generic noun. */
  providerName?: string
  /** Client display name — falls back to a generic noun. */
  clientName?: string
  /** Mirror of the underlying mutation's isPending state. */
  pending?: boolean
}

export function EndIntroConfirmationModal({
  open,
  onClose,
  onConfirm,
  providerName,
  clientName,
  pending = false,
}: EndIntroConfirmationModalProps) {
  const t = useTranslations("referralEndIntro.modal")

  const provider = providerName ?? t("fallbackProvider")
  const client = clientName ?? t("fallbackClient")

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={t("title")}
      maxWidthClassName="max-w-md"
    >
      <div className="space-y-4">
        <p className="text-[13.5px] leading-relaxed text-foreground">
          {t("body", { provider, client })}
        </p>
        <div className="flex flex-col gap-2 pt-1 sm:flex-row sm:justify-end">
          <Button
            variant="outline"
            size="md"
            type="button"
            onClick={onClose}
            disabled={pending}
            data-testid="end-intro-cancel"
          >
            {t("cancel")}
          </Button>
          <Button
            variant="destructive"
            size="md"
            type="button"
            onClick={onConfirm}
            disabled={pending}
            data-testid="end-intro-confirm"
            className={cn("min-w-[180px]")}
          >
            {t("confirm")}
          </Button>
        </div>
      </div>
    </Modal>
  )
}
