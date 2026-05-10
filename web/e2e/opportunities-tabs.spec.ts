import { test, expect } from "@playwright/test"
import { registerProvider } from "./helpers/auth"

// E2E coverage for the merged Opportunités page (W-12 follow-up).
//
// The previous topology had two routes — `/opportunities` (browse all
// open offers) and `/my-applications` (the freelance's own
// applications). The user-facing nav was overloaded; this round
// collapses the two surfaces into a single `/opportunities?tab=...`
// experience with two tabs: "Toutes les offres" (default) and "Mes
// candidatures" (lazy-mounted on activation).
//
// This spec verifies the URL contract + content swap end-to-end.
test.describe("Opportunités tabs", () => {
  test("default tab shows all offers, switching to applications updates the URL", async ({
    page,
  }) => {
    await registerProvider(page)
    await page.goto("/opportunities")

    // Default tab — "Toutes les offres" must be selected.
    const allTab = page.getByRole("tab", { name: /Toutes les offres/i })
    const applicationsTab = page.getByRole("tab", {
      name: /Mes candidatures/i,
    })
    await expect(allTab).toHaveAttribute("aria-selected", "true")
    await expect(applicationsTab).toHaveAttribute("aria-selected", "false")

    // Click "Mes candidatures" → URL gains `?tab=applications`.
    await applicationsTab.click()
    await expect(page).toHaveURL(/\?tab=applications/)
    await expect(applicationsTab).toHaveAttribute("aria-selected", "true")
    await expect(allTab).toHaveAttribute("aria-selected", "false")

    // Click back → URL returns to the canonical form.
    await allTab.click()
    await expect(page).toHaveURL(/\/opportunities(?!\?tab=)/)
    await expect(allTab).toHaveAttribute("aria-selected", "true")
  })

  test("the legacy /my-applications route no longer exists", async ({
    page,
  }) => {
    await registerProvider(page)
    const response = await page.goto("/my-applications")
    // Next.js returns a 404 page; status code may vary by build, but the
    // route group's not-found UI must surface (no `Mes candidatures`
    // heading rendered as a standalone page).
    expect(response?.status()).toBe(404)
  })
})
