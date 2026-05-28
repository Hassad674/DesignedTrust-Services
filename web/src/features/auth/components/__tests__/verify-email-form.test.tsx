import { describe, it, expect, vi, beforeEach } from "vitest"
import { render, screen, waitFor } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { NextIntlClientProvider } from "next-intl"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import messages from "@/../messages/en.json"

// Mock next-intl navigation (useRouter.push) so we can assert the
// post-verify redirect without a real router.
const mockPush = vi.fn()
vi.mock("@i18n/navigation", () => ({
  useRouter: () => ({
    push: mockPush,
    replace: vi.fn(),
    back: vi.fn(),
    prefetch: vi.fn(),
  }),
}))

// The verify-email screen reads the email being verified from the
// session (/auth/me). Mock the user hook to a fixed unverified account.
vi.mock("@/shared/hooks/use-user", () => ({
  useUser: () => ({ data: { email: "new@user.com", email_verified: false } }),
}))

// Mock the API surface — real VerifyEmailError is kept so the form's
// instanceof branch still works.
vi.mock("@/features/auth/api/verify-email-api", async (importOriginal) => {
  const actual =
    await importOriginal<
      typeof import("@/features/auth/api/verify-email-api")
    >()
  return {
    ...actual,
    verifyEmail: vi.fn(),
    resendVerification: vi.fn(),
  }
})

import {
  verifyEmail,
  resendVerification,
  VerifyEmailError,
} from "@/features/auth/api/verify-email-api"
import { VerifyEmailForm } from "../verify-email-form"

const mockVerify = vi.mocked(verifyEmail)
const mockResend = vi.mocked(resendVerification)

let invalidateSpy: ReturnType<typeof vi.fn>

function renderForm() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  invalidateSpy = vi.fn(() => Promise.resolve())
  // Spy on the instance method the form calls after a successful verify.
  queryClient.invalidateQueries = invalidateSpy as never
  return render(
    <QueryClientProvider client={queryClient}>
      <NextIntlClientProvider locale="en" messages={messages}>
        <VerifyEmailForm />
      </NextIntlClientProvider>
    </QueryClientProvider>,
  )
}

beforeEach(() => {
  vi.clearAllMocks()
})

describe("VerifyEmailForm", () => {
  it("shows which email is being verified", () => {
    renderForm()
    expect(screen.getByText(/new@user.com/)).toBeInTheDocument()
  })

  it("submits the code, busts the session cache, then redirects to the dashboard", async () => {
    mockVerify.mockResolvedValueOnce(undefined)
    const user = userEvent.setup()
    renderForm()

    await user.type(
      screen.getByLabelText(messages.verifyEmail.codeLabel),
      "654321",
    )
    await user.click(
      screen.getByRole("button", { name: messages.verifyEmail.verifyCta }),
    )

    await waitFor(() => {
      expect(mockVerify).toHaveBeenCalledWith("654321")
    })
    await waitFor(() => {
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ["session"] })
    })
    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith("/dashboard")
    })
  })

  it("keeps the submit button disabled until 6 digits are entered", async () => {
    const user = userEvent.setup()
    renderForm()
    const button = screen.getByRole("button", {
      name: messages.verifyEmail.verifyCta,
    })
    expect(button).toBeDisabled()
    await user.type(
      screen.getByLabelText(messages.verifyEmail.codeLabel),
      "123456",
    )
    expect(button).toBeEnabled()
  })

  it("strips non-digits from the code input", async () => {
    const user = userEvent.setup()
    renderForm()
    const input = screen.getByLabelText(
      messages.verifyEmail.codeLabel,
    ) as HTMLInputElement
    await user.type(input, "12ab34")
    expect(input.value).toBe("1234")
  })

  it("maps an invalid_code error to the localized message and clears the input", async () => {
    mockVerify.mockRejectedValueOnce(
      new VerifyEmailError("invalid_code", "nope"),
    )
    const user = userEvent.setup()
    renderForm()

    const input = screen.getByLabelText(
      messages.verifyEmail.codeLabel,
    ) as HTMLInputElement
    await user.type(input, "000000")
    await user.click(
      screen.getByRole("button", { name: messages.verifyEmail.verifyCta }),
    )

    expect(
      await screen.findByText(messages.verifyEmail.errors.invalidCode),
    ).toBeInTheDocument()
    expect(input.value).toBe("")
    expect(mockPush).not.toHaveBeenCalled()
  })

  it("resends a code and starts the cooldown countdown", async () => {
    mockResend.mockResolvedValueOnce({ email_verified: false, message: "sent" })
    const user = userEvent.setup()
    renderForm()

    await user.click(
      screen.getByRole("button", { name: messages.verifyEmail.resend }),
    )

    await waitFor(() => {
      expect(mockResend).toHaveBeenCalledTimes(1)
    })
    expect(
      await screen.findByText(messages.verifyEmail.resendSuccess),
    ).toBeInTheDocument()
    // The resend button flips to the cooldown label (and is disabled).
    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: /Resend code \(\d+s\)/ }),
      ).toBeDisabled()
    })
  })

  it("redirects when resend reports the account is already verified", async () => {
    mockResend.mockResolvedValueOnce({ email_verified: true, message: "already" })
    const user = userEvent.setup()
    renderForm()

    await user.click(
      screen.getByRole("button", { name: messages.verifyEmail.resend }),
    )

    await waitFor(() => {
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ["session"] })
    })
    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith("/dashboard")
    })
  })

  it("shows a retry message when resend fails", async () => {
    mockResend.mockRejectedValueOnce(
      new VerifyEmailError("internal_error", "boom"),
    )
    const user = userEvent.setup()
    renderForm()

    await user.click(
      screen.getByRole("button", { name: messages.verifyEmail.resend }),
    )

    expect(
      await screen.findByText(messages.verifyEmail.errors.resendFailed),
    ).toBeInTheDocument()
  })
})
