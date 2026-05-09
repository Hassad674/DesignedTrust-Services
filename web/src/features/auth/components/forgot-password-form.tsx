"use client"

import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { useState } from "react"
import { Link } from "@i18n/navigation"
import { CheckCircle2 } from "lucide-react"
import { useTranslations } from "next-intl"
import { forgotPassword } from "@/features/auth/api/auth-api"
import { Button } from "@/shared/components/ui/button"
import { Input } from "@/shared/components/ui/input"

// Schema kept inline (extracting to features/auth/schemas/ is a
// separate refactor under OFF-LIMITS for this UI batch). Validation
// rule and the literal "Invalid email address" string are unchanged
// — the e2e test (`web/e2e/auth.spec.ts`) pins the resolved French
// translation "Adresse email invalide" and stays green.
const forgotPasswordSchema = z.object({
  email: z.string().email("Invalid email address"),
})

type ForgotPasswordValues = z.infer<typeof forgotPasswordSchema>

export function ForgotPasswordForm() {
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState(false)
  const t = useTranslations("auth")
  const tCommon = useTranslations("common")

  const {
    register: registerField,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ForgotPasswordValues>({
    resolver: zodResolver(forgotPasswordSchema),
  })

  async function onSubmit(values: ForgotPasswordValues) {
    setError(null)
    try {
      await forgotPassword(values.email)
      setSuccess(true)
    } catch (err) {
      setError(err instanceof Error ? err.message : tCommon("errorOccurred"))
    }
  }

  if (success) {
    return (
      <div
        className="space-y-4 text-center"
        role="status"
        aria-live="polite"
      >
        <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-full bg-success-soft">
          <CheckCircle2 className="h-7 w-7 text-success" aria-hidden="true" />
        </div>
        <h2 className="font-serif text-[22px] font-medium text-foreground">
          {tCommon("emailSent")}
        </h2>
        <p className="text-sm leading-relaxed text-muted-foreground">
          {t("resetEmailSent")}
        </p>
        <Link
          href="/login"
          className="inline-block text-[13px] font-semibold text-primary transition-colors hover:text-primary-deep"
        >
          {t("backToLogin")}
          <span aria-hidden="true"> →</span>
        </Link>
      </div>
    )
  }

  return (
    <>
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        {error && (
          <div
            className="rounded-xl border border-destructive/30 bg-primary-soft/40 p-3 text-sm text-destructive"
            role="alert"
          >
            {error}
          </div>
        )}

        {/* Email */}
        <div className="space-y-1.5">
          <label
            htmlFor="email"
            className="block text-[13px] font-semibold text-foreground"
          >
            {t("email")}
          </label>
          <Input
            id="email"
            type="email"
            autoComplete="email"
            placeholder={t("emailPlaceholder")}
            aria-invalid={errors.email ? true : undefined}
            aria-describedby={errors.email ? "email-error" : undefined}
            className={[
              "block w-full rounded-xl border bg-card px-4 py-[13px] text-[14.5px] text-foreground",
              "transition-colors duration-150 placeholder:text-subtle-foreground",
              "focus:border-primary focus:outline-none focus:ring-4 focus:ring-primary/15",
              errors.email
                ? "border-destructive focus:ring-destructive/15"
                : "border-border-strong",
            ].join(" ")}
            {...registerField("email")}
          />
          {errors.email?.message && (
            <p id="email-error" className="text-xs text-destructive">
              {errors.email.message}
            </p>
          )}
        </div>

        {/* Submit */}
        <Button
          variant="primary"
          size="auto"
          type="submit"
          disabled={isSubmitting}
          className={[
            "mt-2 w-full rounded-full px-4 py-3.5 text-[14.5px] font-semibold",
            "active:scale-[0.99]",
            "focus:outline-none focus:ring-4 focus:ring-primary/30",
            "disabled:cursor-not-allowed disabled:opacity-60",
          ].join(" ")}
          style={{ boxShadow: "0 4px 14px rgba(232, 93, 74, 0.3)" }}
        >
          {isSubmitting ? t("sending") : t("sendResetLink")}
        </Button>
      </form>

      {/* Back to sign in */}
      <p className="mt-7 text-center text-[13px] text-muted-foreground">
        <Link
          href="/login"
          className="font-semibold text-primary transition-colors hover:text-primary-deep"
        >
          {t("backToLogin")}
          <span aria-hidden="true"> →</span>
        </Link>
      </p>
    </>
  )
}
