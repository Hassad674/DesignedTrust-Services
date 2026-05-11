package postgres

// sqlmock-backed tests for BatchAuditWriter's actual flush path.
// These exercise the real executeBatchInsert SQL, the threshold
// trigger, the interval trigger, the shutdown drain, and the retry
// chain on a transient BEGIN failure.
//
// Why a separate file: keeps the lightweight unit tests in the main
// test file fast (no sqlmock setup), and concentrates the SQL-level
// assertions here where the expected query string is the load-bearing
// piece.

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newMockSink builds a *sql.DB / sqlmock pair and returns both. The
// DB satisfies auditBatchSink because *sql.DB has a BeginTx method
// with the same signature.
func newMockSink(t *testing.T) (auditBatchSink, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	return db, mock, func() { _ = db.Close() }
}

// expectMultiRowInsert sets up sqlmock to accept ONE multi-row INSERT
// transaction with `rowCount` rows. Returns the regex used so the
// test can assert the SQL shape if needed.
func expectMultiRowInsert(mock sqlmock.Sqlmock, rowCount int) *regexp.Regexp {
	// Build the expected SQL prefix. The exact query is constructed
	// in executeBatchInsert; we anchor on the deterministic prefix and
	// trust that the row-count expansion fed sqlmock the right number
	// of args.
	mock.ExpectBegin()
	// Match the query loosely — we configured QueryMatcherEqual at
	// New() time, but the multi-row VALUES list varies. Switch the
	// matcher to regex per-call.
	pattern := regexp.MustCompile(`(?s)INSERT INTO audit_logs.*VALUES\s*\(\$1.*`)
	exec := mock.ExpectExec(pattern.String()).WillReturnResult(sqlmock.NewResult(0, int64(rowCount)))
	_ = exec
	mock.ExpectCommit()
	return pattern
}

// ─── 1. Threshold-triggered flush ────────────────────────────────────

