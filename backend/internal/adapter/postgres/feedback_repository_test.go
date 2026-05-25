package postgres_test

// Integration tests for the platform feedback repository (migration
// 156: bug_reports, bug_report_attachments, bug_report_notes). Gated on
// MARKETPLACE_TEST_DATABASE_URL via the shared testDB helper — the
// whole suite skips when the variable is unset (CI / fresh checkouts).
//
//	MARKETPLACE_TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5435/marketplace_go_feat_feedback?sslmode=disable \
//	  go test ./internal/adapter/postgres/ -run TestFeedbackRepository -count=1
//
// Every test creates rows with random ids and cleans them up in
// t.Cleanup so reruns stay isolated; it never touches existing rows.

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"marketplace-backend/internal/adapter/postgres"
	"marketplace-backend/internal/domain/feedback"
	"marketplace-backend/internal/port/repository"
)

func newFeedbackRepo(t *testing.T) (*postgres.FeedbackRepository, func(reportID uuid.UUID)) {
	t.Helper()
	db := testDB(t)
	repo := postgres.NewFeedbackRepository(db)
	cleanup := func(reportID uuid.UUID) {
		// ON DELETE CASCADE removes attachments + notes with the parent.
		_, _ = db.Exec(`DELETE FROM bug_reports WHERE id = $1`, reportID)
	}
	return repo, cleanup
}

func mustReport(t *testing.T, in feedback.NewReportInput) *feedback.Report {
	t.Helper()
	r, err := feedback.NewReport(in)
	require.NoError(t, err)
	return r
}

func TestFeedbackRepository_CreateAnonymousTextReport(t *testing.T) {
	repo, cleanup := newFeedbackRepo(t)
	ctx := context.Background()

	report := mustReport(t, feedback.NewReportInput{
		Type:        feedback.TypeBug,
		Title:       "Anonymous bug report",
		Description: "An anonymous visitor reports a bug with enough detail.",
		Context:     map[string]any{"role": "anonymous", "locale": "fr"},
		IPHash:      "deadbeefdeadbeef",
	})
	t.Cleanup(func() { cleanup(report.ID) })

	require.NoError(t, repo.CreateReport(ctx, report, nil))

	got, err := repo.GetReport(ctx, report.ID)
	require.NoError(t, err)
	assert.True(t, got.IsAnonymous(), "anonymous report has nil reporter")
	assert.Equal(t, feedback.TypeBug, got.Type)
	assert.Equal(t, feedback.StatusNew, got.Status)
	assert.Equal(t, feedback.Severity(""), got.Severity, "fresh report has no severity")
	assert.Equal(t, "deadbeefdeadbeef", got.IPHash)
	assert.Equal(t, "fr", got.Context["locale"])
}

func TestFeedbackRepository_CreateWithAttachments(t *testing.T) {
	repo, cleanup := newFeedbackRepo(t)
	ctx := context.Background()
	db := testDB(t)
	reporterID := insertTestUser(t, db)

	report := mustReport(t, feedback.NewReportInput{
		ReporterID:  &reporterID,
		Type:        feedback.TypeSecurity,
		Title:       "Security report with media",
		Description: "Authenticated reporter attaches a screenshot and a clip.",
	})
	t.Cleanup(func() { cleanup(report.ID) })

	img, err := feedback.NewAttachment(feedback.NewAttachmentInput{
		ReportID: report.ID, Kind: feedback.AttachmentImage,
		ObjectKey: "reports/" + report.ID.String() + "/a.png", ContentType: "image/png", SizeBytes: 2048,
	})
	require.NoError(t, err)
	vid, err := feedback.NewAttachment(feedback.NewAttachmentInput{
		ReportID: report.ID, Kind: feedback.AttachmentVideo,
		ObjectKey: "reports/" + report.ID.String() + "/b.mp4", ContentType: "video/mp4", SizeBytes: 5 << 20,
	})
	require.NoError(t, err)

	require.NoError(t, repo.CreateReport(ctx, report, []*feedback.Attachment{img, vid}))

	attachments, err := repo.ListAttachments(ctx, report.ID)
	require.NoError(t, err)
	require.Len(t, attachments, 2)
	assert.Equal(t, feedback.AttachmentImage, attachments[0].Kind)
	assert.Equal(t, feedback.AttachmentVideo, attachments[1].Kind)

	got, err := repo.GetReport(ctx, report.ID)
	require.NoError(t, err)
	require.NotNil(t, got.ReporterID)
	assert.Equal(t, reporterID, *got.ReporterID)
}

func TestFeedbackRepository_GetNotFound(t *testing.T) {
	repo, _ := newFeedbackRepo(t)
	_, err := repo.GetReport(context.Background(), uuid.New())
	assert.ErrorIs(t, err, feedback.ErrNotFound)
}

func TestFeedbackRepository_UpdateResolves(t *testing.T) {
	repo, cleanup := newFeedbackRepo(t)
	ctx := context.Background()
	db := testDB(t)
	adminID := insertTestUser(t, db)

	report := mustReport(t, feedback.NewReportInput{
		Type: feedback.TypeBug, Title: "To be resolved",
		Description: "This report will be triaged and then resolved by an admin.",
	})
	t.Cleanup(func() { cleanup(report.ID) })
	require.NoError(t, repo.CreateReport(ctx, report, nil))

	resolved := feedback.StatusResolved
	high := feedback.SeverityHigh
	require.NoError(t, feedback.Update{Status: &resolved, Severity: &high}.Apply(report, adminID, time.Now()))
	require.NoError(t, repo.UpdateReport(ctx, report))

	got, err := repo.GetReport(ctx, report.ID)
	require.NoError(t, err)
	assert.Equal(t, feedback.StatusResolved, got.Status)
	assert.Equal(t, feedback.SeverityHigh, got.Severity)
	require.NotNil(t, got.ResolvedAt)
	require.NotNil(t, got.ResolvedBy)
	assert.Equal(t, adminID, *got.ResolvedBy)
}

