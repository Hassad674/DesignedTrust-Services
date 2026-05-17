"use client"

import { ArrowRight, Loader2, Wallet } from "lucide-react"
import { useTranslations } from "next-intl"

import { Button } from "@/shared/components/ui/button"
import { cn } from "@/shared/lib/utils"

import { WalletQuickLinks } from "./wallet-quick-links"

/**
 * WalletUnifiedHeader — purely presentational header for the
 * refonte /wallet page (WALLET-UNIFY Run C). Owns the consolidated
 * "Portefeuille" hero card (title + total + single Retirer button)
 * AND the 3 stat cards row below (en séquestre / disponible /
 * transmis). The host (`wallet-unified-page`) owns ALL state — KYC
 * modal toggling, mutation pending flag, partial-success modal — so
 * this file stays renderable in tests without a QueryClient.
 *
 * Soleil v2:
 *   - Hero: ivoire surface, corail-soft icon square, Fraunces title,
 *     big Fraunces 56px total, single corail-deep "Retirer" pill.
 *   - 3 stat cards: ivoire, sable border, Fraunces label + Geist Mono
 *     value.
 *   - All copy goes through useTranslations("walletUnified").
 */

export type WalletUnifiedHeaderProps = {
  /** Total earned across both legs, cents. */
  totalCents: number
  escrowedCents: number
  availableCents: number
  transmittedCents: number
  payoutPending: boolean
  onWithdraw: () => void
}

export function WalletUnifiedHeader({
  totalCents,
  escrowedCents,
  availableCents,
  transmittedCents,
  payoutPending,
  onWithdraw,
}: WalletUnifiedHeaderProps) {
  const t = useTranslations("walletUnified")
  const canClick = availableCents > 0 && !payoutPending

  return (
    <div className="space-y-4">
      <HeroCard
        title={t("title")}
        subtitle={t("subtitle")}
        totalLabel={t("totalEarned")}
        totalCents={totalCents}
        ctaLabel={t("withdraw")}
        ctaPendingLabel={t("withdrawing")}
        ctaNoFundsLabel={t("noFunds")}
        canClick={canClick}
        availableCents={availableCents}
        payoutPending={payoutPending}
        onWithdraw={onWithdraw}
      />
      <StatCardsRow
        escrowedCents={escrowedCents}
        availableCents={availableCents}
        transmittedCents={transmittedCents}
        labels={{
          escrowed: t("card.escrowed"),
          escrowedHint: t("card.escrowedHint"),
          available: t("card.available"),
          availableHint: t("card.availableHint"),
          transmitted: t("card.transmitted"),
          transmittedHint: t("card.transmittedHint"),
        }}
      />
    </div>
  )
}

// ─── Hero ─────────────────────────────────────────────────────────────────

type HeroCardProps = {
  title: string
  subtitle: string
  totalLabel: string
  totalCents: number
  ctaLabel: string
  ctaPendingLabel: string
  ctaNoFundsLabel: string
  canClick: boolean
  availableCents: number
  payoutPending: boolean
  onWithdraw: () => void
}

function HeroCard(props: HeroCardProps) {
  return (
    <section
      data-testid="wallet-unified-hero"
      className={cn(
        "relative overflow-hidden rounded-2xl border border-border bg-card p-5 md:p-7",
      )}
      style={{ boxShadow: "var(--shadow-card)" }}
    >
      <div
        aria-hidden="true"
        className="pointer-events-none absolute -right-16 -top-16 h-56 w-56 rounded-full"
        style={{
          background:
            "radial-gradient(circle, rgba(232,93,74,0.07), transparent 65%)",
        }}
      />
      <div className="relative">
        <HeroIntro title={props.title} subtitle={props.subtitle} />
        <HeroAmountRow
          totalLabel={props.totalLabel}
          totalCents={props.totalCents}
          availableCents={props.availableCents}
          ctaLabel={props.ctaLabel}
          ctaPendingLabel={props.ctaPendingLabel}
          ctaNoFundsLabel={props.ctaNoFundsLabel}
          canClick={props.canClick}
          payoutPending={props.payoutPending}
          onWithdraw={props.onWithdraw}
        />
        {/* Permanent editable shortcuts — always visible, discreet,
            under the Retirer CTA. Lets the provider edit / re-edit
            billing + Stripe payment info proactively, independently of
            the withdraw flow's gating modals. */}
        <WalletQuickLinks />
      </div>
    </section>
  )
}

function HeroIntro({ title, subtitle }: { title: string; subtitle: string }) {
  return (
    <div className="mb-5 flex items-start gap-3.5">
      <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-primary-soft text-primary">
        <Wallet className="h-5 w-5" strokeWidth={1.6} aria-hidden="true" />
      </div>
      <div className="min-w-0">
        <h1 className="font-serif text-[26px] font-medium leading-tight tracking-[-0.02em] text-foreground">
          {title}
        </h1>
        <p className="mt-0.5 text-[13px] text-muted-foreground">{subtitle}</p>
      </div>
    </div>
  )
}

