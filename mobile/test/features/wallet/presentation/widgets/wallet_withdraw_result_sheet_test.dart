import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:marketplace_mobile/features/wallet/domain/entities/wallet_summary_entity.dart';
import 'package:marketplace_mobile/features/wallet/presentation/widgets/wallet_withdraw_result_sheet.dart';
import 'package:marketplace_mobile/l10n/app_localizations.dart';

Widget _wrap(WithdrawResult result) {
  return MaterialApp(
    localizationsDelegates: const [
      AppLocalizations.delegate,
      GlobalMaterialLocalizations.delegate,
      GlobalWidgetsLocalizations.delegate,
      GlobalCupertinoLocalizations.delegate,
    ],
    supportedLocales: const [Locale('fr'), Locale('en')],
    locale: const Locale('fr'),
    home: Scaffold(
      body: WalletWithdrawResultSheetBody(result: result),
    ),
  );
}

void main() {
  group('WalletWithdrawResultSheetBody', () {
    testWidgets('shows the title, drained total and per-leg lines',
        (tester) async {
      await tester.pumpWidget(_wrap(const WithdrawResult(
        drainedCents: 50000,
        missionsCents: 30000,
        commissionsCents: 20000,
        currency: 'EUR',
      ),),);
      await tester.pumpAndSettle();
      expect(find.text('Détails du retrait'), findsOneWidget);
      expect(find.textContaining('retirés'), findsOneWidget);
      expect(find.textContaining('Missions'), findsOneWidget);
      expect(find.textContaining('Commissions'), findsOneWidget);
    });

    testWidgets('renders the errors block when populated', (tester) async {
      await tester.pumpWidget(_wrap(const WithdrawResult(
        drainedCents: 30000,
        missionsCents: 30000,
        commissionsCents: 0,
        currency: 'EUR',
        errors: [
          WithdrawLegError(
            source: 'commissions',
            code: 'failed',
            message: 'Stripe rejected the transfer',
          ),
        ],
      ),),);
      await tester.pumpAndSettle();
      expect(find.text('Erreurs'), findsOneWidget);
      expect(find.text('Côté commissions'), findsOneWidget);
      expect(find.text('Stripe rejected the transfer'), findsOneWidget);
    });

    testWidgets('omits the errors block on full success', (tester) async {
      await tester.pumpWidget(_wrap(const WithdrawResult(
        drainedCents: 30000,
        missionsCents: 30000,
        commissionsCents: 0,
        currency: 'EUR',
      ),),);
      await tester.pumpAndSettle();
      expect(find.text('Erreurs'), findsNothing);
    });

    testWidgets('Fermer button dismisses the sheet via Navigator.pop',
        (tester) async {
      var closed = false;
      await tester.pumpWidget(
        MaterialApp(
          localizationsDelegates: const [
            AppLocalizations.delegate,
            GlobalMaterialLocalizations.delegate,
            GlobalWidgetsLocalizations.delegate,
            GlobalCupertinoLocalizations.delegate,
          ],
          supportedLocales: const [Locale('fr'), Locale('en')],
          locale: const Locale('fr'),
          home: Builder(builder: (context) {
            return Scaffold(
              body: ElevatedButton(
                onPressed: () async {
                  await showWalletWithdrawResultSheet(
                    context: context,
                    result: const WithdrawResult(
                      drainedCents: 1000,
                      missionsCents: 600,
                      commissionsCents: 400,
                    ),
                  );
                  closed = true;
                },
                child: const Text('open'),
              ),
            );
          },),
        ),
      );
      await tester.tap(find.text('open'));
      await tester.pumpAndSettle();
      expect(find.text('Détails du retrait'), findsOneWidget);
      await tester.tap(find.text('Fermer'));
      await tester.pumpAndSettle();
      expect(closed, isTrue);
    });
  });
}
