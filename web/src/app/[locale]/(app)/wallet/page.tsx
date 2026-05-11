"use client"

import { WalletUnifiedPage } from "@/features/wallet/components/wallet-unified-page"

// WALLET-UNIFY Run C — the /wallet route now consumes the unified
// /wallet/summary endpoint (Run B backend). The legacy
// `wallet-page.tsx` orchestrator is kept in the feature folder so
// dependant surfaces (mobile parity in Run D, future admin views)
// can still import the per-leg components without ripples — only
// the page-level composition has switched.
export default function Page() {
  return <WalletUnifiedPage />
}