func TestBatchAuditWriter_FlushesOnThreshold(t *testing.T) {
	sink, mock, closeFn := newMockSink(t)
	defer closeFn()
	mock.MatchExpectationsInOrder(true)

	// Configure: threshold=5, very long interval. Only the threshold
	// can trigger the flush.
	cfg := BatchAuditConfig{
		FlushInterval:   1 * time.Hour,
		FlushThreshold:  5,
		ChannelCapacity: 10,
		FlushTimeout:    1 * time.Second,
	}

	// Switch sqlmock to regex matcher so it accepts the multi-row INSERT.
	// We rebuild the mock with the regex matcher.
	db, mock2, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()
	sink = db

	mock2.ExpectBegin()
	mock2.ExpectExec(`INSERT INTO audit_logs.*VALUES`).WillReturnResult(sqlmock.NewResult(0, 5))
	mock2.ExpectCommit()

	inner := &fakeAuditRepo{}
	w := NewBatchAuditWriter(inner, sink, cfg)

	var flushed atomic.Int32
	done := make(chan struct{}, 1)
	w.SetOnFlush(func(n int) {
		flushed.Add(int32(n))
		select {
		case done <- struct{}{}:
		default:
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w.Start(ctx)

	for i := 0; i < 5; i++ {
		require.NoError(t, w.Log(ctx, newTestEntry(i)))
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("flush did not fire within 2s")
	}

	w.Stop(2 * time.Second)
	assert.Equal(t, int32(5), flushed.Load())
	assert.NoError(t, mock2.ExpectationsWereMet())
}

// ─── 2. Interval-triggered flush ─────────────────────────────────────

func TestBatchAuditWriter_FlushesOnInterval(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	cfg := BatchAuditConfig{
		FlushInterval:   50 * time.Millisecond,
		FlushThreshold:  100,
		ChannelCapacity: 10,
		FlushTimeout:    1 * time.Second,
	}

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO audit_logs.*VALUES`).WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	inner := &fakeAuditRepo{}
	w := NewBatchAuditWriter(inner, db, cfg)

	var flushed atomic.Int32
	done := make(chan struct{}, 1)
	w.SetOnFlush(func(n int) {
		flushed.Add(int32(n))
		select {
		case done <- struct{}{}:
		default:
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w.Start(ctx)

	// Log only 2 entries — below threshold. Wait for the interval to fire.
	require.NoError(t, w.Log(ctx, newTestEntry(1)))
	require.NoError(t, w.Log(ctx, newTestEntry(2)))

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("interval flush did not fire within 2s")
	}

	w.Stop(2 * time.Second)
	assert.Equal(t, int32(2), flushed.Load())
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ─── 3. Shutdown drains buffered entries ─────────────────────────────

func TestBatchAuditWriter_ShutdownDrainsBuffer(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	cfg := BatchAuditConfig{
		FlushInterval:   1 * time.Hour, // disabled
		FlushThreshold:  100,           // disabled
		ChannelCapacity: 10,
		FlushTimeout:    1 * time.Second,
	}

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO audit_logs.*VALUES`).WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectCommit()

	inner := &fakeAuditRepo{}
	w := NewBatchAuditWriter(inner, db, cfg)

	var flushed atomic.Int32
	w.SetOnFlush(func(n int) { flushed.Add(int32(n)) })

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w.Start(ctx)

	require.NoError(t, w.Log(ctx, newTestEntry(1)))
	require.NoError(t, w.Log(ctx, newTestEntry(2)))
	require.NoError(t, w.Log(ctx, newTestEntry(3)))

	// Stop must drain remaining entries before returning.
	w.Stop(2 * time.Second)

	assert.Equal(t, int32(3), flushed.Load(), "Stop must drain buffered entries")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ─── 4. Retry on transient BEGIN failure ─────────────────────────────

func TestBatchAuditWriter_RetriesOnTransientBeginFailure(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	cfg := BatchAuditConfig{
		FlushInterval:            1 * time.Hour,
		FlushThreshold:           2,
		ChannelCapacity:          10,
		MaxRetriesOnFlushFailure: 3,
		FlushTimeout:             3 * time.Second,
	}

	// First Begin fails, second succeeds.
	mock.ExpectBegin().WillReturnError(errors.New("transient: connection reset"))
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO audit_logs.*VALUES`).WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	inner := &fakeAuditRepo{}
	w := NewBatchAuditWriter(inner, db, cfg)

	var flushed atomic.Int32
	done := make(chan struct{}, 1)
	w.SetOnFlush(func(n int) {
		flushed.Add(int32(n))
		select {
		case done <- struct{}{}:
		default:
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w.Start(ctx)

	require.NoError(t, w.Log(ctx, newTestEntry(1)))
	require.NoError(t, w.Log(ctx, newTestEntry(2)))

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("retry-then-success flush did not complete within 3s")
	}

	w.Stop(2 * time.Second)
	assert.Equal(t, int32(2), flushed.Load())
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ─── 5. Drop after exhausting retries ────────────────────────────────

func TestBatchAuditWriter_DropsBatchAfterExhaustingRetries(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	cfg := BatchAuditConfig{
		FlushInterval:            1 * time.Hour,
		FlushThreshold:           1,
		ChannelCapacity:          10,
		MaxRetriesOnFlushFailure: 1,
		FlushTimeout:             1 * time.Second,
	}

	// All BEGINs fail.
	for i := 0; i < cfg.MaxRetriesOnFlushFailure+1; i++ {
		mock.ExpectBegin().WillReturnError(fmt.Errorf("persistent fail #%d", i))
	}

	inner := &fakeAuditRepo{}
	w := NewBatchAuditWriter(inner, db, cfg)

	var flushed atomic.Int32
	w.SetOnFlush(func(n int) { flushed.Add(int32(n)) })

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w.Start(ctx)
	require.NoError(t, w.Log(ctx, newTestEntry(1)))

	// Wait for the retries to exhaust.
	time.Sleep(500 * time.Millisecond)
	w.Stop(2 * time.Second)

	assert.Equal(t, int32(0), flushed.Load(), "no successful flush should be reported")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ─── 6. Order preservation under buffered flush ──────────────────────

// args layout for assertion: each entry's 8 columns are appended in
// order to the args slice. So entry seq=k must appear at positions
// (k*8)..(k*8+7) in the args list.

func TestBatchAuditWriter_PreservesOrderInBatch(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	cfg := BatchAuditConfig{
		FlushInterval:   1 * time.Hour,
		FlushThreshold:  5,
		ChannelCapacity: 10,
		FlushTimeout:    1 * time.Second,
	}

	// We can't easily inspect the args from sqlmock without a custom
	// matcher. Instead, we capture the args via the SetOnFlush hook
	// indirectly: we configure a custom-matcher expectation that
	// pulls out the action labels and asserts ordering.
	var capturedActions []string
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO audit_logs.*VALUES`).
		WithArgs(matchingArgs(&capturedActions, 5)...).
		WillReturnResult(sqlmock.NewResult(0, 5))
	mock.ExpectCommit()

	inner := &fakeAuditRepo{}
	w := NewBatchAuditWriter(inner, db, cfg)

	var done = make(chan struct{}, 1)
	w.SetOnFlush(func(n int) {
		select {
		case done <- struct{}{}:
		default:
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w.Start(ctx)

	for i := 0; i < 5; i++ {
		require.NoError(t, w.Log(ctx, newTestEntry(i)))
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("flush did not fire within 2s")
	}
	w.Stop(2 * time.Second)

	// Verify the captured action labels are in sequential order
	// "test.seq_000000", "test.seq_000001", ... — the test entries
	// were appended in that order.
	require.Len(t, capturedActions, 5)
	for i, action := range capturedActions {
		assert.Equal(t, fmt.Sprintf("test.seq_%06d", i), action,
			"entry %d action must match insertion order", i)
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

// matchingArgs returns an []driver.Value with len = 8*rowCount that
// accepts any value via AnyArg, but captures the 3rd column (the
// action string) of each row into *captured for assertion.
func matchingArgs(captured *[]string, rowCount int) []driver.Value {
	const cols = 8
	args := make([]driver.Value, rowCount*cols)
	for r := 0; r < rowCount; r++ {
		for c := 0; c < cols; c++ {
			idx := r*cols + c
			if c == 2 {
				// Action column — capture and accept.
				args[idx] = capturingArg{captured: captured}
			} else {
				args[idx] = sqlmock.AnyArg()
			}
		}
	}
	return args
}

// capturingArg is a sqlmock argument matcher that records the value
// of the column it sees and accepts any value. Implementation pulled
// directly from sqlmock's ArgumentMatcher pattern.
type capturingArg struct {
	captured *[]string
}

// Match implements sqlmock.ArgumentMatcher.
func (c capturingArg) Match(v driver.Value) bool {
	if s, ok := v.(string); ok {
		*c.captured = append(*c.captured, s)
		return true
	}
	return false
}

// ─── 7. Compile-time guard for the AuditRepository interface ─────────

func TestBatchAuditWriter_SatisfiesAuditRepositoryInterface(t *testing.T) {
	// Compile-time test: assigning a *BatchAuditWriter to an
	// interface variable will fail to compile if the interface is
	// not satisfied. var _ repository.AuditRepository =
	// (*BatchAuditWriter)(nil) at the package level already enforces
	// this; the test exists to give a visible test name.
	inner := &fakeAuditRepo{}
	w := NewBatchAuditWriter(inner, newFakeSink(), BatchAuditConfig{})
	_ = w
}

// silence the uuid import when no other tests in this file use it
var _ = uuid.New
