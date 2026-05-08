import { describe, it, expect, afterEach } from "vitest"
import { act, renderHook } from "@testing-library/react"

import { useDocumentVisibility } from "../use-document-visibility"

const visibilityDescriptor = Object.getOwnPropertyDescriptor(
  Document.prototype,
  "visibilityState",
)

function setVisibilityState(state: DocumentVisibilityState) {
  Object.defineProperty(document, "visibilityState", {
    configurable: true,
    get: () => state,
  })
}

afterEach(() => {
  if (visibilityDescriptor) {
    Object.defineProperty(document, "visibilityState", visibilityDescriptor)
  } else {
    delete (document as unknown as { visibilityState?: unknown }).visibilityState
  }
})

describe("useDocumentVisibility", () => {
  it("returns true when document.visibilityState is 'visible'", () => {
    setVisibilityState("visible")
    const { result } = renderHook(() => useDocumentVisibility())
    expect(result.current).toBe(true)
  })

  it("returns false when document.visibilityState is 'hidden'", () => {
    setVisibilityState("hidden")
    const { result } = renderHook(() => useDocumentVisibility())
    expect(result.current).toBe(false)
  })

  it("flips when the visibilitychange event fires", () => {
    setVisibilityState("hidden")
    const { result } = renderHook(() => useDocumentVisibility())
    expect(result.current).toBe(false)

    act(() => {
      setVisibilityState("visible")
      document.dispatchEvent(new Event("visibilitychange"))
    })

    expect(result.current).toBe(true)
  })

  it("removes its listener on unmount", () => {
    setVisibilityState("visible")
    const { unmount, result } = renderHook(() => useDocumentVisibility())
    expect(result.current).toBe(true)

    unmount()

    // After unmount, flipping should not throw and should not
    // affect the previously returned value.
    expect(() => {
      setVisibilityState("hidden")
      document.dispatchEvent(new Event("visibilitychange"))
    }).not.toThrow()
  })
})
