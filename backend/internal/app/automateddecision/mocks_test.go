package automateddecision_test

import (
	"context"

	"marketplace-backend/internal/port/service"
)

// recordingEmail is a small struct that records the parameters of the
// last SendNotification call. Wrapped by fakeEmailFull so the mock
// satisfies the full service.EmailService surface but only the
// notification path is exercised in the tests.
type recordingEmail struct {
	calls    int
	lastTo   string
	lastSub  string
	lastBody string
	failWith error
}

// fakeEmailFull is the inline mock for service.EmailService. Per
// backend/CLAUDE.md, mocks live in mocks_test.go next to the service
// test, NOT in a generated mock directory.
type fakeEmailFull struct {
	f *recordingEmail
}

// newEmailFake returns a fresh fakeEmailFull with a zeroed recorder.
func newEmailFake() *fakeEmailFull {
	return &fakeEmailFull{f: &recordingEmail{}}
}

// Compile-time assertion: the mock implements the port interface.
var _ service.EmailService = (*fakeEmailFull)(nil)

func (m *fakeEmailFull) SendPasswordReset(ctx context.Context, to, resetURL string) error {
	return nil
}

func (m *fakeEmailFull) SendNotification(ctx context.Context, to, subject, body string) error {
	m.f.calls++
	m.f.lastTo = to
	m.f.lastSub = subject
	m.f.lastBody = body
	return m.f.failWith
}

func (m *fakeEmailFull) SendTeamInvitation(
	ctx context.Context,
	in service.TeamInvitationEmailInput,
) error {
	return nil
}

func (m *fakeEmailFull) SendRolePermissionsChanged(
	ctx context.Context,
	in service.RolePermissionsChangedEmailInput,
) error {
	return nil
}

// Test-only ergonomics for the test file: read the recorded call data
// without exposing the recorder struct directly.
func (m *fakeEmailFull) calls() int       { return m.f.calls }
func (m *fakeEmailFull) lastTo() string   { return m.f.lastTo }
func (m *fakeEmailFull) lastSub() string  { return m.f.lastSub }
func (m *fakeEmailFull) lastBody() string { return m.f.lastBody }
