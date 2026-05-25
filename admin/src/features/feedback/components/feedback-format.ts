// formatBytes renders a byte count as a human-readable size in French
// units (o / Ko / Mo), mirroring the media feature's convention.
export function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} o`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} Ko`
  return `${(bytes / (1024 * 1024)).toFixed(1)} Mo`
}

// CONTEXT_LABELS maps the known feedback context keys (the SubmitFeedback
// context object: platform, locale, role, app_version, user_agent,
// viewport) to French labels. Unknown keys fall back to the raw key so
// future additions still render.
const CONTEXT_LABELS: Record<string, string> = {
  platform: "Plateforme",
  locale: "Langue",
  role: "Rôle",
  app_version: "Version",
  user_agent: "Navigateur",
  viewport: "Fenêtre",
}

export type ContextEntry = { key: string; label: string; value: string }

// formatContext flattens the report's free-form context object into a
// stable, label-resolved list of string entries for display. Non-string
// values are JSON-stringified; empty values are dropped.
export function formatContext(
  context: Record<string, unknown> | null | undefined,
): ContextEntry[] {
  if (!context) return []
  return Object.entries(context)
    .map(([key, raw]) => {
      const value =
        typeof raw === "string" ? raw : raw == null ? "" : JSON.stringify(raw)
      return { key, label: CONTEXT_LABELS[key] ?? key, value }
    })
    .filter((entry) => entry.value !== "")
}
