import { describe, it, expect, vi, beforeEach } from "vitest"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { NextIntlClientProvider } from "next-intl"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import messages from "@/../messages/fr.json"
import { OpportunitiesTabs } from "../opportunities-tabs"

// Drive the tab via mocked search params + router so we can assert the
// URL contract without spinning up the App Router runtime. Pattern
// mirrors `invoices-tabs.test.tsx` so reviewers see one tab convention
// across the codebase.
const mockReplace = vi.fn()
const tabValueRef = { current: "" as string | null }

vi.mock("next/navigation", () => ({
  useSearchParams: () => ({
    get: (key: string) => (key === "tab" ? tabValueRef.current : null),
  }),
}))

vi.mock("@i18n/navigation", () => ({
  useRouter: () => ({ replace: mockReplace }),
  Link: ({ children, ...rest }: { children: React.ReactNode } & Record<string, unknown>) =>
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    <a {...(rest as any)}>{children}</a>,
}))

// Stub the embedded feature components so the test focuses on the tab
// shell. Each replacement renders a unique sentinel + a `data-testid`
// counter so we can assert lazy mount / unmount semantics.
let opportunityRenderCount = 0
let applicationRenderCount = 0

vi.mock("@/features/job/components/opportunity-list", () => ({
  OpportunityList: () => {
    opportunityRenderCount += 1
    return (
      <div data-testid="opportunity-list-stub">
        Toutes les offres content
      </div>
    )
  },
}))

vi.mock("@/features/job/components/application-list", () => ({
  ApplicationList: () => {
    applicationRenderCount += 1
    return (
      <div data-testid="application-list-stub">
        Mes candidatures content
      </div>
    )
  },
}))

function renderTabs(initialTab: string | null = null) {
  tabValueRef.current = initialTab
  const client = new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0 },
      mutations: { retry: false },
    },
  })
  return render(
    <NextIntlClientProvider locale="fr" messages={messages}>
      <QueryClientProvider client={client}>
        <OpportunitiesTabs />
      </QueryClientProvider>
    </NextIntlClientProvider>,
  )
}

beforeEach(() => {
  vi.clearAllMocks()
  opportunityRenderCount = 0
  applicationRenderCount = 0
})

describe("OpportunitiesTabs", () => {
  it("defaults to the all-offers tab when no `?tab=` is set", () => {
    renderTabs(null)
    const allTab = screen.getByRole("tab", { name: /Toutes les offres/i })
    const applicationsTab = screen.getByRole("tab", {
      name: /Mes candidatures/i,
    })
    expect(allTab).toHaveAttribute("aria-selected", "true")
    expect(applicationsTab).toHaveAttribute("aria-selected", "false")
  })

  it("activates the applications tab when `?tab=applications`", () => {
    renderTabs("applications")
    const applicationsTab = screen.getByRole("tab", {
      name: /Mes candidatures/i,
    })
    expect(applicationsTab).toHaveAttribute("aria-selected", "true")
  })

  it("falls back to the default tab when the param value is unknown", () => {
    renderTabs("garbage")
    const allTab = screen.getByRole("tab", { name: /Toutes les offres/i })
    expect(allTab).toHaveAttribute("aria-selected", "true")
  })

  it("calls router.replace with the applications query on click", async () => {
    const user = userEvent.setup()
    renderTabs(null)
    await user.click(screen.getByRole("tab", { name: /Mes candidatures/i }))
    expect(mockReplace).toHaveBeenCalledWith("/opportunities?tab=applications")
  })

  it("strips the query string when navigating back to the default tab", async () => {
    const user = userEvent.setup()
    renderTabs("applications")
    await user.click(screen.getByRole("tab", { name: /Toutes les offres/i }))
    expect(mockReplace).toHaveBeenCalledWith("/opportunities")
  })

  it("does NOT mount ApplicationList until the applications tab is active", () => {
    renderTabs(null)
    // Default tab → only OpportunityList rendered.
    expect(screen.getByTestId("opportunity-list-stub")).toBeInTheDocument()
    expect(screen.queryByTestId("application-list-stub")).toBeNull()
    expect(opportunityRenderCount).toBe(1)
    expect(applicationRenderCount).toBe(0)
  })

  it("mounts ApplicationList only when applications tab is active", () => {
    renderTabs("applications")
    expect(screen.queryByTestId("opportunity-list-stub")).toBeNull()
    expect(screen.getByTestId("application-list-stub")).toBeInTheDocument()
    expect(opportunityRenderCount).toBe(0)
    expect(applicationRenderCount).toBe(1)
  })

  it("exposes both tabpanels with correct ARIA wiring", () => {
    renderTabs(null)
    const allPanel = document.getElementById("panel-opportunities-all")
    const applicationsPanel = document.getElementById(
      "panel-opportunities-applications",
    )
    expect(allPanel).toHaveAttribute(
      "aria-labelledby",
      "tab-opportunities-all",
    )
    expect(applicationsPanel).toHaveAttribute(
      "aria-labelledby",
      "tab-opportunities-applications",
    )
    // The inactive panel must be `hidden` so screen readers skip it.
    expect(applicationsPanel).toHaveAttribute("hidden")
    expect(allPanel).not.toHaveAttribute("hidden")
  })
})
