import { describe, it, expect, vi, beforeEach } from "vitest"
import {
  listFeedback,
  getFeedback,
  updateFeedback,
  addFeedbackNote,
} from "../api/feedback-api"
import { EMPTY_FEEDBACK_FILTERS } from "../types"

const mockAdminApi = vi.fn()

vi.mock("@/shared/lib/api-client", () => ({
  adminApi: (...a: unknown[]) => mockAdminApi(...a),
}))

beforeEach(() => {
  vi.clearAllMocks()
  mockAdminApi.mockResolvedValue({})
})

describe("admin feedback api", () => {
  it("listFeedback appends limit=20 on empty filters", () => {
    listFeedback(EMPTY_FEEDBACK_FILTERS)
    expect(mockAdminApi).toHaveBeenCalledWith("/api/v1/admin/feedback?limit=20")
  })

  it("listFeedback appends type when present", () => {
    listFeedback({ ...EMPTY_FEEDBACK_FILTERS, type: "bug" })
    expect(mockAdminApi).toHaveBeenCalledWith(
      "/api/v1/admin/feedback?type=bug&limit=20",
    )
  })

  it("listFeedback appends status, severity and search", () => {
    listFeedback({
      type: "vulnerability",
      status: "triaged",
      severity: "high",
      search: "crash",
      cursor: "",
    })
    expect(mockAdminApi).toHaveBeenCalledWith(
      "/api/v1/admin/feedback?type=vulnerability&status=triaged&severity=high&search=crash&limit=20",
    )
  })

  it("listFeedback appends cursor when present", () => {
    listFeedback({ ...EMPTY_FEEDBACK_FILTERS, cursor: "tok" })
    expect(mockAdminApi).toHaveBeenCalledWith(
      "/api/v1/admin/feedback?cursor=tok&limit=20",
    )
  })

  it("getFeedback GETs by id", () => {
    getFeedback("f-1")
    expect(mockAdminApi).toHaveBeenCalledWith("/api/v1/admin/feedback/f-1")
  })

  it("updateFeedback PATCHes the status payload", () => {
    updateFeedback("f-1", { status: "resolved" })
    expect(mockAdminApi).toHaveBeenCalledWith("/api/v1/admin/feedback/f-1", {
      method: "PATCH",
      body: { status: "resolved" },
    })
  })

  it("updateFeedback PATCHes the severity payload", () => {
    updateFeedback("f-1", { severity: "critical" })
    expect(mockAdminApi).toHaveBeenCalledWith("/api/v1/admin/feedback/f-1", {
      method: "PATCH",
      body: { severity: "critical" },
    })
  })

  it("addFeedbackNote POSTs the body wrapped in { body }", () => {
    addFeedbackNote("f-1", "looks like a real bug")
    expect(mockAdminApi).toHaveBeenCalledWith(
      "/api/v1/admin/feedback/f-1/notes",
      { method: "POST", body: { body: "looks like a real bug" } },
    )
  })

  it("propagates errors", async () => {
    mockAdminApi.mockRejectedValueOnce(new Error("403"))
    await expect(getFeedback("f-1")).rejects.toThrow("403")
  })
})
