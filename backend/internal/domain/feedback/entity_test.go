package feedback_test

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"marketplace-backend/internal/domain/feedback"
)

func ptrUUID(id uuid.UUID) *uuid.UUID { return &id }

func TestReportType_IsValid(t *testing.T) {
	tests := []struct {
		name string
		in   feedback.ReportType
		want bool
	}{
		{"bug", feedback.TypeBug, true},
		{"security", feedback.TypeSecurity, true},
		{"empty", feedback.ReportType(""), false},
		{"unknown", feedback.ReportType("feature"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.in.IsValid())
		})
	}
}

func TestReportStatus_IsValid_AndTerminal(t *testing.T) {
	tests := []struct {
		name         string
		in           feedback.ReportStatus
		wantValid    bool
		wantTerminal bool
	}{
		{"new", feedback.StatusNew, true, false},
		{"triaged", feedback.StatusTriaged, true, false},
		{"in_progress", feedback.StatusInProgress, true, false},
		{"resolved", feedback.StatusResolved, true, true},
		{"rejected", feedback.StatusRejected, true, true},
		{"empty", feedback.ReportStatus(""), false, false},
		{"unknown", feedback.ReportStatus("closed"), false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantValid, tt.in.IsValid())
			assert.Equal(t, tt.wantTerminal, tt.in.IsTerminal())
		})
	}
}

func TestSeverity_IsValid(t *testing.T) {
	tests := []struct {
		name string
		in   feedback.Severity
		want bool
	}{
		{"low", feedback.SeverityLow, true},
		{"medium", feedback.SeverityMedium, true},
		{"high", feedback.SeverityHigh, true},
		{"critical", feedback.SeverityCritical, true},
		{"empty is not valid", feedback.Severity(""), false},
		{"unknown", feedback.Severity("blocker"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.in.IsValid())
		})
	}
}

func TestAttachmentKind_IsValid(t *testing.T) {
	assert.True(t, feedback.AttachmentImage.IsValid())
	assert.True(t, feedback.AttachmentVideo.IsValid())
	assert.False(t, feedback.AttachmentKind("").IsValid())
	assert.False(t, feedback.AttachmentKind("audio").IsValid())
}

func TestNewReport(t *testing.T) {
	validTitle := "Login button does nothing"
	validDesc := "When I click the login button on the homepage nothing happens at all."

	tests := []struct {
		name    string
		in      feedback.NewReportInput
		wantErr error
	}{
		{
			name: "valid anonymous bug",
			in: feedback.NewReportInput{
				Type:        feedback.TypeBug,
				Title:       validTitle,
				Description: validDesc,
			},
		},
		{
			name: "valid authenticated security report",
			in: feedback.NewReportInput{
				ReporterID:  ptrUUID(uuid.New()),
				Type:        feedback.TypeSecurity,
				Title:       validTitle,
				Description: validDesc,
				PageURL:     "https://app.example.com/login",
				Context:     map[string]any{"role": "agency"},
			},
		},
		{
			name: "invalid type",
			in: feedback.NewReportInput{
				Type:        feedback.ReportType("feature"),
				Title:       validTitle,
				Description: validDesc,
			},
			wantErr: feedback.ErrInvalidType,
		},
		{
			name: "title too short",
			in: feedback.NewReportInput{
				Type:        feedback.TypeBug,
				Title:       "ab",
				Description: validDesc,
			},
			wantErr: feedback.ErrTitleTooShort,
		},
		{
			name: "title whitespace only is too short",
			in: feedback.NewReportInput{
				Type:        feedback.TypeBug,
				Title:       "      ",
				Description: validDesc,
			},
			wantErr: feedback.ErrTitleTooShort,
		},
		{
			name: "title too long",
			in: feedback.NewReportInput{
				Type:        feedback.TypeBug,
				Title:       strings.Repeat("x", feedback.MaxTitleLength+1),
				Description: validDesc,
			},
			wantErr: feedback.ErrTitleTooLong,
		},
		{
			name: "description too short",
			in: feedback.NewReportInput{
				Type:        feedback.TypeBug,
				Title:       validTitle,
				Description: "short",
			},
			wantErr: feedback.ErrDescriptionTooShort,
		},
		{
			name: "description too long",
			in: feedback.NewReportInput{
				Type:        feedback.TypeBug,
				Title:       validTitle,
				Description: strings.Repeat("y", feedback.MaxDescriptionLength+1),
			},
			wantErr: feedback.ErrDescriptionTooLong,
		},
		{
			name: "page url too long",
			in: feedback.NewReportInput{
				Type:        feedback.TypeBug,
				Title:       validTitle,
				Description: validDesc,
				PageURL:     "https://x/" + strings.Repeat("p", feedback.MaxPageURLLength),
			},
			wantErr: feedback.ErrPageURLTooLong,
		},
		{
			name: "reporter email too long",
			in: feedback.NewReportInput{
				Type:          feedback.TypeBug,
				Title:         validTitle,
				Description:   validDesc,
				ReporterEmail: strings.Repeat("e", feedback.MaxReporterEmailLength+1),
			},
			wantErr: feedback.ErrReporterEmailTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := feedback.NewReport(tt.in)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, r)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, r)
			assert.NotEqual(t, uuid.Nil, r.ID)
			assert.Equal(t, feedback.StatusNew, r.Status)
			assert.Equal(t, feedback.Severity(""), r.Severity)
			assert.NotNil(t, r.Context, "context must always be non-nil")
			assert.False(t, r.CreatedAt.IsZero())
			assert.Equal(t, r.CreatedAt, r.UpdatedAt)
			assert.Equal(t, tt.in.ReporterID == nil, r.IsAnonymous())
		})
	}
}

