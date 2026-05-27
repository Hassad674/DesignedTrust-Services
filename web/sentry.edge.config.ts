// Edge runtime Sentry bootstrap, loaded by instrumentation.ts when
// NEXT_RUNTIME === "edge" (middleware + edge routes). Env-gated:
// initSentry() does nothing unless a DSN is configured, so the edge
// runtime stays inert before the DSN is set on Vercel.
import * as Sentry from "@sentry/nextjs"

import { initSentry } from "@/shared/lib/sentry"

initSentry(Sentry)
