"use client"

import { Pencil, ShieldCheck } from "lucide-react"
import { useTranslations } from "next-intl"

import { Link } from "@i18n/navigation"

/**
 * WalletQuickLinks — two permanent, discreet shortcuts under the
 * "Retirer" CTA letting the provider proactively edit (and re-edit)
 * their billing + Stripe payment info, INDEPENDENTLY of the withdraw
 * flow's gating modals.
 *
 * Restored in fix/wallet-kyc-billing-regression: the WALLET-UNIFY
 * Run C refonte dropped the two-link footer originally added in
 * `feat(wallet): quick links to billing-profile + payment-info`
 * (commit 02793494). The links are always visible — not only when a
 * withdrawal is blocked — so the user can fix KYC/billing before ever
 * clicking Retirer.
 *
 * Soleil v2: semantic tokens only (text-muted-foreground →
 * text-foreground on hover), lucide icons size-3.5, separated by a
 * subtle "·". No hex, no inline style.
 */
export function WalletQuickLinks() {
  const t = useTranslations("walletUnified.quickLinks")

  return (
    <div className="mt-5 flex flex-wrap items-center gap-x-4 gap-y-2 border-t border-border pt-4 text-xs">
      <Link
        href="/settings/billing-profile?return_to=/wallet"
        className="inline-flex items-center gap-1.5 text-muted-foreground transition-colors hover:text-foreground"
      >
        <Pencil className="size-3.5" aria-hidden="true" />
        {t("editBilling")}
      </Link>
      <span aria-hidden="true" className="text-subtle-foreground">
        ·
      </span>
      <Link
        href="/payment-info"
        className="inline-flex items-center gap-1.5 text-muted-foreground transition-colors hover:text-foreground"
      >
        <ShieldCheck className="size-3.5" aria-hidden="true" />
        {t("stripePaymentInfo")}
      </Link>
    </div>
  )
}
