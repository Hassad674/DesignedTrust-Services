// Package feedback is the pure domain for the platform feedback feature
// — user-submitted bug reports and security/vulnerability disclosures.
//
// This is PLATFORM feedback, NOT business state: a Report is owned by
// the user who authored it (ReporterUserID), never by an organization.
// Anonymous submissions are allowed (ReporterUserID is nil); media
// attachments are reserved for authenticated submitters and are
// enforced at the application + handler layers, never here.
//
// The package imports only the Go stdlib + the shared uuid helper, in
// line with the domain-layer dependency rule. It carries no persistence,
// HTTP, or transport concerns. Every entity validates itself through a
// constructor that returns a sentinel domain error on invalid input.
package feedback

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// ReportType discriminates the two kinds of platform feedback a single
// unified model carries: a functional bug or a security/vulnerability
// disclosure. One field, one table — no separate "security_reports".
type ReportType string

const (
	// TypeBug is a functional defect report.
	TypeBug ReportType = "bug"
	// TypeSecurity is a security / vulnerability disclosure.
	TypeSecurity ReportType = "security"
)

// IsValid reports whether the report type is one of the known values.
func (t ReportType) IsValid() bool {
	switch t {
	case TypeBug, TypeSecurity:
		return true
	default:
		return false
	}
}

// ReportStatus is the triage lifecycle state of a report. New reports
// start at StatusNew; an admin moves them through the pipeline.
type ReportStatus string

const (
	StatusNew        ReportStatus = "new"
	StatusTriaged    ReportStatus = "triaged"
	StatusInProgress ReportStatus = "in_progress"
	StatusResolved   ReportStatus = "resolved"
	StatusRejected   ReportStatus = "rejected"
)

// IsValid reports whether the status is one of the known lifecycle
// values.
func (s ReportStatus) IsValid() bool {
	switch s {
	case StatusNew, StatusTriaged, StatusInProgress, StatusResolved, StatusRejected:
		return true
	default:
		return false
	}
}

// IsTerminal reports whether the status is an end state. Resolved and
// rejected reports carry a resolved_at / resolved_by stamp; the other
// states do not.
func (s ReportStatus) IsTerminal() bool {
	return s == StatusResolved || s == StatusRejected
}

// Severity is the admin-assigned impact rating. Optional: a freshly
// submitted report has no severity until an admin triages it, so the
// zero value ("") is meaningful and distinct from any valid rating.
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// IsValid reports whether the severity is one of the known values.
// The empty severity is NOT valid here — callers that allow "unset"
// must check for "" separately (see Report.SetSeverity).
func (s Severity) IsValid() bool {
	switch s {
	case SeverityLow, SeverityMedium, SeverityHigh, SeverityCritical:
		return true
	default:
		return false
	}
}

// AttachmentKind discriminates the two media kinds a logged-in reporter
// may attach. Text-only (anonymous) reports carry no attachments.
type AttachmentKind string

const (
	AttachmentImage AttachmentKind = "image"
	AttachmentVideo AttachmentKind = "video"
)

// IsValid reports whether the attachment kind is known.
func (k AttachmentKind) IsValid() bool {
	switch k {
	case AttachmentImage, AttachmentVideo:
		return true
	default:
		return false
	}
}

// Length bounds for the free-text fields. Enforced in the domain so the
// rule lives in exactly one place and both the public submit path and
// any future caller share it.
const (
	// MinTitleLength rejects empty / whitespace-only titles.
	MinTitleLength = 3
	// MaxTitleLength caps the title at a single tweet-ish line.
	MaxTitleLength = 200
	// MinDescriptionLength forces a minimally useful description.
	MinDescriptionLength = 10
	// MaxDescriptionLength caps the body so a hostile client cannot
	// stuff megabytes of text through the public endpoint. The handler
	// also bounds the raw request body; this is the post-sanitisation
	// character cap.
	MaxDescriptionLength = 5000
	// MaxPageURLLength bounds the optional page_url field.
	MaxPageURLLength = 2048
	// MaxReporterEmailLength bounds the optional contact email.
	MaxReporterEmailLength = 320
	// MaxNoteBodyLength bounds an internal admin note.
	MaxNoteBodyLength = 5000
)

