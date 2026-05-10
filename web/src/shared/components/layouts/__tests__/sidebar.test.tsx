/**
 * Sidebar tests — PERF-FIX-W-SIDEBAR-POLL.
 *
 * The sidebar used to start a `setInterval(updateSearch, 300)` per
 * `<NavLink>` instance. With ~17 navigation items rendered for an
 * agency user, that fired ~57 callbacks/second on idle and was the
 * dominant CPU draw for an otherwise-quiet dev tab.
 *
 * The regression guards below assert two invariants:
 *   1. SOURCE-LEVEL: `setInterval` must not appear in the file. Any
 *      future regression that re-introduces a polling loop here will
 *      flip the test red.
 *   2. RUNTIME: rendering the sidebar with a fully mocked auth tree
 *      does not register any timers (vi.useFakeTimers + getTimerCount).
 *      A passing test means the sidebar relies on `useSearchParams()`
 *      — which is what we want.
 */

import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import { render } from "@testing-library/react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { createElement, type ReactNode } from "react"

// Hoist mocks so the import order does not matter.
vi.mock("@/shared/hooks/use-user", () => ({
  useUser: () => ({ data: { id: "u-1", role: "agency", display_name: "Agence" } }),
  useOrganization: () => ({ data: { id: "o-1", type: "agency" } }),
  useLogout: () => vi.fn(),
}))

vi.mock("@/shared/hooks/use-workspace", () => ({
  useWorkspace: () => ({
    isReferrerMode: false,
    setReferrerMode: vi.fn(),
    switchToReferrer: vi.fn(() => "/dashboard"),
    switchToFreelance: vi.fn(() => "/dashboard"),
  }),
}))

vi.mock("@/shared/hooks/use-unread-count", () => ({
  useUnreadCount: () => ({ data: { count: 0 } }),
  unreadCountQueryKey: () => ["messaging", "unread-count"],
}))

vi.mock("@/features/profile-completion/components/profile-completion-bar", () => ({
  ProfileCompletionBar: () => null,
}))

vi.mock("@/shared/components/ui/user-avatar", () => ({
  UserAvatar: () => null,
}))

vi.mock("@/shared/components/layouts/logout-confirm-dialog", () => ({
  LogoutConfirmDialog: () => null,
}))

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}))

vi.mock("next/navigation", () => ({
  useSearchParams: () => new URLSearchParams(""),
}))

vi.mock("@i18n/navigation", () => ({
  Link: ({ children, ...rest }: React.ComponentProps<"a">) =>
    createElement("a", rest, children),
  usePathname: () => "/dashboard",
  useRouter: () => ({ push: vi.fn(), replace: vi.fn(), back: vi.fn() }),
}))

import { Sidebar } from "../sidebar"

function withQueryClient(ui: ReactNode) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return createElement(QueryClientProvider, { client }, ui)
}

describe("Sidebar — PERF-FIX-W-SIDEBAR-POLL", () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.clearAllMocks()
  })

  it("does not register timers per NavLink (regression guard)", async () => {
    render(
      withQueryClient(
        <Sidebar collapsed={false} onToggleCollapse={vi.fn()} onClose={vi.fn()} />,
      ),
    )

    // Source-level invariant — the file must not contain `setInterval`
    // anywhere. Any future re-introduction of polling will flip this red.
    const fs = await import("node:fs")
    const pathMod = await import("node:path")
    const source = fs.readFileSync(
      pathMod.resolve(__dirname, "../sidebar.tsx"),
      "utf-8",
    )
    const codeOnly = source
      .replace(/\/\/.*$/gm, "")
      .replace(/\/\*[\s\S]*?\*\//g, "")
    expect(codeOnly.includes("setInterval")).toBe(false)

    // Runtime invariant — no pending timers after the initial paint.
    // A regression that re-introduces `setInterval` would trip this.
    expect(vi.getTimerCount()).toBe(0)
  })

  it("renders the agency nav without crashing", () => {
    const { container } = render(
      withQueryClient(
        <Sidebar collapsed={false} onToggleCollapse={vi.fn()} onClose={vi.fn()} />,
      ),
    )
    // Smoke check that nav links rendered (at least the dashboard).
    expect(container.querySelector("nav")).toBeTruthy()
  })
})

// --- Fix 1 (referrer nav: opportunities) -------------------------------

describe("Sidebar — referrer nav has Opportunités", () => {
  it("includes the opportunities entry in the REFERRER_NAV source", async () => {
    // We assert at the source level so the test does not depend on the
    // full sidebar render path (the referrer mode requires both the
    // user.role=provider mock AND the workspace cookie to flip; both
    // already land on the freelance branch in the existing mocks).
    const fs = await import("node:fs")
    const pathMod = await import("node:path")
    const source = fs.readFileSync(
      pathMod.resolve(__dirname, "../sidebar.tsx"),
      "utf-8",
    )
    // Find the REFERRER_NAV table and slice up to its closing bracket.
    // A regex grabs the contents between `REFERRER_NAV: NavItem[] = [`
    // and the next top-level `]` so we are sure we are inside the
    // referrer table (not the freelance one).
    const match = source.match(
      /const REFERRER_NAV: NavItem\[\] = \[([\s\S]*?)\n\]/,
    )
    expect(match).not.toBeNull()
    const referrerBlock = match![1]
    expect(referrerBlock).toContain('labelKey: "opportunities"')
    expect(referrerBlock).toContain('href: "/opportunities"')
    // Roles must include 'provider' so the entry is visible to the
    // referrer-enabled user (who is always role=provider).
    expect(referrerBlock).toContain('roles: ["provider"]')
  })
})

// --- Opportunities-tabs merge: legacy /my-applications entry removed ---

describe("Sidebar — /my-applications entry removed (merged into /opportunities)", () => {
  it("no longer references the legacy myApplications nav item", async () => {
    // Source-level guard: the dedicated route + entry are gone, so no
    // FREELANCE or REFERRER nav table should ship a myApplications
    // labelKey or `/my-applications` href any more.
    const fs = await import("node:fs")
    const pathMod = await import("node:path")
    const source = fs.readFileSync(
      pathMod.resolve(__dirname, "../sidebar.tsx"),
      "utf-8",
    )
    const codeOnly = source
      .replace(/\/\/.*$/gm, "")
      .replace(/\/\*[\s\S]*?\*\//g, "")
    expect(codeOnly).not.toContain('"myApplications"')
    expect(codeOnly).not.toContain("/my-applications")
  })
})
