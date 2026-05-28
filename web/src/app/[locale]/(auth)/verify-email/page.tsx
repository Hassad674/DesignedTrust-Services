import { useTranslations } from "next-intl"
import { VerifyEmailForm } from "@/features/auth/components/verify-email-form"
import { AuthPageShell } from "@/features/auth/components/auth-page-shell"

// Signup email-verification screen. The user lands here straight after
// /register (the account exists but is email_verified=false and a
// 6-digit code was emailed) OR is funneled here by the api-client when a
// gated route answers 403 email_not_verified. Reuses the shared Soleil
// v2 split layout so the funnel stays inside the editorial auth chrome.
export default function VerifyEmailPage() {
  const t = useTranslations("verifyEmail")

  return (
    <AuthPageShell
      eyebrow={t("eyebrow")}
      titlePrefix={t("titlePrefix")}
      titleAccent={t("titleAccent")}
      subtitle={t("subtitle")}
    >
      <VerifyEmailForm />
    </AuthPageShell>
  )
}
