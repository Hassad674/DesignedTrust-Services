import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:marketplace_mobile/features/wallet/domain/entities/wallet_summary_entity.dart';
import 'package:marketplace_mobile/features/wallet/presentation/widgets/wallet_unified_header.dart';
import 'package:marketplace_mobile/l10n/app_localizations.dart';

Widget _wrap(Widget child) {
  return MaterialApp(
    localizationsDelegates: const [
      AppLocalizations.delegate,
      GlobalMaterialLocalizations.delegate,
      GlobalWidgetsLocalizations.delegate,
      GlobalCupertinoLocalizations.delegate,
    ],
    supportedLocales: const [Locale('fr'), Locale('en')],
    locale: const Locale('fr'),
    home: Scaffold(body: child),
  );
}

WalletSummary _summary({
  int total = 0,
  int available = 0,
  int escrowed = 0,
  int transmitted = 0,
}) {
  return WalletSummary(
    totalCents: total,
    availableCents: available,
    escrowedCents: escrowed,
    transmittedCents: transmitted,
  );
}

void main() {
  group('WalletUnifiedHeader', () {
    testWidgets('renders title, total, 3 stat cards and the Retirer CTA',
        (tester) async {
      // Wide test window so the IntrinsicHeight + 3-up Row layout
      // does not crash on a too-narrow surface.
      tester.view.physicalSize = const Size(800, 1600);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.reset);

      await tester.pumpWidget(_wrap(
        WalletUnifiedHeader(
          summary: _summary(
            total: 311600,
            available: 41000,
            escrowed: 150000,
            transmitted: 120600,
          ),
          payoutPending: false,
          canWithdraw: true,
          onWithdraw: () {},
        ),
      ),);
      await tester.pumpAndSettle();

      // Title + total visible.
      expect(find.text('Portefeuille'), findsOneWidget);
      expect(find.byKey(const ValueKey('wallet-unified-total')),
          findsOneWidget,);
      // Three stat cards keyed for testability.
      expect(find.byKey(const ValueKey('wallet-stat-escrowed')), findsOneWidget);
      expect(find.byKey(const ValueKey('wallet-stat-available')), findsOneWidget);
      expect(find.byKey(const ValueKey('wallet-stat-transmitted')),
          findsOneWidget,);
      // CTA visible + enabled (availableCents > 0 + canWithdraw).
      final cta = find.byKey(const ValueKey('wallet-unified-withdraw'));
      expect(cta, findsOneWidget);
      final button = tester.widget<ElevatedButton>(cta);
      expect(button.onPressed, isNotNull);
    });

    testWidgets('disables Retirer when availableCents == 0',
        (tester) async {
      tester.view.physicalSize = const Size(800, 1600);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.reset);
      await tester.pumpWidget(_wrap(
        WalletUnifiedHeader(
          summary: _summary(total: 100, available: 0),
          payoutPending: false,
          canWithdraw: true,
          onWithdraw: () {},
        ),
      ),);
      await tester.pumpAndSettle();

      final cta = find.byKey(const ValueKey('wallet-unified-withdraw'));
      final button = tester.widget<ElevatedButton>(cta);
      expect(button.onPressed, isNull,
          reason: 'CTA must be disabled when available_cents = 0',);
      expect(find.text('Aucun fonds disponible'), findsOneWidget);
    });

    testWidgets('disables Retirer when canWithdraw=false', (tester) async {
      tester.view.physicalSize = const Size(800, 1600);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.reset);
      await tester.pumpWidget(_wrap(
        WalletUnifiedHeader(
          summary: _summary(total: 1000, available: 500),
          payoutPending: false,
          canWithdraw: false,
          onWithdraw: () {},
        ),
      ),);
      await tester.pumpAndSettle();
      final cta = find.byKey(const ValueKey('wallet-unified-withdraw'));
      final button = tester.widget<ElevatedButton>(cta);
      expect(button.onPressed, isNull);
      // Permission-denied helper line surfaces.
      expect(
        find.textContaining("n'avez pas la permission"),
        findsOneWidget,
      );
    });

    testWidgets('shows spinner + pending label when payoutPending',
        (tester) async {
      tester.view.physicalSize = const Size(800, 1600);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.reset);
      await tester.pumpWidget(_wrap(
        WalletUnifiedHeader(
          summary: _summary(total: 1000, available: 500),
          payoutPending: true,
          canWithdraw: true,
          onWithdraw: () {},
        ),
      ),);
      // `pumpAndSettle` would never settle — the spinner spins
      // forever. Pump a single frame so the build runs and the
      // pending state surfaces, then immediately assert.
      await tester.pump();
      expect(find.byType(CircularProgressIndicator), findsOneWidget);
      expect(find.text('Retrait en cours…'), findsOneWidget);
      final cta = find.byKey(const ValueKey('wallet-unified-withdraw'));
      final button = tester.widget<ElevatedButton>(cta);
      expect(button.onPressed, isNull,
          reason: 'CTA must be disabled while a withdraw is in flight',);
    });

    testWidgets('Retirer tap fires onWithdraw', (tester) async {
      tester.view.physicalSize = const Size(800, 1600);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.reset);
      var taps = 0;
      await tester.pumpWidget(_wrap(
        WalletUnifiedHeader(
          summary: _summary(total: 1000, available: 500),
          payoutPending: false,
          canWithdraw: true,
          onWithdraw: () => taps++,
        ),
      ),);
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const ValueKey('wallet-unified-withdraw')));
      await tester.pumpAndSettle();
      expect(taps, 1);
    });
  });
}
