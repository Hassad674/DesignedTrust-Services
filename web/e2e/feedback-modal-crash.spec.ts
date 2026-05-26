import { test, expect, type Page } from "@playwright/test"

// Regression for the "Signaler" feedback modal crash.
//
// BUG: the floating bottom-left "Signaler" FAB lives in the root locale
// layout. It was mounted as a SIBLING of <Providers> (the
// QueryClientProvider), so opening the modal — whose ReportForm reads
// the session via useUser() (TanStack Query) — threw "No QueryClient
// set, use QueryClientProvider to set one". Because the FAB renders
// directly in the root layout (above every per-segment error.tsx), the
// throw escalated to the html-level global-error.tsx boundary and the
// WHOLE page was replaced by "We hit a snag".
//
// The throw is wrapped in production-only code paths in some libs, so
// the React dev overlay can mask it — this spec runs against a
// PRODUCTION build (next build && next start) to pin the real failure.
//
// We intercept GET /api/v1/auth/me so the page believes a real user is
// logged in (the attachments zone only renders for authenticated
// reporters), then drive: open → fill → toggle type → attach an image
// + a video → submit, asserting no uncaught error, no global-error
// boundary, and a working success state.

const FAKE_SESSION = {
  user: {
    id: "11111111-1111-1111-1111-111111111111",
    email: "tester@example.com",
    first_name: "Test",
    last_name: "Reporter",
    display_name: "Test Reporter",
    role: "agency",
    referrer_enabled: false,
    email_verified: true,
    kyc_status: "completed",
    created_at: "2026-01-01T00:00:00Z",
  },
  organization: {
    id: "22222222-2222-2222-2222-222222222222",
    type: "agency",
    owner_user_id: "11111111-1111-1111-1111-111111111111",
    member_role: "owner",
    member_title: "Owner",
    permissions: [],
  },
}

// The exact heading rendered by src/app/global-error.tsx — the
// html-level boundary the crash escalated to. Anchored to avoid
// matching unrelated landing-page copy.
const GLOBAL_ERROR_HEADING = /^We hit a snag$/

async function installLoggedInSession(page: Page) {
  await page.route("**/api/v1/auth/me", (route) =>
    route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify(FAKE_SESSION),
    }),
  )
  // presign returns the BARE PresignFeedbackAttachmentResponse — apiClient
  // does not unwrap a { data } envelope for these endpoints.
  await page.route("**/api/v1/feedback/attachments/presign", (route) => {
    const kind = route.request().postDataJSON?.()?.kind ?? "image"
    return route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        url: "https://example.r2.dev/feedback/abc",
        object_key: `feedback/${kind}-abc`,
        kind,
      }),
    })
  })
  // The direct PUT to R2 — return 200 so the upload resolves "uploaded".
  await page.route("https://example.r2.dev/**", (route) =>
    route.fulfill({ status: 200, body: "" }),
  )
  // submit returns the bare report summary.
  await page.route("**/api/v1/feedback", (route) =>
    route.fulfill({
      status: 201,
      contentType: "application/json",
      body: JSON.stringify({
        id: "fb-1",
        type: "bug",
        status: "received",
        created_at: "2026-01-01T00:00:00Z",
      }),
    }),
  )
}

test("logged-in reporter can open, attach image+video, and submit without crashing", async ({
  page,
}) => {
  const pageErrors: string[] = []
  const queryClientErrors: string[] = []
  page.on("pageerror", (err) => {
    pageErrors.push(`${err.name}: ${err.message}\n${err.stack ?? ""}`)
  })
  page.on("console", (msg) => {
    if (msg.type() === "error" && /QueryClient/i.test(msg.text())) {
      queryClientErrors.push(msg.text())
    }
  })

  await installLoggedInSession(page)
  await page.goto("/fr")

  // The bottom-left "Signaler" FAB.
  const fab = page.getByRole("button", { name: /signaler/i })
  await fab.waitFor({ state: "visible", timeout: 15000 })
  await fab.click()

  // Modal opens (would never appear if the QueryClient throw crashed it).
  const dialog = page.getByRole("dialog")
  await dialog.waitFor({ state: "visible", timeout: 10000 })

  // The global-error boundary must NOT have taken over the page.
  await expect(page.getByRole("heading", { name: GLOBAL_ERROR_HEADING })).toHaveCount(0)

  // Fill title + description.
  await page.getByLabel(/titre/i).fill("Le bouton ne marche pas")
  await page
    .getByLabel(/description/i)
    .fill("Rien ne se passe quand je clique sur Envoyer.")

  // Toggle to "Faille de sécurité" then back to "Bug" to exercise the toggle.
  await page.getByRole("radio", { name: /faille|sécurité/i }).click()
  await page.getByRole("radio", { name: /bug/i }).click()

  // Attach an image File and a video File — the logged-in path that
  // mounts the next/image blob preview inside the QueryClient-dependent form.
  const fileInput = page.locator('input[type="file"]')
  await fileInput.setInputFiles([
    {
      name: "screenshot.png",
      mimeType: "image/png",
      buffer: Buffer.from([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]),
    },
    {
      name: "clip.mp4",
      mimeType: "video/mp4",
      buffer: Buffer.from([0x00, 0x00, 0x00, 0x18, 0x66, 0x74, 0x79, 0x70]),
    },
  ])

  // Both files settle to "uploaded" (image preview + video icon render).
  // "Ajoutée" is the fr `attachments_uploaded` confirmation.
  await expect(page.getByText("screenshot.png")).toBeVisible()
  await expect(page.getByText("clip.mp4")).toBeVisible()
  await expect(page.getByText("Ajoutée").first()).toBeVisible({ timeout: 10000 })

  // No uncaught errors and no QueryClient throw at any point.
  expect(pageErrors, "no uncaught page errors").toEqual([])
  expect(queryClientErrors, "no 'No QueryClient set' error").toEqual([])

  // Submit and assert the success state ("Merci, c'est envoyé").
  await page.getByRole("button", { name: /Envoyer le signalement/i }).click()
  await expect(
    page.getByRole("heading", { name: /Merci/i }),
  ).toBeVisible({ timeout: 10000 })

  expect(pageErrors, "no uncaught page errors after submit").toEqual([])
})
