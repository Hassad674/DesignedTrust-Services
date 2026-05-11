package redis_test

// SEC-AUDIT-CACHE-PART2 — adversarial security tests for the QW-HARDENING
// post-merge state of session_version_cache.go. Every test:
//
//   * Lives under the "Security" prefix so it can be selected with
//     `-run "Security"` per the audit's validation pipeline.
//   * Runs under `-race` (the package is exercised with `-race` in CI).
//   * Avoids modifying production code — these tests only OBSERVE the
//     post-merge behaviour and flag any unsafe pattern.
//
// Vectors covered (from sec-audit-cache.md "Attack vectors — test plan"):
//
//   A — Race between bump and read (1000 iterations)
//   B — Invalidate fails but Bump succeeds
//   C — Cache poisoning after Redis recovery
//   D — Singleflight key isolation across uids
//   E — Negative cache poisoning (ErrUserNotFound never cached)
//   F — N/A for session_version (TTL is server-internal, no user input
//       path can influence the cache write call). Covered indirectly
//       by TestSessionVersionCache_DefaultTTLAppliedWhenZero in the
//       existing happy-path suite.
//   G — Concurrent invalidation idempotency (100 parallel Invalidate)

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	adapter "marketplace-backend/internal/adapter/redis"
	"marketplace-backend/internal/domain/user"
)

// versionedSessionChecker is a thread-safe inner that returns whatever
// version the test sets, with optional per-call latency and per-uid
// inner-call accounting. Concurrent callers must coordinate via the
// internal mutex.
type versionedSessionChecker struct {
	mu       sync.Mutex
	versions map[uuid.UUID]int
	calls    map[uuid.UUID]int
	delay    time.Duration
	err      map[uuid.UUID]error
}

func newVersionedSessionChecker() *versionedSessionChecker {
	return &versionedSessionChecker{
		versions: make(map[uuid.UUID]int),
		calls:    make(map[uuid.UUID]int),
		err:      make(map[uuid.UUID]error),
	}
}

func (v *versionedSessionChecker) GetSessionVersion(_ context.Context, uid uuid.UUID) (int, error) {
	if v.delay > 0 {
		time.Sleep(v.delay)
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	v.calls[uid]++
	if e, ok := v.err[uid]; ok && e != nil {
		return 0, e
	}
	return v.versions[uid], nil
}

func (v *versionedSessionChecker) set(uid uuid.UUID, version int) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.versions[uid] = version
}

func (v *versionedSessionChecker) setErr(uid uuid.UUID, err error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.err[uid] = err
}

func (v *versionedSessionChecker) callsTo(uid uuid.UUID) int {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.calls[uid]
}

// newSecSessionCache wires a fresh miniredis + cache for each test.
func newSecSessionCache(
	t *testing.T,
	inner *versionedSessionChecker,
	ttl time.Duration,
) (*adapter.CachedSessionVersionChecker, *miniredis.Miniredis, *goredis.Client) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	cache := adapter.NewCachedSessionVersionChecker(client, inner, ttl)
	return cache, mr, client
}

// ----------------------------------------------------------------------------
// Vector A — Race between bump and read.
//
// Production pattern (post QW-HARDENING):
//   1. inner.BumpSessionVersion(uid)   → DB: session_version++
//   2. cache.Invalidate(ctx, uid)      → Redis: DEL session_version:{uid}
//
// Invariant: AFTER step 2 returns, every subsequent GetSessionVersion(uid)
// MUST return the new (bumped) version. Stale reads BEFORE step 2 are
// inherent to a TTL cache and are not violations.
//
// Test shape: 1000 iterations, in each iteration a writer goroutine
// performs (Bump+Invalidate) while a reader goroutine spins on
// GetSessionVersion. After the writer finishes, a fresh
// GetSessionVersion is issued and must equal the new version.
// ----------------------------------------------------------------------------

func TestSecuritySessionVersion_A_RaceBumpVsRead(t *testing.T) {
	t.Parallel()
	inner := newVersionedSessionChecker()
	cache, _, _ := newSecSessionCache(t, inner, 30*time.Second)
	uid := uuid.New()
	inner.set(uid, 0)
	ctx := context.Background()

	// Warm the cache with version 0.
	v, err := cache.GetSessionVersion(ctx, uid)
	require.NoError(t, err)
	require.Equal(t, 0, v)

	const iterations = 1000
	for i := 1; i <= iterations; i++ {
		// Concurrent reader: must not crash or return a parse error
		// during the bump+invalidate window. The value it sees can be
		// stale (i-1) or fresh (i) — both legal.
		var wg sync.WaitGroup
		wg.Add(1)
		readErrs := make(chan error, 1)
		go func() {
			defer wg.Done()
			_, e := cache.GetSessionVersion(ctx, uid)
			readErrs <- e
		}()

		// Writer: bump + invalidate.
		inner.set(uid, i)
		invErr := cache.Invalidate(ctx, uid)
		require.NoError(t, invErr, "iter %d: invalidate must not error", i)

		wg.Wait()
		select {
		case e := <-readErrs:
			require.NoError(t, e, "iter %d: concurrent read returned error", i)
		default:
		}

		// Post-invalidate read MUST return the new version.
		fresh, ferr := cache.GetSessionVersion(ctx, uid)
		require.NoError(t, ferr)
		require.Equalf(t, i, fresh,
			"iter %d: stale read AFTER Invalidate — got %d, want %d", i, fresh, i)
	}
}

