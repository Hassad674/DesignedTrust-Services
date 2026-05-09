/**
 * Tests for the persona-aware apply modal radio (Fix 2).
 *
 * Asserts:
 *  - radio is hidden for non-referrer-enabled providers + agencies
 *  - radio shows two options for a referrer-enabled provider
 *  - default selection matches the workspace_mode cookie state
 *  - submit forwards the selected applicant_kind to useApplyToJob
 */
import { describe, expect, it, vi, beforeEach } from "vitest"
import { render, screen, fireEvent } from "@testing-library/react"
import { NextIntlClientProvider } from "next-intl"

import { ApplyModal } from "../apply-modal"

const messages = {
  opportunity: {
    apply: "Postuler",
    close: "Fermer",
    yourMessage: "Votre message",
    messagePlaceholder: "Pourquoi êtes-vous le bon candidat ?",
    optionalVideo: "Vidéo de présentation (optionnel)",
    uploadVideo: "Ajouter une vidéo",
    uploading: "Envoi en cours...",
    error: "Une erreur est survenue",
    applyAsLegend: "Tu postules en tant que",
    applyAsFreelance: "Freelance",
    applyAsReferrer: "Apporteur d'affaires",
  },
}

const mockMutate = vi.fn()
let mockUserData: { role: string; referrer_enabled: boolean } | null = null
let mockReferrerMode = false

vi.mock("../../hooks/use-job-applications", () => ({
  useApplyToJob: () => ({
    mutate: mockMutate,
    isPending: false,
    isError: false,
    error: null,
  }),
}))

vi.mock("@/shared/hooks/use-user", () => ({
  useUser: () => ({ data: mockUserData }),
}))

vi.mock("@/shared/hooks/use-workspace", () => ({
  useWorkspace: () => ({ isReferrerMode: mockReferrerMode }),
}))

vi.mock("@/shared/lib/upload-api", () => ({
  uploadVideo: vi.fn(),
}))

function renderModal() {
  return render(
    <NextIntlClientProvider locale="fr" messages={messages}>
      <ApplyModal open onClose={() => {}} jobId="job-1" />
    </NextIntlClientProvider>,
  )
}

beforeEach(() => {
  mockMutate.mockClear()
  mockUserData = null
  mockReferrerMode = false
})

describe("ApplyModal · persona radio", () => {
  it("hides the radio when the user is an agency", () => {
    mockUserData = { role: "agency", referrer_enabled: false }
    renderModal()
    expect(screen.queryByText("Tu postules en tant que")).toBeNull()
  })

  it("hides the radio when the user is a non-referrer provider", () => {
    mockUserData = { role: "provider", referrer_enabled: false }
    renderModal()
    expect(screen.queryByText("Tu postules en tant que")).toBeNull()
  })

  it("shows the two-option radio for a referrer-enabled provider", () => {
    mockUserData = { role: "provider", referrer_enabled: true }
    renderModal()
    expect(screen.getByText("Tu postules en tant que")).toBeInTheDocument()
    expect(screen.getByLabelText("Freelance")).toBeInTheDocument()
    expect(screen.getByLabelText("Apporteur d'affaires")).toBeInTheDocument()
  })

  it("defaults to freelance when not in referrer workspace", () => {
    mockUserData = { role: "provider", referrer_enabled: true }
    mockReferrerMode = false
    renderModal()
    const freelanceRadio = screen.getByLabelText("Freelance") as HTMLInputElement
    const referrerRadio = screen.getByLabelText("Apporteur d'affaires") as HTMLInputElement
    expect(freelanceRadio.checked).toBe(true)
    expect(referrerRadio.checked).toBe(false)
  })

  it("defaults to referrer when workspace_mode is referrer", () => {
    mockUserData = { role: "provider", referrer_enabled: true }
    mockReferrerMode = true
    renderModal()
    const freelanceRadio = screen.getByLabelText("Freelance") as HTMLInputElement
    const referrerRadio = screen.getByLabelText("Apporteur d'affaires") as HTMLInputElement
    expect(referrerRadio.checked).toBe(true)
    expect(freelanceRadio.checked).toBe(false)
  })

  it("forwards applicantKind=referrer when the referrer radio is selected", () => {
    mockUserData = { role: "provider", referrer_enabled: true }
    renderModal()
    fireEvent.click(screen.getByLabelText("Apporteur d'affaires"))
    fireEvent.click(screen.getByRole("button", { name: "Postuler" }))
    expect(mockMutate).toHaveBeenCalledTimes(1)
    const args = mockMutate.mock.calls[0][0]
    expect(args.applicantKind).toBe("referrer")
    expect(args.jobId).toBe("job-1")
  })

  it("forwards applicantKind=freelance for the default selection", () => {
    mockUserData = { role: "provider", referrer_enabled: true }
    renderModal()
    fireEvent.click(screen.getByRole("button", { name: "Postuler" }))
    expect(mockMutate).toHaveBeenCalledTimes(1)
    expect(mockMutate.mock.calls[0][0].applicantKind).toBe("freelance")
  })

  it("does NOT forward applicantKind for a non-referrer provider (legacy compat)", () => {
    mockUserData = { role: "provider", referrer_enabled: false }
    renderModal()
    fireEvent.click(screen.getByRole("button", { name: "Postuler" }))
    expect(mockMutate).toHaveBeenCalledTimes(1)
    expect(mockMutate.mock.calls[0][0].applicantKind).toBeUndefined()
  })
})
