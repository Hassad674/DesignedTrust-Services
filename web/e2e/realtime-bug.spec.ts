// E2E reproduction of the 2026-05-12/13 realtime regression:
// "user A sends a message, user B never sees it without refresh".
//
// Uses the existing test users:
//   provider/sender:    testadmin@gmail.com
//   enterprise/recipient: pmpm@gmail.com
// Both share password Test1234! (set via SQL update in setup).
// Both participate in conversation a82ca4b1-ef04-481b-b164-20f60ec57546.
//
// If this test PASSES, the rendering chain works end-to-end and the
// bug is specific to the user's actual browser environment. If it
// FAILS, we have a deterministic repro to debug from.

import { test, expect } from "@playwright/test"
import type { BrowserContext, Page } from "@playwright/test"

const CONV_ID = "a82ca4b1-ef04-481b-b164-20f60ec57546"
const SENDER = { email: "testadmin@gmail.com", password: "Test1234!" }
const RECIPIENT = { email: "pmpm@gmail.com", password: "Test1234!" }

async function loginViaUI(page: Page, email: string, password: string) {
  await page.goto("/fr/login")
  // Try french or fall back to english label
  const emailInput = page.getByLabel(/email/i).first()
  await emailInput.fill(email)
  const passwordInput = page.getByLabel(/mot de passe|password/i).first()
  await passwordInput.fill(password)
  // Submit button - prefer FR text, fall back to EN
  const submit = page.locator("button[type='submit']").first()
  await submit.click()
  await page.waitForURL((u) => !u.toString().includes("/login"), { timeout: 30_000 })
}

test.describe("realtime regression — 2026-05-13", () => {
  test("user A sends, user B sees the message without manual refresh", async ({
    browser,
  }) => {
    test.setTimeout(120_000)

    // Two isolated contexts so cookies/storage don't leak
    const ctxA = await browser.newContext()
    const ctxB = await browser.newContext()

    const pageA = await ctxA.newPage()
    const pageB = await ctxB.newPage()

    // Capture console logs from both pages — surfaces our [WS-DEBUG] traces
    pageA.on("console", (msg) => {
      console.log(`[A console.${msg.type()}] ${msg.text()}`)
    })
    pageB.on("console", (msg) => {
      console.log(`[B console.${msg.type()}] ${msg.text()}`)
    })

    // Capture WS frames flowing into each browser
    const framesA: string[] = []
    const framesB: string[] = []
    pageA.on("websocket", (ws) => {
      console.log(`[A WS open] ${ws.url()}`)
      ws.on("framereceived", (e) => {
        const data = typeof e.payload === "string" ? e.payload : e.payload.toString("utf8")
        framesA.push(data)
        if (data.includes("new_message")) {
          console.log(`[A WS framereceived] new_message: ${data.slice(0, 120)}`)
        }
      })
      ws.on("close", () => console.log(`[A WS closed] ${ws.url()}`))
    })
    pageB.on("websocket", (ws) => {
      console.log(`[B WS open] ${ws.url()}`)
      ws.on("framereceived", (e) => {
        const data = typeof e.payload === "string" ? e.payload : e.payload.toString("utf8")
        framesB.push(data)
        if (data.includes("new_message")) {
          console.log(`[B WS framereceived] new_message: ${data.slice(0, 200)}`)
        }
      })
      ws.on("close", () => console.log(`[B WS closed] ${ws.url()}`))
    })

    // Login both users in parallel
    console.log("=== Logging in users ===")
    await Promise.all([
      loginViaUI(pageA, SENDER.email, SENDER.password),
      loginViaUI(pageB, RECIPIENT.email, RECIPIENT.password),
    ])
    console.log("=== Both logged in ===")

    // Navigate both to the conversation
    await Promise.all([
      pageA.goto(`/fr/messages?id=${CONV_ID}`),
      pageB.goto(`/fr/messages?id=${CONV_ID}`),
    ])

    // Wait for the conversation to be loaded — look for the message input
    const inputSelector = "textarea, input[type='text']"
    await pageA.waitForSelector(inputSelector, { timeout: 30_000 })
    await pageB.waitForSelector(inputSelector, { timeout: 30_000 })

    // Give WS time to connect + handler to mount
    console.log("=== Waiting 3s for WS to settle ===")
    await pageA.waitForTimeout(3_000)

    // Sender (A) types and sends a unique message
    const uniqueMsg = `E2E-RT-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`
    console.log(`=== Sender posts message: "${uniqueMsg}" ===`)

    const inputA = pageA.locator("textarea").first()
    await inputA.fill(uniqueMsg)
    await inputA.press("Enter")

    // Recipient (B) MUST see the message within 10 s WITHOUT reloading
    console.log("=== Waiting for recipient to see the message ===")
    const messageVisibleStart = Date.now()
    try {
      await expect(pageB.getByText(uniqueMsg)).toBeVisible({ timeout: 10_000 })
      const elapsed = Date.now() - messageVisibleStart
      console.log(`✅ RECIPIENT SAW MESSAGE in ${elapsed}ms`)
    } catch (err) {
      console.log("❌ RECIPIENT DID NOT SEE THE MESSAGE within 10s")
      console.log("--- Last 3 frames received by B ---")
      framesB.slice(-3).forEach((f, i) => console.log(`  frame${i}: ${f.slice(0, 250)}`))
      console.log("--- Total frames A: " + framesA.length + ", B: " + framesB.length)
      // Take screenshots for debugging
      await pageA.screenshot({ path: "/tmp/realtime-a.png", fullPage: true })
      await pageB.screenshot({ path: "/tmp/realtime-b.png", fullPage: true })
      throw err
    }

    await ctxA.close()
    await ctxB.close()
  })
})
