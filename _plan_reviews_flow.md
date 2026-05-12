# Plan — Double-blind reviews flow entry points

Branch: `feat/reviews-flow-entries` (created from `main` @ a651d87b)

## Three entry points to wire

### 1. Dispute resolution → review system message + notification (backend)

**Audit findings**
- `Service.RespondToCounter` (amiable resolution) and `Service.AdminResolve`
  (mediation decision) BOTH already call `restoreProposalAndDistribute`
  which already emits `proposal_completed` + `evaluation_request` system
  messages AND `proposal_completed` notifications. Good.
- `Service.RespondToCancellation(accept=true)` does NOT — but that path
  cancels the dispute and restores the proposal to `active`, not
  `completed`. No review CTA needed.
- `Scheduler.autoResolve` (the 7-day ghost path) has its OWN local
  `restoreAndDistribute` (see `scheduler.go` lines 155-181) that is
  MISSING the `proposal_completed` + `evaluation_request` system
  messages and the user-facing notifications.
  → **This is the regression.** Auto-resolved missions land in
  `completed` status but never get the review CTA in the chat.
- The `proposal_completed` notification payload in
  `service_helpers.go` currently carries only `dispute_id`. To let the
  frontend navigate to `/messages?id={conv}&openReview=1`, the
  notification payload must carry `proposal_id` AND `conversation_id`
  (matching the shape emitted by `proposal/service_create.go::buildNotificationData`).

**Fixes**
- In `backend/internal/app/dispute/service_helpers.go`, change
  `restoreProposalAndDistribute` so the `proposal_completed`
  notification carries `proposal_id`, `conversation_id`,
  `proposal_title` (i.e. the same shape proposal's
  `buildNotificationData` uses) instead of only `dispute_id`.
- In `backend/internal/app/dispute/scheduler.go`, make
  `Scheduler.autoResolve` emit the `proposal_completed` +
  `evaluation_request` system messages and the user-facing
  `proposal_completed` notifications (with the proposal-shaped payload)
  after the proposal is moved to `completed`. Reuse the existing
  `buildProposalCompletedMetadata` helper for the messages.
- Verify by table-driven unit tests on every dispute resolution branch.

### 2. Notification click → open review modal (web)

**Plan**
- Extend `web/src/features/notification/types.ts` `NotificationType` with
  the two existing review notification kinds emitted by the backend:
  `"review_received"` and `"review_revealed"` are already in the union;
  add no new kind — the existing `proposal_completed` notification IS
  the "leave a review" CTA notification (it's what the proposal flow
  fires with title "Mission completed" / body "Leave a review…"). 
- Make `notification-item.tsx` route clicks to:
  - `proposal_completed` notifications → navigate to
    `/messages?id={conversation_id}&openReview=1` (with `proposal_id`
    passed in URL so messaging-page can target the right review).
    Use the existing `useRouter()` from `@i18n/navigation`.
  - `review_received` (counterpart left you a review but reveal not
    yet triggered) → navigate to messaging with `openReview=1` so the
    user submits theirs and unlocks the reveal.
  - Other types → no-op navigation (just mark-as-read like today).
- In `web/src/features/messaging/components/messaging-page.tsx`, on
  mount read `?openReview=1` + `?reviewProposalId=…` from
  `useSearchParams()`. When present + `useCanReview(proposalId)`
  returns `can_review === true`, open the modal once. When
  `can_review === false`, silently swallow (or toast — keep it quiet
  per brief). After the modal opens, strip the params from the URL so
  refresh doesn't reopen the modal.

### 3. Profile project history row → review modal (web)

**Plan**
- The backend `project_history` endpoint returns the counterparty's
  review only (e.g. for a provider profile: the client→provider review).
  It does NOT carry "did the viewer review this?". To decide row
  interactivity per entry, call `useCanReview(proposal_id)` per row.
