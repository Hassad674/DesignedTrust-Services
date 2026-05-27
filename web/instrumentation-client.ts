// Browser-side Sentry bootstrap. Next.js 16 runs this module on the
// client before hydration (the App Router replacement for the old
// `sentry.client.config.ts`). Initialisation is fully env-gated by
// initSentry(): with no NEXT_PUBLIC_SENTRY_DSN it is a no-op, so the
// client bundle ships Sentry as dead code with zero runtime effect.
import * as Sentry from "@sentry/nextjs"

import { initSentry } from "@/shared/lib/sentry"

initSentry(Sentry, true)

// Instruments App Router client-side navigations so transactions are
// named after the route. Exported unconditionally — the hook is inert
// when Sentry never initialised (no DSN), so this is safe to ship
// before the DSN is configured on Vercel.
export const onRouterTransitionStart = Sentry.captureRouterTransitionStart
