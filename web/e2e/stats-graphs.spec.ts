import { test, expect } from "@playwright/test"
import { registerProvider } from "./helpers/auth"

// stats-graphs.spec.ts — D3 coverage. The user complaint was that the
// graphs lacked precision: no clear distinction between unique vs
// total counts, no 1-year window, and a flimsy empty state. This
// spec exercises the three new contracts end-to-end with stubbed
// stats endpoints so the assertions are deterministic.
//
// The stub returns a 7-bucket series carrying both `count` (total)
// and `unique` so the chart renders the corail solid + dashed pair.

const SERIES = Array.from({ length: 7 }).map((_, i) => ({
  date: `2026-04-${(10 + i).toString().padStart(2, "0")}T00:00:00Z`,
  count: 4 + i * 2,
  unique: 2 + i,
}))

test.describe("/stats — D3 graph improvements", () => {
  test("renders unique + total split, supports 1 an, hides empty card when data", async ({
    page,
  }) => {
    await registerProvider(page)

    await page.route("**/api/v1/me/stats/visibility**", (route) =>
      route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          data: {
            organization_id: "org-e2e",
            period_days: 30,
            total_views: 56,
            unique_viewers: 21,
            search_appearances: 15,
            avg_search_position: 3.4,
            series: SERIES,
          },
        }),
      }),
    )
    await page.route("**/api/v1/me/stats/keywords**", (route) =>
      route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ data: [] }),
      }),
    )

    await page.goto("/fr/stats")
    const strip = page.getByTestId("stats-metric-strip")
    await expect(strip).toBeVisible({ timeout: 15000 })

    // D3: unique + total surfaced as separate cards.
    await expect(strip).toContainText("21")
    await expect(strip).toContainText("56")
    await expect(strip).toContainText(/Visiteurs uniques/i)
    await expect(strip).toContainText(/visiteurs distincts/i)

    // Empty card MUST NOT appear when the org has views.
    await expect(page.getByTestId("stats-empty-card")).toHaveCount(0)

    // 1-year chip is reachable, clicking it updates the URL filter.
    const oneYear = page.getByRole("button", { name: /1 an/ })
    await expect(oneYear).toBeVisible()
    await oneYear.click()
    await expect(page).toHaveURL(/period=365/)

    // Period change rerenders without breaking the strip.
    await expect(strip).toBeVisible()
  })

  test("shows the friendly empty card when total_views is 0", async ({
    page,
  }) => {
    await registerProvider(page)
    await page.route("**/api/v1/me/stats/visibility**", (route) =>
      route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          data: {
            organization_id: "org-e2e",
            period_days: 30,
            total_views: 0,
            unique_viewers: 0,
            search_appearances: 0,
            avg_search_position: null,
            series: [],
          },
        }),
      }),
    )
    await page.route("**/api/v1/me/stats/keywords**", (route) =>
      route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ data: [] }),
      }),
    )

    await page.goto("/fr/stats")
    const empty = page.getByTestId("stats-empty-card")
    await expect(empty).toBeVisible({ timeout: 15000 })
    await expect(empty).toContainText(/Personne n'a encore consulté/)
    await expect(empty).toContainText(/LinkedIn/)
  })
})
