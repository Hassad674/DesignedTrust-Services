import { API_BASE_URL } from "@/shared/lib/api-client"
import { uploadVideoDirect } from "@/shared/lib/upload/direct-video-upload"

// Per-persona video-upload boundary for the referrer aggregate.
// Hits the dedicated endpoints under /api/v1/referrer-profile/video
// which write directly to referrer_profiles.video_url. The legacy
// /api/v1/upload/referrer-video path is unreachable for
// provider_personal orgs since migration 104 removed their legacy
// profiles row.

type UploadVideoResponse = { video_url: string }

// Direct-to-R2 presigned upload (presign + complete) so videos bypass
// the Vercel proxy body cap. /video/complete persists video_url and
// fires the SAME moderation pipeline as the legacy multipart POST.
export async function uploadReferrerVideo(
  file: File,
): Promise<UploadVideoResponse> {
  return uploadVideoDirect<UploadVideoResponse>(
    {
      presign: "/api/v1/referrer-profile/video/presign",
      complete: "/api/v1/referrer-profile/video/complete",
    },
    file,
  )
}

export async function deleteReferrerVideo(): Promise<void> {
  const res = await fetch(
    `${API_BASE_URL}/api/v1/referrer-profile/video`,
    {
      method: "DELETE",
      credentials: "include",
    },
  )
  if (!res.ok) {
    throw new Error("delete_failed")
  }
}
