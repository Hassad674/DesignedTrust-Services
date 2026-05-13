import { defineRouting } from 'next-intl/routing'

/**
 * Locale-aware URL segments for the legal / compliance corpus.
 *
 * The KEY is the canonical (FR) slug used as the on-disk route under
 * `app/[locale]/(public)/...`. The VALUE map gives the user-facing
 * segment that should appear in the URL for each supported locale.
 *
 * Why not feed this into `defineRouting({ pathnames })`:
 *   next-intl's typed `Link` becomes strictly bound to `keyof pathnames`
 *   when the option is set — every other `<Link href="/...">` in the
 *   codebase would error out unless every single route is enumerated
 *   here. The legal corpus is the only family that needs locale-aware
 *   segments today, so we keep this table purely declarative and apply
 *   it via:
 *     1. `next.config.ts` `rewrites()` — the EN-named URL renders the
 *        same on-disk page as the FR slug (e.g. `/en/legal/terms`
 *        serves the page from `app/[locale]/(public)/legal/cgu`);
 *     2. helpers (`legalHref()`) used by the legal-link surfaces
 *        (footer, dashboard legal bar, cookie consent banner) so links
 *        rendered on the EN locale point at the EN-named URL.
 *
 * If you add a new legal route, update this table AND drop the new
 * file under `app/[locale]/(public)/...` using the FR slug, then add
 * the corresponding rewrite in `next.config.ts`.
 */
export const legalPathnames: Record<string, Record<'en' | 'fr', string>> = {
  '/legal/cgu': { fr: '/legal/cgu', en: '/legal/terms' },
  '/legal/cgv': { fr: '/legal/cgv', en: '/legal/sales-terms' },
  '/legal/politique-confidentialite': {
    fr: '/legal/politique-confidentialite',
    en: '/legal/privacy',
  },
  '/sous-processeurs': { fr: '/sous-processeurs', en: '/subprocessors' },
  '/legal/registre': { fr: '/legal/registre', en: '/legal/processing-register' },
  '/legal/aipd': { fr: '/legal/aipd', en: '/legal/dpia' },
  '/legal/code-de-conduite': {
    fr: '/legal/code-de-conduite',
    en: '/legal/code-of-conduct',
  },
  '/decisions-automatisees': {
    fr: '/decisions-automatisees',
    en: '/automated-decisions',
  },
}

export const routing = defineRouting({
  locales: ['en', 'fr'],
  defaultLocale: 'en',
  localePrefix: 'as-needed',
})

/**
 * Resolve a canonical legal route (FR slug) to its locale-aware URL,
 * including the `/<locale>` prefix when needed. Mirrors next-intl's
 * `Link` behaviour but produces a plain string suitable for surfaces
 * that cannot mount a React component (e.g. vanilla-cookieconsent's
 * `footer` field, raw `<a href>` markup in MDX).
 *
 * Tiny + pure so unit tests can pin "this canonical path on this
 * locale resolves to this URL" without bootstrapping next-intl.
 */
export function legalHref(canonical: string, locale: string): string {
  const entry = legalPathnames[canonical]
  const localized = entry?.[locale as 'en' | 'fr'] ?? canonical
  // `localePrefix: 'as-needed'` — the default locale renders without a
  // prefix; every other locale takes its `/<locale>` prefix.
  if (locale === routing.defaultLocale) return localized
  return `/${locale}${localized}`
}
