"use client"

import { useState } from "react"
import { useTranslations } from "next-intl"
import { useQueryClient } from "@tanstack/react-query"
import { useRouter } from "@i18n/navigation"
import { MailCheck } from "lucide-react"

import { Input } from "@/shared/components/ui/input"
import { Button } from "@/shared/components/ui/button"
import { useUser } from "@/shared/hooks/use-user"
import {
  verifyEmail,
  resendVerification,
  VerifyEmailError,
} from "@/features/auth/api/verify-email-api"
import { useResendCooldown } from "@/features/auth/hooks/use-resend-cooldown"

/**
 * VerifyEmailForm — the signup email-verification screen. After
 * /auth/register the account exists but is `email_verified: false` and
 * the backend already emailed a 6-digit code. The user recopies it
 * here; on success the backend re-issues the session (fresh Set-Cookie
 * carrying the verified state), so we bust the ["session"] cache and
 * push to /dashboard.
 *
 * Visual rhythm mirrors VerifyTwoFactorForm (single big mono input,
 * corail CTA, resend link) but adds a 60-second resend cooldown so a
 * user cannot inbox-bomb themselves (the backend rate-limits too — this
 * is just the UI affordance).
 *
 * The email being verified comes from the session (/auth/me) — the user
 * is authenticated (the register cookie is set), just not yet verified.
 */
const RESEND_COOLDOWN_SECONDS = 60

const ERROR_CODE_KEYS: Record<string, string> = {
  no_challenge: "errors.noChallenge",
  challenge_expired: "errors.challengeExpired",
  invalid_code: "errors.invalidCode",
  too_many_attempts: "errors.tooManyAttempts",
  session_invalid: "errors.sessionInvalid",
}

export function VerifyEmailForm() {
  const t = useTranslations("verifyEmail")
  const router = useRouter()
  const queryClient = useQueryClient()
  const { data: user } = useUser()
  const [code, setCode] = useState("")
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const cooldown = useResendCooldown()

  function mapError(err: unknown): string {
    if (err instanceof VerifyEmailError) {
      return t(ERROR_CODE_KEYS[err.code] ?? "errors.generic")
    }
    return t("errors.generic")
  }

  async function goToDashboard() {
    // Bust the stale "unverified" /auth/me verdict so the dashboard
    // refetches against the fresh, verified session cookie.
    await queryClient.invalidateQueries({ queryKey: ["session"] })
    router.push("/dashboard")
  }

  async function onSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    if (submitting) return
    setError(null)
    setNotice(null)
    const trimmed = code.trim()
    if (trimmed.length !== 6) {
      setError(t("errors.codeLength"))
      return
    }
    setSubmitting(true)
    try {
      await verifyEmail(trimmed)
      await goToDashboard()
    } catch (err) {
      setError(mapError(err))
      setCode("")
    } finally {
      setSubmitting(false)
    }
  }

  async function onResend() {
    if (cooldown.active) return
    setError(null)
    setNotice(null)
    try {
      const resp = await resendVerification()
      if (resp.email_verified) {
        // Already-verified no-op — just send the user through.
        setNotice(t("alreadyVerified"))
        await goToDashboard()
        return
      }
      cooldown.start(RESEND_COOLDOWN_SECONDS)
      setNotice(t("resendSuccess"))
    } catch (err) {
      setError(
        err instanceof VerifyEmailError && err.code === "too_many_attempts"
          ? t("errors.tooManyAttempts")
          : t("errors.resendFailed"),
      )
    }
  }

  return (
    <form onSubmit={onSubmit} className="space-y-4" noValidate>
      <div
        className="flex items-center gap-2 rounded-xl border border-border bg-primary-soft/40 px-4 py-3 text-sm text-foreground"
        role="status"
      >
        <MailCheck
          className="h-5 w-5 flex-shrink-0 text-primary"
          aria-hidden="true"
        />
        <span>{t("emailHint", { email: user?.email ?? "" })}</span>
      </div>

      {error && (
        <div
          id="verify-email-error"
          role="alert"
          className="rounded-xl border border-destructive/30 bg-primary-soft/40 p-3 text-sm text-destructive"
        >
          {error}
        </div>
      )}

      {notice && !error && (
        <div
          role="status"
          className="rounded-xl border border-success/30 bg-amber-soft/40 p-3 text-sm text-foreground"
        >
          {notice}
        </div>
      )}

      <div className="space-y-1.5">
        <label
          htmlFor="verify-email-code"
          className="block text-[13px] font-semibold text-foreground"
        >
          {t("codeLabel")}
        </label>
        <Input
          id="verify-email-code"
          name="verify-email-code"
          type="text"
          inputMode="numeric"
          autoComplete="one-time-code"
          autoFocus
          maxLength={6}
          pattern="[0-9]*"
          placeholder={t("codePlaceholder")}
          value={code}
          onChange={(e) => setCode(e.target.value.replace(/[^0-9]/g, ""))}
          disabled={submitting}
          aria-invalid={Boolean(error) || undefined}
          aria-describedby={error ? "verify-email-error" : undefined}
          className={[
            "block w-full rounded-xl border bg-card px-4 py-[13px] text-center font-mono text-[18px] tracking-[0.4em] text-foreground",
            "transition-colors duration-150 placeholder:text-subtle-foreground placeholder:tracking-normal placeholder:font-sans",
            "focus:border-primary focus:outline-none focus:ring-4 focus:ring-primary/15",
            error
              ? "border-destructive focus:ring-destructive/15"
              : "border-border-strong",
          ].join(" ")}
        />
      </div>

      <Button
        variant="primary"
        size="auto"
        type="submit"
        disabled={submitting || code.length !== 6}
        className={[
          "mt-2 w-full rounded-full px-4 py-3.5 text-[14.5px] font-semibold",
          "active:scale-[0.99]",
          "focus:outline-none focus:ring-4 focus:ring-primary/30",
          "disabled:cursor-not-allowed disabled:opacity-60",
        ].join(" ")}
      >
        {submitting ? t("verifying") : t("verifyCta")}
      </Button>

      <div className="text-center text-[13px] text-muted-foreground">
        {t("resendQuestion")}{" "}
        <Button
          variant="ghost"
          size="auto"
          type="button"
          onClick={onResend}
          disabled={cooldown.active || submitting}
          className="h-auto bg-transparent p-0 font-semibold text-[var(--text-link)] underline-offset-4 transition-colors hover:bg-transparent hover:text-primary hover:underline disabled:cursor-not-allowed disabled:opacity-60"
        >
          {cooldown.active
            ? t("resendIn", { seconds: cooldown.remaining })
            : t("resend")}
        </Button>
      </div>
    </form>
  )
}