func TestFeedbackRepository_UpdateNotFound(t *testing.T) {
	repo, _ := newFeedbackRepo(t)
	ghost := mustReport(t, feedback.NewReportInput{
		Type: feedback.TypeBug, Title: "ghost report",
		Description: "this report was never persisted to the database at all",
	})
	assert.ErrorIs(t, repo.UpdateReport(context.Background(), ghost), feedback.ErrNotFound)
}

func TestFeedbackRepository_NotesNewestFirst(t *testing.T) {
	repo, cleanup := newFeedbackRepo(t)
	ctx := context.Background()
	db := testDB(t)
	adminID := insertTestUser(t, db)

	report := mustReport(t, feedback.NewReportInput{
		Type: feedback.TypeBug, Title: "With notes",
		Description: "An admin will add two internal notes to this report.",
	})
	t.Cleanup(func() { cleanup(report.ID) })
	require.NoError(t, repo.CreateReport(ctx, report, nil))

	first, err := feedback.NewNote(feedback.NewNoteInput{ReportID: report.ID, AdminUserID: adminID, Body: "first note"})
	require.NoError(t, err)
	require.NoError(t, repo.AddNote(ctx, first))
	second, err := feedback.NewNote(feedback.NewNoteInput{ReportID: report.ID, AdminUserID: adminID, Body: "second note"})
	require.NoError(t, err)
	require.NoError(t, repo.AddNote(ctx, second))

	notes, err := repo.ListNotes(ctx, report.ID)
	require.NoError(t, err)
	require.Len(t, notes, 2)
	// Newest first — the second note created sorts ahead.
	assert.Equal(t, "second note", notes[0].Body)
	assert.Equal(t, "first note", notes[1].Body)
}

func TestFeedbackRepository_ListFiltersAndCounts(t *testing.T) {
	repo, cleanup := newFeedbackRepo(t)
	ctx := context.Background()
	db := testDB(t)
	reporterID := insertTestUser(t, db)

	// A bug report with one attachment + one note.
	bug := mustReport(t, feedback.NewReportInput{
		ReporterID: &reporterID, Type: feedback.TypeBug, Title: "ListFilter bug uniquetoken",
		Description: "A bug report used to assert the list filter + counts.",
	})
	t.Cleanup(func() { cleanup(bug.ID) })
	img, _ := feedback.NewAttachment(feedback.NewAttachmentInput{
		ReportID: bug.ID, Kind: feedback.AttachmentImage,
		ObjectKey: "reports/" + bug.ID.String() + "/a.png", ContentType: "image/png", SizeBytes: 1024,
	})
	require.NoError(t, repo.CreateReport(ctx, bug, []*feedback.Attachment{img}))
	note, _ := feedback.NewNote(feedback.NewNoteInput{ReportID: bug.ID, AdminUserID: reporterID, Body: "a note"})
	require.NoError(t, repo.AddNote(ctx, note))

	// A security report (different type) with no attachments/notes.
	sec := mustReport(t, feedback.NewReportInput{
		Type: feedback.TypeSecurity, Title: "ListFilter security uniquetoken",
		Description: "A security report used to assert the type filter excludes it.",
	})
	t.Cleanup(func() { cleanup(sec.ID) })
	require.NoError(t, repo.CreateReport(ctx, sec, nil))

	// Filter type=bug + search on the shared token: only the bug matches.
	summaries, _, err := repo.ListReports(ctx, repository.ReportFilter{
		Type: string(feedback.TypeBug), Search: "uniquetoken",
	}, "", 50)
	require.NoError(t, err)

	var found *repository.ReportSummary
	for _, s := range summaries {
		if s.Report.ID == bug.ID {
			found = s
		}
		assert.NotEqual(t, sec.ID, s.Report.ID, "type=bug filter must exclude the security report")
	}
	require.NotNil(t, found, "the bug report must appear in the filtered list")
	assert.Equal(t, 1, found.AttachmentCount)
	assert.Equal(t, 1, found.NoteCount)
}

func TestFeedbackRepository_ListPaginationCursor(t *testing.T) {
	repo, cleanup := newFeedbackRepo(t)
	ctx := context.Background()

	// Create 3 reports sharing a unique search token so the listing is
	// deterministic and isolated from other rows.
	token := "pgtoken-" + uuid.New().String()[:8]
	ids := make([]uuid.UUID, 0, 3)
	for i := 0; i < 3; i++ {
		r := mustReport(t, feedback.NewReportInput{
			Type: feedback.TypeBug, Title: token + " report",
			Description: "Pagination probe report number with sufficient length.",
		})
		require.NoError(t, repo.CreateReport(ctx, r, nil))
		ids = append(ids, r.ID)
		t.Cleanup(func() { cleanup(r.ID) })
	}

	// Page size 2 → first page returns 2 + a next cursor.
	page1, cursor, err := repo.ListReports(ctx, repository.ReportFilter{Search: token}, "", 2)
	require.NoError(t, err)
	require.Len(t, page1, 2)
	require.NotEmpty(t, cursor, "a third row exists so a next cursor must be issued")

	// Second page returns the remaining row, no further cursor.
	page2, cursor2, err := repo.ListReports(ctx, repository.ReportFilter{Search: token}, cursor, 2)
	require.NoError(t, err)
	require.Len(t, page2, 1)
	assert.Empty(t, cursor2, "no further rows so the cursor must be empty")
	assert.Len(t, ids, 3)
}
