import { apiClient } from "@/shared/lib/api-client"

// Shared client for the DIRECT-to-R2 video upload flow.
//
// WHY THIS EXISTS — in production the web app reaches the Go backend
// through Next.js `rewrites()`, proxied by Vercel. The Vercel proxy
// caps request bodies at ~4.5 MB, so any video over that size used to
// get a 413 at the edge before reaching the backend (which allows
// 50-100 MB). This helper mirrors the messaging-attachment flow: it
// asks the backend for a short-lived presigned PUT URL (a tiny JSON
// request, no body cap) then PUTs the file BYTES DIRECTLY to the R2
// origin — bypassing the Vercel proxy entirely — and finally tells the
// backend the upload is done so it can persist the URL and run the
// moderation pipeline.
//
// The R2 origin is already whitelisted in the CSP `connect-src`
// (see src/shared/lib/csp.ts), so the cross-origin PUT is allowed.

/** Backend presign envelope — identical across every video surface. */
export type PresignedUploadResponse = {
  upload_url: string
  file_key: string
  public_url: string
}

/** Endpoints for one video surface's two-step direct-upload flow. */
export type DirectVideoEndpoints = {
  /** POST endpoint that returns a presigned PUT URL. */
  presign: string
  /** POST endpoint that confirms the upload + triggers persistence/moderation. */
  complete: string
}

/**
 * uploadVideoDirect runs the full presign → PUT → complete sequence and
 * returns the parsed JSON body of the complete endpoint (shape varies
 * per surface: `{ url }` or `{ video_url }`). Callers narrow the return
 * type via the generic `T`.
 *
 * Throwing semantics mirror the messaging flow: `fetch` does NOT throw
 * on a non-2xx PUT, so we explicitly guard `uploadRes.ok` to avoid
 * confirming an upload that never landed in R2.
 */
export async function uploadVideoDirect<T>(
  endpoints: DirectVideoEndpoints,
  file: File,
): Promise<T> {
  const contentType = file.type || "application/octet-stream"

  // 1. Ask the backend for a presigned PUT URL (tiny JSON, no body cap).
  const presign = await apiClient<PresignedUploadResponse>(endpoints.presign, {
    method: "POST",
    body: { filename: file.name, content_type: contentType },
  })

  // 2. PUT the bytes DIRECTLY to R2 (absolute URL → bypasses the proxy).
  const uploadRes = await fetch(presign.upload_url, {
    method: "PUT",
    body: file,
    headers: { "Content-Type": contentType },
  })
  if (!uploadRes.ok) {
    throw new Error(`upload failed: ${uploadRes.status}`)
  }

  // 3. Confirm: backend persists the URL + fires the moderation pipeline.
  return apiClient<T>(endpoints.complete, {
    method: "POST",
    body: {
      file_key: presign.file_key,
      filename: file.name,
      content_type: contentType,
      file_size: file.size,
    },
  })
}
