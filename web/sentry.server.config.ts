// Node.js runtime Sentry bootstrap, loaded by instrumentation.ts when
// NEXT_RUNTIME === "nodejs". Env-gated: initSentry() does nothing
// unless SENTRY_DSN (or the NEXT_PUBLIC_ fallback) is set, so the
// server stays inert before the DSN is configured on Vercel.
import * as Sentry from "@sentry/nextjs"

import { initSentry } from "@/shared/lib/sentry"

initSentry(Sentry)
