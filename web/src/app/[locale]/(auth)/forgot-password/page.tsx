import { useTranslations } from "next-intl"
import { ForgotPasswordForm } from "@/features/auth/components/forgot-password-form"
import { AuthPageShell } from "@/features/auth/components/auth-page-shell"

// W-03 · Mot de passe oublié · Soleil v2.
// Reuses the AuthPageShell extracted from the login page so the
// split layout (logo + form left, editorial hero right) stays
// pixel-identical across all auth surfaces.
//
// Source: design/assets/sources/phase1/soleil-lotE.jsx (W-01 layout)
// adapted with a recovery-specific eyebrow + H1 (auth.forgotPassword.*).
//
// e2e contract: heading still resolves to "Mot de passe oublié" via
// the eyebrow text — kept stable in `auth.forgotPassword.eyebrow`.
export default function ForgotPasswordPage() {
  const t = useTranslations("auth")

  return (
    <AuthPageShell
      eyebrow={t("forgotShell.eyebrow")}
      titlePrefix={t("forgotShell.titlePrefix")}
      titleAccent={t("forgotShell.titleAccent")}
      subtitle={t("forgotShell.subtitle")}
    >
      <ForgotPasswordForm />
    </AuthPageShell>
  )
}
