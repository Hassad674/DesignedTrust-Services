import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:marketplace_mobile/features/profile_completion/domain/entities/profile_completion_report.dart';
import 'package:marketplace_mobile/features/profile_completion/domain/repositories/profile_completion_repository.dart';
import 'package:marketplace_mobile/features/profile_completion/presentation/providers/profile_completion_providers.dart';
import 'package:marketplace_mobile/core/theme/app_theme.dart';
import 'package:marketplace_mobile/features/profile_completion/presentation/widgets/profile_completion_bar.dart';
import 'package:marketplace_mobile/l10n/app_localizations.dart';

class _FakeRepo implements ProfileCompletionRepository {
  _FakeRepo(this._report);
  final ProfileCompletionReport _report;
  @override
  Future<ProfileCompletionReport> getMy() async => _report;
}

class _ThrowingRepo implements ProfileCompletionRepository {
  @override
  Future<ProfileCompletionReport> getMy() async => throw Exception('boom');
}

Widget _wrap({
  required ProfileCompletionRepository repo,
  bool hideWhenComplete = false,
}) {
  return ProviderScope(
    overrides: [
      profileCompletionRepositoryProvider.overrideWithValue(repo),
    ],
    child: MaterialApp(
      theme: AppTheme.light,
      localizationsDelegates: const [
        AppLocalizations.delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      supportedLocales: const [Locale('en'), Locale('fr')],
      locale: const Locale('fr'),
      home: Scaffold(
        body: ProfileCompletionBar(hideWhenComplete: hideWhenComplete),
      ),
    ),
  );
}

const _baseReport = ProfileCompletionReport(
  role: 'provider',
  persona: 'freelance',
  percent: 50,
  totalSections: 10,
  filledSections: 5,
  sections: [
    ProfileCompletionSection(
      key: 'title',
      filled: true,
      labelKey: 'profile.completion.section.title',
      completionPath: '/dashboard/profile/edit',
    ),
    ProfileCompletionSection(
      key: 'about',
      filled: false,
      labelKey: 'profile.completion.section.about',
      completionPath: '/dashboard/profile/edit',
    ),
  ],
);

void main() {
  testWidgets('renders the percent and filled/total subtitle',
      (tester) async {
    await tester.pumpWidget(_wrap(repo: _FakeRepo(_baseReport)));
    await tester.pumpAndSettle();

    expect(find.text('Profil rempli à 50%'), findsOneWidget);
    expect(find.text('5/10 sections complétées'), findsOneWidget);
  });

  testWidgets('opens the bottom sheet listing sections on tap',
      (tester) async {
    await tester.pumpWidget(_wrap(repo: _FakeRepo(_baseReport)));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Profil rempli à 50%'));
    await tester.pumpAndSettle();

    expect(find.text('Profil complété à 50%'), findsOneWidget);
    expect(find.text('Titre professionnel'), findsOneWidget);
    expect(find.text('À propos'), findsOneWidget);
  });

  testWidgets('hides itself at 100 percent when hideWhenComplete is true',
      (tester) async {
    const completed = ProfileCompletionReport(
      role: 'provider',
      persona: 'freelance',
      percent: 100,
      totalSections: 5,
      filledSections: 5,
      sections: [],
    );
    await tester.pumpWidget(_wrap(repo: _FakeRepo(completed), hideWhenComplete: true));
    await tester.pumpAndSettle();

    expect(find.textContaining('Profil rempli'), findsNothing);
  });

  testWidgets('renders nothing on repository error (silent fallback)',
      (tester) async {
    await tester.pumpWidget(_wrap(repo: _ThrowingRepo()));
    await tester.pumpAndSettle();

    expect(find.textContaining('Profil rempli'), findsNothing);
  });
}
