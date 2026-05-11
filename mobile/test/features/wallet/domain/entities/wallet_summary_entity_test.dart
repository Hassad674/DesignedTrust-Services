import 'package:flutter_test/flutter_test.dart';
import 'package:marketplace_mobile/features/wallet/domain/entities/wallet_summary_entity.dart';

void main() {
  group('WalletSummary.fromJson', () {
    test('parses a full payload with breakdown + transactions', () {
      final summary = WalletSummary.fromJson({
        'currency': 'EUR',
        'total_cents': 3000,
        'available_cents': 1500,
        'escrowed_cents': 1000,
        'transmitted_cents': 500,
        'breakdown': {
          'missions': {
            'total_cents': 2000,
            'available_cents': 1000,
            'escrowed_cents': 600,
            'transmitted_cents': 400,
          },
          'commissions': {
            'total_cents': 1000,
            'available_cents': 500,
            'escrowed_cents': 400,
            'transmitted_cents': 100,
          },
        },
        'recent_transactions': [
          {
            'type': 'mission',
            'amount_cents': 500,
            'currency': 'EUR',
            'status': 'paid',
            'mission_title': 'Logo design',
            'occurred_at': '2026-05-01T10:00:00Z',
            'reference_id': 'r1',
          },
          {
            'type': 'commission',
            'amount_cents': 200,
            'currency': 'EUR',
            'status': 'pending',
            'occurred_at': '2026-05-02T10:00:00Z',
            'reference_id': 'r2',
          },
        ],
        'next_cursor': 'CURSOR',
      });

      expect(summary.currency, 'EUR');
      expect(summary.totalCents, 3000);
      expect(summary.missions.totalCents, 2000);
      expect(summary.commissions.totalCents, 1000);
      expect(summary.recentTransactions, hasLength(2));
      expect(summary.recentTransactions.first.isMission, isTrue);
      expect(summary.recentTransactions.last.isCommission, isTrue);
      expect(summary.recentTransactions.last.missionTitle, isNull);
      expect(summary.nextCursor, 'CURSOR');
    });

    test('falls back to defaults when breakdown is missing', () {
      final summary = WalletSummary.fromJson({});
      expect(summary.currency, 'EUR');
      expect(summary.totalCents, 0);
      expect(summary.missions.totalCents, 0);
      expect(summary.commissions.transmittedCents, 0);
      expect(summary.recentTransactions, isEmpty);
      expect(summary.nextCursor, isNull);
    });

    test('skips malformed transaction entries', () {
      final summary = WalletSummary.fromJson({
        'recent_transactions': [
          {'type': 'mission', 'amount_cents': 1, 'reference_id': 'ok'},
          'not-a-map',
          null,
          {'type': 'commission', 'amount_cents': 2, 'reference_id': 'ok2'},
        ],
      });
      expect(summary.recentTransactions, hasLength(2));
      expect(summary.recentTransactions.first.referenceId, 'ok');
      expect(summary.recentTransactions.last.referenceId, 'ok2');
    });

    test('null/missing optional fields parse without throwing', () {
      final summary = WalletSummary.fromJson({
        'total_cents': null,
        'breakdown': {
          'missions': {'total_cents': null},
          'commissions': null,
        },
        'recent_transactions': null,
      });
      expect(summary.totalCents, 0);
      expect(summary.missions.totalCents, 0);
      expect(summary.commissions.totalCents, 0);
      expect(summary.recentTransactions, isEmpty);
    });
  });

  group('WithdrawResult.fromJson', () {
    test('full success — empty errors', () {
      final res = WithdrawResult.fromJson({
        'drained_cents': 500,
        'missions_cents': 300,
        'commissions_cents': 200,
        'stripe_transfer_ids': ['tr_1', 'tr_2'],
        'currency': 'EUR',
        'errors': [],
      });
      expect(res.isFullSuccess, isTrue);
      expect(res.isPartialSuccess, isFalse);
      expect(res.stripeTransferIds, ['tr_1', 'tr_2']);
    });

    test('partial success — populated errors', () {
      final res = WithdrawResult.fromJson({
        'drained_cents': 300,
        'missions_cents': 300,
        'commissions_cents': 0,
        'errors': [
          {
            'source': 'commissions',
            'code': 'failed',
            'message': 'Stripe rejected',
          },
        ],
      });
      expect(res.isPartialSuccess, isTrue);
      expect(res.errors, hasLength(1));
      expect(res.errors.first.source, 'commissions');
      expect(res.errors.first.message, 'Stripe rejected');
    });

    test('no-op result when nothing drained and no errors', () {
      final res = WithdrawResult.fromJson({});
      expect(res.isNoOp, isTrue);
      expect(res.isFullSuccess, isFalse);
      expect(res.isPartialSuccess, isFalse);
    });
  });

  group('formatWalletSummaryCents', () {
    // The format helper inserts a NBSP (U+00A0) between the number
    // and the euro sign so the currency unit never wraps onto a new
    // line on narrow widgets — same convention as
    // WalletOverview.formatCents.
    const nbsp = ' ';
    test('formats cents with French-style grouping', () {
      expect(formatWalletSummaryCents(100), '1$nbsp€');
      expect(formatWalletSummaryCents(123456), '1 234$nbsp€');
      expect(formatWalletSummaryCents(0), '0$nbsp€');
    });

    test('handles negative values', () {
      expect(formatWalletSummaryCents(-100), '-1$nbsp€');
    });
  });

  group('walletStatusTone', () {
    test('maps known statuses to tones', () {
      expect(walletStatusTone('paid'), WalletStatusTone.paid);
      expect(walletStatusTone('transferred'), WalletStatusTone.paid);
      expect(walletStatusTone('pending'), WalletStatusTone.pending);
      expect(walletStatusTone('pending_kyc'), WalletStatusTone.pending);
      expect(walletStatusTone('escrowed'), WalletStatusTone.escrowed);
      expect(walletStatusTone('failed'), WalletStatusTone.failed);
      expect(walletStatusTone('clawed_back'), WalletStatusTone.failed);
    });

    test('falls back to pending for unknown status', () {
      expect(walletStatusTone('weird_unknown'), WalletStatusTone.pending);
    });
  });
}