func TestNewReport_TrimsTitleAndDescription(t *testing.T) {
	r, err := feedback.NewReport(feedback.NewReportInput{
		Type:        feedback.TypeBug,
		Title:       "   Trimmed title   ",
		Description: "   This is a description that needs trimming.   ",
	})
	require.NoError(t, err)
	assert.Equal(t, "Trimmed title", r.Title)
	assert.Equal(t, "This is a description that needs trimming.", r.Description)
}

func TestNewAttachment(t *testing.T) {
	reportID := uuid.New()
	tests := []struct {
		name    string
		in      feedback.NewAttachmentInput
		wantErr error
	}{
		{
			name: "valid image",
			in: feedback.NewAttachmentInput{
				ReportID:    reportID,
				Kind:        feedback.AttachmentImage,
				ObjectKey:   "reports/abc/def.png",
				ContentType: "image/png",
				SizeBytes:   1024,
			},
		},
		{
			name: "missing report id",
			in: feedback.NewAttachmentInput{
				Kind:      feedback.AttachmentImage,
				ObjectKey: "reports/abc/def.png",
			},
			wantErr: feedback.ErrMissingReport,
		},
		{
			name: "invalid kind",
			in: feedback.NewAttachmentInput{
				ReportID:  reportID,
				Kind:      feedback.AttachmentKind("audio"),
				ObjectKey: "reports/abc/def.png",
			},
			wantErr: feedback.ErrInvalidAttachmentKind,
		},
		{
			name: "missing object key",
			in: feedback.NewAttachmentInput{
				ReportID:  reportID,
				Kind:      feedback.AttachmentImage,
				ObjectKey: "   ",
			},
			wantErr: feedback.ErrMissingObjectKey,
		},
		{
			name: "negative size",
			in: feedback.NewAttachmentInput{
				ReportID:  reportID,
				Kind:      feedback.AttachmentVideo,
				ObjectKey: "reports/abc/def.mp4",
				SizeBytes: -5,
			},
			wantErr: feedback.ErrInvalidAttachmentSize,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := feedback.NewAttachment(tt.in)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, a)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, a)
			assert.NotEqual(t, uuid.Nil, a.ID)
			assert.Equal(t, reportID, a.ReportID)
		})
	}
}

func TestNewNote(t *testing.T) {
	reportID := uuid.New()
	adminID := uuid.New()
	tests := []struct {
		name    string
		in      feedback.NewNoteInput
		wantErr error
	}{
		{
			name: "valid note",
			in:   feedback.NewNoteInput{ReportID: reportID, AdminUserID: adminID, Body: "Reproduced on staging."},
		},
		{
			name:    "missing report",
			in:      feedback.NewNoteInput{AdminUserID: adminID, Body: "x"},
			wantErr: feedback.ErrMissingReport,
		},
		{
			name:    "missing admin",
			in:      feedback.NewNoteInput{ReportID: reportID, Body: "x"},
			wantErr: feedback.ErrMissingAdmin,
		},
		{
			name:    "empty body",
			in:      feedback.NewNoteInput{ReportID: reportID, AdminUserID: adminID, Body: "   "},
			wantErr: feedback.ErrNoteBodyEmpty,
		},
		{
			name:    "body too long",
			in:      feedback.NewNoteInput{ReportID: reportID, AdminUserID: adminID, Body: strings.Repeat("z", feedback.MaxNoteBodyLength+1)},
			wantErr: feedback.ErrNoteBodyTooLong,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, err := feedback.NewNote(tt.in)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, n)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, n)
			assert.Equal(t, "Reproduced on staging.", n.Body)
		})
	}
}

