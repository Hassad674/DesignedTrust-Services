"use client"

import { useCallback, useEffect, useRef, useState } from "react"

// useResendCooldown drives the "resend code" cooldown affordance: after
// a successful resend the caller starts a countdown, and the button
// stays disabled until it elapses. This is a pure UI timer (not server
// state), so a setInterval in an effect is the right tool — TanStack
// Query has nothing to fetch here.
//
// The interval is torn down on unmount and whenever the countdown hits
// zero, so no orphan timers leak across navigations.
export type ResendCooldown = {
  /** Seconds left before the resend action is allowed again. */
  remaining: number
  /** True while the countdown is running (button stays disabled). */
  active: boolean
  /** Begin (or restart) the countdown from `seconds`. */
  start: (seconds: number) => void
}

export function useResendCooldown(): ResendCooldown {
  const [remaining, setRemaining] = useState(0)
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const clear = useCallback(() => {
    if (intervalRef.current !== null) {
      clearInterval(intervalRef.current)
      intervalRef.current = null
    }
  }, [])

  const start = useCallback(
    (seconds: number) => {
      clear()
      setRemaining(seconds)
      intervalRef.current = setInterval(() => {
        setRemaining((prev) => {
          if (prev <= 1) {
            clear()
            return 0
          }
          return prev - 1
        })
      }, 1000)
    },
    [clear],
  )

  // Belt-and-braces unmount cleanup — `clear` is also called when the
  // countdown reaches zero, but a navigation mid-countdown must not
  // leave the interval running.
  useEffect(() => clear, [clear])

  return { remaining, active: remaining > 0, start }
}
