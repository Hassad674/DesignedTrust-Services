import { describe, it, expect, vi, beforeEach } from "vitest"
import { submitFeedback, presignFeedbackAttachment } from "../feedback-api"

const mockApiClient = vi.fn()
vi.mock("@/shared/lib/api-client", () => ({
  apiClient: (...a: unknown[]) => mockApiClient(...a),
}))

beforeEach(() => {
  vi.clearAllMocks()
})

describe("submitFeedback", () => {
  it("POSTs the report body to /api/v1/feedback", async () => {
    mockApiClient.mockResolvedValueOnce({
      id: "rep-1",
      type: "bug",
      status: "new",
      created_at: "2026-05-25T10:00:00Z",
    })

    const body = {
      type: "bug",
      title: "Button broken",
      description: "The submit button does nothing on Safari.",
      page_url: "https://app.test/projects",
      context: {
        role: "agency",
        locale: "fr",
        platform: "web",
        app_version: "",
        viewport: "1280x720",
        user_agent: "UA",
      },
      reporter_email: "",
      attachment_keys: [],
      hp: "",
    }

    const res = await submitFeedback(body)

    expect(res.id).toBe("rep-1")
    expect(mockApiClient).toHaveBeenCalledTimes(1)
    const [path, options] = mockApiClient.mock.calls[0]
    expect(path).toBe("/api/v1/feedback")
    expect(options.method).toBe("POST")
    expect(options.body).toEqual(body)
  })

  it("propagates the API error to the caller", async () => {
    mockApiClient.mockRejectedValueOnce(new Error("boom"))
    await expect(
      submitFeedback({
        type: "security",
        title: "t",
        description: "d",
        page_url: "",
        context: null,
        reporter_email: "",
        attachment_keys: [],
        hp: "",
      }),
    ).rejects.toThrow("boom")
  })
})

describe("presignFeedbackAttachment", () => {
  it("POSTs the presign request to the presign endpoint", async () => {
    mockApiClient.mockResolvedValueOnce({
      url: "https://r2/put/abc",
      object_key: "feedback/abc.png",
      kind: "image",
    })

    const res = await presignFeedbackAttachment({
      kind: "image",
      content_type: "image/png",
      size_bytes: 1234,
      filename: "shot.png",
    })

    expect(res.object_key).toBe("feedback/abc.png")
    const [path, options] = mockApiClient.mock.calls[0]
    expect(path).toBe("/api/v1/feedback/attachments/presign")
    expect(options.method).toBe("POST")
    expect(options.body).toEqual({
      kind: "image",
      content_type: "image/png",
      size_bytes: 1234,
      filename: "shot.png",
    })
  })
})
