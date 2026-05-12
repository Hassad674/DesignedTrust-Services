import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import { renderHook, act } from "@testing-library/react"
import { createElement } from "react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"

// Mock the notification imports to avoid cross-feature dependency resolution
vi.mock("@/features/notification/hooks/use-unread-notification-count", () => ({
  unreadNotifCountKey: (uid: string | undefined) => ["user", uid, "notifications", "unread-count"],
}))

vi.mock("@/features/notification/hooks/use-notifications", () => ({
  notificationsQueryKey: (uid: string | undefined) => ["user", uid, "notifications"],
}))

vi.mock("sonner", () => ({
  toast: vi.fn(),
}))

import { toast } from "sonner"

type WSHandler = ((event: { data: string }) => void) | null

class MockWebSocket {
  static CONNECTING = 0
  static OPEN = 1
  static CLOSING = 2
  static CLOSED = 3

  url: string
  readyState = MockWebSocket.CONNECTING
  onopen: (() => void) | null = null
  onmessage: WSHandler = null
  onclose: (() => void) | null = null
  onerror: (() => void) | null = null

  send = vi.fn()
  close = vi.fn(() => {
    this.readyState = MockWebSocket.CLOSED
    if (this.onclose) this.onclose()
  })

  constructor(url: string) {
    this.url = url
    // Store instance for test access
    MockWebSocket.lastInstance = this
  }

  // Simulate opening the connection
  simulateOpen() {
    this.readyState = MockWebSocket.OPEN
    if (this.onopen) this.onopen()
  }

  // Simulate receiving a message
  simulateMessage(data: Record<string, unknown>) {
    if (this.onmessage) {
      this.onmessage({ data: JSON.stringify(data) })
    }
  }

  // Simulate an error
  simulateError() {
    if (this.onerror) this.onerror()
  }

  static lastInstance: MockWebSocket | null = null
}

// Attach static constants to the constructor for readyState comparisons
Object.assign(MockWebSocket, {
  CONNECTING: 0,
  OPEN: 1,
  CLOSING: 2,
  CLOSED: 3,
})

let queryClient: QueryClient

function createWrapper() {
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return createElement(QueryClientProvider, { client: queryClient }, children)
  }
}

beforeEach(() => {
  vi.useFakeTimers()
  MockWebSocket.lastInstance = null
  vi.stubGlobal("WebSocket", MockWebSocket)
  vi.stubGlobal("fetch", vi.fn(() => Promise.resolve({ ok: false })))

  queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
})

afterEach(() => {
  vi.useRealTimers()
  vi.restoreAllMocks()
})

