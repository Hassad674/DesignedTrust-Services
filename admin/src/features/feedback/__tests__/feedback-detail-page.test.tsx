import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import { render, screen, waitFor, fireEvent } from "@testing-library/react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { MemoryRouter, Routes, Route } from "react-router-dom"
import { FeedbackDetailPage } from "../components/feedback-detail-page"
import * as api from "../api/feedback-api"
import type { FeedbackReportDetail } from "../types"

const REPORT_ID = "11111111-1111-1111-1111-111111111111"

function renderDetail() {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={[`/feedback/${REPORT_ID}`]}>
        <Routes>
          <Route path="/feedback/:id" element={<FeedbackDetailPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

const sampleDetail: FeedbackReportDetail = {
  id: REPORT_ID,
  type: "bug",
  title: "Le formulaire plante",
  description: "Erreur 500 à la soumission.",
  status: "new",
  severity: "medium",
  page_url: "https://app.test/contact",
  reporter_email: "reporter@test.com",
  reporter_user_id: "99999999-9999-9999-9999-999999999999",
  context: { platform: "web", locale: "fr", role: "agency" },
  attachment_count: 2,
  note_count: 1,
  created_at: "2026-05-20T10:00:00Z",
  updated_at: "2026-05-20T10:00:00Z",
  attachments: [
    {
      id: "att-img",
      kind: "image",
      content_type: "image/png",
      size_bytes: 2048,
      url: "https://signed/img.png",
      created_at: "2026-05-20T10:00:00Z",
    },
    {
      id: "att-vid",
      kind: "video",
      content_type: "video/mp4",
      size_bytes: 1048576,
      url: "https://signed/clip.mp4",
      created_at: "2026-05-20T10:00:00Z",
    },
  ],
  notes: [
    {
      id: "note-1",
      admin_user_id: "admin-1",
      body: "Reproduit en staging.",
      created_at: "2026-05-20T11:00:00Z",
    },
  ],
}

describe("FeedbackDetailPage", () => {
  let getSpy: ReturnType<typeof vi.spyOn>

  beforeEach(() => {
    getSpy = vi.spyOn(api, "getFeedback")
  })
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it("renders the report info, context, attachments and notes", async () => {
    getSpy.mockResolvedValue(sampleDetail)
    renderDetail()

    await waitFor(() => {
      expect(screen.getByText("Le formulaire plante")).toBeInTheDocument()
    })
    expect(screen.getByText("Erreur 500 à la soumission.")).toBeInTheDocument()
    // Context flattened + labelled
    expect(screen.getByText("Plateforme")).toBeInTheDocument()
    expect(screen.getByText("web")).toBeInTheDocument()

    // Attachments: image rendered as <img>, video as <video>
    const image = screen.getByAltText(
      "Aperçu de la pièce jointe",
    ) as HTMLImageElement
    expect(image.src).toBe("https://signed/img.png")
    const video = document.querySelector("video") as HTMLVideoElement
    expect(video).not.toBeNull()
    expect(video.getAttribute("src")).toBe("https://signed/clip.mp4")

    // Notes thread
    expect(screen.getByText("Reproduit en staging.")).toBeInTheDocument()
  })

  it("clicking an image thumbnail opens the lightbox", async () => {
    getSpy.mockResolvedValue(sampleDetail)
    renderDetail()

    await waitFor(() => {
      expect(screen.getByText("Le formulaire plante")).toBeInTheDocument()
    })

    fireEvent.click(
      screen.getByRole("button", { name: "Agrandir la pièce jointe" }),
    )

    await waitFor(() => {
      expect(screen.getByRole("dialog")).toBeInTheDocument()
    })
    const full = screen.getByAltText(
      "Pièce jointe en taille réelle",
    ) as HTMLImageElement
    expect(full.src).toBe("https://signed/img.png")
  })

  it("changing the status select PATCHes with the new status", async () => {
    getSpy.mockResolvedValue(sampleDetail)
    const updateSpy = vi
      .spyOn(api, "updateFeedback")
      .mockResolvedValue({ ...sampleDetail })
    renderDetail()

    await waitFor(() => {
      expect(screen.getByLabelText("Statut")).toBeInTheDocument()
    })

    fireEvent.change(screen.getByLabelText("Statut"), {
      target: { value: "in_progress" },
    })

    await waitFor(() => {
      expect(updateSpy).toHaveBeenCalledWith(REPORT_ID, {
        status: "in_progress",
      })
    })
  })

  it("changing the severity select PATCHes with the new severity", async () => {
    getSpy.mockResolvedValue(sampleDetail)
    const updateSpy = vi
      .spyOn(api, "updateFeedback")
      .mockResolvedValue({ ...sampleDetail })
    renderDetail()

    await waitFor(() => {
      expect(screen.getByLabelText("Gravité")).toBeInTheDocument()
    })

    fireEvent.change(screen.getByLabelText("Gravité"), {
      target: { value: "critical" },
    })

    await waitFor(() => {
      expect(updateSpy).toHaveBeenCalledWith(REPORT_ID, {
        severity: "critical",
      })
    })
  })

  it("submitting the note form POSTs the trimmed body", async () => {
    getSpy.mockResolvedValue(sampleDetail)
    const noteSpy = vi.spyOn(api, "addFeedbackNote").mockResolvedValue({
      id: "note-2",
      admin_user_id: "admin-1",
      body: "Note ajoutée",
      created_at: "2026-05-20T12:00:00Z",
    })
    renderDetail()

    await waitFor(() => {
      expect(screen.getByLabelText("Ajouter une note")).toBeInTheDocument()
    })

    fireEvent.change(screen.getByLabelText("Ajouter une note"), {
      target: { value: "  Investigation en cours  " },
    })
    fireEvent.click(screen.getByRole("button", { name: "Ajouter" }))

    await waitFor(() => {
      expect(noteSpy).toHaveBeenCalledWith(REPORT_ID, "Investigation en cours")
    })
  })

  it("renders the error state when the detail fails to load", async () => {
    getSpy.mockRejectedValue(new Error("boom"))
    renderDetail()

    await waitFor(() => {
      expect(
        screen.getByText("Erreur lors du chargement du signalement"),
      ).toBeInTheDocument()
    })
  })
})