- In `web/src/features/provider/components/project-history-card.tsx`,
  when the viewer is logged in AND `useCanReview` returns
  `can_review === true`, render the "awaiting review" placeholder as
  a `<button>` whose `onClick` calls a parent callback to open the
  review modal for that proposal.
- Wire the modal at the `profile-history.tsx` level (the section
  component) so we mount one shared `ReviewModal` and pass the open
  state down. Side derivation: a provider profile shows the
  client→provider history → the viewer is a client → side is
  `client_to_provider`. A client profile (`ListByClientOrganization`)
  → side is `provider_to_client`. Since the section component already
  receives the `orgId` of the profile owner, we hint via a new
  optional prop `side?: ReviewSide` on `ProfileHistory`.

## Files to modify

### Backend
- `backend/internal/app/dispute/service_helpers.go` (notif payload shape
  + reuse for autoResolve)
- `backend/internal/app/dispute/scheduler.go` (autoResolve emits
  review system messages + notifications)
- New tests:
  - extend `backend/internal/app/dispute/scheduler_*_test.go` or add a
    new test file with `TestScheduler_AutoResolve_EmitsReviewMessages`
  - extend dispute action tests to assert the new notification payload
    shape on all resolution branches

### Web
- `web/src/features/notification/components/notification-item.tsx`
  (click handler → navigate + openReview)
- `web/src/features/messaging/components/messaging-page.tsx`
  (read query params on mount, auto-open modal)
- `web/src/features/provider/components/profile-history.tsx`
  (mount modal at section level + side prop)
- `web/src/features/provider/components/project-history-card.tsx`
  (clickable row when eligible + onOpenReview callback)
- Tests:
  - `web/src/features/notification/components/__tests__/notification-item.test.tsx`
  - `web/src/features/provider/components/__tests__/project-history-card.test.tsx`
  - `web/src/features/provider/components/__tests__/profile-history.test.tsx`
  - `web/src/features/messaging/components/__tests__/messaging-page-open-review.test.tsx`

### Mobile
- `mobile/lib/features/notification/...` — tap on review-related
  notifications navigates to the conversation route with a flag that
  the conversation screen reads to auto-open the review sheet.
- `mobile/lib/features/provider/...` — project history list rows are
  tappable when the viewer is eligible to review.

## Decision log

- **URL param**: `?openReview=1&reviewProposalId={uuid}`. The
  `reviewProposalId` carries the target proposal so we don't have to
  resolve it from the conversation list (messaging-page already
  caches conversations, but the user might land on `/messages?id=X`
  while the proposal is on a different conv — unlikely but safe).
- **"Already reviewed" guard**: the existing `useCanReview` hook (and
  its `GET /api/v1/reviews/can-review/{proposalId}` endpoint) is the
  single source of truth. We always check it before opening the modal.
- **Notification payload shape**: align dispute resolution
  notifications with the proposal-flow shape (`proposal_id`,
  `conversation_id`, `proposal_title`) so the frontend handler is
  generic.
- **No new notification kinds**: the existing `proposal_completed`
  type already plays the "leave a review" role; we just teach the
  frontend that this kind opens the review modal.

## Test count

- Backend: 4 new tests (AutoResolve emits messages, AutoResolve emits
  notifications, AmiableResolve notification payload shape, AdminResolve
  notification payload shape)
- Web: 6 new tests (notification-item route table, project-history-card
  eligibility states, profile-history mount, messaging-page query param
  auto-open, idempotency: refresh doesn't reopen, already-reviewed
  swallows)
- Mobile: minimum 3 widget tests covering the 3 entry points

## Out of scope / flagged

- Not touching `backend/internal/app/payment/wallet.go`,
  `backend/internal/handler/referral_handler.go`,
  `web/src/features/referral/`, LiveKit, presence — those belong to
  parallel agents.
- Not touching the existing review modal internals or the double-blind
  privacy logic — we only wire entry points.
