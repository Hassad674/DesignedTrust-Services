import { apiClient } from "@/shared/lib/api-client"

import type { Get, Post } from "@/shared/lib/api-paths"
export type WalletRecord = {
  id: string
  proposal_id: string
  milestone_id?: string
  proposal_amount: number
  platform_fee: number
  provider_payout: number
  payment_status: string
  transfer_status: string
  mission_status: string
  created_at: string
}

export type CommissionWallet = {
  pending_cents: number
  pending_kyc_cents: number
  paid_cents: number
  clawed_back_cents: number
  /**
   * Rolling 30-day sum of commissions paid out. Drives the
   * "Versées 30j" tile on the apporteur wallet (WALLET-UX). Falls
   * back to 0 when the backend hasn't been redeployed yet — every
   * consumer must treat this field as optional-at-runtime.
   */
  paid_30d_cents?: number
  /**
   * Cumulative paid-out lifetime total (paid + clawed_back). Drives
   * the "Cumul lifetime" tile. Also optional-at-runtime so the UI
   * degrades to the locally-derived sum on older deployments.
   */
  lifetime_cents?: number
  currency: string
}

export type WalletCommissionRecord = {
  id: string
  referral_id?: string
  proposal_id?: string
  milestone_id?: string
  gross_amount_cents: number
  commission_cents: number
  currency: string
  status: string
  stripe_transfer_id?: string
  paid_at?: string
  clawed_back_at?: string
  created_at: string
}

export type WalletOverview = {
  stripe_account_id: string
  charges_enabled: boolean
  payouts_enabled: boolean
  escrow_amount: number
  available_amount: number
  transferred_amount: number
  records: WalletRecord[] | null
  commissions: CommissionWallet
  commission_records: WalletCommissionRecord[] | null
}

export type PayoutResult = {
  status: string
  message: string
}

export function getWallet(): Promise<WalletOverview> {
  return apiClient<Get<"/api/v1/wallet"> & WalletOverview>("/api/v1/wallet")
}

export function requestPayout(): Promise<PayoutResult> {
  return apiClient<Post<"/api/v1/wallet/payout"> & PayoutResult>("/api/v1/wallet/payout", { method: "POST" })
}

/**
 * Re-issues the Stripe transfer for a single record stuck in
 * transfer_status="failed". Takes the payment record id — NOT the
 * proposal id — because a proposal can own N records (one per
 * milestone) and only the record id is unambiguous. The backend
 * enforces the same guards as the global payout (mission completed,
 * Stripe account present) and returns 409 when the row is no longer
 * retriable (e.g. someone else retried it).
 */
export function retryFailedTransfer(recordId: string): Promise<PayoutResult> {
  return apiClient<Post<"/api/v1/wallet/transfers/{record_id}/retry"> & PayoutResult>(
    `/api/v1/wallet/transfers/${recordId}/retry`,
    { method: "POST" },
  )
}
