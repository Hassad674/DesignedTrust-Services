package feedback

import "errors"

// Sentinel domain errors for the feedback feature. They are matched at
// the handler layer with errors.Is and mapped to HTTP status codes. The
// domain layer never wraps these — wrapping happens in the app layer.
var (
	// ErrNotFound is returned when a report lookup misses.
	ErrNotFound = errors.New("feedback: report not found")

	// Report validation
	ErrInvalidType          = errors.New("feedback: invalid report type")
	ErrTitleTooShort        = errors.New("feedback: title too short")
	ErrTitleTooLong         = errors.New("feedback: title too long")
	ErrDescriptionTooShort  = errors.New("feedback: description too short")
	ErrDescriptionTooLong   = errors.New("feedback: description too long")
	ErrPageURLTooLong       = errors.New("feedback: page_url too long")
	ErrReporterEmailTooLong = errors.New("feedback: reporter_email too long")

	// Status / severity transitions
	ErrInvalidStatus   = errors.New("feedback: invalid status")
	ErrInvalidSeverity = errors.New("feedback: invalid severity")
	ErrNoUpdateFields  = errors.New("feedback: no update fields provided")

	// Attachment validation
	ErrMissingReport         = errors.New("feedback: report id is required")
	ErrInvalidAttachmentKind = errors.New("feedback: invalid attachment kind")
	ErrMissingObjectKey      = errors.New("feedback: object key is required")
	ErrInvalidAttachmentSize = errors.New("feedback: invalid attachment size")
	ErrAttachmentNotAllowed  = errors.New("feedback: attachments require authentication")

	// Note validation
	ErrMissingAdmin    = errors.New("feedback: admin user id is required")
	ErrNoteBodyEmpty   = errors.New("feedback: note body is required")
	ErrNoteBodyTooLong = errors.New("feedback: note body too long")

	// Presign validation (logged-in-only media upload)
	ErrUnsupportedContentType = errors.New("feedback: unsupported content type")
	ErrAttachmentTooLarge     = errors.New("feedback: attachment too large")
)
