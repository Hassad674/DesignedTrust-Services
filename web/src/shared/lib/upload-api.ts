import { API_BASE_URL } from "@/shared/lib/api-client"
import { uploadVideoDirect } from "@/shared/lib/upload/direct-video-upload"

const API_URL = API_BASE_URL

type UploadResponse = {
  url: string
}

// Photos stay on the multipart path: they are capped at 5 MB, well
// under the Vercel proxy's ~4.5 MB... actually photos can brush the
// cap, but the photo flow is out of scope for this fix — only videos
// regressed in production. Left as-is per scope.
async function uploadFile(
  endpoint: string,
  file: File,
): Promise<UploadResponse> {
  const formData = new FormData()
  formData.append("file", file)

  const res = await fetch(`${API_URL}${endpoint}`, {
    method: "POST",
    credentials: "include",
    body: formData,
  })

  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: "Upload failed" }))
    throw new Error(err.message || "Upload failed")
  }

  return res.json()
}

export async function uploadPhoto(
  file: File,
): Promise<UploadResponse> {
  return uploadFile("/api/v1/upload/photo", file)
}

// Legacy agency intro video — DIRECT-to-R2 presigned upload (presign +
// complete) so videos > ~4.5 MB no longer 413 at the Vercel proxy. The
// /video/complete endpoint persists the URL onto the agency profile and
// fires the SAME moderation pipeline as the legacy multipart POST.
export async function uploadVideo(
  file: File,
): Promise<UploadResponse> {
  return uploadVideoDirect<UploadResponse>(
    {
      presign: "/api/v1/upload/video/presign",
      complete: "/api/v1/upload/video/complete",
    },
    file,
  )
}

// Legacy agency referrer video — direct-to-R2 presigned upload.
export async function uploadReferrerVideo(
  file: File,
): Promise<UploadResponse> {
  return uploadVideoDirect<UploadResponse>(
    {
      presign: "/api/v1/upload/referrer-video/presign",
      complete: "/api/v1/upload/referrer-video/complete",
    },
    file,
  )
}

export async function deleteVideo(): Promise<void> {
  const res = await fetch(`${API_URL}/api/v1/upload/video`, {
    method: "DELETE",
    credentials: "include",
  })
  if (!res.ok) throw new Error("Failed to delete video")
}

export async function deleteReferrerVideo(): Promise<void> {
  const res = await fetch(`${API_URL}/api/v1/upload/referrer-video`, {
    method: "DELETE",
    credentials: "include",
  })
  if (!res.ok) throw new Error("Failed to delete referrer video")
}
