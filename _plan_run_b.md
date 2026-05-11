# WALLET-UNIFY Run B — Backend Wallet Surfaces — Implementation Plan

Branch: `feat/wallet-unify-run-b` (parent: `main` @ dca95a14)

## Scope summary

Run B is **backend-only**. It delivers three new app/handler surfaces on top of the
foundation Run A merged (`attribution.ended_at` + gate). Web (Run C) and mobile
(Run D) consume the new endpoints later.

Touch list (preview, by layer):

```
domain    : (no changes — milestone/referral/commission entities are already enough)
port      : port/service/referral_wallet.go        — extend with ProjectedCommissions contract
            port/repository/organization_member_repository.go — already has ListMemberUserIDsByOrgIDs
app       : app/referral/wallet_reader.go          — add ProjectedCommissions + ProjectionStatus/Source
            app/referral/wallet_reader_projections.go (new file ≤ 200 LOC for the projection algorithm + helpers)
            app/referral/service.go                — wire MilestoneProjections port (small interface) + OrgMemberLister
handler   : handler/wallet_summary.go              — new file, GET /api/v1/wallet/summary
            handler/wallet_withdraw.go             — new file, POST /api/v1/wallet/withdraw (unified)
            handler/wallet_handler.go              — Deprecation header on RetryCommission (back-compat)
            handler/routes_billing.go              — register summary + withdraw routes; keep commissions/retry
audit     : domain/audit/entity.go                 — add ActionWalletWithdrawExecuted constant
tests     : new tests next to each new file (≥ 90% coverage on new code)
golden    : update testdata/routes.golden + openapi catalog
```

No migrations. No frontend changes. No changes to milestone domain or proposal logic.

## B.1 — ProjectedCommissions

**Location:** new file `backend/internal/app/referral/wallet_reader_projections.go`
(keeps wallet_reader.go under 600 LOC).

**Public API (referral.Service method):**

```go
func (s *Service) ProjectedCommissions(ctx context.Context, orgID uuid.UUID) ([]ProjectedCommission, error)
```

**Types in `wallet_reader.go` (additions):**

```go
type ProjectionStatus string
const (
    ProjectionEscrowed ProjectionStatus = "escrowed"
    ProjectionPending  ProjectionStatus = "pending"
    ProjectionPaid     ProjectionStatus = "paid"
    ProjectionFailed   ProjectionStatus = "failed"
)

type ProjectionSource string
const (
    SourceProjection ProjectionSource = "projection"
    SourceRow        ProjectionSource = "row"
)

type ProjectedCommission struct {
    AttributionID  uuid.UUID
    MilestoneID    uuid.UUID
    ProposalID     uuid.UUID
    MissionTitle   string
    ProjectedCents int64
    Currency       string
    Status         ProjectionStatus
    Source         ProjectionSource
    ProjectedAt    time.Time
}
```

**New ports the projection algorithm depends on (`port/service/`):**

- `MilestonesByProposalReader` (1 method) — `ListByProposals(ctx, []uuid.UUID) (map[uuid.UUID][]MilestoneSnapshot, error)`
- `OrgMemberLister` (1 method) — `ListMemberUserIDsByOrgIDs(ctx, []uuid.UUID) (map[uuid.UUID][]uuid.UUID, error)`
- Re-use existing `repository.ReferralRepository`:
  - `ListByReferrer` for the apporteur's referrals
  - `ListAttributionsByReferralIDs` for batch attribution fetch
  - `ListCommissionsByReferral` for batch commission fetch (one call per referral)
- `ProposalSummaryResolver` (already wired) for mission titles

**Algorithm (`ProjectedCommissions`):**

1. `userIDs := orgMemberLister.ListMemberUserIDsByOrgIDs(ctx, []orgID)[orgID]`
   - Defensive: if empty, return `nil, nil`.
2. For each userID:
   - `referrals, _, _ := repo.ListByReferrer(ctx, userID, {Limit: 200})`
