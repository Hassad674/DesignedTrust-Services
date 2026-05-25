package feedback

import (
	"path"
	"strings"
)

// Attachment upload caps. Media is logged-in-only; these are the size
// ceilings the presign boundary enforces before issuing a PUT URL.
const (
	// MaxImageSizeBytes caps a single image attachment at 10 MB.
	MaxImageSizeBytes int64 = 10 << 20
	// MaxVideoSizeBytes caps a single video attachment at 50 MB.
	MaxVideoSizeBytes int64 = 50 << 20
)

// allowedContentTypes maps each accepted upload MIME type to the safe
// file extension the SERVER stamps onto the storage key and the
// attachment kind it belongs to. The client-declared content type is
// validated against this table; anything absent is rejected at the
// presign boundary. The extension is never taken from a client-supplied
// filename — only from this server-side allowlist.
var allowedContentTypes = map[string]struct {
	kind AttachmentKind
	ext  string
}{
	"image/png":  {AttachmentImage, "png"},
	"image/jpeg": {AttachmentImage, "jpg"},
	"image/webp": {AttachmentImage, "webp"},
	"video/mp4":  {AttachmentVideo, "mp4"},
	"video/webm": {AttachmentVideo, "webm"},
}

// normaliseContentType lowercases, trims, and strips any "; codecs=..."
// parameter a browser may append so the allowlist lookup is robust.
func normaliseContentType(contentType string) string {
	base := strings.TrimSpace(strings.ToLower(contentType))
	if idx := strings.IndexByte(base, ';'); idx >= 0 {
		base = strings.TrimSpace(base[:idx])
	}
	return base
}

// MaxSizeForKind returns the byte ceiling for a given attachment kind.
func MaxSizeForKind(kind AttachmentKind) int64 {
	if kind == AttachmentVideo {
		return MaxVideoSizeBytes
	}
	return MaxImageSizeBytes
}

// ValidatePresign validates a presign request against the content-type
// allowlist, the declared kind, and the per-kind size cap. It returns
// the resolved attachment kind and the server-controlled file extension
// to stamp onto the storage key. The declared kind must agree with the
// kind implied by the content type (an image MIME with kind=video is
// rejected) so a client cannot smuggle a video under an image cap.
//
// This is pure policy — it mints no key and touches no storage. The
// adapter/handler layer composes the randomized key from the returned
// extension.
func ValidatePresign(kind AttachmentKind, contentType string, sizeBytes int64) (resolvedKind AttachmentKind, ext string, err error) {
	if !kind.IsValid() {
		return "", "", ErrInvalidAttachmentKind
	}
	spec, ok := allowedContentTypes[normaliseContentType(contentType)]
	if !ok {
		return "", "", ErrUnsupportedContentType
	}
	if spec.kind != kind {
		// Declared kind disagrees with the kind implied by the MIME.
		return "", "", ErrUnsupportedContentType
	}
	if sizeBytes <= 0 {
		return "", "", ErrInvalidAttachmentSize
	}
	if sizeBytes > MaxSizeForKind(spec.kind) {
		return "", "", ErrAttachmentTooLarge
	}
	return spec.kind, spec.ext, nil
}

// IsAllowedObjectExtension reports whether the extension of the given
// object key is one the presign allowlist could have minted. Used as a
// defensive guard when persisting an attachment whose key came back
// from a presign round-trip — a tampered key carrying e.g. ".html" is
// refused.
func IsAllowedObjectExtension(objectKey string) bool {
	ext := strings.TrimPrefix(strings.ToLower(path.Ext(objectKey)), ".")
	if ext == "" {
		return false
	}
	for _, spec := range allowedContentTypes {
		if spec.ext == ext {
			return true
		}
	}
	return false
}
