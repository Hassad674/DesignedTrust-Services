import { describe, it, expect, vi, beforeEach } from "vitest"
import { renderHook, waitFor, act } from "@testing-library/react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { createElement } from "react"
import { useSubmitFeedback } from "../use-submit-feedback"
import type { SubmitFeedbackRequest } from "../../types"

const mockSubmitFeedback = vi.fn()
vi.mock("../../api/feedback-api", () => ({
  submitFeedback: (...a: unknown[]) => mockSubmitFeedback(...a),
}))

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0 },
      mutations: { retry: false },
    },
  })
  const Wrapper = ({ children }: { children: React.ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children)
  return Wrapper
}

const body: SubmitFeedbackRequest = {
  type: "bug",
  title: "Crash on save",
  description: "Saving the profile throws a 500 on Firefox.",
  page_url: "https://app.test/account",
  context: null,
  reporter_email: "",
  attachment_keys: [],
  hp: "",
}

beforeEach(() => {
  vi.clearAllMocks()
})

describe("useSubmitFeedback", () => {
  it("calls submitFeedback with the body and resolves with the response", async () => {
    mockSubmitFeedback.mockResolvedValueOnce({
      id: "rep-9",
      type: "bug",
      status: "new",
      created_at: "2026-05-25T12:00:00Z",
    })

    const { result } = renderHook(() => useSubmitFeedback(), {
      wrapper: createWrapper(),
    })

    await act(async () => {
      result.current.mutate(body)
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(mockSubmitFeedback).toHaveBeenCalledWith(body)
    expect(result.current.data?.id).toBe("rep-9")
  })

  it("surfaces an error state when the submission fails", async () => {
    mockSubmitFeedback.mockRejectedValueOnce(new Error("429 rate limited"))

    const { result } = renderHook(() => useSubmitFeedback(), {
      wrapper: createWrapper(),
    })

    await act(async () => {
      result.current.mutate(body)
    })

    await waitFor(() => expect(result.current.isError).toBe(true))
    expect(result.current.error?.message).toBe("429 rate limited")
  })
})
