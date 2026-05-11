"use client"

import { Briefcase, Handshake } from "lucide-react"
import { useTranslations } from "next-intl"
import { useEffect, useState } from "react"

import { Button } from "@/shared/components/ui/button"
import { cn } from "@/shared/lib/utils"

import { useWalletSummary } from "../hooks/use-wallet"
import type { WalletSummaryTransaction } from "../api/wallet-api"
import {
  WalletStatusBadge,
  resolveWalletStatusTone,
} from "./wallet-status-badge"

/**
 * WalletUnifiedHistory — the merged mission + commission timeline
 * below the WALLET-UNIFY hero. Each row carries:
 *
 *   - a type icon (💼 missions / 🤝 commissions)
 *   - the amount in tabular Geist Mono, right-aligned
 *   - a status badge in one of four tones (paid / pending / escrowed
 *     / failed) — see `WalletStatusBadge`.
 *
 * Pagination is cursor-driven: the host hook caches one page per
 * cursor key, and a "Charger plus" ghost button advances by setting
 * the local cursor state.
 *
 * The component owns its OWN cursor state — keeping it out of the
 * URL avoids a noisy history stack and matches the existing wallet
 * UX. Initial mount hits the cursor-less page.
 */
export function WalletUnifiedHistory() {
  const t = useTranslations("walletUnified.history")
  const [cursor, setCursor] = useState<string | undefined>(undefined)
  const [accumulated, setAccumulated] = useState<WalletSummaryTransaction[]>([])

  const { data, isLoading, isError } = useWalletSummary(cursor)

  // First mount: seed `accumulated` from the first page once the
  // hook lands. Subsequent "Charger plus" advances both `accumulated`
  // and `cursor` together in handleLoadMore. The effect only fires
  // on initial seed — guarded by accumulated.length === 0 + no
  // cursor — so an empty subsequent page never re-seeds.
  useEffect(() => {
    if (
      data &&
      cursor === undefined &&
      accumulated.length === 0 &&
      data.recent_transactions.length > 0
    ) {
      setAccumulated(data.recent_transactions)
    }
  }, [data, cursor, accumulated.length])

  function handleLoadMore() {
    if (!data?.next_cursor) return
    setAccumulated((prev) => [...prev, ...data.recent_transactions])
    setCursor(data.next_cursor)
  }

  if (isLoading && accumulated.length === 0) {
    return <HistorySkeleton title={t("title")} subtitle={t("subtitle")} />
  }
  if (isError) {
    return null
  }

  const rows =
    accumulated.length > 0 ? accumulated : data?.recent_transactions ?? []
  const hasMore = Boolean(data?.next_cursor)

  return (
    <section
      data-testid="wallet-unified-history"
      className={cn("rounded-2xl border border-border bg-card p-4 md:p-5")}
      style={{ boxShadow: "var(--shadow-card)" }}
    >
      <HistoryHeader title={t("title")} subtitle={t("subtitle")} />
      {rows.length === 0 ? (
        <p className="px-2 py-6 text-center text-[13px] text-muted-foreground">
          {t("empty")}
        </p>
      ) : (
        <ul className="mt-3 divide-y divide-border">
          {rows.map((row) => (
            <li key={row.reference_id}>
              <TransactionRow row={row} />
            </li>
          ))}
        </ul>
      )}
      {hasMore && (
        <div className="mt-3 flex justify-center">
          <Button
            variant="outline"
            size="md"
            type="button"
            onClick={handleLoadMore}
            data-testid="wallet-history-load-more"
          >
            {t("loadMore")}
          </Button>
        </div>
      )}
    </section>
  )
}

function HistoryHeader({
  title,
  subtitle,
}: {
  title: string
  subtitle: string
}) {
  return (
    <header className="space-y-0.5">
      <h2 className="font-serif text-[18px] font-medium tracking-[-0.015em] text-foreground">
        {title}
      </h2>
      <p className="text-[12.5px] text-muted-foreground">{subtitle}</p>
    </header>
  )
}

function TransactionRow({ row }: { row: WalletSummaryTransaction }) {
  const t = useTranslations("walletUnified.history.row")
  const tStatus = useTranslations("walletUnified.history.status")
  const tone = resolveWalletStatusTone(row.status)
  const typeLabel = row.type === "mission" ? t("mission") : t("commission")
  const title = row.mission_title ?? t("untitled")
  const amount = formatEurCents(row.amount_cents)
  return (
    <div
      className={cn("flex items-center gap-3 px-2 py-3")}
      role="listitem"
      aria-label={t("ariaLabel", {
        type: typeLabel,
        amount,
        status: tStatus(tone),
      })}
      data-type={row.type}
      data-testid={`wallet-history-row-${row.reference_id}`}
    >
      <TypeIcon type={row.type} />
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-semibold text-foreground">
          {title}
        </p>
        <p className="mt-0.5 text-[11px] text-muted-foreground">
          {typeLabel} · {formatFrenchDate(row.occurred_at)}
        </p>
      </div>
      <div className="flex flex-col items-end gap-1.5">
        <span className="font-mono text-sm font-semibold tabular-nums text-foreground">
          {row.type === "mission" ? "+" : "+"}
          {amount}
        </span>
        <WalletStatusBadge status={row.status} label={tStatus(tone)} />
      </div>
    </div>
  )
}

function TypeIcon({ type }: { type: "mission" | "commission" }) {
  const Icon = type === "mission" ? Briefcase : Handshake
  return (
    <span
      className={cn(
        "grid h-9 w-9 shrink-0 place-items-center rounded-lg",
        type === "mission"
          ? "bg-primary-soft text-primary"
          : "bg-success-soft text-success",
      )}
      aria-hidden="true"
    >
      <Icon className="h-4 w-4" />
    </span>
  )
}

function HistorySkeleton({
  title,
  subtitle,
}: {
  title: string
  subtitle: string
}) {
  return (
    <section
      className="rounded-2xl border border-border bg-card p-4 md:p-5"
      style={{ boxShadow: "var(--shadow-card)" }}
    >
      <HistoryHeader title={title} subtitle={subtitle} />
      <ul className="mt-3 divide-y divide-border">
        {[0, 1, 2].map((i) => (
          <li key={i} className="flex items-center gap-3 px-2 py-3">
            <div className="h-9 w-9 shrink-0 animate-shimmer rounded-lg bg-muted" />
            <div className="flex-1 space-y-1.5">
              <div className="h-3 w-1/2 animate-shimmer rounded bg-muted" />
              <div className="h-2.5 w-1/3 animate-shimmer rounded bg-muted" />
            </div>
            <div className="h-5 w-16 animate-shimmer rounded bg-muted" />
          </li>
        ))}
      </ul>
    </section>
  )
}

function formatEurCents(cents: number): string {
  return new Intl.NumberFormat("fr-FR", {
    style: "currency",
    currency: "EUR",
    maximumFractionDigits: 0,
  }).format(cents / 100)
}

function formatFrenchDate(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ""
  return new Intl.DateTimeFormat("fr-FR", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  }).format(d)
}
