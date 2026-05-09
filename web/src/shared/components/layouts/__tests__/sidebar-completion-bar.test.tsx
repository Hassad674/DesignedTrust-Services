/**
 * Profile completion bar in the sidebar (2026-05-09).
 *
 * The sidebar must:
 *   1. Render the bar with persona=undefined (default freelance/agency
 *      checklist) when the workspace cookie is in freelance mode.
 *   2. Render the bar with persona="referrer" when the workspace
 *      cookie flips to referrer mode — surfaces the apporteur
 *      checklist instead of the freelance one.
 *   3. NOT render the bar at all when the user role is enterprise —
 *      enterprise users have a 4-section checklist that is not
 *      actionable from the sidebar nudge.
 */

import { describe, it, expect, vi, beforeEach } from "vitest"
import { render } from "@testing-library/react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { createElement, type ReactNode } from "react"
import type {
  CurrentOrganization,
  CurrentUser,
} from "@/shared/hooks/use-user"

// Capture the ProfileCompletionBar props so tests can assert what
// the sidebar passes (variant, persona, hideWhenComplete).
const completionBarSpy = vi.fn<(props: unknown) => null>(() => null)

vi.mock("@/shared/hooks/use-user", () => ({
  useUser: vi.fn(),
  useOrganization: vi.fn(),
  useLogout: () => vi.fn(),
}))

vi.mock("@/shared/hooks/use-workspace", () => ({
  useWorkspace: vi.fn(),
}))

vi.mock("@/shared/hooks/use-unread-count", () => ({
  useUnreadCount: () => ({ data: { count: 0 } }),
  unreadCountQueryKey: () => ["messaging", "unread-count"],
}))

vi.mock("@/features/profile-completion/components/profile-completion-bar", () => ({
  ProfileCompletionBar: (props: unknown) => completionBarSpy(props),
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
import { useUser, useOrganization } from "@/shared/hooks/use-user"
import { useWorkspace } from "@/shared/hooks/use-workspace"

const mockedUseUser = vi.mocked(useUser)
const mockedUseOrganization = vi.mocked(useOrganization)
const mockedUseWorkspace = vi.mocked(useWorkspace)

function withQueryClient(ui: ReactNode) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return createElement(QueryClientProvider, { client }, ui)
}

// Cast a partial fixture into the typed return value of useUser /
// useOrganization. The sidebar only reads `data.id`, `data.role`,
// `data.display_name` and `data.type`, so the partial object is
// enough at runtime — we just need the type cast to compile.
function makeUserResult(id: string, role: CurrentUser["role"]) {
  return {
    data: { id, role, display_name: role },
  } as unknown as ReturnType<typeof useUser>
}

function makeOrgResult(id: string, type: string) {
  return {
    data: { id, type },
  } as unknown as ReturnType<typeof useOrganization>
}

function setupWorkspace(isReferrerMode: boolean) {
  mockedUseWorkspace.mockReturnValue({
    isReferrerMode,
    setReferrerMode: vi.fn(),
    toggleMode: vi.fn(),
    switchToReferrer: vi.fn(() => "/dashboard"),
    switchToFreelance: vi.fn(() => "/dashboard"),
  })
}

beforeEach(() => {
  completionBarSpy.mockClear()
})

describe("Sidebar — profile completion bar wiring", () => {
  it("mounts the bar with no persona override for an agency user", () => {
    mockedUseUser.mockReturnValue(makeUserResult("u-1", "agency"))
    mockedUseOrganization.mockReturnValue(makeOrgResult("o-1", "agency"))
    setupWorkspace(false)

    render(
      withQueryClient(
        <Sidebar collapsed={false} onToggleCollapse={vi.fn()} onClose={vi.fn()} />,
      ),
    )

    expect(completionBarSpy).toHaveBeenCalled()
    const props = completionBarSpy.mock.calls[0]![0] as {
      variant: string
      persona: string | undefined
      hideWhenComplete?: boolean
    }
    expect(props.variant).toBe("sidebar")
    expect(props.persona).toBeUndefined()
    expect(props.hideWhenComplete).toBe(true)
  })

  it("mounts the bar with persona=referrer when in referrer mode", () => {
    mockedUseUser.mockReturnValue(makeUserResult("u-2", "provider"))
    mockedUseOrganization.mockReturnValue(makeOrgResult("o-2", "provider_personal"))
    setupWorkspace(true)

    render(
      withQueryClient(
        <Sidebar collapsed={false} onToggleCollapse={vi.fn()} onClose={vi.fn()} />,
      ),
    )

    expect(completionBarSpy).toHaveBeenCalled()
    const props = completionBarSpy.mock.calls[0]![0] as { persona: string }
    expect(props.persona).toBe("referrer")
  })

  it("mounts the bar with no persona override when in freelance mode", () => {
    mockedUseUser.mockReturnValue(makeUserResult("u-3", "provider"))
    mockedUseOrganization.mockReturnValue(makeOrgResult("o-3", "provider_personal"))
    setupWorkspace(false)

    render(
      withQueryClient(
        <Sidebar collapsed={false} onToggleCollapse={vi.fn()} onClose={vi.fn()} />,
      ),
    )

    expect(completionBarSpy).toHaveBeenCalled()
    const props = completionBarSpy.mock.calls[0]![0] as {
      persona: string | undefined
    }
    expect(props.persona).toBeUndefined()
  })

  it("does NOT mount the bar at all for an enterprise user", () => {
    mockedUseUser.mockReturnValue(makeUserResult("u-4", "enterprise"))
    mockedUseOrganization.mockReturnValue(makeOrgResult("o-4", "enterprise"))
    setupWorkspace(false)

    render(
      withQueryClient(
        <Sidebar collapsed={false} onToggleCollapse={vi.fn()} onClose={vi.fn()} />,
      ),
    )

    expect(completionBarSpy).not.toHaveBeenCalled()
  })
})

// Avoid an unused-import warning when CurrentOrganization is referenced
// only by the cast utility above.
type _Unused = CurrentOrganization
