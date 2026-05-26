import type { Metadata } from "next"
import { Fraunces, Inter_Tight, Geist_Mono } from "next/font/google"
import { hasLocale } from "next-intl"
import { NextIntlClientProvider } from "next-intl"
import { getMessages, getTranslations } from "next-intl/server"
import { notFound } from "next/navigation"
import { routing } from "@i18n/routing"
import { Toaster } from "sonner"
import { ReportButton } from "@/features/feedback/components/report-button"
import { Providers } from "./providers"

// Soleil v2 typography stack:
//   • Fraunces — display, editorial accents, italic-quoted citations
//   • Inter Tight — UI, body, labels (the workhorse)
//   • Geist Mono — numbers, IDs, mono labels
const fraunces = Fraunces({
  subsets: ["latin"],
  variable: "--font-fraunces",
  display: "swap",
  axes: ["opsz"],
})
const interTight = Inter_Tight({
  subsets: ["latin"],
  variable: "--font-inter-tight",
  display: "swap",
})
const geistMono = Geist_Mono({
  subsets: ["latin"],
  variable: "--font-geist-mono",
  display: "swap",
})

export async function generateMetadata({
  params,
}: {
  params: Promise<{ locale: string }>
}): Promise<Metadata> {
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: "metadata" })

  return {
    title: t("title"),
    description: t("description"),
    openGraph: {
      type: "website",
      siteName: "DesignedTrust Services",
      title: t("title"),
      description: t("description"),
      locale: locale === "fr" ? "fr_FR" : "en_US",
      alternateLocale: locale === "fr" ? "en_US" : "fr_FR",
    },
    alternates: {
      languages: {
        en: "/en",
        fr: "/fr",
      },
    },
  }
}

export function generateStaticParams() {
  return routing.locales.map((locale) => ({ locale }))
}

export default async function LocaleLayout({
  children,
  params,
}: {
  children: React.ReactNode
  params: Promise<{ locale: string }>
}) {
  const { locale } = await params

  if (!hasLocale(routing.locales, locale)) {
    notFound()
  }

  const messages = await getMessages()

  return (
    <html lang={locale} className={`${fraunces.variable} ${interTight.variable} ${geistMono.variable}`}>
      <body className="font-sans antialiased">
        <NextIntlClientProvider messages={messages}>
          <Providers>
            {children}
            {/* Always-visible "Signaler" shortcut on every page (public +
             * app + auth). Anchored bottom-LEFT so it never collides with
             * the messaging ChatWidget (bottom-right, desktop-only).
             *
             * MUST live inside <Providers> (the QueryClientProvider): the
             * report form reads the session via useUser() (TanStack Query)
             * to gate the logged-in attachments zone. Mounted as a sibling
             * of <Providers> it had no QueryClient in context, so opening
             * the modal threw "No QueryClient set" and tripped the global
             * error boundary ("We hit a snag") in production. */}
            <ReportButton />
          </Providers>
          <Toaster position="top-right" richColors closeButton duration={3000} />
        </NextIntlClientProvider>
      </body>
    </html>
  )
}
