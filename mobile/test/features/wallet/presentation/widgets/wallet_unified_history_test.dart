import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:marketplace_mobile/features/wallet/domain/entities/wallet_entity.dart';
import 'package:marketplace_mobile/features/wallet/domain/entities/wallet_summary_entity.dart';
import 'package:marketplace_mobile/features/wallet/domain/repositories/wallet_repository.dart';
import 'package:marketplace_mobile/features/wallet/presentation/providers/wallet_provider.dart';
import 'package:marketplace_mobile/features/wallet/presentation/widgets/wallet_unified_history.dart';
import 'package:marketplace_mobile/l10n/app_localizations.dart';

// Stub repository — never hits the network. The test injects
// WalletSummary fixtures via `setSummary` and the provider returns
// them synchronously.
class _StubRepository implements WalletRepository {
  WalletSummary _summary = const WalletSummary();
  void setSummary(WalletSummary s) => _summary = s;

  @override
  Future<WalletSummary> fetchSummary({String? cursor}) async => _summary;

  // Unused below — throw to flag accidental usage in this test.
  @override
  Future<void> requestPayout() async => throw UnimplementedError();
  @override
  Future<WalletOverview> getWallet() async => throw UnimplementedError();
  @override
  Future<void> retryFailedTransfer(String recordId) async =>
      throw UnimplementedError();
  @override
  Future<void> retryCommission(String commissionId) async =>
      throw UnimplementedError();
  @override
  Future<WithdrawResult> withdraw({int? amountCents}) async =>
      throw UnimplementedError();
}

Widget _wrap(_StubRepository repo) {
  return ProviderScope(
    overrides: [
      walletRepositoryProvider.overrideWithValue(repo),
    ],
    child: const MaterialApp(
      localizationsDelegates: [
        AppLocalizations.delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      supportedLocales: [Locale('fr'), Locale('en')],
      locale: Locale('fr'),
      home: Scaffold(body: WalletUnifiedHistory()),
    ),
  );
}

WalletSummaryTransaction _tx({
  String type = 'mission',
  int amountCents = 500,
  String status = 'paid',
  String? title,
  String referenceId = 'ref-1',
}) {
  return WalletSummaryTransaction(
    type: type,
    amountCents: amountCents,
    currency: 'EUR',
    status: status,
    missionTitle: title,
    occurredAt: '2026-05-01T10:00:00Z',
    referenceId: referenceId,
  );
}

void main() {
  group('WalletUnifiedHistory', () {
    testWidgets('renders the empty state when no transactions',
        (tester) async {
      final repo = _StubRepository();
      await tester.pumpWidget(_wrap(repo));
      await tester.pumpAndSettle();
      expect(
        find.text("Pas encore de mouvement sur ton portefeuille."),
        findsOneWidget,
      );
    });

    testWidgets('renders mission row with paid badge', (tester) async {
      final repo = _StubRepository()
        ..setSummary(WalletSummary(
          recentTransactions: [
            _tx(
              type: 'mission',
              amountCents: 50000,
              status: 'paid',
              title: 'Logo design',
              referenceId: 'r1',
            ),
          ],
        ),);
      await tester.pumpWidget(_wrap(repo));
      await tester.pumpAndSettle();
      expect(find.text('Logo design'), findsOneWidget);
      // Type label + date + "Reçu" status badge.
      expect(find.textContaining('Mission ·'), findsOneWidget);
      expect(find.text('Reçu'), findsOneWidget);
    });

    testWidgets('renders commission row with pending badge', (tester) async {
      final repo = _StubRepository()
        ..setSummary(WalletSummary(
          recentTransactions: [
            _tx(
              type: 'commission',
              amountCents: 20000,
              status: 'pending',
              referenceId: 'r2',
            ),
          ],
        ),);
      await tester.pumpWidget(_wrap(repo));
      await tester.pumpAndSettle();
      expect(find.text('Sans titre'), findsOneWidget,
          reason:
              'commission row without explicit title falls back to i18n placeholder',);
      expect(find.textContaining('Commission ·'), findsOneWidget);
      expect(find.text('En attente'), findsOneWidget);
    });

    testWidgets('shows "Charger plus" when next_cursor present',
        (tester) async {
      final repo = _StubRepository()
        ..setSummary(WalletSummary(
          recentTransactions: [_tx(referenceId: 'r1')],
          nextCursor: 'CURSOR',
        ),);
      await tester.pumpWidget(_wrap(repo));
      await tester.pumpAndSettle();
      expect(
        find.byKey(const ValueKey('wallet-history-load-more')),
        findsOneWidget,
      );
    });

    testWidgets('hides "Charger plus" when next_cursor is empty',
        (tester) async {
      final repo = _StubRepository()
        ..setSummary(WalletSummary(
          recentTransactions: [_tx(referenceId: 'r1')],
        ),);
      await tester.pumpWidget(_wrap(repo));
      await tester.pumpAndSettle();
      expect(
        find.byKey(const ValueKey('wallet-history-load-more')),
        findsNothing,
      );
    });

    testWidgets('Charger plus tap appends the next page', (tester) async {
      final repo = _StubRepository()
        ..setSummary(WalletSummary(
          recentTransactions: [
            _tx(referenceId: 'page1-row1', title: 'Mission un'),
          ],
          nextCursor: 'CURSOR2',
        ),);
      await tester.pumpWidget(_wrap(repo));
      await tester.pumpAndSettle();
      // Swap the fixture so the SECOND fetch returns a different row.
      repo.setSummary(WalletSummary(
        recentTransactions: [
          _tx(referenceId: 'page2-row1', title: 'Mission deux'),
        ],
      ),);
      await tester.tap(
        find.byKey(const ValueKey('wallet-history-load-more')),
      );
      await tester.pumpAndSettle();
      // Both rows visible after load-more — the first stays accumulated.
      expect(find.text('Mission un'), findsOneWidget);
      expect(find.text('Mission deux'), findsOneWidget);
    });

    testWidgets('renders the failed status badge tone', (tester) async {
      final repo = _StubRepository()
        ..setSummary(WalletSummary(
          recentTransactions: [_tx(status: 'failed', referenceId: 'r1')],
        ),);
      await tester.pumpWidget(_wrap(repo));
      await tester.pumpAndSettle();
      expect(find.text('Échoué'), findsOneWidget);
    });

    testWidgets('renders the escrowed status badge tone', (tester) async {
      final repo = _StubRepository()
        ..setSummary(WalletSummary(
          recentTransactions: [_tx(status: 'escrowed', referenceId: 'r1')],
        ),);
      await tester.pumpWidget(_wrap(repo));
      await tester.pumpAndSettle();
      expect(find.text('En séquestre'), findsOneWidget);
    });
  });
}
