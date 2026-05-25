// Package feedback orchestrates the platform feedback use cases —
// submitting bug / security reports (anonymous text or authenticated
// with media), presigning media uploads for logged-in reporters, and
// the admin triage surface (list / get / update / note).
//
// The service depends only on port interfaces (FeedbackRepository,
// StorageService) plus an injected anonymization salt — never on a
// concrete adapter. It returns domain types and domain errors, never
// HTTP concepts. Audit logging of admin mutations is best-effort and
// never blocks the caller, mirroring the platform-wide audit pattern.
package feedback

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	domain "marketplace-backend/internal/domain/feedback"
	"marketplace-backend/internal/port/repository"
	"marketplace-backend/internal/port/service"
)

// ServiceDeps groups the feedback service collaborators. Bundled into a
// struct so the constructor stays under the 4-parameter rule and new
// optional collaborators can be added without breaking call sites.
type ServiceDeps struct {
	Reports repository.FeedbackRepository
	Storage service.StorageService
	// Audit is optional. When non-nil, admin mutations (update / note)
	// emit an append-only audit row. Best-effort — a failure is logged,
	// never surfaced to the caller.
	Audit auditRecorder
	// AnonymizationSalt is the secret salt used to hash submitter IPs
	// before storage (RGPD — the raw IP is never persisted). Injected
	// from config at wiring time; the service fails the submit closed
	// if it is empty in a context that requires hashing.
	AnonymizationSalt string
	// PresignExpiry bounds the lifetime of issued upload / download
	// URLs. Defaults to 15 minutes when zero.
	PresignExpiry time.Duration
}

// Service implements the feedback use cases.
type Service struct {
	reports       repository.FeedbackRepository
	storage       service.StorageService
	audit         auditRecorder
	salt          string
	presignExpiry time.Duration
}

const defaultPresignExpiry = 15 * time.Minute

// NewService wires a feedback service from its dependencies.
func NewService(deps ServiceDeps) *Service {
	expiry := deps.PresignExpiry
	if expiry <= 0 {
		expiry = defaultPresignExpiry
	}
	return &Service{
		reports:       deps.Reports,
		storage:       deps.Storage,
		audit:         deps.Audit,
		salt:          deps.AnonymizationSalt,
		presignExpiry: expiry,
	}
}

// AttachmentRef is a server-validated reference to media a logged-in
// reporter previously presigned and uploaded. The kind/content-type are
// re-validated by SubmitReport before persistence.
type AttachmentRef struct {
	Kind        string
	ObjectKey   string
	ContentType string
	SizeBytes   int64
}

// SubmitReportInput carries everything the public submit endpoint
// collects. ReporterID is nil for anonymous submissions; Attachments
// are accepted ONLY when ReporterID is non-nil (media is logged-in
// only). RawIP is hashed with the configured salt — never stored raw.
type SubmitReportInput struct {
	ReporterID    *uuid.UUID
	Type          string
	Title         string
	Description   string
	PageURL       string
	Context       map[string]any
	ReporterEmail string
	RawIP         string
	Attachments   []AttachmentRef
}

// SubmitReport validates the input, builds the domain report (+ any
// attachments), and persists everything atomically. Business rules
// enforced here on top of the domain validation:
//
//   - Anonymous submissions MUST NOT carry attachments (media is
//     logged-in only). An anonymous submit with attachments is rejected
//     with domain.ErrAttachmentNotAllowed.
//   - Each attachment is re-validated against the content-type allowlist
//     and per-kind size cap, and its object key extension is checked —
//     a client cannot persist an object outside the allowlist by
//     tampering with the key it PUT to.
//
// The submitter IP is hashed (sha256(ip + salt)) before storage so the
// raw IP never lands in the database (RGPD). When the IP is empty the
// hash is left empty rather than hashing the bare salt.
func (s *Service) SubmitReport(ctx context.Context, in SubmitReportInput) (*domain.Report, error) {
	report, err := domain.NewReport(domain.NewReportInput{
		ReporterID:    in.ReporterID,
		Type:          domain.ReportType(in.Type),
		Title:         in.Title,
		Description:   in.Description,
		PageURL:       in.PageURL,
		Context:       in.Context,
		ReporterEmail: in.ReporterEmail,
		IPHash:        s.hashIP(in.RawIP),
	})
	if err != nil {
		return nil, err
	}

	attachments, err := s.buildAttachments(report, in.ReporterID, in.Attachments)
	if err != nil {
		return nil, err
	}

	if err := s.reports.CreateReport(ctx, report, attachments); err != nil {
		return nil, fmt.Errorf("submit report: persist: %w", err)
	}
	return report, nil
}

// buildAttachments validates and constructs the attachment rows for a
// submission. Returns domain.ErrAttachmentNotAllowed when an anonymous
// reporter supplies any attachment. Returns nil (no error) when there
// are no attachments.
func (s *Service) buildAttachments(
	report *domain.Report,
	reporterID *uuid.UUID,
	refs []AttachmentRef,
) ([]*domain.Attachment, error) {
	if len(refs) == 0 {
		return nil, nil
	}
	if reporterID == nil {
		// Media is logged-in only — an anonymous payload carrying
		// attachment keys is a contract violation.
		return nil, domain.ErrAttachmentNotAllowed
	}

	out := make([]*domain.Attachment, 0, len(refs))
	for _, ref := range refs {
		kind := domain.AttachmentKind(ref.Kind)
		// Re-validate the declared kind / content-type / size against
		// the same allowlist the presign step used, so a tampered
		// submit body cannot smuggle an oversized or disallowed object.
		if _, _, err := domain.ValidatePresign(kind, ref.ContentType, ref.SizeBytes); err != nil {
			return nil, err
		}
		if !domain.IsAllowedObjectExtension(ref.ObjectKey) {
			return nil, domain.ErrUnsupportedContentType
		}
		attachment, err := domain.NewAttachment(domain.NewAttachmentInput{
			ReportID:    report.ID,
			Kind:        kind,
			ObjectKey:   ref.ObjectKey,
			ContentType: ref.ContentType,
			SizeBytes:   ref.SizeBytes,
		})
		if err != nil {
			return nil, err
		}
		out = append(out, attachment)
	}
	return out, nil
}

// hashIP returns sha256(ip + salt) as lowercase hex, or "" when the IP
// is empty. The raw IP is never persisted (RGPD); the salted hash lets
// an operator correlate repeat submissions from the same source without
// retaining the address itself.
func (s *Service) hashIP(rawIP string) string {
	ip := strings.TrimSpace(rawIP)
	if ip == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(ip + s.salt))
	return hex.EncodeToString(sum[:])
}