type HeroAmountRowProps = {
  totalLabel: string
  totalCents: number
  availableCents: number
  ctaLabel: string
  ctaPendingLabel: string
  ctaNoFundsLabel: string
  canClick: boolean
  payoutPending: boolean
  onWithdraw: () => void
}

function HeroAmountRow(props: HeroAmountRowProps) {
  return (
    <div className="grid items-end gap-6 md:grid-cols-[1fr_auto] md:items-center">
      <div>
        <p className="mb-1.5 font-mono text-[11px] font-bold uppercase tracking-[0.12em] text-subtle-foreground">
          {props.totalLabel}
        </p>
        <p
          data-testid="wallet-unified-total"
          className="font-serif text-[40px] font-normal leading-none tracking-[-0.035em] text-foreground md:text-[56px]"
        >
          {formatEurCents(props.totalCents)}
        </p>
      </div>
      <WithdrawPill
        ctaLabel={props.ctaLabel}
        ctaPendingLabel={props.ctaPendingLabel}
        ctaNoFundsLabel={props.ctaNoFundsLabel}
        canClick={props.canClick}
        availableCents={props.availableCents}
        payoutPending={props.payoutPending}
        onWithdraw={props.onWithdraw}
      />
    </div>
  )
}

type WithdrawPillProps = {
  ctaLabel: string
  ctaPendingLabel: string
  ctaNoFundsLabel: string
  canClick: boolean
  availableCents: number
  payoutPending: boolean
  onWithdraw: () => void
}

function WithdrawPill(props: WithdrawPillProps) {
  return (
    <div className="flex flex-col items-stretch gap-1.5 md:items-end">
      <Button
        variant="ghost"
        size="auto"
        type="button"
        onClick={props.onWithdraw}
        disabled={!props.canClick}
        data-testid="wallet-unified-withdraw"
        aria-label={props.ctaLabel}
        className={cn(
          "inline-flex items-center justify-center gap-2 rounded-full",
          "w-full px-5 py-2.5 text-sm font-bold text-primary-foreground md:w-auto",
          "bg-primary transition-all duration-200 ease-out",
          "hover:bg-primary-deep hover:shadow-[0_4px_14px_rgba(232,93,74,0.28)]",
          "active:scale-[0.98]",
          "disabled:cursor-not-allowed disabled:bg-pink disabled:opacity-60",
          "disabled:hover:bg-pink disabled:hover:shadow-none",
        )}
      >
        {props.payoutPending ? (
          <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" />
        ) : (
          <ArrowRight className="h-4 w-4" aria-hidden="true" />
        )}
        {props.payoutPending
          ? props.ctaPendingLabel
          : `${props.ctaLabel} ${formatEurCents(props.availableCents)}`}
      </Button>
      {props.availableCents === 0 && (
        <span className="text-[11px] text-subtle-foreground md:text-right">
          {props.ctaNoFundsLabel}
        </span>
      )}
    </div>
  )
}

// ─── 3 stat cards ──────────────────────────────────────────────────────────

type StatLabels = {
  escrowed: string
  escrowedHint: string
  available: string
  availableHint: string
  transmitted: string
  transmittedHint: string
}

function StatCardsRow({
  escrowedCents,
  availableCents,
  transmittedCents,
  labels,
}: {
  escrowedCents: number
  availableCents: number
  transmittedCents: number
  labels: StatLabels
}) {
  return (
    <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
      <StatCard
        label={labels.escrowed}
        hint={labels.escrowedHint}
        cents={escrowedCents}
        toneClass="bg-muted text-muted-foreground"
        testId="wallet-stat-escrowed"
      />
      <StatCard
        label={labels.available}
        hint={labels.availableHint}
        cents={availableCents}
        toneClass="bg-success-soft text-success"
        testId="wallet-stat-available"
      />
      <StatCard
        label={labels.transmitted}
        hint={labels.transmittedHint}
        cents={transmittedCents}
        toneClass="bg-primary-soft text-primary"
        testId="wallet-stat-transmitted"
      />
    </div>
  )
}

function StatCard({
  label,
  hint,
  cents,
  toneClass,
  testId,
}: {
  label: string
  hint: string
  cents: number
  toneClass: string
  testId: string
}) {
  return (
    <div
      data-testid={testId}
      className={cn(
        "flex flex-col gap-2 rounded-2xl border border-border bg-card p-4",
      )}
      style={{ boxShadow: "var(--shadow-card)" }}
    >
      <span
        className={cn(
          "inline-flex w-fit items-center rounded-full px-2.5 py-1",
          "text-[11px] font-semibold uppercase tracking-wide",
          toneClass,
        )}
      >
        {label}
      </span>
      <p className="font-serif text-[28px] font-medium leading-none tracking-[-0.02em] text-foreground">
        {formatEurCents(cents)}
      </p>
      <p className="text-[12px] text-muted-foreground">{hint}</p>
    </div>
  )
}

// formatEurCents matches the existing wallet-overview-card pattern —
// kept local to keep the component pure and free of imports outside
// React + shared/ui.
function formatEurCents(cents: number): string {
  return new Intl.NumberFormat("fr-FR", {
    style: "currency",
    currency: "EUR",
    maximumFractionDigits: 0,
  }).format(cents / 100)
}
