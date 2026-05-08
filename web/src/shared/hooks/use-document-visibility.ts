"use client"

import { useCallback, useSyncExternalStore } from "react"

/**
 * SSR-safe hook that tracks `document.visibilityState`.
 * Returns true on the server and during hydration (the safe default
 * is "visible" so guards do not silently disable themselves), then
 * updates from the real `document.visibilityState` once mounted.
 *
 * Used to short-circuit one-shot reads (e.g. the call-reconcile probe)
 * when the tab is hidden — typical case is a middle-click opening a
 * background tab. Combined with `refetchIntervalInBackground: false`
 * on TanStack Query polls, this stops dev rate-limit thrash from
 * background tabs that the user isn't even looking at.
 *
 * Uses useSyncExternalStore so the initial value comes straight from
 * `document.visibilityState` on the client (no setState-in-effect
 * cascade) and re-renders only when the visibility actually flips.
 */
export function useDocumentVisibility(): boolean {
  const subscribe = useCallback((onStoreChange: () => void) => {
    document.addEventListener("visibilitychange", onStoreChange)
    return () => document.removeEventListener("visibilitychange", onStoreChange)
  }, [])

  const getSnapshot = useCallback(
    () => document.visibilityState !== "hidden",
    [],
  )

  // SSR fallback: assume visible so the gate never silently disables
  // itself during the very first render before hydration.
  const getServerSnapshot = useCallback(() => true, [])

  return useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot)
}