// Report is a single platform feedback submission. ReporterUserID is
// nil for anonymous submissions. Severity / ResolvedAt / ResolvedBy are
// populated by admin triage and are nil/empty on a fresh report.
//
// Fields are public so the repository can scan directly into the
// struct, but callers outside this package construct a Report through
// NewReport so the validation rules are always enforced.
type Report struct {
	ID            uuid.UUID
	ReporterID    *uuid.UUID // nil = anonymous submission
	Type          ReportType
	Title         string
	Description   string
	Status        ReportStatus
	Severity      Severity // "" until triaged
	PageURL       string
	Context       map[string]any // free-form client context, always non-nil
	ReporterEmail string         // optional contact for anonymous reporters
	IPHash        string         // salted hash of the submitter IP — never the raw IP (RGPD)
	ResolvedAt    *time.Time
	ResolvedBy    *uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewReportInput groups the constructor arguments for NewReport. A
// struct keeps the parameter count under the project's 4-arg limit and
// makes call sites self-documenting.
type NewReportInput struct {
	ReporterID    *uuid.UUID
	Type          ReportType
	Title         string
	Description   string
	PageURL       string
	Context       map[string]any
	ReporterEmail string
	IPHash        string
}

// NewReport validates the input and returns a Report ready to persist.
// Title and Description are trimmed and length-checked; the caller is
// responsible for having already stripped HTML from them (the handler
// boundary does this) — the domain enforces the post-sanitisation
// bounds. A new report always starts at StatusNew with no severity.
func NewReport(in NewReportInput) (*Report, error) {
	if !in.Type.IsValid() {
		return nil, ErrInvalidType
	}

	title := strings.TrimSpace(in.Title)
	if len(title) < MinTitleLength {
		return nil, ErrTitleTooShort
	}
	if len(title) > MaxTitleLength {
		return nil, ErrTitleTooLong
	}

	description := strings.TrimSpace(in.Description)
	if len(description) < MinDescriptionLength {
		return nil, ErrDescriptionTooShort
	}
	if len(description) > MaxDescriptionLength {
		return nil, ErrDescriptionTooLong
	}

	pageURL := strings.TrimSpace(in.PageURL)
	if len(pageURL) > MaxPageURLLength {
		return nil, ErrPageURLTooLong
	}

	email := strings.TrimSpace(in.ReporterEmail)
	if len(email) > MaxReporterEmailLength {
		return nil, ErrReporterEmailTooLong
	}

	context := in.Context
	if context == nil {
		context = map[string]any{}
	}

	now := time.Now().UTC()
	return &Report{
		ID:            uuid.New(),
		ReporterID:    in.ReporterID,
		Type:          in.Type,
		Title:         title,
		Description:   description,
		Status:        StatusNew,
		PageURL:       pageURL,
		Context:       context,
		ReporterEmail: email,
		IPHash:        in.IPHash,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// IsAnonymous reports whether the report was filed without an
// authenticated identity. Media attachments are forbidden on anonymous
// reports — enforced by the application layer.
func (r *Report) IsAnonymous() bool {
	return r.ReporterID == nil
}

// Attachment is a single piece of media (image or video) attached to a
// report by an authenticated reporter. ObjectKey is the randomized R2
// storage key minted server-side at presign time — never a public URL
// and never a client-supplied filename.
type Attachment struct {
	ID          uuid.UUID
	ReportID    uuid.UUID
	Kind        AttachmentKind
	ObjectKey   string
	ContentType string
	SizeBytes   int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewAttachmentInput groups the constructor arguments for NewAttachment.
type NewAttachmentInput struct {
	ReportID    uuid.UUID
	Kind        AttachmentKind
	ObjectKey   string
	ContentType string
	SizeBytes   int64
}

// NewAttachment validates the input and returns an Attachment ready to
// persist. The object key must be non-empty (it is minted server-side)
// and the kind must be a known value.
func NewAttachment(in NewAttachmentInput) (*Attachment, error) {
	if in.ReportID == uuid.Nil {
		return nil, ErrMissingReport
	}
	if !in.Kind.IsValid() {
		return nil, ErrInvalidAttachmentKind
	}
	if strings.TrimSpace(in.ObjectKey) == "" {
		return nil, ErrMissingObjectKey
	}
	if in.SizeBytes < 0 {
		return nil, ErrInvalidAttachmentSize
	}
	now := time.Now().UTC()
	return &Attachment{
		ID:          uuid.New(),
		ReportID:    in.ReportID,
		Kind:        in.Kind,
		ObjectKey:   strings.TrimSpace(in.ObjectKey),
		ContentType: strings.TrimSpace(in.ContentType),
		SizeBytes:   in.SizeBytes,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Note is an internal admin triage note attached to a report. Notes are
// authored by admins (AdminUserID) and are never exposed to reporters.
type Note struct {
	ID          uuid.UUID
	ReportID    uuid.UUID
	AdminUserID uuid.UUID
	Body        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewNoteInput groups the constructor arguments for NewNote.
type NewNoteInput struct {
	ReportID    uuid.UUID
	AdminUserID uuid.UUID
	Body        string
}

// NewNote validates the input and returns a Note ready to persist. The
// body is trimmed and bounded; both the report and admin ids are
// required.
func NewNote(in NewNoteInput) (*Note, error) {
	if in.ReportID == uuid.Nil {
		return nil, ErrMissingReport
	}
	if in.AdminUserID == uuid.Nil {
		return nil, ErrMissingAdmin
	}
	body := strings.TrimSpace(in.Body)
	if body == "" {
		return nil, ErrNoteBodyEmpty
	}
	if len(body) > MaxNoteBodyLength {
		return nil, ErrNoteBodyTooLong
	}
	now := time.Now().UTC()
	return &Note{
		ID:          uuid.New(),
		ReportID:    in.ReportID,
		AdminUserID: in.AdminUserID,
		Body:        body,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}
