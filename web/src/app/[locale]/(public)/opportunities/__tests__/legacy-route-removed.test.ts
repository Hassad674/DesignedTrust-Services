import { describe, it, expect } from "vitest"
import { existsSync } from "node:fs"
import { resolve } from "node:path"

// Regression guard: the dedicated /my-applications route was deleted
// when "Mes candidatures" moved into the Opportunités tab system. If
// the route file ever reappears, the sidebar nav + composition story
// breaks (we'd ship two ways to reach the same data + a duplicate
// fetch). Cheap source-level test.
describe("legacy /my-applications route", () => {
  it("is removed from the (app) route group", () => {
    const projectRoot = resolve(__dirname, "../../../../../..")
    const legacyRoute = resolve(
      projectRoot,
      "src/app/[locale]/(app)/my-applications/page.tsx",
    )
    expect(existsSync(legacyRoute)).toBe(false)
  })
})
