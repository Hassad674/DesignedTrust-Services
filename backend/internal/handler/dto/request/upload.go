package request

// PresignVideoRequest is the body of the video-presign endpoints
// (POST .../video/presign across the profile / referrer / portfolio /
// review surfaces). The browser sends the original filename (used only
// for the human-facing media row, never for the storage key) and the
// content type, which is validated against the server-side video
// allowlist before a presigned PUT URL is issued.
//
// This mirrors the messaging PresignedURLRequest shape: the bytes never
// transit the backend — the client PUTs them straight to R2 — so this
// request carries no file payload, only metadata.
type PresignVideoRequest struct {
	Filename    string `json:"filename" validate:"required,min=1,max=255"`
	ContentType string `json:"content_type" validate:"required,min=1,max=128"`
}

// CompleteVideoUploadRequest is the body of the video-complete
// endpoints (POST .../video/complete). After the browser has PUT the
// file directly to R2 using the presigned URL, it calls complete with
// the file_key returned by the presign step so the backend can persist
// the video_url AND trigger the moderation pipeline on the stored
// object — exactly as the legacy multipart handler did after
// storage.Upload.
//
// file_key is re-validated server-side against the caller's namespace
// prefix (ownership guard) before anything is persisted. content_type
// and file_size feed the media row that the moderation pipeline reads.
type CompleteVideoUploadRequest struct {
	FileKey     string `json:"file_key" validate:"required,min=1,max=512"`
	Filename    string `json:"filename" validate:"required,min=1,max=255"`
	ContentType string `json:"content_type" validate:"required,min=1,max=128"`
	FileSize    int64  `json:"file_size" validate:"required,gt=0"`
}
