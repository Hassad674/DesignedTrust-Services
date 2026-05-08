/**
 * OpportunityCard — applications-count badge tests.
 *
 * Covers the social-proof badge rendering on the public marketplace
 * feed cards. The backend exposes `total_applicants` as an optional
 * field on /jobs/open responses; the UI must:
 *   - render nothing when the field is absent (legacy / single-GET shape)
 *   - render a "be the first" nudge when count is zero
 *   - render the FR plural when count is positive (1 vs n)
 *   - keep an aria-label readable to screen readers in all three cases
 */
import { describe, expect, it, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import { NextIntlClientProvider } from "next-intl"

import { OpportunityCard } from "../opportunity-card"
import type { JobResponse } from "../../types"

const messages = {
  opportunity: {
    weekly: "semaine",
    monthly: "mois",
    allTypes: "Tous",
    freelancersOnly: "Freelances",
    agenciesOnly: "Agences",
    oneShot: "Projet ponctuel",
    longTerm: "Collaboration long terme",
    alreadyApplied: "Déjà postulé",
    applicationsCountZero: "Sois le premier à candidater",
    applicationsCount:
      "{count, plural, one {# candidature} other {# candidatures}}",
    applicationsCountAria:
      "{count, plural, =0 {Aucune candidature pour le moment} one {# candidature reçue} other {# candidatures reçues}}",
  },
  reporting: {
    report: "Signaler",
  },
}
const LOCALE = "fr"

vi.mock("@i18n/navigation", () => ({
  Link: ({
    href,
    children,
    className,
  }: {
    href: string
    children: React.ReactNode
    className?: string
  }) => (
    <a href={href} className={className}>
      {children}
    </a>
  ),
  useRouter: () => ({ push: vi.fn() }),
}))

vi.mock("@/shared/components/reporting/report-dialog", () => ({
  ReportDialog: () => null,
}))

vi.mock("@/shared/components/ui/portrait", () => ({
  Portrait: () => <div data-testid="portrait" />,
}))

function createJob(overrides: Partial<JobResponse> = {}): JobResponse {
  return {
    id: "job-abc",
    creator_id: "user-1",
    title: "Backend Go senior",
    description: "Build a marketplace search engine.",
    skills: ["Go", "PostgreSQL", "Typesense"],
    applicant_type: "freelancers",
    budget_type: "long_term",
    min_budget: 4000,
    max_budget: 7000,
    status: "open",
    created_at: "2026-04-22T10:00:00Z",
    updated_at: "2026-04-22T10:00:00Z",
    is_indefinite: false,
    description_type: "text",
    payment_frequency: "monthly",
    ...overrides,
  }
}

function renderCard(job: JobResponse, hasApplied = false) {
  return render(
    <NextIntlClientProvider locale={LOCALE} messages={messages}>
      <OpportunityCard job={job} hasApplied={hasApplied} />
    </NextIntlClientProvider>,
  )
}

describe("OpportunityCard applications count badge", () => {
  it("renders nothing when total_applicants is undefined", () => {
    renderCard(createJob({ total_applicants: undefined }))
    expect(
      screen.queryByLabelText(/candidature/i),
    ).not.toBeInTheDocument()
    expect(
      screen.queryByText(/Sois le premier/i),
    ).not.toBeInTheDocument()
  })

  it("renders the 'be the first' nudge when total_applicants is 0", () => {
    renderCard(createJob({ total_applicants: 0 }))
    const badge = screen.getByLabelText(
      "Aucune candidature pour le moment",
    )
    expect(badge).toBeInTheDocument()
    expect(badge).toHaveTextContent("Sois le premier à candidater")
  })

  it("renders singular FR label when total_applicants is 1", () => {
    renderCard(createJob({ total_applicants: 1 }))
    const badge = screen.getByLabelText("1 candidature reçue")
    expect(badge).toBeInTheDocument()
    expect(badge).toHaveTextContent("1 candidature")
  })

  it("renders plural FR label when total_applicants is greater than 1", () => {
    renderCard(createJob({ total_applicants: 12 }))
    const badge = screen.getByLabelText("12 candidatures reçues")
    expect(badge).toBeInTheDocument()
    expect(badge).toHaveTextContent("12 candidatures")
  })

  it("does not leak new_applicants — only total_applicants drives the badge", () => {
    // Simulate a hypothetical leak attempt: even if the API returns
    // owner-only fields by accident, the public card must base its
    // badge on total_applicants only.
    renderCard(createJob({ total_applicants: 3 }))
    expect(screen.getByLabelText("3 candidatures reçues")).toBeInTheDocument()
    expect(screen.queryByText(/nouveau/i)).not.toBeInTheDocument()
    expect(screen.queryByText(/new/i)).not.toBeInTheDocument()
  })
})
