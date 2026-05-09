/**
 * Issue 3 — hide profile completion bar for enterprise role on
 * /client-profile (2026-05-09).
 *
 * Enterprises only have 4 sections in their checklist; the bar adds
 * noise without an actionable nudge. Agencies (which also reach this
 * page via their client persona) keep it.
 */

import { describe, it, expect, vi, beforeEach } from "vitest"
import { render } from "@testing-library/react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { NextIntlClientProvider } from "next-intl"
import { createElement } from "react"

import messages from "@/../messages/fr.json"

const completionBarSpy = vi.fn<(props: unknown) => null>(() => null)

vi.mock("@/shared/hooks/use-user", () => ({
  useUser: vi.fn(),
  useOrganization: vi.fn(),
  useLogout: () => vi.fn(),
}))

vi.mock("@/shared/hooks/use-permissions", () => ({
  useHasPermission: () => true,
}))

vi.mock("@/shared/hooks/use-upload-photo", () => ({
  useUploadPhoto: () => ({ mutateAsync: vi.fn(), isPending: false }),
}))

vi.mock("@/features/profile-completion/components/profile-completion-bar", () => ({
  ProfileCompletionBar: (props: unknown) => completionBarSpy(props),
}))

vi.mock("../../hooks/use-my-client-profile", () => ({
  useMyClientProfile: () => ({
    data: {
      organization_id: "org-1",
      company_name: "ACME",
      client_description: "We sell things",
      avatar_url: "",
      total_spent: 0,
      review_count: 0,
      average_rating: null,
      projects_completed_as_client: 0,
      project_history: [],
    },
    isLoading: false,
    isError: false,
  }),
}))

vi.mock("../../hooks/use-update-client-profile", () => ({
  useUpdateClientProfile: () => ({ mutateAsync: vi.fn(), isPending: false }),
}))

// Stub heavy children to keep this test focused on the bar gate.
vi.mock("../client-project-history-section", () => ({
  ClientProjectHistorySection: () => null,
}))

vi.mock("../client-profile-header", () => ({
  ClientProfileHeader: () => null,
}))

vi.mock("../client-profile-editor", () => ({
  ClientProfileEditor: () => null,
}))

vi.mock("../client-profile-description", () => ({
  ClientProfileDescription: () => null,
}))

import { ClientProfilePage } from "../client-profile-page"
import { useOrganization } from "@/shared/hooks/use-user"

const mockedUseOrganization = vi.mocked(useOrganization)

function makeOrgResult(id: string, type: string) {
  return {
    data: { id, type },
  } as unknown as ReturnType<typeof useOrganization>
}

function renderPage() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    createElement(
      QueryClientProvider,
      { client },
      createElement(
        NextIntlClientProvider,
        { locale: "fr", messages, children: createElement(ClientProfilePage) },
      ),
    ),
  )
}

beforeEach(() => {
  completionBarSpy.mockClear()
})

describe("ClientProfilePage — completion bar gate", () => {
  it("renders the bar for an agency org", () => {
    mockedUseOrganization.mockReturnValue(makeOrgResult("org-1", "agency"))
    renderPage()
    expect(completionBarSpy).toHaveBeenCalled()
  })

  it("HIDES the bar for an enterprise org", () => {
    mockedUseOrganization.mockReturnValue(makeOrgResult("org-2", "enterprise"))
    renderPage()
    expect(completionBarSpy).not.toHaveBeenCalled()
  })
})
