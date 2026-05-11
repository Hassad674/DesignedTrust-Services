/**
 * D4 (GDPR Phase C) — metadata regression for the 6 new /legal/* docs.
 *
 * Each page exposes generateMetadata that:
 *   1. Pulls title + subtitle from the corresponding legal.docs.<key>
 *      namespace.
 *   2. Sets robots noindex/nofollow — legal pages are not SEO targets.
 *   3. Suffixes the title with " | Marketplace Service".
 *
 * The pages are imported lazily via dynamic import inside each test to
 * avoid hoisting issues with the next-intl/server mock.
 */

import { describe, it, expect, vi } from "vitest"

vi.mock("next-intl/server", () => ({
  getTranslations: async ({ namespace }: { namespace: string }) => {
    return (key: string) => `[${namespace}.${key}]`
  },
}))

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}))

vi.mock("@i18n/navigation", () => ({
  Link: ({ children }: { children: React.ReactNode }) => children,
}))

const CASES: ReadonlyArray<{ path: string; namespace: string }> = [
  {
    path: "@/app/[locale]/(public)/legal/registre/page",
    namespace: "legal.docs.registre",
  },
  {
    path: "@/app/[locale]/(public)/legal/aipd/page",
    namespace: "legal.docs.aipd",
  },
  {
    path: "@/app/[locale]/(public)/legal/dpa-template/page",
    namespace: "legal.docs.dpaTemplate",
  },
  {
    path: "@/app/[locale]/(public)/legal/politique-confidentialite/page",
    namespace: "legal.docs.politiqueConfidentialite",
  },
  {
    path: "@/app/[locale]/(public)/legal/cgu/page",
    namespace: "legal.docs.cgu",
  },
  {
    path: "@/app/[locale]/(public)/legal/cgv/page",
    namespace: "legal.docs.cgv",
  },
]

describe("/legal/* doc pages metadata (D4)", () => {
  for (const { path, namespace } of CASES) {
    it(`${namespace}: generateMetadata sets noindex + localized title`, async () => {
      const mod = await import(path)
      const generate = (mod as { generateMetadata?: unknown })
        .generateMetadata
      expect(typeof generate).toBe("function")
      const meta = await (
        generate as (args: {
          params: Promise<{ locale: string }>
        }) => Promise<Record<string, unknown>>
      )({ params: Promise.resolve({ locale: "fr" }) })

      expect(meta.title).toBe(`[${namespace}.title] | Marketplace Service`)
      expect(meta.description).toBe(`[${namespace}.subtitle]`)
      expect(meta.robots).toEqual({ index: false, follow: false })
    })
  }
})
