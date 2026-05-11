import { describe, it, expect, vi } from "vitest"
import { render, screen } from "@testing-library/react"

import { AnonymizedProviderCard } from "../anonymized-provider-card"
import { AnonymizedClientCard } from "../anonymized-client-card"
import type {
  ProviderSnapshot,
  ClientSnapshot,
} from "@/shared/types/referral"

vi.mock("next-intl", () => ({
  useTranslations:
    (namespace?: string) =>
    (key: string) =>
      namespace ? `${namespace}.${key}` : key,
}))

vi.mock("next/link", () => ({
  default: ({
    children,
    href,
    ...rest
  }: {
    children: React.ReactNode
    href: string
  }) => (
    <a href={href} {...rest}>
      {children}
    </a>
  ),
}))

const providerFixture: ProviderSnapshot = {
  expertise_domains: ["dev"],
  years_experience: 5,
}

const clientFixture: ClientSnapshot = {
  industry: "SaaS",
}

describe("AnonymizedProviderCard — reveal toggle", () => {
  it("hides the reveal link by default (masked behaviour, viewer is not owner)", () => {
    render(<AnonymizedProviderCard snapshot={providerFixture} />)
    expect(
      screen.queryByTestId("anonymized-provider-reveal-link"),
    ).not.toBeInTheDocument()
    // The eyebrow text matches the masked variant.
    expect(
      screen.getByText(/Identité révélée à l'acceptation/),
    ).toBeInTheDocument()
  })

  it("shows the reveal link when revealed AND providerId is set", () => {
    render(
      <AnonymizedProviderCard
        snapshot={providerFixture}
        revealed
        providerId="user-123"
      />,
    )
    const link = screen.getByTestId("anonymized-provider-reveal-link")
    expect(link).toBeInTheDocument()
    expect(link.getAttribute("href")).toBe("/freelances/user-123")
    // Eyebrow flips to the revealed variant.
    expect(
      screen.getByText(/Identité visible \(tu es l'apporteur\)/),
    ).toBeInTheDocument()
  })

  it("does NOT show the reveal link when revealed=true but providerId is missing (defensive)", () => {
    render(<AnonymizedProviderCard snapshot={providerFixture} revealed />)
    expect(
      screen.queryByTestId("anonymized-provider-reveal-link"),
    ).not.toBeInTheDocument()
  })
})

describe("AnonymizedClientCard — reveal toggle", () => {
  it("hides the reveal link by default", () => {
    render(<AnonymizedClientCard snapshot={clientFixture} />)
    expect(
      screen.queryByTestId("anonymized-client-reveal-link"),
    ).not.toBeInTheDocument()
    expect(
      screen.getByText(/Identité révélée à l'acceptation/),
    ).toBeInTheDocument()
  })

  it("shows the reveal link when revealed AND clientId is set", () => {
    render(
      <AnonymizedClientCard
        snapshot={clientFixture}
        revealed
        clientId="org-999"
      />,
    )
    const link = screen.getByTestId("anonymized-client-reveal-link")
    expect(link).toBeInTheDocument()
    expect(link.getAttribute("href")).toBe("/enterprises/org-999")
    expect(
      screen.getByText(/Identité visible \(tu es l'apporteur\)/),
    ).toBeInTheDocument()
  })

  it("does NOT show the reveal link when revealed=true but clientId is missing", () => {
    render(<AnonymizedClientCard snapshot={clientFixture} revealed />)
    expect(
      screen.queryByTestId("anonymized-client-reveal-link"),
    ).not.toBeInTheDocument()
  })
})
