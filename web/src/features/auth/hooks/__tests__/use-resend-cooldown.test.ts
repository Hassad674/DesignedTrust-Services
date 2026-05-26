import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import { act, renderHook } from "@testing-library/react"
import { useResendCooldown } from "../use-resend-cooldown"

beforeEach(() => {
  vi.useFakeTimers()
})

afterEach(() => {
  vi.useRealTimers()
})

describe("useResendCooldown", () => {
  it("starts inactive at zero", () => {
    const { result } = renderHook(() => useResendCooldown())
    expect(result.current.remaining).toBe(0)
    expect(result.current.active).toBe(false)
  })

  it("counts down once per second and goes inactive at zero", () => {
    const { result } = renderHook(() => useResendCooldown())

    act(() => result.current.start(3))
    expect(result.current.remaining).toBe(3)
    expect(result.current.active).toBe(true)

    act(() => vi.advanceTimersByTime(1000))
    expect(result.current.remaining).toBe(2)

    act(() => vi.advanceTimersByTime(2000))
    expect(result.current.remaining).toBe(0)
    expect(result.current.active).toBe(false)
  })

  it("does not drift below zero when more time elapses", () => {
    const { result } = renderHook(() => useResendCooldown())
    act(() => result.current.start(1))
    act(() => vi.advanceTimersByTime(5000))
    expect(result.current.remaining).toBe(0)
  })

  it("restarting resets the countdown", () => {
    const { result } = renderHook(() => useResendCooldown())
    act(() => result.current.start(2))
    act(() => vi.advanceTimersByTime(1000))
    expect(result.current.remaining).toBe(1)
    act(() => result.current.start(5))
    expect(result.current.remaining).toBe(5)
  })

  it("clears the interval on unmount (no orphan timer)", () => {
    const clearSpy = vi.spyOn(globalThis, "clearInterval")
    const { result, unmount } = renderHook(() => useResendCooldown())
    act(() => result.current.start(10))
    unmount()
    expect(clearSpy).toHaveBeenCalled()
    clearSpy.mockRestore()
  })
})
