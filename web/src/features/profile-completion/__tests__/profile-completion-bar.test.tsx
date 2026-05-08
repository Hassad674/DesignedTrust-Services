import { describe, expect, it, vi, beforeEach } from "vitest"
import { render, screen, fireEvent } from "@testing-library/react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { NextIntlClientProvider } from "next-intl"

import messages from "@/../messages/fr.json"

vi.mock("@i18n/navigation", () => ({
  Link: ({ href, children, className, onClick }: {
    href: string
    children: React.ReactNode
    className?: string
    onClick?: () => void
  }) => (
    <a href={href} className={className} onClick={onClick}>
      {children}
    </a>
  ),
}))

vi.mock("@/shared/hooks/use-current-user-id", () => ({
  useCurrentUserId: vi.fn(() => "user-1"),
}))

vi.mock("@/shared/lib/api-client", () => ({
  apiClient: vi.fn(),
  API_BASE_URL: "http://localhost:8080",
}))

import { apiClient } from "@/shared/lib/api-client"
import { ProfileCompletionBar } from "../components/profile-completion-bar"
import type { ProfileCompletionReport } from "../api/profile-completion-api"

const mockedApiClient = vi.mocked(apiClient)

function renderBar(report: ProfileCompletionReport, props: Parameters<typeof ProfileCompletionBar>[0] = {}) {
  mockedApiClient.mockResolvedValue(report)
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0 } },
  })
  return render(
    <QueryClientProvider client={client}>
      <NextIntlClientProvider locale="fr" messages={messages}>
        <ProfileCompletionBar {...props} />
      </NextIntlClientProvider>
    </QueryClientProvider>,
  )
}

const baseReport: ProfileCompletionReport = {
  role: "provider",
  persona: "freelance",
  percent: 50,
  total_sections: 10,
  filled_sections: 5,
  sections: [
    {
      key: "title",
      filled: true,
      label_key: "profile.completion.section.title",
      completion_path: "/dashboard/profile/edit",
    },
    {
      key: "about",
      filled: false,
      label_key: "profile.completion.section.about",
      completion_path: "/dashboard/profile/edit",
    },
  ],
}

beforeEach(() => {
  vi.clearAllMocks()
})

describe("ProfileCompletionBar", () => {
  it("renders title with the percent and the filled/total subtitle once data loads", async () => {
    renderBar(baseReport)
    expect(await screen.findByText(/Profil rempli à 50%/)).toBeInTheDocument()
    expect(screen.getByText(/5\/10 sections complétées/)).toBeInTheDocument()
  })

  it("renders progressbar with correct aria values", async () => {
    renderBar(baseReport)
    const bar = await screen.findByRole("progressbar")
    expect(bar.getAttribute("aria-valuenow")).toBe("50")
    expect(bar.getAttribute("aria-valuemin")).toBe("0")
    expect(bar.getAttribute("aria-valuemax")).toBe("100")
  })

  it("renders zero percent state with 0/N subtitle", async () => {
    renderBar({ ...baseReport, percent: 0, filled_sections: 0 })
    const bar = await screen.findByRole("progressbar")
    expect(bar.getAttribute("aria-valuenow")).toBe("0")
    expect(screen.getByText(/0\/10 sections complétées/)).toBeInTheDocument()
  })

  it("renders the complete subtitle at 100%", async () => {
    renderBar({
      ...baseReport,
      percent: 100,
      filled_sections: baseReport.total_sections,
    })
    expect(await screen.findByText(/Toutes les sections sont complètes/)).toBeInTheDocument()
  })

  it("hides itself at 100% when hideWhenComplete is true", async () => {
    const { container } = renderBar(
      { ...baseReport, percent: 100, filled_sections: baseReport.total_sections },
      { hideWhenComplete: true },
    )
    // Wait a tick for the query to resolve and the component to re-render.
    await new Promise((r) => setTimeout(r, 30))
    expect(container.firstChild).toBeNull()
  })

  it("opens the modal listing the missing sections on click", async () => {
    renderBar(baseReport)
    const button = await screen.findByRole("button", {
      name: /Profil rempli à 50%, ouvrir la liste des sections/,
    })
    fireEvent.click(button)

    expect(await screen.findByRole("dialog")).toBeInTheDocument()
    expect(screen.getByTestId("completion-section-list")).toBeInTheDocument()
    // Filled section appears struck through; missing section appears as a Link.
    expect(screen.getByText(/Titre professionnel/)).toBeInTheDocument()
    expect(screen.getByText(/À propos/)).toBeInTheDocument()
  })

  it("renders the compact pill when sidebar variant is collapsed", async () => {
    renderBar(baseReport, { variant: "sidebar", collapsed: true })
    const pill = await screen.findByLabelText(/Profil rempli à 50%/)
    expect(pill.textContent).toContain("50%")
  })
})
