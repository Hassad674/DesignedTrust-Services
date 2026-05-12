"use client"

import { MutationCache, QueryCache, QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { useEffect, useState } from "react"
import { toast } from "sonner"
import { CookieConsentProvider } from "@/shared/components/analytics/cookie-consent-provider"
import { GoogleAnalyticsProvider } from "@/shared/components/analytics/google-analytics-provider"
import { PostHogProvider } from "@/shared/components/analytics/posthog-provider"
import { useTheme } from "@/shared/hooks/use-theme"
import { ApiError } from "@/shared/lib/api-client"

/** Map 403 error codes to user-friendly French messages for global toast. */
function getPermissionErrorMessage(error: ApiError): string | null {
  if (error.status !== 403) return null
  if (error.code === "no_organization") {
    return "Vous devez appartenir à une organisation pour effectuer cette action"
  }
  if (error.code === "permission_denied" || error.code === "forbidden") {
    return "Permission refusée — vous n'avez pas accès à cette fonctionnalité"
  }
  return null
}

/**
 * Reads `meta.suppressGlobalErrorToast` from a query/mutation's options.
 * When `true` the global QueryCache / MutationCache toast handler skips
 * the error — the local consumer is expected to surface its own message.
 *
 * This is the surgical escape hatch for flows where a 403 from a
 * background refetch (or a chained invalidation) is EXPECTED and would
 * otherwise spuriously alarm the user — e.g. the role-permissions editor,
 * which fans out a `["session"]` invalidation immediately after a
 * successful save and used to surface a false "permission denied" toast
 * as the post-success refetch raced the now-stale permission snapshot.
 */
function shouldSuppressGlobalErrorToast(meta: unknown): boolean {
  if (typeof meta !== "object" || meta === null) return false
  const flag = (meta as { suppressGlobalErrorToast?: unknown })
    .suppressGlobalErrorToast
  return flag === true
}

function ThemeInitializer() {
  const { theme } = useTheme()

  useEffect(() => {
    document.documentElement.classList.toggle("dark", theme === "dark")
  }, [theme])

  return null
}

export function Providers({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 2 * 60 * 1000, // 2 minutes — prevents refetching on every component mount
            gcTime: 10 * 60 * 1000, // 10 minutes — keep unused cache entries longer
            retry: (failureCount, error) => {
              // Never retry 403 — permission errors are not transient
              if (error instanceof ApiError && error.status === 403) return false
              return failureCount < 1
            },
            refetchOnWindowFocus: false, // avoid surprise refetches when switching tabs
          },
        },
        queryCache: new QueryCache({
          onError: (error, query) => {
            if (!(error instanceof ApiError)) return
            if (shouldSuppressGlobalErrorToast(query.meta)) return
            const message = getPermissionErrorMessage(error)
            if (message) {
              toast.error(message)
            }
          },
        }),
        mutationCache: new MutationCache({
          onError: (error, _vars, _onMutateResult, mutation) => {
            if (!(error instanceof ApiError)) return
            if (shouldSuppressGlobalErrorToast(mutation.meta)) return
            const message = getPermissionErrorMessage(error)
            if (message) {
              toast.error(message)
            }
          },
        }),
      }),
  )

  return (
    <QueryClientProvider client={queryClient}>
      <ThemeInitializer />
      {/*
        PostHogProvider must live INSIDE QueryClientProvider so it can
        consume useSession() to identify the logged-in user. It renders
        nothing — pure side-effect on the SDK lifecycle. The banner is
        rendered last so it floats above page content without forcing
        anyone to wrap their layouts.
      */}
      <PostHogProvider />
      {/*
        GoogleAnalyticsProvider mounts the gtag.js script via
        `@next/third-parties/google`. It renders nothing until BOTH
        NEXT_PUBLIC_GA_MEASUREMENT_ID is set AND the user opted in
        through the cookie banner. RGPD-compatible: no script loaded
        before consent.
      */}
      <GoogleAnalyticsProvider />
      {children}
      {/*
        CookieConsentProvider mounts vanilla-cookieconsent (CMP) which
        injects the consent dialog + preferences modal into
        `document.body`. The PostHog + GA4 providers above are gated on
        the `analytics` category — neither fires a network call before
        the user's first interaction with this CMP.
      */}
      <CookieConsentProvider />
    </QueryClientProvider>
  )
}
