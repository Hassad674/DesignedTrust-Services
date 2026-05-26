import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import { render } from "@testing-library/react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { createElement, type ReactNode } from "react"
import { ReportForm } from "../report-form"

// Regression: the "Signaler" feedback modal crashed the whole page in
// production with "No QueryClient set, use QueryClientProvider to set
// one". Root cause — the ReportButton FAB was mounted as a SIBLING of
// <Providers> (the QueryClientProvider) in the root locale layout, so
// ReportForm's useUser() (a TanStack Query `useQuery`) had no client in
// context. The throw escalated to the html-level global-error.tsx
// boundary ("We hit a snag").
//
// This test pins the contract that motivated the fix: ReportForm is a
// QueryClient consumer (via the REAL useUser, not a mock), therefore
//   1. rendering it WITHOUT a QueryClientProvider must throw, and
//   2. rendering it WITHIN one must NOT throw.
// If a future refactor re-mounts the FAB outside the provider, (1)
// documents exactly why the page crashes; (2) is the green path.

// next-intl + the leaf UI deps are mocked to keep the test focused on
// the provider dependency, not i18n/icons.
vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
  useLocale: () => "fr",
}))

vi.mock("lucide-react", () => ({
  Bug: (p: Record<string, unknown>) => createElement("span", p),
  ShieldAlert: (p: Record<string, unknown>) => createElement("span", p),
  Paperclip: (p: Record<string, unknown>) => createElement("span", p),
  Trash2: (p: Record<string, unknown>) => createElement("span", p),
  Video: (p: Record<string, unknown>) => createElement("span", p),
  Loader2: (p: Record<string, unknown>) => createElement("span", p),
}))

// next/image → plain img so the preview renders in jsdom (it must not be
// the thing that throws; the QueryClient is).
vi.mock("next/image", () => ({
  default: (props: Record<string, unknown>) => createElement("img", props),
}))

// The /auth/me fetch must never hit the network in a unit test — but the
// hook still needs a QueryClient in context to even register the query.
const originalFetch = globalThis.fetch
beforeEach(() => {
  globalThis.fetch = vi.fn(
    async () =>
      new Response(JSON.stringify({ user: null, organization: null }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
  ) as typeof fetch
  // Silence React's expected error logging for the throwing render.
  vi.spyOn(console, "error").mockImplementation(() => {})
})

afterEach(() => {
  globalThis.fetch = originalFetch
  vi.restoreAllMocks()
})

function withQueryClient(children: ReactNode) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0 } },
  })
  return createElement(QueryClientProvider, { client: queryClient }, children)
}

describe("ReportForm — QueryClient dependency (modal-crash regression)", () => {
  it("throws 'No QueryClient set' when rendered outside a QueryClientProvider", () => {
    expect(() => render(<ReportForm onSuccess={vi.fn()} />)).toThrow(
      /No QueryClient set/,
    )
  })

  it("renders without throwing when wrapped in a QueryClientProvider", () => {
    expect(() =>
      render(withQueryClient(<ReportForm onSuccess={vi.fn()} />)),
    ).not.toThrow()
  })
})
