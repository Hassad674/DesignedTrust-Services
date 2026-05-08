import { describe, it, expect, vi, beforeEach } from "vitest"
import { renderHook, waitFor } from "@testing-library/react"
import { createElement } from "react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"

vi.mock("@/shared/hooks/use-current-user-id", () => ({
  useCurrentUserId: vi.fn(),
}))

vi.mock("@/shared/lib/api-client", () => ({
  apiClient: vi.fn(),
  API_BASE_URL: "http://localhost:8080",
}))

import { useCurrentUserId } from "@/shared/hooks/use-current-user-id"
import { apiClient } from "@/shared/lib/api-client"
import {
  profileCompletionQueryKey,
  useProfileCompletion,
} from "../hooks/use-profile-completion"

const mockedUseCurrentUserId = vi.mocked(useCurrentUserId)
const mockedApiClient = vi.mocked(apiClient)

function createWrapper() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0 } },
  })
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return createElement(QueryClientProvider, { client }, children)
  }
}

beforeEach(() => {
  vi.clearAllMocks()
})

describe("profileCompletionQueryKey", () => {
  it("returns a user-scoped key", () => {
    expect(profileCompletionQueryKey("user-1")).toEqual([
      "user",
      "user-1",
      "profile-completion",
    ])
  })
})

describe("useProfileCompletion", () => {
  it("fetches the report and calls /api/v1/me/profile/completion", async () => {
    mockedUseCurrentUserId.mockReturnValue("user-1")
    mockedApiClient.mockResolvedValue({
      role: "provider",
      persona: "freelance",
      percent: 60,
      total_sections: 10,
      filled_sections: 6,
      sections: [],
    })

    const { result } = renderHook(() => useProfileCompletion(), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(mockedApiClient).toHaveBeenCalledWith("/api/v1/me/profile/completion")
    expect(result.current.data?.percent).toBe(60)
  })

  it("disables the query when no user id is available", async () => {
    mockedUseCurrentUserId.mockReturnValue(undefined)

    const { result } = renderHook(() => useProfileCompletion(), {
      wrapper: createWrapper(),
    })

    // Query is disabled — fetch is never called and isLoading stays false.
    expect(result.current.isLoading).toBe(false)
    expect(mockedApiClient).not.toHaveBeenCalled()
  })

  it("propagates API errors", async () => {
    mockedUseCurrentUserId.mockReturnValue("user-2")
    mockedApiClient.mockRejectedValue(new Error("kaboom"))

    const { result } = renderHook(() => useProfileCompletion(), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isError).toBe(true))
    expect(result.current.error?.message).toBe("kaboom")
  })
})