func TestUpdate_Apply(t *testing.T) {
	adminID := uuid.New()
	now := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)

	statusPtr := func(s feedback.ReportStatus) *feedback.ReportStatus { return &s }
	sevPtr := func(s feedback.Severity) *feedback.Severity { return &s }

	t.Run("no changes returns error", func(t *testing.T) {
		r := &feedback.Report{Status: feedback.StatusNew}
		err := feedback.Update{}.Apply(r, adminID, now)
		assert.ErrorIs(t, err, feedback.ErrNoUpdateFields)
	})

	t.Run("set severity only", func(t *testing.T) {
		r := &feedback.Report{Status: feedback.StatusNew}
		err := feedback.Update{Severity: sevPtr(feedback.SeverityHigh)}.Apply(r, adminID, now)
		require.NoError(t, err)
		assert.Equal(t, feedback.SeverityHigh, r.Severity)
		assert.Equal(t, feedback.StatusNew, r.Status, "status unchanged")
		assert.Nil(t, r.ResolvedAt)
	})

	t.Run("invalid severity rejected", func(t *testing.T) {
		r := &feedback.Report{Status: feedback.StatusNew}
		err := feedback.Update{Severity: sevPtr(feedback.Severity("blocker"))}.Apply(r, adminID, now)
		assert.ErrorIs(t, err, feedback.ErrInvalidSeverity)
	})

	t.Run("clear severity with empty string", func(t *testing.T) {
		r := &feedback.Report{Status: feedback.StatusNew, Severity: feedback.SeverityHigh}
		err := feedback.Update{Severity: sevPtr(feedback.Severity(""))}.Apply(r, adminID, now)
		require.NoError(t, err)
		assert.Equal(t, feedback.Severity(""), r.Severity)
	})

	t.Run("invalid status rejected", func(t *testing.T) {
		r := &feedback.Report{Status: feedback.StatusNew}
		err := feedback.Update{Status: statusPtr(feedback.ReportStatus("closed"))}.Apply(r, adminID, now)
		assert.ErrorIs(t, err, feedback.ErrInvalidStatus)
	})

	t.Run("transition to resolved stamps resolver", func(t *testing.T) {
		r := &feedback.Report{Status: feedback.StatusInProgress}
		err := feedback.Update{Status: statusPtr(feedback.StatusResolved)}.Apply(r, adminID, now)
		require.NoError(t, err)
		assert.Equal(t, feedback.StatusResolved, r.Status)
		require.NotNil(t, r.ResolvedAt)
		assert.Equal(t, now, *r.ResolvedAt)
		require.NotNil(t, r.ResolvedBy)
		assert.Equal(t, adminID, *r.ResolvedBy)
		assert.Equal(t, now, r.UpdatedAt)
	})

	t.Run("transition to rejected stamps resolver", func(t *testing.T) {
		r := &feedback.Report{Status: feedback.StatusNew}
		err := feedback.Update{Status: statusPtr(feedback.StatusRejected)}.Apply(r, adminID, now)
		require.NoError(t, err)
		require.NotNil(t, r.ResolvedAt)
		require.NotNil(t, r.ResolvedBy)
	})

	t.Run("leaving terminal state clears resolution stamp", func(t *testing.T) {
		resolvedAt := now.Add(-time.Hour)
		resolver := uuid.New()
		r := &feedback.Report{Status: feedback.StatusResolved, ResolvedAt: &resolvedAt, ResolvedBy: &resolver}
		err := feedback.Update{Status: statusPtr(feedback.StatusInProgress)}.Apply(r, adminID, now)
		require.NoError(t, err)
		assert.Equal(t, feedback.StatusInProgress, r.Status)
		assert.Nil(t, r.ResolvedAt)
		assert.Nil(t, r.ResolvedBy)
	})

	t.Run("set status and severity together", func(t *testing.T) {
		r := &feedback.Report{Status: feedback.StatusNew}
		err := feedback.Update{
			Status:   statusPtr(feedback.StatusTriaged),
			Severity: sevPtr(feedback.SeverityCritical),
		}.Apply(r, adminID, now)
		require.NoError(t, err)
		assert.Equal(t, feedback.StatusTriaged, r.Status)
		assert.Equal(t, feedback.SeverityCritical, r.Severity)
	})
}