// ----------------------------------------------------------------------------
// Vector B — Invalidate fails but Bump succeeds.
//
// Setup: close the miniredis BEFORE calling Invalidate so every Redis
// command returns a connection error. The inner (DB) is in-memory so
// Bump always succeeds.
//
// Invariant (per the audit plan):
//   GREEN: cache.Invalidate returns an error AND the DB bump remains
//   authoritative AND stale cache disappears after TTL (or via a
//   force-write fallback if implemented).
//   YELLOW: Invalidate returns error; stale cache survives until TTL.
//     Acceptable per current doctrine.
//   RED: Invalidate swallows the error OR returns nil despite a failed DEL.
//
// This test asserts the YELLOW invariant — the BARE MINIMUM the cache
// owes its callers is to surface the DEL failure as a non-nil error so
// the call-site can log/alert.
// ----------------------------------------------------------------------------

func TestSecuritySessionVersion_B_InvalidateFailsSurfacesError(t *testing.T) {
	t.Parallel()
	inner := newVersionedSessionChecker()
	cache, mr, _ := newSecSessionCache(t, inner, 30*time.Second)
	uid := uuid.New()
	inner.set(uid, 1)
	ctx := context.Background()

	// Prime the cache with the pre-bump version.
	v, err := cache.GetSessionVersion(ctx, uid)
	require.NoError(t, err)
	require.Equal(t, 1, v)

	// Simulate Redis going down between Bump and Invalidate.
	mr.Close()

	// Inner Bump: succeeds (in-memory, no Redis dependency).
	inner.set(uid, 2)

	// Cache Invalidate: must surface the Redis DEL failure.
	invErr := cache.Invalidate(ctx, uid)
	assert.Error(t, invErr,
		"vector B: Invalidate must surface Redis DEL failures so the call site can log/alert; "+
			"silently returning nil would hide a stampede of stale-read leaks for up to 30s")
}

// ----------------------------------------------------------------------------
// Vector C — Cache poisoning after Redis recovery.
//
// Setup: Redis fails (SetError) → every cache read is a miss → cache
// falls through to the inner reader → returns correct DB state. Then
// Redis "recovers" (SetError cleared) → next read must be sourced from
// the inner OR from the cache value that was correctly written. In
// either case it MUST be the DB value, NEVER a stale user-supplied
// value.
//
// Invariant: there is no path that lets user-controlled state poison
// the cache after a Redis blip.
// ----------------------------------------------------------------------------

func TestSecuritySessionVersion_C_NoPoisoningAfterRedisRecovery(t *testing.T) {
	t.Parallel()
	inner := newVersionedSessionChecker()
	cache, mr, _ := newSecSessionCache(t, inner, 30*time.Second)
	uid := uuid.New()
	inner.set(uid, 42)
	ctx := context.Background()

	// 1. Redis "fails" — every command returns an error string.
	mr.SetError("ECONNREFUSED simulated")

	// 2. While Redis is down, calls must fall through to inner.
	v, err := cache.GetSessionVersion(ctx, uid)
	require.NoError(t, err)
	assert.Equal(t, 42, v, "vector C: under Redis blip, must fall through to inner")
	assert.GreaterOrEqual(t, inner.callsTo(uid), 1)

	// 3. Redis "recovers".
	mr.SetError("")

	// 4. The next read must reflect the inner state (no stale poison).
	// First call after recovery may be a miss (cache empty since SET
	// failed during the blip) → inner is called again. Then it caches.
	v2, err := cache.GetSessionVersion(ctx, uid)
	require.NoError(t, err)
	assert.Equal(t, 42, v2,
		"vector C: post-recovery read must equal DB value, never a stale poison")

	// 5. Now mutate inner to a new value; cache should serve the
	// CACHED post-recovery value (not the new inner one) because the
	// post-recovery call wrote-through.
	inner.set(uid, 99)
	v3, err := cache.GetSessionVersion(ctx, uid)
	require.NoError(t, err)
	assert.Equal(t, 42, v3,
		"vector C: post-recovery write-through cached the correct value")
}

// ----------------------------------------------------------------------------
// Vector D — Singleflight key isolation across uids.
//
// Setup: 100 concurrent GetSessionVersion(uidA) + 100 concurrent
// GetSessionVersion(uidB), with the inner having a 50ms delay so all
// 200 goroutines land inside the singleflight window. Distinct
// per-uid versions.
//
// Invariants:
//   1. Each uid's inner is called AT MOST 1× (singleflight per-key).
//   2. Every caller receives the version for the uid it asked for —
//      NO cross-key value leak.
// ----------------------------------------------------------------------------

