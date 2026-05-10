/**
 * server-fetchers.ts — server-only SEO fetchers used by the public
 * profile and listing pages.
 *
 * These fetchers do NOT forward cookies — every endpoint they hit is
 * publicly readable. They return null on any error so a transient
 * backend hiccup never tanks the route render or the crawl budget.
 *
 * ISR is enabled via `next.revalidate` so re-rendering the page on a
 * subsequent request is cheap (cached HTML). The 120s window is the
 * same one used by the existing profile-metadata fetchers.
 */

import { API_BASE_URL } from "@/shared/lib/api-client"
import type { Review, AverageRating } from "@/shared/types/review"

const REVALIDATE_SECONDS = 120

function apiBase(): string {
  return API_BASE_URL || "http://localhost:8080"
}

/**
 * fetchPublicReviews fetches the most recent N reviews for an
 * organization. Reviews come back already filtered to the published
 * subset (the backend does not return pending double-blind reviews).
 */
export async function fetchPublicReviews(
  orgId: string,
  limit = 5,
): Promise<Review[] | null> {
  try {
    const url = `${apiBase()}/api/v1/reviews/org/${orgId}?limit=${limit}`
    const res = await fetch(url, { next: { revalidate: REVALIDATE_SECONDS } })
    if (!res.ok) return null
    const json = (await res.json()) as { data?: Review[] }
    return Array.isArray(json.data) ? json.data : null
  } catch {
    return null
  }
}

/**
 * fetchPublicAverageRating returns the org's published-review
 * aggregate. Returns null on any failure path so the caller can omit
 * the rating from JSON-LD without crashing.
 */
export async function fetchPublicAverageRating(
  orgId: string,
): Promise<AverageRating | null> {
  try {
    const url = `${apiBase()}/api/v1/reviews/average/${orgId}`
    const res = await fetch(url, { next: { revalidate: REVALIDATE_SECONDS } })
    if (!res.ok) return null
    const json = (await res.json()) as { data?: AverageRating }
    if (!json.data) return null
    return json.data
  } catch {
    return null
  }
}
