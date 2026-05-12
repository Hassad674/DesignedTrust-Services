// Unit tests for the mobile review deep-link helper.
//
// Mirrors the web vitest coverage for buildReviewDeepLink — every
// review notification kind produces a /chat/<conv>?openReview=1&
// reviewProposalId=<id> path, every other kind returns null, and
// malformed payloads (missing/empty/non-string fields, legacy
// dispute_id-only payloads) silently degrade to null instead of
// throwing.

import 'package:flutter_test/flutter_test.dart';
import 'package:marketplace_mobile/features/notification/domain/entities/app_notification.dart';
import 'package:marketplace_mobile/features/notification/domain/review_deep_link.dart';

AppNotification _make(String type, Map<String, dynamic> data) {
  return AppNotification(
    id: 'n-1',
    userId: 'u-1',
    type: type,
    title: 'title',
    body: 'body',
    data: data,
    createdAt: DateTime.utc(2026, 5, 12),
  );
}

void main() {
  group('buildReviewDeepLink', () {
    const proposalId = '11111111-1111-1111-1111-111111111111';
    const conversationId = '22222222-2222-2222-2222-222222222222';

    test('proposal_completed with full payload → returns the deep link', () {
      final link = buildReviewDeepLink(_make('proposal_completed', {
        'proposal_id': proposalId,
        'conversation_id': conversationId,
        'proposal_title': 'Site web Acme',
      }));

      expect(link, isNotNull);
      expect(link!.conversationId, conversationId);
      expect(link.proposalId, proposalId);
      expect(
        link.toRoute('/chat'),
        '/chat/$conversationId?openReview=1&reviewProposalId=$proposalId',
      );
    });

    test('review_received with full payload → returns the deep link', () {
      final link = buildReviewDeepLink(_make('review_received', {
        'proposal_id': proposalId,
        'conversation_id': conversationId,
      }));

      expect(link, isNotNull);
      expect(link!.proposalId, proposalId);
    });

    test('non-review kinds → returns null', () {
      const nonReviewKinds = <String>[
        'proposal_received',
        'proposal_accepted',
        'proposal_declined',
        'proposal_modified',
        'proposal_paid',
        'completion_requested',
        'new_message',
        'system_announcement',
      ];
      for (final kind in nonReviewKinds) {
        final link = buildReviewDeepLink(_make(kind, {
          'proposal_id': proposalId,
          'conversation_id': conversationId,
        }));
        expect(link, isNull, reason: 'kind $kind must not deep-link');
      }
    });

    test('missing proposal_id → null', () {
      final link = buildReviewDeepLink(_make('proposal_completed', {
        'conversation_id': conversationId,
      }));
      expect(link, isNull);
    });

    test('missing conversation_id → null', () {
      final link = buildReviewDeepLink(_make('proposal_completed', {
        'proposal_id': proposalId,
      }));
      expect(link, isNull);
    });

    test('legacy dispute payload (only dispute_id) → null (regression guard)', () {
      // Pre-fix, dispute resolution notifications shipped only with
      // dispute_id. A stale notification arriving post-fix must not
      // crash the tap handler — it must silently fall back to
      // "mark-as-read only".
      final link = buildReviewDeepLink(_make('proposal_completed', {
        'dispute_id': 'd-1',
      }));
      expect(link, isNull);
    });

    test('empty string values → null', () {
      final link = buildReviewDeepLink(_make('proposal_completed', {
        'proposal_id': '',
        'conversation_id': conversationId,
      }));
      expect(link, isNull);
    });

    test('non-string values → null', () {
      final link = buildReviewDeepLink(_make('proposal_completed', {
        'proposal_id': 42,
        'conversation_id': conversationId,
      }));
      expect(link, isNull);
    });
  });
}