func TestSecuritySessionVersion_D_SingleflightKeyIsolation(t *testing.T) {
	t.Parallel()
	inner := newVersionedSessionChecker()
	inner.delay = 50 * time.Millisecond
	cache, _, _ := newSecSessionCache(t, inner, 30*time.Second)
	uidA := uuid.New()
	uidB := uuid.New()
	inner.set(uidA, 111)
	inner.set(uidB, 222)
	ctx := context.Background()

	const callers = 100
	var gotA, gotB sync.Map
	var eg errgroup.Group
	// Barrier so all 200 goroutines start in parallel.
	start := make(chan struct{})

	for i := 0; i < callers; i++ {
		i := i
		eg.Go(func() error {
			<-start
			v, err := cache.GetSessionVersion(ctx, uidA)
			if err != nil {
				return err
			}
			gotA.Store(i, v)
			return nil
		})
		eg.Go(func() error {
			<-start
			v, err := cache.GetSessionVersion(ctx, uidB)
			if err != nil {
				return err
			}
			gotB.Store(i, v)
			return nil
		})
	}
	close(start)
	require.NoError(t, eg.Wait())

	// Inner called exactly once per uid.
	assert.Equal(t, 1, inner.callsTo(uidA),
		"vector D: singleflight must collapse uidA's 100 concurrent misses to 1 inner call")
	assert.Equal(t, 1, inner.callsTo(uidB),
		"vector D: singleflight must collapse uidB's 100 concurrent misses to 1 inner call")

	// Every caller receives the correct per-uid value (no cross-leak).
	gotA.Range(func(_, v interface{}) bool {
		assert.Equal(t, 111, v, "vector D: uidA caller got wrong value")
		return true
	})
	gotB.Range(func(_, v interface{}) bool {
		assert.Equal(t, 222, v, "vector D: uidB caller got wrong value")
		return true
	})
}

// ----------------------------------------------------------------------------
// Vector E — Negative cache poisoning.
//
// Setup: inner returns user.ErrUserNotFound. Issue 100 concurrent
// GetSessionVersion. Verify NO Redis key exists for the uid (negative
// answers must NEVER be cached) and that every caller receives the
// not-found error.
// ----------------------------------------------------------------------------

func TestSecuritySessionVersion_E_NegativeCacheNotPoisoned(t *testing.T) {
	t.Parallel()
	inner := newVersionedSessionChecker()
	cache, mr, _ := newSecSessionCache(t, inner, 30*time.Second)
	uid := uuid.New()
	inner.setErr(uid, user.ErrUserNotFound)
	ctx := context.Background()

	const callers = 100
	var eg errgroup.Group
	start := make(chan struct{})
	for i := 0; i < callers; i++ {
		eg.Go(func() error {
			<-start
			_, err := cache.GetSessionVersion(ctx, uid)
			if !errors.Is(err, user.ErrUserNotFound) {
				return fmt.Errorf("expected ErrUserNotFound, got %v", err)
			}
			return nil
		})
	}
	close(start)
	require.NoError(t, eg.Wait())

	// Critical: no Redis key was written for the not-found user.
	_, redisErr := mr.Get("session_version:" + uid.String())
	assert.ErrorIs(t, redisErr, miniredis.ErrKeyNotFound,
		"vector E: ErrUserNotFound MUST NOT poison the cache — Redis key must be absent")

	// Singleflight is allowed to coalesce error-returning loads as
	// well (the inner may be called as few as 1× across the 100
	// concurrent callers). The HARD invariant is "no Redis key
	// written" + "every caller sees the error".
	assert.GreaterOrEqual(t, inner.callsTo(uid), 1,
		"vector E: inner must be called at least once")
}

// ----------------------------------------------------------------------------
// Vector G — Concurrent invalidation idempotency.
//
// Setup: 100 goroutines each call cache.Invalidate(uid). Then 1
// final read must observe a fresh inner load (cache absent).
//
// Invariant: 100 concurrent DELs must produce zero errors and leave
// the cache key absent.
// ----------------------------------------------------------------------------

func TestSecuritySessionVersion_G_ConcurrentInvalidateIdempotent(t *testing.T) {
	t.Parallel()
	inner := newVersionedSessionChecker()
	cache, mr, _ := newSecSessionCache(t, inner, 30*time.Second)
	uid := uuid.New()
	inner.set(uid, 7)
	ctx := context.Background()

	// Prime the cache.
	_, err := cache.GetSessionVersion(ctx, uid)
	require.NoError(t, err)

	const invalidators = 100
	var eg errgroup.Group
	var errCount atomic.Int64
	start := make(chan struct{})
	for i := 0; i < invalidators; i++ {
		eg.Go(func() error {
			<-start
			if e := cache.Invalidate(ctx, uid); e != nil {
				errCount.Add(1)
				return e
			}
			return nil
		})
	}
	close(start)
	require.NoError(t, eg.Wait())
	assert.Equal(t, int64(0), errCount.Load(),
		"vector G: 100 concurrent Invalidate calls must all succeed")

	// Final state: key absent.
	_, redisErr := mr.Get("session_version:" + uid.String())
	assert.ErrorIs(t, redisErr, miniredis.ErrKeyNotFound,
		"vector G: after concurrent invalidate burst, the key must be gone")
}
