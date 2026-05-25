import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import { render, screen, waitFor, fireEvent } from "@testing-library/react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { MemoryRouter } from "react-router-dom"
import { FeedbackPage } from "../components/feedback-page"
import * as api from "../api/feedback-api"
import type { FeedbackListResponse } from "../types"

function renderPage() {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <FeedbackPage />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

const sampleResponse: FeedbackListResponse = {
  data: [
    {
      id: "11111111-1111-1111-1111-111111111111",
      type: "bug",
      title: "Le bouton de connexion ne répond pas",
      description: "Rien ne se passe au clic.",
      status: "new",
      severity: "medium",
      page_url: "https://app.test/login",
      reporter_email: "user@test.com",
      reporter_user_id: null,
      context: {},
      attachment_count: 2,
      note_count: 0,
      created_at: "2026-05-20T10:00:00Z",
      updated_at: "2026-05-20T10:00:00Z",
    },
    {
      id: "22222222-2222-2222-2222-222222222222",
      type: "vulnerability",
      title: "Faille XSS sur le profil",
      description: "Injection possible.",
      status: "triaged",
      severity: "critical",
      page_url: "https://app.test/profile",
      reporter_email: "",
      reporter_user_id: null,
      context: {},
      attachment_count: 0,
      note_count: 3,
      created_at: "2026-05-21T10:00:00Z",
      updated_at: "2026-05-21T10:00:00Z",
    },
  ],
  next_cursor: "next-tok",
  has_more: true,
}

describe("FeedbackPage", () => {
  let listSpy: ReturnType<typeof vi.spyOn>

  beforeEach(() => {
    listSpy = vi.spyOn(api, "listFeedback")
  })
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it("renders rows from the API", async () => {
    listSpy.mockResolvedValue(sampleResponse)
    renderPage()

    await waitFor(() => {
      expect(
        screen.getByText("Le bouton de connexion ne répond pas"),
      ).toBeInTheDocument()
    })
    expect(screen.getByText("Faille XSS sur le profil")).toBeInTheDocument()
    // Type / status / severity labels also appear as filter <option>s,
    // so we assert the rendered badge <span> specifically (not the
    // <option> in the dropdowns).
    const asBadge = (label: string) =>
      screen.getAllByText(label).find((el) => el.tagName === "SPAN")
    expect(asBadge("Bug")).toBeInTheDocument()
    expect(asBadge("Sécurité")).toBeInTheDocument()
    expect(asBadge("Nouveau")).toBeInTheDocument()
    expect(asBadge("Critique")).toBeInTheDocument()
    // Reporter email + anonymous fallback
    expect(screen.getByText("user@test.com")).toBeInTheDocument()
    expect(screen.getByText("Anonyme")).toBeInTheDocument()
  })

  it("renders empty state when API returns no rows", async () => {
    listSpy.mockResolvedValue({ data: [], next_cursor: "", has_more: false })
    renderPage()

    await waitFor(() => {
      expect(screen.getByText("Aucun signalement")).toBeInTheDocument()
    })
  })

  it("changing the type filter refetches with the new type and reset cursor", async () => {
    listSpy.mockResolvedValue(sampleResponse)
    renderPage()

    await waitFor(() => {
      expect(listSpy).toHaveBeenCalledTimes(1)
    })

    const typeSelect = screen.getByLabelText("Filtrer par type") as HTMLSelectElement
    fireEvent.change(typeSelect, { target: { value: "vulnerability" } })

    await waitFor(() => {
      expect(listSpy).toHaveBeenCalledTimes(2)
    })
    const lastFilters = listSpy.mock.calls[1][0]
    expect(lastFilters.type).toBe("vulnerability")
    expect(lastFilters.cursor).toBe("")
  })

  it("changing the status filter refetches with the new status", async () => {
    listSpy.mockResolvedValue(sampleResponse)
    renderPage()

    await waitFor(() => expect(listSpy).toHaveBeenCalledTimes(1))

    const statusSelect = screen.getByLabelText("Filtrer par statut") as HTMLSelectElement
    fireEvent.change(statusSelect, { target: { value: "resolved" } })

    await waitFor(() => expect(listSpy).toHaveBeenCalledTimes(2))
    expect(listSpy.mock.calls[1][0].status).toBe("resolved")
  })

  it("clicking Suivant advances the cursor to next_cursor", async () => {
    listSpy.mockResolvedValue(sampleResponse)
    renderPage()

    // Wait for the rows (and therefore the pager) to render before
    // clicking — the pager only mounts once a non-empty page is loaded.
    const nextButton = await screen.findByRole("button", { name: /Suivant/ })
    fireEvent.click(nextButton)

    await waitFor(() => expect(listSpy).toHaveBeenCalledTimes(2))
    expect(listSpy.mock.calls[1][0].cursor).toBe("next-tok")
  })
})
