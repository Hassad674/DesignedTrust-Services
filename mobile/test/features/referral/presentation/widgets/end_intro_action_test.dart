import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:marketplace_mobile/features/referral/domain/entities/referral_entity.dart';
import 'package:marketplace_mobile/features/referral/domain/repositories/referral_repository.dart';
import 'package:marketplace_mobile/features/referral/presentation/providers/referral_provider.dart';
import 'package:marketplace_mobile/features/referral/presentation/widgets/end_intro_action.dart';
import 'package:marketplace_mobile/l10n/app_localizations.dart';

class _StubReferralRepository implements ReferralRepository {
  bool fail = false;
  String? endedAt;

  @override
  Future<String?> endAttribution(String attributionId) async {
    if (fail) {
      throw Exception('boom');
    }
    return endedAt ?? '2026-05-11T14:30:00Z';
  }

  // Unused — throw to flag accidental usage.
  @override
  Future<List<Referral>> listMyReferrals() async => throw UnimplementedError();
  @override
  Future<List<Referral>> listIncomingReferrals() async =>
      throw UnimplementedError();
  @override
  Future<Referral> getReferral(String id) async => throw UnimplementedError();
  @override
  Future<Referral> createReferral({
    required String providerId,
    required String clientId,
    required double ratePct,
    required int durationMonths,
    required String introMessageProvider,
    required String introMessageClient,
    Map<String, bool>? snapshotToggles,
  }) async =>
      throw UnimplementedError();
  @override
  Future<Referral> respondToReferral({
    required String id,
    required String action,
    double? newRatePct,
    String? message,
  }) async =>
      throw UnimplementedError();
  @override
  Future<List<ReferralNegotiation>> listNegotiations(String id) async =>
      throw UnimplementedError();
  @override
  Future<List<ReferralAttribution>> listAttributions(String id) async =>
      throw UnimplementedError();
  @override
  Future<List<ReferralCommission>> listCommissions(String id) async =>
      throw UnimplementedError();
}

Widget _wrap({
  required _StubReferralRepository repo,
  String? initialEndedAt,
}) {
  return ProviderScope(
    overrides: [
      referralRepositoryProvider.overrideWithValue(repo),
    ],
    child: MaterialApp(
      localizationsDelegates: const [
        AppLocalizations.delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      supportedLocales: const [Locale('fr'), Locale('en')],
      locale: const Locale('fr'),
      home: Scaffold(
        body: EndIntroAction(
          referralId: 'r1',
          attributionId: 'a1',
          providerName: 'Acme',
          clientName: 'Globex',
          initialEndedAt: initialEndedAt,
        ),
      ),
    ),
  );
}

void main() {
  group('EndIntroAction', () {
    testWidgets('renders trigger when initialEndedAt is null',
        (tester) async {
      await tester
          .pumpWidget(_wrap(repo: _StubReferralRepository()));
      await tester.pumpAndSettle();
      expect(
        find.byKey(const ValueKey('end-intro-trigger')),
        findsOneWidget,
      );
      expect(find.byKey(const ValueKey('end-intro-badge')), findsNothing);
    });

    testWidgets('renders badge directly when initialEndedAt is set',
        (tester) async {
      await tester.pumpWidget(_wrap(
        repo: _StubReferralRepository(),
        initialEndedAt: '2026-05-11T14:30:00Z',
      ),);
      await tester.pumpAndSettle();
      expect(
        find.byKey(const ValueKey('end-intro-badge')),
        findsOneWidget,
      );
      expect(find.textContaining('Intro terminée'), findsOneWidget);
      expect(find.textContaining('11/05/2026'), findsOneWidget);
    });

    testWidgets('confirm → swaps trigger for badge on success',
        (tester) async {
      final repo = _StubReferralRepository();
      await tester.pumpWidget(_wrap(repo: repo));
      await tester.pumpAndSettle();

      await tester.tap(find.byKey(const ValueKey('end-intro-trigger')));
      await tester.pumpAndSettle();
      // Dialog visible.
      expect(find.text('Terminer la mise en relation'), findsOneWidget);
      await tester.tap(find.byKey(const ValueKey('end-intro-confirm')));
      await tester.pumpAndSettle();
      // Badge swapped in.
      expect(
        find.byKey(const ValueKey('end-intro-badge')),
        findsOneWidget,
      );
    });

    testWidgets('error path keeps trigger and shows the snackbar',
        (tester) async {
      final repo = _StubReferralRepository()..fail = true;
      await tester.pumpWidget(_wrap(repo: repo));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const ValueKey('end-intro-trigger')));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const ValueKey('end-intro-confirm')));
      await tester.pumpAndSettle();
      // Trigger still visible (badge did not replace it).
      expect(
        find.byKey(const ValueKey('end-intro-trigger')),
        findsOneWidget,
      );
      // Generic error copy surfaces via snackbar.
      expect(
        find.textContaining(
          'Impossible de terminer la mise en relation',
        ),
        findsOneWidget,
      );
    });

    testWidgets('cancel keeps the trigger button visible', (tester) async {
      final repo = _StubReferralRepository();
      await tester.pumpWidget(_wrap(repo: repo));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const ValueKey('end-intro-trigger')));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const ValueKey('end-intro-cancel')));
      await tester.pumpAndSettle();
      expect(
        find.byKey(const ValueKey('end-intro-trigger')),
        findsOneWidget,
      );
      expect(find.byKey(const ValueKey('end-intro-badge')), findsNothing);
    });
  });
}
