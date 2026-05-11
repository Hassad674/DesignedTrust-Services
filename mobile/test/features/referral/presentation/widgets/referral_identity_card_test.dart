import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:marketplace_mobile/features/referral/domain/entities/referral_entity.dart';
import 'package:marketplace_mobile/features/referral/presentation/widgets/anonymized_client_card.dart';
import 'package:marketplace_mobile/features/referral/presentation/widgets/anonymized_provider_card.dart';
import 'package:marketplace_mobile/features/referral/presentation/widgets/referral_identity_card.dart';
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

Referral _referral() {
  return const Referral(
    id: 'r1',
    referrerId: 'ref-1',
    providerId: 'prov-1',
    clientId: 'cli-1',
    durationMonths: 6,
    status: 'active',
    version: 1,
    introSnapshot: IntroSnapshot(
      provider: ProviderSnapshot(
        expertiseDomains: ['SEO', 'Tech'],
      ),
      client: ClientSnapshot(industry: 'SaaS B2B'),
    ),
    lastActionAt: '2026-05-01T10:00:00Z',
    createdAt: '2026-05-01T10:00:00Z',
    updatedAt: '2026-05-01T10:00:00Z',
  );
}

void main() {
  group('ReferralIdentityCard', () {
    testWidgets('owner viewer → renders revealed identity tiles',
        (tester) async {
      await tester.pumpWidget(_wrap(ReferralIdentityCard(
        referral: _referral(),
        isOwner: true,
        providerName: 'Acme Consulting',
        clientName: 'Globex Corp',
      ),),);
      await tester.pumpAndSettle();
      // Clear provider + client names are visible.
      expect(find.text('Acme Consulting'), findsOneWidget);
      expect(find.text('Globex Corp'), findsOneWidget);
      // Reveal tile keys are wired for navigation.
      expect(
        find.byKey(const ValueKey('referral-identity-provider')),
        findsOneWidget,
      );
      expect(
        find.byKey(const ValueKey('referral-identity-client')),
        findsOneWidget,
      );
      // Anonymized cards are NOT used in the owner branch.
      expect(find.byType(AnonymizedProviderCard), findsNothing);
      expect(find.byType(AnonymizedClientCard), findsNothing);
    });

    testWidgets('non-owner viewer → renders anonymized cards',
        (tester) async {
      await tester.pumpWidget(_wrap(ReferralIdentityCard(
        referral: _referral(),
        isOwner: false,
      ),),);
      await tester.pumpAndSettle();
      // Both anonymized cards render — the wrapper component does not
      // need to pick a side; the parent page is in charge of layout
      // for non-owner viewers.
      expect(find.byType(AnonymizedProviderCard), findsOneWidget);
      expect(find.byType(AnonymizedClientCard), findsOneWidget);
      // No reveal tiles in the masked branch.
      expect(
        find.byKey(const ValueKey('referral-identity-provider')),
        findsNothing,
      );
    });

    testWidgets('owner viewer with no names → falls back to expertise + industry',
        (tester) async {
      await tester.pumpWidget(_wrap(ReferralIdentityCard(
        referral: _referral(),
        isOwner: true,
      ),),);
      await tester.pumpAndSettle();
      // First expertise domain becomes the provider fallback.
      expect(find.text('SEO'), findsOneWidget);
      // Client industry becomes the client fallback.
      expect(find.text('SaaS B2B'), findsOneWidget);
    });
  });
}
