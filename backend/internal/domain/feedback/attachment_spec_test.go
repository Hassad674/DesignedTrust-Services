package feedback_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"marketplace-backend/internal/domain/feedback"
)

func TestValidatePresign(t *testing.T) {
	tests := []struct {
		name        string
		kind        feedback.AttachmentKind
		contentType string
		size        int64
		wantKind    feedback.AttachmentKind
		wantExt     string
		wantErr     error
	}{
		{
			name:        "png image ok",
			kind:        feedback.AttachmentImage,
			contentType: "image/png",
			size:        1024,
			wantKind:    feedback.AttachmentImage,
			wantExt:     "png",
		},
		{
			name:        "jpeg image ok",
			kind:        feedback.AttachmentImage,
			contentType: "image/jpeg",
			size:        feedback.MaxImageSizeBytes,
			wantKind:    feedback.AttachmentImage,
			wantExt:     "jpg",
		},
		{
			name:        "webp image ok",
			kind:        feedback.AttachmentImage,
			contentType: "image/webp",
			size:        2048,
			wantKind:    feedback.AttachmentImage,
			wantExt:     "webp",
		},
		{
			name:        "mp4 video ok",
			kind:        feedback.AttachmentVideo,
			contentType: "video/mp4",
			size:        feedback.MaxVideoSizeBytes,
			wantKind:    feedback.AttachmentVideo,
			wantExt:     "mp4",
		},
		{
			name:        "webm video ok",
			kind:        feedback.AttachmentVideo,
			contentType: "video/webm",
			size:        5 << 20,
			wantKind:    feedback.AttachmentVideo,
			wantExt:     "webm",
		},
		{
			name:        "content type with codecs parameter is normalised",
			kind:        feedback.AttachmentVideo,
			contentType: "video/webm; codecs=vp9",
			size:        5 << 20,
			wantKind:    feedback.AttachmentVideo,
			wantExt:     "webm",
		},
		{
			name:        "uppercase content type is normalised",
			kind:        feedback.AttachmentImage,
			contentType: "IMAGE/PNG",
			size:        1024,
			wantKind:    feedback.AttachmentImage,
			wantExt:     "png",
		},
		{
			name:        "invalid kind",
			kind:        feedback.AttachmentKind("audio"),
			contentType: "image/png",
			size:        1024,
			wantErr:     feedback.ErrInvalidAttachmentKind,
		},
		{
			name:        "unsupported content type (gif)",
			kind:        feedback.AttachmentImage,
			contentType: "image/gif",
			size:        1024,
			wantErr:     feedback.ErrUnsupportedContentType,
		},
		{
			name:        "executable content type rejected",
			kind:        feedback.AttachmentImage,
			contentType: "application/x-msdownload",
			size:        1024,
			wantErr:     feedback.ErrUnsupportedContentType,
		},
		{
			name:        "kind mismatch image MIME with video kind",
			kind:        feedback.AttachmentVideo,
			contentType: "image/png",
			size:        1024,
			wantErr:     feedback.ErrUnsupportedContentType,
		},
		{
			name:        "kind mismatch video MIME with image kind (smuggle under image cap)",
			kind:        feedback.AttachmentImage,
			contentType: "video/mp4",
			size:        20 << 20,
			wantErr:     feedback.ErrUnsupportedContentType,
		},
		{
			name:        "zero size rejected",
			kind:        feedback.AttachmentImage,
			contentType: "image/png",
			size:        0,
			wantErr:     feedback.ErrInvalidAttachmentSize,
		},
		{
			name:        "image over cap rejected",
			kind:        feedback.AttachmentImage,
			contentType: "image/png",
			size:        feedback.MaxImageSizeBytes + 1,
			wantErr:     feedback.ErrAttachmentTooLarge,
		},
		{
			name:        "video over cap rejected",
			kind:        feedback.AttachmentVideo,
			contentType: "video/mp4",
			size:        feedback.MaxVideoSizeBytes + 1,
			wantErr:     feedback.ErrAttachmentTooLarge,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kind, ext, err := feedback.ValidatePresign(tt.kind, tt.contentType, tt.size)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantKind, kind)
			assert.Equal(t, tt.wantExt, ext)
		})
	}
}

func TestMaxSizeForKind(t *testing.T) {
	assert.Equal(t, feedback.MaxImageSizeBytes, feedback.MaxSizeForKind(feedback.AttachmentImage))
	assert.Equal(t, feedback.MaxVideoSizeBytes, feedback.MaxSizeForKind(feedback.AttachmentVideo))
	// Unknown kind defaults to the (smaller) image cap.
	assert.Equal(t, feedback.MaxImageSizeBytes, feedback.MaxSizeForKind(feedback.AttachmentKind("x")))
}

func TestIsAllowedObjectExtension(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{"reports/abc/def.png", true},
		{"reports/abc/def.jpg", true},
		{"reports/abc/def.webp", true},
		{"reports/abc/def.mp4", true},
		{"reports/abc/def.webm", true},
		{"reports/abc/def.PNG", true},
		{"reports/abc/def.html", false},
		{"reports/abc/def.exe", false},
		{"reports/abc/def", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			assert.Equal(t, tt.want, feedback.IsAllowedObjectExtension(tt.key))
		})
	}
}
