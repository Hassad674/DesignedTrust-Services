"use client"

import { useEffect } from "react"
import { useLocale, useTranslations } from "next-intl"
import * as CookieConsent from "vanilla-cookieconsent"

import "vanilla-cookieconsent/dist/cookieconsent.css"
import "@/styles/cookie-consent.css"

import { applyCustomConsent } from "@/shared/lib/posthog-consent"

/**
 * Mounts the vanilla-cookieconsent CMP and gates analytics opt-in on
 * the `analytics` category. Renders nothing — the CMP injects its own
 * modal/banner DOM into `document.body` on `run()`.
 *
 * Why this sits outside any feature folder: every public/auth/dash
 * page must surface the CMP, and the cookie banner straddles
 * analytics + legal concerns that don't belong to a single feature.
 *
 * Why we keep the legacy `applyCustomConsent()` glue: the GA4 provider
 * + the PostHog SDK glue still read the legacy localStorage flag (see
 * `shared/lib/posthog-consent.ts` header). Mirroring the CMP state into
 * that flag avoids touching every analytics consumer in this dispatch.
 */
export function CookieConsentProvider() {
  const t = useTranslations("cookieConsent")
  const locale = useLocale()

  useEffect(() => {
    // Run is idempotent — the library guards against double init via
    // an internal flag. We still wrap in a try/catch to never crash
    // the host app on a third-party hiccup.
    try {
      const localeFooter = buildBannerFooter(t, locale)
      void CookieConsent.run({
        // RGPD-compliant default: nothing tracks before the user
        // makes a choice.
        mode: "opt-in",
        autoShow: true,
        // We surface our own logging via onChange/onFirstConsent —
        // the CMP cookie itself is sufficient persistence.
        revision: 1,
        // Disable the CMP's <script type="text/plain"> auto-runner.
        // Our analytics SDKs are ES module imports gated on the
        // category flag, not inline <script> tags, so this would be
        // a no-op anyway and can be safely turned off.
        manageScriptTags: false,
        guiOptions: {
          // CNIL deliberation 2020-091 + 2024 lignes directrices —
          // the "Refuse all" button MUST carry the exact same visual
          // weight as "Accept all" (size, contrast, font weight).
          // Anything else is the #1 motive of CNIL cookie-banner
          // sanctions (TF1, Decathlon, Carrefour, Yahoo, …).
          // `equalWeightButtons: true` is the structural switch;
          // matching CSS in `web/src/styles/cookie-consent.css`
          // neutralises the residual style asymmetries shipped by
          // vanilla-cookieconsent's default theme.
          consentModal: {
            layout: "box",
            position: "bottom right",
            equalWeightButtons: true,
            flipButtons: false,
          },
          preferencesModal: {
            layout: "box",
            position: "right",
            equalWeightButtons: true,
            flipButtons: false,
          },
        },
        categories: {
          necessary: {
            enabled: true,
            readOnly: true,
          },
          analytics: {
            enabled: false,
            readOnly: false,
            autoClear: {
              cookies: [
                { name: /^_ga/ },
                { name: /^ph_/ },
                { name: "_gid" },
              ],
            },
          },
        },
        language: {
          default: locale,
          translations: {
            fr: buildTranslation(t, localeFooter),
            en: buildTranslation(t, localeFooter),
          },
        },
        onFirstConsent: ({ cookie }) => syncConsentToAnalytics(cookie.categories),
        onChange: ({ cookie }) => syncConsentToAnalytics(cookie.categories),
      })
    } catch {
      // best-effort — never block the app on a CMP boot failure
    }
    // Only re-init when the locale changes; the translations function is
    // recomputed alongside the locale via next-intl, so no need to add
    // `t` to the deps and trigger spurious reboots.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [locale])

  return null
}

/**
 * Translate the CMP modals from the same `cookieConsent.*` i18n
 * namespace used everywhere else in the app. Pulled out so the
 * (heavy) literal table is not inlined in the effect.
 */
function buildTranslation(
  t: (k: string) => string,
  footer: string,
): CookieConsent.Translation {
  return {
    consentModal: {
      title: t("banner.title"),
      description: t("banner.description"),
      acceptAllBtn: t("banner.acceptAll"),
      acceptNecessaryBtn: t("banner.refuseAll"),
      showPreferencesBtn: t("banner.preferences"),
      footer,
    },
    preferencesModal: {
      title: t("preferences.title"),
      acceptAllBtn: t("banner.acceptAll"),
      acceptNecessaryBtn: t("banner.refuseAll"),
      savePreferencesBtn: t("preferences.save"),
      closeIconLabel: t("preferences.close"),
      sections: [
        {
          title: t("preferences.intro.title"),
          description: t("preferences.intro.description"),
        },
        {
          title: t("preferences.necessary.title"),
          description: t("preferences.necessary.description"),
          linkedCategory: "necessary",
        },
        {
          title: t("preferences.analytics.title"),
          description: t("preferences.analytics.description"),
          linkedCategory: "analytics",
        },
      ],
    },
  }
}

/**
 * Build the CMP banner footer as a locale-aware HTML string with the
 * four required legal references (CNIL Recommendation 2020 point 6.3:
 * the consent banner must surface a clear path to the long privacy
 * policy, cookies notice, legal notice, and the sub-processors list).
 *
 * The legacy i18n value used to hardcode `/fr/privacy` + `/fr/cookies`
 * — broken for English visitors and missing two of the four mandatory
 * links. We now build the footer at runtime from locale-prefixed
 * canonical routes, keeping the rendered HTML minimal so
 * vanilla-cookieconsent's default styling stays consistent.
 */
function buildBannerFooter(
  t: (k: string) => string,
  locale: string,
): string {
  const prefix = `/${locale}`
  const links: ReadonlyArray<{ href: string; label: string }> = [
    {
      href: `${prefix}/legal/politique-confidentialite`,
      label: t("banner.linkPrivacy"),
    },
    { href: `${prefix}/cookies`, label: t("banner.linkCookies") },
    { href: `${prefix}/legal`, label: t("banner.linkMentions") },
    {
      href: `${prefix}/sous-processeurs`,
      label: t("banner.linkSubprocessors"),
    },
  ]
  return links
    .map(({ href, label }) => `<a href="${href}">${escapeHtml(label)}</a>`)
    .join(" · ")
}

/**
 * Minimal HTML-attribute escaper for the banner footer. The labels
 * come from our i18n bundle (trusted), but we still escape defensively
 * because the CMP renders this string via `innerHTML`.
 */
function escapeHtml(value: string): string {
  return value
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;")
}

/**
 * Mirror the CMP's chosen categories into the legacy analytics
 * surface (PostHog opt-in flag + GA4 conditional mount + audit log
 * receipt). Keeps a single localStorage flag in sync with the CMP
 * cookie so legacy consumers don't need to learn the CMP API.
 */
function syncConsentToAnalytics(categories: string[]): void {
  const analyticsAccepted = categories.includes("analytics")
  applyCustomConsent(analyticsAccepted, categories)
}
