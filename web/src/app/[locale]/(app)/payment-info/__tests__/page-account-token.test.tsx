import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import { render, screen, fireEvent, waitFor } from "@testing-library/react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import type { ReactElement } from "react"

// The page reads NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY into a module-level
// const at import time. vi.hoisted runs before the (hoisted) import below,
// so the key is present when the page module evaluates — otherwise
// initializeConnect() short-circuits and never wires fetchClientSecret.
vi.hoisted(() => {
  process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY = "pk_test_page_wiring"
})

import PaymentInfoV2Page from "../page"

// next/navigation — no mobile token, default desktop flow.
const mockSearchParams = new Map<string, string>()
vi.mock("next/navigation", () => ({
  useParams: () => ({ locale: "fr" }),
  useSearchParams: () => ({
    get: (key: string) => mockSearchParams.get(key) ?? null,
  }),
}))

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}))

// Capture the fetchClientSecret callback handed to Stripe Connect so the
// test can invoke it directly and inspect the account-session POST body.
let capturedFetchClientSecret: (() => Promise<string>) | null = null
vi.mock("@stripe/connect-js", () => ({
  loadConnectAndInitialize: vi.fn(
    (params: { fetchClientSecret: () => Promise<string> }) => {
      capturedFetchClientSecret = params.fetchClientSecret
      return {}
    },
  ),
}))

vi.mock("@stripe/react-connect-js", () => ({
  ConnectAccountManagement: () => null,
  ConnectAccountOnboarding: () => null,
  ConnectComponentsProvider: ({ children }: { children: React.ReactNode }) => (
    <>{children}</>
  ),
  ConnectNotificationBanner: () => null,
}))

// Control the client-side account-token generation per test.
const createAccountTokenSafeMock = vi.fn()
vi.mock("../lib/account-token", () => ({
  createAccountTokenSafe: (key: string) => createAccountTokenSafeMock(key),
}))

// Permission granted so the wizard renders.
vi.mock("@/shared/hooks/use-permissions", () => ({
  usePermissionStatus: () => ({
    status: "granted",
    granted: true,
    isLoading: false,
    isError: false,
  }),
}))

// The onboarding wizard exposes a submit button we can click to drive the
// flow that initialises Stripe Connect (and thus captures fetchClientSecret).
vi.mock("../components/onboarding-wizard", () => ({
  OnboardingWizard: ({ onSubmit }: { onSubmit: (country: string) => void }) => (
    <button type="button" onClick={() => onSubmit("FR")}>
      wizard-submit
    </button>
  ),
}))

vi.mock("../components/account-status-card", () => ({
  AccountStatusCard: () => null,
}))

function renderPage(): ReactElement {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0 } },
  })
  return (
    <QueryClientProvider client={client}>
      <PaymentInfoV2Page />
    </QueryClientProvider>
  )
}

async function driveToFetchClientSecret(): Promise<() => Promise<string>> {
  render(renderPage())
  // Initial load: account-status 404 → wizard mode.
  const submit = await screen.findByText("wizard-submit")
  fireEvent.click(submit)
  await waitFor(() => {
    expect(capturedFetchClientSecret).not.toBeNull()
  })
  // Non-null asserted above.
  return capturedFetchClientSecret as () => Promise<string>
}

describe("PaymentInfoV2Page — account-session account_token wiring", () => {
  beforeEach(() => {
    mockSearchParams.clear()
    capturedFetchClientSecret = null
    createAccountTokenSafeMock.mockReset()

    // Default network: account-status 404 (no account), account-session OK.
    const fetchMock = vi.fn((url: string | URL, init?: RequestInit) => {
      const href = String(url)
      if (href.includes("/account-status")) {
        return Promise.resolve(
          new Response(null, { status: 404 }),
        ) as Promise<Response>
      }
      if (href.includes("/account-session") && init?.method === "POST") {
        return Promise.resolve(
          new Response(
            JSON.stringify({
              client_secret: "cs_live_secret",
              account_id: "acct_1",
              expires_at: 9999,
            }),
            { status: 200, headers: { "Content-Type": "application/json" } },
          ),
        ) as Promise<Response>
      }
      return Promise.resolve(new Response(null, { status: 204 })) as Promise<Response>
    })
    vi.stubGlobal("fetch", fetchMock)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it("includes account_token in the POST body when token generation succeeds", async () => {
    createAccountTokenSafeMock.mockResolvedValue("ct_test_success")

    const fetchClientSecret = await driveToFetchClientSecret()
    const secret = await fetchClientSecret()

    expect(secret).toBe("cs_live_secret")
    expect(createAccountTokenSafeMock).toHaveBeenCalled()

    const fetchMock = vi.mocked(globalThis.fetch)
    const postCall = fetchMock.mock.calls.find(
      ([url, init]) =>
        String(url).includes("/account-session") &&
        (init as RequestInit | undefined)?.method === "POST",
    )
    expect(postCall).toBeDefined()
    const body = JSON.parse(
      ((postCall?.[1] as RequestInit).body as string) ?? "{}",
    )
    expect(body).toMatchObject({ country: "FR", account_token: "ct_test_success" })
  })

  it("omits account_token but still posts when token generation returns null", async () => {
    createAccountTokenSafeMock.mockResolvedValue(null)

    const fetchClientSecret = await driveToFetchClientSecret()
    const secret = await fetchClientSecret()

    // The session call still succeeds — token failure must not block KYC.
    expect(secret).toBe("cs_live_secret")

    const fetchMock = vi.mocked(globalThis.fetch)
    const postCall = fetchMock.mock.calls.find(
      ([url, init]) =>
        String(url).includes("/account-session") &&
        (init as RequestInit | undefined)?.method === "POST",
    )
    expect(postCall).toBeDefined()
    const body = JSON.parse(
      ((postCall?.[1] as RequestInit).body as string) ?? "{}",
    )
    expect(body).toMatchObject({ country: "FR" })
    expect(body).not.toHaveProperty("account_token")
  })
})