3. Collect all referral IDs → `repo.ListAttributionsByReferralIDs(referralIDs)` (single batch).
4. Group attributions by proposalID → `milestonesByProposal.ListByProposals(proposalIDs)` (single batch).
5. For each referral → `repo.ListCommissionsByReferral(referralID)` — N calls (N = number of referrals; typically ≤ 10). NOT N+1 per milestone (that's the rule that matters).
6. For each attribution:
   - If ended AND milestone approved after ended_at → SKIP (consistent with Run A gate).
   - Else: dispatch milestone status:
     - `draft / rejected / cancelled` → skip (note: milestone domain uses `pending_funding / cancelled` — map to brief's wording).
     - `funded / submitted / disputed` (active) → emit projection, status=Escrowed, source=projection.
     - `approved / released` → look up commission row by milestoneID. If found, emit from row (source=row, status from commission.Status). If not found AND attribution not ended → safety projection (status=Pending).
     - `refunded` → SKIP (clawback territory).
7. Cap output at 200, sorted by ProjectedAt DESC (use milestone.CreatedAt as proxy).
8. Resolve mission titles via `ProposalSummaryResolver` (single batch).

**Status mapping (brief vs. our domain):**

| Brief states         | Our milestone status         | Decision    |
|----------------------|------------------------------|-------------|
| draft                | pending_funding              | skip        |
| in_progress          | funded, submitted, disputed  | escrowed (projection) |
| pending_release      | approved                     | row lookup; pending fallback |
| approved             | approved (post-row creation) | row lookup  |
| released             | released                     | row lookup  |
| paid                 | (commission.Status=paid)     | row, status=paid |
| failed               | (commission.Status=failed)   | row, status=failed |
| rejected             | (no native status)           | skip        |
| cancelled            | cancelled                    | skip        |

**Tests (table-driven, ≥ 90% coverage on new code):**

- `TestProjectedCommissions_AllStatuses` — 7+ rows × commission-exists × ended-attribution mix (~15 cases).
- `TestProjectedCommissions_BatchFetchAvoidsNPlusOne` — query counter via fake; assert ≤ N_REFERRALS+3 calls (1 list refs + 1 list att + 1 list ms + N commissions per ref + 0 mission summaries).
  - Concretely: 5 referrals × 3 milestones → assert ≤ 4 + 5 = 9 calls.
- `TestProjectedCommissions_Cap200` — seed 250 → assert len ≤ 200.
- `TestProjectedCommissions_RespectsEndedGate` — milestone approved AFTER ended_at → skipped; BEFORE ended_at → kept.
- `TestProjectedCommissions_EmptyOrg` — no members → returns nil.

## B.2 — `GET /api/v1/wallet/summary`

**Location:** new file `backend/internal/handler/wallet_summary.go` (~250 LOC).
**Method on `*WalletHandler`:** `Summary(w, r)`.

**Composition:**
- `paymentSvc.GetWalletOverview(ctx, userID, orgID)` — mission side + commission summary (uses existing wallet_reader hooks).
- `referralSvc.ProjectedCommissions(ctx, orgID)` — escrowed_cents for commissions = sum of `Escrowed` + `Pending` projections.
- Build unified `recent_transactions` from:
  - `WalletOverview.Records` (mission), mapped to {type:"mission", amount_cents:r.ProviderPayout, status from PaymentStatus/TransferStatus, occurred_at:CreatedAt, reference_id:r.ID}.
  - `WalletOverview.CommissionRecords` (commission), mapped to {type:"commission", amount_cents:r.CommissionCents, status:r.Status, occurred_at:r.CreatedAt, reference_id:r.ID}.
- Merge sort by occurred_at DESC.
- Cursor: opaque base64 of `{"occurred_at":"…","id":"…"}`. Defaults: limit=20, max=100.

**Response shape:** as specified in brief.

**Wire:** new `commissionProjector` interface dep on WalletHandler:
```go
type commissionProjector interface {
    ProjectedCommissions(ctx context.Context, orgID uuid.UUID) ([]referralapp.ProjectedCommission, error)
}
```
Builder method `WithCommissionProjector` (matches existing pattern: `WithCommissionRetrier`, `WithKYCOnboardingURLResolver`). When nil, summary degrades: commission breakdown.escrowed_cents = 0.

**Tests (integration-style with stub services):**

- `TestWalletHandler_Summary_HappyPath` — seed 2 missions (1 paid, 1 escrowed) + 1 commission paid + 1 projected escrowed → verify breakdown sums, recent_transactions ordered DESC.
- `TestWalletHandler_Summary_Pagination` — 50 transactions, paginate by 20, verify no dup/gap.
- `TestWalletHandler_Summary_Unauthorized` — no userID/orgID context → 401.
- `TestWalletHandler_Summary_EmptyWallet` — zero everything → currency:"EUR", all-zero breakdown, empty transactions.
- `TestWalletHandler_Summary_ProjectorNil_DegradedMode` — no projector wired → escrowed_cents=0 in commission breakdown but missions still work.

## B.3 — `POST /api/v1/wallet/withdraw` (unified)

**Location:** new file `backend/internal/handler/wallet_withdraw.go` (~200 LOC).
**Method:** `Withdraw(w, r)`.

**Body:**
```json
{ "amount_cents": 12345 }   // optional; omitted = drain all available
```

**Implementation flow:**
1. KYC pre-check (reuse `kycProbe.CanProviderReceivePayouts`) — 422 `kyc_required` if not ready (with onboarding_url).
2. Optional billing-profile gate (reuse `invoicingSvc.IsBillingProfileComplete`) — 403 if incomplete.
3. Source order: missions first, then commissions.
4. Call `paymentSvc.RequestPayout(ctx, userID, orgID)` for missions → captures missions_cents drained.
5. If amount remaining > 0 (or omitted), iterate eligible commissions (`referralSvc.RecentCommissions` filtered to status=pending_kyc/failed) and call `referralSvc.RetryCommission(ctx, userID, commissionID)` for each → captures commissions_cents drained.
6. Audit: emit `ActionWalletWithdrawExecuted` with metadata `{missions_cents, commissions_cents, stripe_transfer_ids}`.

**Partial-failure decision (documented in commit msg per brief):**

I will return **207 Multi-Status** when one leg succeeds and the other fails after at least one transfer was actually executed. Returning 500 would erase any successful work the client did. A pure 500 only returns when no transfer succeeded.

Body always includes `drained_cents`, `missions_cents`, `commissions_cents`, `stripe_transfer_ids`, `currency`, plus `errors[]` array if partial.

**Idempotency:** keyed on `Idempotency-Key` header (existing middleware on the route).

**Tests:**

- `TestWalletHandler_Withdraw_DrainAll` — happy path with both legs.
- `TestWalletHandler_Withdraw_PartialAmount` — prefer missions, then commissions.
- `TestWalletHandler_Withdraw_KYCRequired` — 422 with onboarding URL.
- `TestWalletHandler_Withdraw_BillingIncomplete` — 403 billing_profile_incomplete.
- `TestWalletHandler_Withdraw_Idempotency` — same Idempotency-Key returns cached, no double Stripe.
- `TestWalletHandler_Withdraw_PartialFailure_207` — missions OK, commissions fail → 207 + errors[].
- `TestWalletHandler_Withdraw_NothingToDrain` — empty wallet → 200 with drained=0.
- `TestWalletHandler_Withdraw_AuditEmitted` — verify audit row written on success.

## B.4 — Audit

- Add `audit.ActionWalletWithdrawExecuted = "wallet.withdraw_executed"` constant.
- Withdraw handler builds an audit.Entry via `audit.NewEntry` and calls `auditRepo.Log`.
- ProjectedCommissions is read-only — no audit.

## B.5 — Backwards compat

- The OLD `POST /api/v1/wallet/commissions/{id}/retry` endpoint stays mounted.
- Wrap its handler with a small middleware that injects `Deprecation: true` + `Sunset: <today+30d>` headers.
- Test: `TestRetryCommission_DeprecationHeaders` verifies both headers present.

## Pipeline

After every commit:

```bash
cd backend
go build ./...
go vet ./...
go test ./internal/app/referral/... ./internal/app/payment/... ./internal/handler/ -count=1 -race -timeout 240s
go test ./internal/adapter/postgres/... -count=1 -timeout 180s
```

All must be green before commit; output pasted into the final report.

## Commit sequence

1. `_plan_run_b.md`
2. `feat(referral): ProjectedCommissions reader + 200-row cap + 15-case matrix`
3. `feat(handler): GET /wallet/summary unified breakdown + pagination`
4. `feat(handler): POST /wallet/withdraw unified missions+commissions + audit`
5. `chore(handler): deprecation headers on legacy /wallet/commissions/{id}/retry`

Each commit: must build + tests green + paste output. Push after each.

## Hard constraints respected

- File ≤ 600 LOC: wallet_handler.go is at 485; splitting NEW endpoints into wallet_summary.go and wallet_withdraw.go avoids breaching.
- Function ≤ 50 LOC: orchestrator methods broken into helpers (loadAttributions, projectMilestones, mergeTransactions).
- Params ≤ 4: use input structs where needed.
- Parameterized SQL only — no new SQL written (extending existing wallet_reader paths).
- 5-second timeouts already enforced by adapter layer.
- DO NOT touch: web/, mobile/, admin/, migrations/, Run A's end-intro logic (extend only).
- Re-run Run A's gate tests as part of pipeline — `commission_distributor_endedgate_test.go`.

## Risks / open questions

- The brief's "draft / in_progress / pending_release" milestone wording doesn't 1:1 match our domain (`pending_funding / funded / submitted / approved / released / disputed / cancelled / refunded`). I documented my mapping above. If this is wrong the projection behaviour will be off but the unit tests pin the dispatch precisely so it'll be visible in CI.
- Brief asks signature `ProjectedCommissions(ctx, orgID)`. Existing wallet code is user-keyed. Resolution: take orgID, fan-out via OrgMemberLister to all member userIDs. Worst-case if the org has 0 members the function returns nil — graceful degradation, no error.
- Brief says "5-second DB context timeout per query" — already handled inside each adapter, no new SQL means no new timeouts to add.
