import { API_BASE_URL } from "@/shared/lib/api-client"
import { uploadVideoDirect } from "@/shared/lib/upload/direct-video-upload"

// Per-persona video-upload boundary for the freelance aggregate.
// Hits the dedicated endpoints under /api/v1/freelance-profile/video
// which write directly to freelance_profiles.video_url — the legacy
// /api/v1/upload/video path still serves agency orgs but cannot
// persist for provider_personal users since migration 104.

type UploadVideoResponse = { video_url: string }

// Videos upload DIRECTLY to R2 via a presigned PUT (presign +
// complete), bypassing the Vercel proxy's ~4.5 MB body cap that 413'd
// large videos in production. The backend's /video/complete endpoint
// persists video_url and runs the SAME moderation pipeline the legacy
// multipart POST did. See shared/lib/upload/direct-video-upload.ts.
export async function uploadFreelanceVideo(
  file: File,
): Promise<UploadVideoResponse> {
  return uploadVideoDirect<UploadVideoResponse>(
    {
      presign: "/api/v1/freelance-profile/video/presign",
      complete: "/api/v1/freelance-profile/video/complete",
    },
    file,
  )
}

export async function deleteFreelanceVideo(): Promise<void> {
  const res = await fetch(`${API_BASE_URL}/api/v1/freelance-profile/video`, {
    method: "DELETE",
    credentials: "include",
  })
  if (!res.ok) {
    throw new Error("delete_failed")
  }
}