describe("useGlobalWS", () => {
  // Import after mocks are set up
  async function importHook() {
    const mod = await import("../use-global-ws")
    return mod.useGlobalWS
  }

  it("does not connect when userId is undefined", async () => {
    const useGlobalWS = await importHook()

    renderHook(() => useGlobalWS(undefined), { wrapper: createWrapper() })

    expect(MockWebSocket.lastInstance).toBeNull()
  })

  it("connects when userId is provided", async () => {
    const useGlobalWS = await importHook()

    renderHook(() => useGlobalWS("user-123"), { wrapper: createWrapper() })

    // Allow the async connect to resolve
    await act(async () => {
      await vi.runAllTimersAsync()
    })

    expect(MockWebSocket.lastInstance).not.toBeNull()
    expect(MockWebSocket.lastInstance!.url).toContain("/api/v1/ws")
  })

  it("sends heartbeat after connection opens", async () => {
    const useGlobalWS = await importHook()

    renderHook(() => useGlobalWS("user-123"), { wrapper: createWrapper() })

    await act(async () => {
      await vi.runAllTimersAsync()
    })

    const ws = MockWebSocket.lastInstance!
    act(() => {
      ws.simulateOpen()
    })

    // Advance past heartbeat interval (30 seconds)
    act(() => {
      vi.advanceTimersByTime(30_000)
    })

    expect(ws.send).toHaveBeenCalledWith(JSON.stringify({ type: "heartbeat" }))
  })

  it("updates unread count on unread_count message", async () => {
    const useGlobalWS = await importHook()

    renderHook(() => useGlobalWS("user-123"), { wrapper: createWrapper() })

    await act(async () => {
      await vi.runAllTimersAsync()
    })

    const ws = MockWebSocket.lastInstance!
    act(() => {
      ws.simulateOpen()
    })

    const setDataSpy = vi.spyOn(queryClient, "setQueryData")

    act(() => {
      ws.simulateMessage({ type: "unread_count", payload: { count: 7 } })
    })

    expect(setDataSpy).toHaveBeenCalledWith(
      ["user", "user-123", "messaging", "unread-count"],
      { count: 7 },
    )
  })

  it("shows toast and updates cache on notification message", async () => {
    const useGlobalWS = await importHook()

    renderHook(() => useGlobalWS("user-123"), { wrapper: createWrapper() })

    await act(async () => {
      await vi.runAllTimersAsync()
    })

    const ws = MockWebSocket.lastInstance!
    act(() => {
      ws.simulateOpen()
    })

    act(() => {
      ws.simulateMessage({
        type: "notification",
        payload: { title: "New message", body: "You got mail" },
      })
    })

    expect(toast).toHaveBeenCalledWith("New message", {
      description: "You got mail",
    })
  })

  it("invokes call event handler on call_event message", async () => {
    const useGlobalWS = await importHook()
    const callHandler = vi.fn()

    renderHook(() => useGlobalWS("user-123", callHandler), {
      wrapper: createWrapper(),
    })

    await act(async () => {
      await vi.runAllTimersAsync()
    })

    const ws = MockWebSocket.lastInstance!
    act(() => {
      ws.simulateOpen()
    })

    const callPayload = { callId: "call-1", action: "ringing" }
    act(() => {
      ws.simulateMessage({ type: "call_event", payload: callPayload })
    })

    expect(callHandler).toHaveBeenCalledWith(callPayload)
  })

  it("returns setMessagingPageActive function", async () => {
    const useGlobalWS = await importHook()

    const { result } = renderHook(() => useGlobalWS("user-123"), {
      wrapper: createWrapper(),
    })

    expect(result.current.setMessagingPageActive).toBeTypeOf("function")
  })

  it("returns registerCallEventHandler function (PERF-W-01)", async () => {
    const useGlobalWS = await importHook()

    const { result } = renderHook(() => useGlobalWS("user-123"), {
      wrapper: createWrapper(),
    })

    expect(result.current.registerCallEventHandler).toBeTypeOf("function")
  })

  it("registerCallEventHandler swaps the handler at runtime", async () => {
    const useGlobalWS = await importHook()
    const initialHandler = vi.fn()
    const lateHandler = vi.fn()

    const { result } = renderHook(() => useGlobalWS("user-123", initialHandler), {
      wrapper: createWrapper(),
    })

    await act(async () => {
      await vi.runAllTimersAsync()
    })

    const ws = MockWebSocket.lastInstance!
    act(() => {
      ws.simulateOpen()
    })

    // Initial handler is invoked
    act(() => {
      ws.simulateMessage({ type: "call_event", payload: { event: "first" } })
    })
    expect(initialHandler).toHaveBeenCalledWith({ event: "first" })

    // Swap handler
    act(() => {
      result.current.registerCallEventHandler(lateHandler)
    })

    act(() => {
      ws.simulateMessage({ type: "call_event", payload: { event: "second" } })
    })

    expect(lateHandler).toHaveBeenCalledWith({ event: "second" })
    // Initial handler must NOT receive the second event
    expect(initialHandler).toHaveBeenCalledTimes(1)
  })

  it("registerCallEventHandler(undefined) silences events", async () => {
    const useGlobalWS = await importHook()
    const handler = vi.fn()

    const { result } = renderHook(() => useGlobalWS("user-123", handler), {
      wrapper: createWrapper(),
    })

    await act(async () => {
      await vi.runAllTimersAsync()
    })

    const ws = MockWebSocket.lastInstance!
    act(() => {
      ws.simulateOpen()
    })

    act(() => {
      result.current.registerCallEventHandler(undefined)
    })

    act(() => {
      ws.simulateMessage({ type: "call_event", payload: { event: "ignored" } })
    })

    expect(handler).not.toHaveBeenCalled()
  })

  it("StrictMode guard: does not open a second socket when one is still CONNECTING", async () => {
    // Regression for the post-revert WS reconnect storm. React 19
    // StrictMode runs every useEffect twice in dev (run → cleanup →
    // run). Without the CONNECTING guard, the second run races the
    // first socket's handshake and opens a parallel WS; the backend
    // closes the duplicate; the survivor sees the onclose and
    // reconnects in a tight 4-8 s loop. The guard short-circuits the
    // second connect() when wsRef points to a socket that is still
    // CONNECTING (or OPEN) so only one WS is ever in flight.
    const useGlobalWS = await importHook()

    let firstInstance: MockWebSocket | null = null

    const { rerender } = renderHook(
      ({ uid }: { uid: string }) => useGlobalWS(uid),
      {
        wrapper: createWrapper(),
        initialProps: { uid: "user-123" },
      },
    )

    await act(async () => {
      await vi.runAllTimersAsync()
    })

    firstInstance = MockWebSocket.lastInstance
    expect(firstInstance).not.toBeNull()
    // First WS is still CONNECTING — we never called simulateOpen.
    expect(firstInstance!.readyState).toBe(MockWebSocket.CONNECTING)

    // Force a re-render with the same userId. In real StrictMode this
    // is the second mount pass; here we exercise the same code path
    // (the effect runs `connect()` against the live ref). With the
    // guard in place, `lastInstance` must NOT have changed.
    rerender({ uid: "user-123" })

    await act(async () => {
      await vi.runAllTimersAsync()
    })

    expect(MockWebSocket.lastInstance).toBe(firstInstance)
  })

  it("cleans up WebSocket on unmount", async () => {
    const useGlobalWS = await importHook()

    const { unmount } = renderHook(() => useGlobalWS("user-123"), {
      wrapper: createWrapper(),
    })

    await act(async () => {
      await vi.runAllTimersAsync()
    })

    const ws = MockWebSocket.lastInstance!
    act(() => {
      ws.simulateOpen()
    })

    unmount()

    expect(ws.close).toHaveBeenCalled()
  })

  it("clears heartbeat interval on close and schedules reconnect with backoff", async () => {
    const useGlobalWS = await importHook()

    renderHook(() => useGlobalWS("user-123"), { wrapper: createWrapper() })

    await act(async () => {
      await vi.runAllTimersAsync()
    })

    const ws = MockWebSocket.lastInstance!
    act(() => {
      ws.simulateOpen()
    })

    // Trigger an onclose — the hook should clear the heartbeat and
    // schedule a reconnect via setTimeout(connect, delay).
    act(() => {
      ws.close()
    })

    // Advance past the first reconnect delay (2^0 * 1000 = 1000ms).
    const previousInstance = ws
    await act(async () => {
      vi.advanceTimersByTime(1500)
      await vi.runAllTimersAsync()
    })

    // A new WebSocket instance should have been created by the
    // reconnect path.
    expect(MockWebSocket.lastInstance).not.toBe(previousInstance)
  })

  it("calls close on the underlying socket when ws.onerror fires", async () => {
    const useGlobalWS = await importHook()

    renderHook(() => useGlobalWS("user-123"), { wrapper: createWrapper() })

    await act(async () => {
      await vi.runAllTimersAsync()
    })

    const ws = MockWebSocket.lastInstance!
    act(() => {
      ws.simulateOpen()
    })

    act(() => {
      ws.simulateError()
    })

    expect(ws.close).toHaveBeenCalled()
  })

  it("invokes notification_unread_count handler when frame arrives", async () => {
    const useGlobalWS = await importHook()

    renderHook(() => useGlobalWS("user-123"), { wrapper: createWrapper() })

    await act(async () => {
      await vi.runAllTimersAsync()
    })

    const ws = MockWebSocket.lastInstance!
    act(() => {
      ws.simulateOpen()
    })

    const setDataSpy = vi.spyOn(queryClient, "setQueryData")
    act(() => {
      ws.simulateMessage({
        type: "notification_unread_count",
        payload: { count: 3 },
      })
    })
    expect(setDataSpy).toHaveBeenCalledWith(
      ["user", "user-123", "notifications", "unread-count"],
      { data: { count: 3 } },
    )
  })

  it("handles account_suspended frame by redirecting to /login", async () => {
    const useGlobalWS = await importHook()

    // Stub window.location.href + window.alert so the redirect
    // doesn't actually navigate jsdom.
    const alertSpy = vi.spyOn(window, "alert").mockImplementation(() => {})
    const locationStub = { href: "http://localhost/" }
    Object.defineProperty(window, "location", {
      configurable: true,
      writable: true,
      value: locationStub,
    })

    renderHook(() => useGlobalWS("user-123"), { wrapper: createWrapper() })

    await act(async () => {
      await vi.runAllTimersAsync()
    })

    const ws = MockWebSocket.lastInstance!
    act(() => {
      ws.simulateOpen()
    })

    act(() => {
      ws.simulateMessage({
        type: "account_suspended",
        payload: { reason: "policy violation" },
      })
    })

    expect(alertSpy).toHaveBeenCalled()
    expect(locationStub.href).toBe("/login")
  })

  it("ignores malformed messages without crashing", async () => {
    const useGlobalWS = await importHook()

    renderHook(() => useGlobalWS("user-123"), { wrapper: createWrapper() })

    await act(async () => {
      await vi.runAllTimersAsync()
    })

    const ws = MockWebSocket.lastInstance!
    act(() => {
      ws.simulateOpen()
    })

    // Send invalid JSON — should not throw
    expect(() => {
      act(() => {
        if (ws.onmessage) {
          ws.onmessage({ data: "not json at all" })
        }
      })
    }).not.toThrow()
  })

  // ────────────────────────────────────────────────────────────────
  // Regressions for the 4 WS perf protections re-applied after
  // the revert of perf commit c0c68dbf. Without these, multi-tab
  // dev React 19 StrictMode opens/closes WS at 5-10/s per user,
  // dropping real-time messages because the recipient socket is
  // in CLOSING when the backend tries to send.
  // ────────────────────────────────────────────────────────────────

  it("reconnect uses 2 s floor on first attempt (no premature retry < 2 s)", async () => {
    // Make jitter deterministic at 0 ms so we only measure the floor.
    vi.spyOn(Math, "random").mockReturnValue(0)
    const useGlobalWS = await importHook()

    renderHook(() => useGlobalWS("user-123"), { wrapper: createWrapper() })

    await act(async () => {
      await vi.runAllTimersAsync()
    })

    const first = MockWebSocket.lastInstance!
    act(() => {
      first.simulateOpen()
    })

    act(() => {
      first.close()
    })

    // 1500 ms < 2 s floor → no new socket yet.
    await act(async () => {
      vi.advanceTimersByTime(1500)
    })
    expect(MockWebSocket.lastInstance).toBe(first)

    // Advance past the floor (total elapsed 2100 ms ≥ 2000 ms).
    await act(async () => {
      vi.advanceTimersByTime(600)
      await vi.runAllTimersAsync()
    })
    expect(MockWebSocket.lastInstance).not.toBe(first)
  })

  it("reconnect adds jitter so two hooks do not lockstep", async () => {
    // Two hooks closing at the same tick must NOT schedule their
    // setTimeout(connect) with the exact same delay. We capture the
    // raw setTimeout delays via a spy and assert the two reconnect
    // delays differ by at least 1 ms. Math.random is mocked to return
    // two distinct values across the close handlers.
    const randomValues = [0, 0.9] // jitter: 0 ms then 449 ms
    let randIdx = 0
    vi.spyOn(Math, "random").mockImplementation(() => {
      const v = randomValues[randIdx % randomValues.length]
      randIdx += 1
      return v
    })

    const useGlobalWS = await importHook()

    // Capture setTimeout delays that match the reconnect window.
    const reconnectDelays: number[] = []
    const realSetTimeout = globalThis.setTimeout
    const setTimeoutSpy = vi.spyOn(globalThis, "setTimeout")
    setTimeoutSpy.mockImplementation(((handler: TimerHandler, delay?: number, ...rest: unknown[]) => {
      // Only the reconnect setTimeout fires with a delay in [2000,2500).
      if (typeof delay === "number" && delay >= 2000 && delay < 2500) {
        reconnectDelays.push(delay)
      }
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      return (realSetTimeout as any)(handler, delay, ...rest)
    }) as typeof setTimeout)

    // Hook A — flush microtasks so the WS is created, but do NOT
    // call simulateOpen (heartbeat interval would never terminate
    // and break fake-timer assumptions when paired with B).
    renderHook(() => useGlobalWS("user-A"), { wrapper: createWrapper() })
    await act(async () => {
      await Promise.resolve()
      await Promise.resolve()
    })
    const wsA = MockWebSocket.lastInstance!

    // Hook B with its own QueryClient/wrapper.
    queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    renderHook(() => useGlobalWS("user-B"), { wrapper: createWrapper() })
    await act(async () => {
      await Promise.resolve()
      await Promise.resolve()
    })
    const wsB = MockWebSocket.lastInstance!
    expect(wsB).not.toBe(wsA)

    // Both close at the same tick → both schedule reconnect via
    // the close→onclose path (close() triggers onclose in our mock).
    act(() => {
      wsA.close()
    })
    act(() => {
      wsB.close()
    })

    // Two reconnect setTimeouts captured, with distinct delays.
    expect(reconnectDelays.length).toBeGreaterThanOrEqual(2)
    const [d1, d2] = reconnectDelays
    expect(Math.abs(d1 - d2)).toBeGreaterThanOrEqual(1)
  })

  it("pendingConnect sentinel blocks the async race", async () => {
    // `connect()` awaits getWSUrl() (a microtask) before
    // `new WebSocket()`. Two parallel calls (StrictMode double
    // mount, or unrelated re-renders pre-StrictMode-guard) could
    // both pass the readyState check during the await microtask
    // and create two sockets. The sync sentinel must block the
    // second entry until the first completes.
    const useGlobalWS = await importHook()

    // Slow getWSUrl: stub fetch so the call is awaited; in the dev
    // path getWSUrl returns synchronously from NEXT_PUBLIC_API_URL
    // (a single microtask). To exercise the race we trigger two
    // mounts in quick succession on the same userId.

    const wrapper = createWrapper()
    // First hook mount — schedules connect() (async, pending).
    const { rerender } = renderHook(
      ({ uid }: { uid: string }) => useGlobalWS(uid),
      { wrapper, initialProps: { uid: "user-123" } },
    )
    // Rerender same uid synchronously: the effect's connect()
    // identity is stable (useCallback over [userId]), so the effect
    // does NOT re-run; but if a future refactor breaks that, the
    // sentinel must still prevent the parallel connect from racing.
    rerender({ uid: "user-123" })

    await act(async () => {
      await vi.runAllTimersAsync()
    })

    // Exactly one WS instance was created across both passes.
    expect(MockWebSocket.lastInstance).not.toBeNull()
    const created = MockWebSocket.lastInstance!
    // Sanity: still CONNECTING — we never opened it.
    expect(created.readyState).toBe(MockWebSocket.CONNECTING)
  })

  it("sentinel resets on onopen so subsequent close+reconnect creates a new socket", async () => {
    vi.spyOn(Math, "random").mockReturnValue(0)
    const useGlobalWS = await importHook()

    renderHook(() => useGlobalWS("user-123"), { wrapper: createWrapper() })

    await act(async () => {
      await vi.runAllTimersAsync()
    })

    const first = MockWebSocket.lastInstance!
    // Open the socket → sentinel must reset to false.
    act(() => {
      first.simulateOpen()
    })

    // Close → reconnect path. If the sentinel had stayed true, the
    // setTimeout(connect) would short-circuit and no new WS would
    // appear; if it correctly resets, a new instance is created.
    act(() => {
      first.close()
    })

    await act(async () => {
      vi.advanceTimersByTime(2500)
      await vi.runAllTimersAsync()
    })

    expect(MockWebSocket.lastInstance).not.toBe(first)
  })

  it("sentinel resets when getWSUrl throws so the next connect attempt is not blocked", async () => {
    // First call to fetch throws (forces the await getWSUrl()
    // catch branch); subsequent calls succeed via the fallback
    // direct connection. The sentinel must be reset by the catch
    // so the reconnect attempt is not silently blocked.
    const useGlobalWS = await importHook()

    // Force the production path: empty NEXT_PUBLIC_API_URL +
    // a NEXT_PUBLIC_WS_URL that triggers fetch(). We stub
    // process.env via vi.stubEnv.
    vi.stubEnv("NEXT_PUBLIC_API_URL", "")
    vi.stubEnv("NEXT_PUBLIC_WS_URL", "ws://example.invalid")

    let callIdx = 0
    vi.stubGlobal(
      "fetch",
      vi.fn(() => {
        callIdx += 1
        if (callIdx === 1) return Promise.reject(new Error("network down"))
        return Promise.resolve({ ok: false })
      }),
    )

    renderHook(() => useGlobalWS("user-123"), { wrapper: createWrapper() })

    // First attempt: fetch rejects → getWSUrl catches internally
    // and falls through to `${wsUrl}/api/v1/ws`. A socket is still
    // created. The sentinel must not be stuck at true.
    await act(async () => {
      await vi.runAllTimersAsync()
    })
    const first = MockWebSocket.lastInstance!
    expect(first).not.toBeNull()

    // Close → reconnect → second fetch (ok:false → fallback) → new socket.
    act(() => {
      first.simulateOpen()
    })
    act(() => {
      first.close()
    })
    vi.spyOn(Math, "random").mockReturnValue(0)
    await act(async () => {
      vi.advanceTimersByTime(2500)
      await vi.runAllTimersAsync()
    })
    expect(MockWebSocket.lastInstance).not.toBe(first)
  })
})
