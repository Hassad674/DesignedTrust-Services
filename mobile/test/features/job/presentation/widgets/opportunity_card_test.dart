/// Pinning tests for the OpportunityCard's applications-count badge.
///
/// Mirrors the web `opportunity-card.test.tsx` expectations so the
/// mobile feed exposes the same social-proof signal:
///   * count == 0 → corail-soft "be the first to apply" pill
///   * count == 1 → singular FR plural ("1 candidature")
///   * count >  1 → cardinal FR plural ("12 candidatures")
import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:marketplace_mobile/core/theme/app_theme.dart';
import 'package:marketplace_mobile/features/job/domain/entities/job_entity.dart';
import 'package:marketplace_mobile/features/job/presentation/widgets/opportunity_card.dart';
import 'package:marketplace_mobile/l10n/app_localizations.dart';

JobEntity _job({int totalApplicants = 0}) {
  return JobEntity(
    id: 'job-1',
    creatorId: 'user-1',
    title: 'Backend Go senior',
    description: 'Build a marketplace search engine.',
    skills: const ['Go', 'PostgreSQL'],
    applicantType: 'freelancers',
    budgetType: 'long_term',
    minBudget: 4000,
    maxBudget: 7000,
    status: 'open',
    createdAt: '2026-04-22T10:00:00Z',
    updatedAt: '2026-04-22T10:00:00Z',
    paymentFrequency: 'monthly',
    totalApplicants: totalApplicants,
  );
}

Widget _wrap(JobEntity job, Locale locale) {
  final router = GoRouter(
    initialLocation: '/',
    routes: [
      GoRoute(
        path: '/',
        builder: (_, __) => Scaffold(
          body: SizedBox(width: 480, child: OpportunityCard(job: job)),
        ),
      ),
      GoRoute(
        path: '/opportunities/detail',
        builder: (_, __) => const SizedBox(),
      ),
    ],
  );
  return ProviderScope(
    child: MaterialApp.router(
      theme: AppTheme.light,
      locale: locale,
      localizationsDelegates: const [
        AppLocalizations.delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      supportedLocales: const [Locale('fr'), Locale('en')],
      routerConfig: router,
    ),
  );
}

void main() {
  group('OpportunityCard applications-count badge (FR)', () {
    testWidgets('renders the "be the first" nudge when total is 0',
        (tester) async {
      tester.view.physicalSize = const Size(1000, 1600);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.reset);

      await tester.pumpWidget(_wrap(_job(totalApplicants: 0), const Locale('fr')));
      await tester.pumpAndSettle();
      expect(find.text('Sois le premier à candidater'), findsOneWidget);
      // Confirms the discrete plural label is hidden when count is zero
      // (we only show the "be the first" nudge, not "0 candidatures").
      expect(find.text('0 candidatures'), findsNothing);
    });

    testWidgets('renders singular FR plural when total is 1', (tester) async {
      tester.view.physicalSize = const Size(1000, 1600);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.reset);

      await tester.pumpWidget(_wrap(_job(totalApplicants: 1), const Locale('fr')));
      await tester.pumpAndSettle();
      expect(find.text('1 candidature'), findsOneWidget);
      expect(find.text('Sois le premier à candidater'), findsNothing);
    });

    testWidgets('renders cardinal FR plural when total is greater than 1',
        (tester) async {
      tester.view.physicalSize = const Size(1000, 1600);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.reset);

      await tester.pumpWidget(_wrap(_job(totalApplicants: 12), const Locale('fr')));
      await tester.pumpAndSettle();
      expect(find.text('12 candidatures'), findsOneWidget);
      expect(find.text('Sois le premier à candidater'), findsNothing);
    });
  });

  group('OpportunityCard applications-count badge (EN)', () {
    testWidgets('renders the EN nudge when total is 0', (tester) async {
      tester.view.physicalSize = const Size(1000, 1600);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.reset);

      await tester.pumpWidget(_wrap(_job(totalApplicants: 0), const Locale('en')));
      await tester.pumpAndSettle();
      expect(find.text('Be the first to apply'), findsOneWidget);
    });

    testWidgets('renders singular EN plural when total is 1', (tester) async {
      tester.view.physicalSize = const Size(1000, 1600);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.reset);

      await tester.pumpWidget(_wrap(_job(totalApplicants: 1), const Locale('en')));
      await tester.pumpAndSettle();
      expect(find.text('1 application'), findsOneWidget);
    });

    testWidgets('renders cardinal EN plural when total is greater than 1',
        (tester) async {
      tester.view.physicalSize = const Size(1000, 1600);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.reset);

      await tester.pumpWidget(_wrap(_job(totalApplicants: 4), const Locale('en')));
      await tester.pumpAndSettle();
      expect(find.text('4 applications'), findsOneWidget);
    });
  });
}
