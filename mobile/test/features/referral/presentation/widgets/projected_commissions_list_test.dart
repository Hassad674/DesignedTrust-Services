import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:marketplace_mobile/features/referral/domain/entities/referral_entity.dart';
import 'package:marketplace_mobile/features/referral/presentation/widgets/projected_commissions_list.dart';
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

ReferralCommission _commission({
  required String status,
  int cents = 10000,
  String id = 'c1',
}) {
  return ReferralCommission(
    id: id,
    attributionId: 'att-1',
    milestoneId: 'm-1',
    grossAmountCents: cents * 5,
    commissionCents: cents,
    status: status,
    createdAt: '2026-05-01T10:00:00Z',
  );
}

void main() {
  group('ProjectedCommissionsList', () {
    testWidgets('renders paid commission with success tone', (tester) async {
      await tester.pumpWidget(_wrap(ProjectedCommissionsList(
        commissions: [_commission(status: 'paid', cents: 50000, id: 'paid-1')],
      ),),);
      await tester.pumpAndSettle();
      expect(find.textContaining('reçue'), findsOneWidget);
      expect(
        find.byKey(const ValueKey('projected-commission-row-paid-1')),
        findsOneWidget,
      );
    });

    testWidgets('renders pending commission with attente label',
        (tester) async {
      await tester.pumpWidget(_wrap(ProjectedCommissionsList(
        commissions: [
          _commission(status: 'pending', cents: 30000, id: 'pen-1'),
        ],
      ),),);
      await tester.pumpAndSettle();
      expect(find.textContaining('en attente'), findsOneWidget);
    });

    testWidgets('renders failed commission with échouée label',
        (tester) async {
      await tester.pumpWidget(_wrap(ProjectedCommissionsList(
        commissions: [_commission(status: 'failed', id: 'fail-1')],
      ),),);
      await tester.pumpAndSettle();
      expect(find.textContaining('échouée'), findsOneWidget);
    });

    testWidgets('renders escrow preview when escrowCents > 0',
        (tester) async {
      await tester.pumpWidget(_wrap(const ProjectedCommissionsList(
        commissions: [],
        escrowCents: 25000,
      ),),);
      await tester.pumpAndSettle();
      expect(
        find.byKey(const ValueKey('projected-commission-escrow-line')),
        findsOneWidget,
      );
      expect(find.textContaining('séquestre'), findsOneWidget);
    });

    testWidgets('filters out cancelled and clawed_back rows',
        (tester) async {
      await tester.pumpWidget(_wrap(ProjectedCommissionsList(
        commissions: [
          _commission(status: 'cancelled', id: 'cancelled-1'),
          _commission(status: 'clawed_back', id: 'clawed-1'),
          _commission(status: 'paid', id: 'paid-2'),
        ],
      ),),);
      await tester.pumpAndSettle();
      expect(
        find.byKey(const ValueKey('projected-commission-row-cancelled-1')),
        findsNothing,
      );
      expect(
        find.byKey(const ValueKey('projected-commission-row-clawed-1')),
        findsNothing,
      );
      expect(
        find.byKey(const ValueKey('projected-commission-row-paid-2')),
        findsOneWidget,
      );
    });

    testWidgets('renders empty state when no rows and no escrow',
        (tester) async {
      await tester.pumpWidget(_wrap(const ProjectedCommissionsList(
        commissions: [],
      ),),);
      await tester.pumpAndSettle();
      expect(
        find.text('Pas encore de jalons commissionnables.'),
        findsOneWidget,
      );
    });
  });
}
