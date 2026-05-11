/**
 * D4 (GDPR Phase C) — LegalDocument unit tests.
 *
 * The component is a server-renderable presentation surface for the
 * long-form legal markdown shipped under /legal/*. The tests assert:
 *
 *   1. Title + subtitle + formatted date render exactly once.
 *   2. Each section renders an H2 with stable id (for in-page anchors)
 *      and aria-labelledby wiring (a11y).
 *   3. Paragraph blocks emit <p>, list blocks emit <ul>/<ol>, table
 *      blocks emit a <table> with caption + thead + tbody.
 *   4. The optional `sourceHref` renders an anchor; the optional
 *      `englishNotice` renders the disclaimer banner.
 *
 * No i18n provider needed — the component is dumb and receives
 * already-resolved text via props.
 */
import { describe, it, expect } from "vitest"
import { render, screen } from "@testing-library/react"
import { LegalDocument } from "../legal-document"
import type { LegalSection } from "../legal-document"

const SECTIONS: LegalSection[] = [
  {
    id: "scope",
    heading: "Périmètre",
    blocks: [{ type: "p", content: "Ce document couvre l'ensemble." }],
  },
  {
    id: "list-section",
    heading: "Liste de mesures",
    blocks: [
      {
        type: "ul",
        items: ["Chiffrement TLS", "Audit log append-only"],
      },
    ],
  },
  {
    id: "table-section",
    heading: "Sous-traitants",
    blocks: [
      {
        type: "table",
        caption: "Liste indicative",
        headers: ["Nom", "Pays"],
        rows: [
          ["Neon", "UE"],
          ["Vercel", "US"],
        ],
      },
    ],
  },
  {
    id: "ordered",
    heading: "Étapes",
    blocks: [{ type: "ol", items: ["Étape 1", "Étape 2"] }],
  },
  {
    id: "callout",
    heading: "Avertissement",
    blocks: [
      {
        type: "callout",
        variant: "warning",
        content: "À valider par le DPO.",
      },
    ],
  },
]

function renderDoc(extra?: Partial<React.ComponentProps<typeof LegalDocument>>) {
  return render(
    <LegalDocument
      title="Document de test"
      subtitle="Sous-titre indicatif"
      lastUpdatedISO="2026-05-11"
      sections={SECTIONS}
      {...extra}
    />,
  )
}

describe("LegalDocument", () => {
  it("renders the title as a single H1", () => {
    renderDoc()
    const headings = screen.getAllByRole("heading", { level: 1 })
    expect(headings).toHaveLength(1)
    expect(headings[0]).toHaveTextContent("Document de test")
  })

  it("renders the subtitle next to the H1", () => {
    renderDoc()
    expect(screen.getByText("Sous-titre indicatif")).toBeInTheDocument()
  })

  it("renders a formatted ISO date for the last update", () => {
    renderDoc()
    expect(screen.getByText(/Version du 2026-05-11/)).toBeInTheDocument()
  })

  it("emits a stable id and aria-labelledby on every section", () => {
    renderDoc()
    for (const section of SECTIONS) {
      const node = document.getElementById(section.id)
      expect(node, `section ${section.id}`).not.toBeNull()
      expect(node?.getAttribute("aria-labelledby")).toBe(
        `${section.id}-heading`,
      )
      const heading = document.getElementById(`${section.id}-heading`)
      expect(heading?.textContent).toBe(section.heading)
    }
  })

  it("renders paragraph blocks as <p> tags", () => {
    renderDoc()
    expect(
      screen.getByText("Ce document couvre l'ensemble."),
    ).toBeInTheDocument()
    const para = screen.getByText("Ce document couvre l'ensemble.")
    expect(para.tagName.toLowerCase()).toBe("p")
  })

  it("renders unordered list blocks as <ul>/<li>", () => {
    renderDoc()
    expect(screen.getByText("Chiffrement TLS").tagName.toLowerCase()).toBe("li")
    expect(
      screen.getByText("Audit log append-only").tagName.toLowerCase(),
    ).toBe("li")
  })

  it("renders ordered list blocks as <ol>", () => {
    renderDoc()
    const item = screen.getByText("Étape 1")
    expect(item.tagName.toLowerCase()).toBe("li")
    expect(item.parentElement?.tagName.toLowerCase()).toBe("ol")
  })

  it("renders table blocks with caption + thead + tbody", () => {
    renderDoc()
    const table = screen.getByRole("table")
    expect(table).toBeInTheDocument()
    expect(table.querySelector("caption")?.textContent).toBe("Liste indicative")
    expect(table.querySelectorAll("thead th")).toHaveLength(2)
    expect(table.querySelectorAll("tbody tr")).toHaveLength(2)
    expect(screen.getByText("Neon")).toBeInTheDocument()
    expect(screen.getByText("US")).toBeInTheDocument()
  })

  it("renders the callout block", () => {
    renderDoc()
    expect(screen.getByText("À valider par le DPO.")).toBeInTheDocument()
  })

  it("renders the optional sourceHref as an anchor when provided", () => {
    renderDoc({
      sourceHref: "https://example.com/source.md",
    })
    const link = document.querySelector('a[href="https://example.com/source.md"]')
    expect(link).not.toBeNull()
  })

  it("renders the optional englishNotice banner when provided", () => {
    renderDoc({
      englishNotice: "English version available on request.",
    })
    expect(
      screen.getByText("English version available on request."),
    ).toBeInTheDocument()
  })

  it("does NOT render the source anchor when sourceHref is omitted", () => {
    renderDoc()
    expect(document.querySelector('a[href$=".md"]')).toBeNull()
  })

  it("does NOT render an English-notice banner when englishNotice is omitted", () => {
    renderDoc()
    expect(
      screen.queryByText(/English version available on request/),
    ).toBeNull()
  })
})
