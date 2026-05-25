import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:marketplace_mobile/core/router/app_router.dart';
import 'package:marketplace_mobile/core/theme/app_theme.dart';
import 'package:marketplace_mobile/features/auth/presentation/providers/auth_provider.dart';
import 'package:marketplace_mobile/features/feedback/domain/entities/feedback_attachment.dart';
import 'package:marketplace_mobile/features/feedback/domain/entities/feedback_attachment_kind.dart';
import 'package:marketplace_mobile/features/feedback/domain/entities/feedback_submission.dart';
import 'package:marketplace_mobile/features/feedback/domain/entities/submit_feedback_input.dart';
import 'package:marketplace_mobile/features/feedback/domain/repositories/feedback_repository.dart';
import 'package:marketplace_mobile/features/feedback/presentation/providers/feedback_providers.dart';
import 'package:marketplace_mobile/features/feedback/presentation/widgets/feedback_fab.dart';
import 'package:marketplace_mobile/features/feedback/presentation/widgets/feedback_overlay.dart';
import 'package:marketplace_mobile/features/feedback/presentation/widgets/feedback_sheet.dart';
import 'package:marketplace_mobile/l10n/app_localizations.dart';

class _NoopFeedbackRepository implements FeedbackRepository {
  @override
  Future<FeedbackSubmission> submit(SubmitFeedbackInput input) async =>
      FeedbackSubmission(
        id: 'x',
        type: input.type,
        status: 'new',
        createdAt: DateTime.utc(2026),
      );

  @override
  Future<FeedbackAttachment> uploadAttachment({
    required File file,
    required FeedbackAttachmentKind kind,
    required String contentType,
    required String filename,
  }) async =>
      FeedbackAttachment(
        kind: kind,
        objectKey: 'k',
        contentType: contentType,
        sizeBytes: 1,
      );
}

class _FakeAuthNotifier extends StateNotifier<AuthState>
    implements AuthNotifier {
  _FakeAuthNotifier() : super(const AuthState(status: AuthStatus.authenticated));

  @override
  dynamic noSuchMethod(Invocation invocation) =>
      super.noSuchMethod(invocation);
}

GoRouter _stubRouter() => GoRouter(
      initialLocation: '/dashboard',
      routes: [
        GoRoute(
          path: '/dashboard',
          builder: (_, __) => const Scaffold(body: Center(child: Text('home'))),
        ),
      ],
    );

Widget _hostFab() {
  return ProviderScope(
    overrides: [
      feedbackRepositoryProvider.overrideWithValue(_NoopFeedbackRepository()),
      authProvider.overrideWith((_) => _FakeAuthNotifier()),
      appRouterProvider.overrideWithValue(_stubRouter()),
    ],
    child: MaterialApp(
      theme: AppTheme.light,
      locale: const Locale('fr'),
      localizationsDelegates: const [
        AppLocalizations.delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      supportedLocales: const [Locale('en'), Locale('fr')],
      home: const Scaffold(body: Center(child: FeedbackFab())),
    ),
  );
}

void main() {
  testWidgets('FAB shows the "Signaler" label', (tester) async {
    await tester.pumpWidget(_hostFab());
    await tester.pumpAndSettle();
    expect(find.text('Signaler'), findsOneWidget);
  });

  testWidgets('tapping the FAB opens the feedback sheet', (tester) async {
    tester.view.physicalSize = const Size(1000, 2600);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(_hostFab());
    await tester.pumpAndSettle();

    expect(find.byType(FeedbackSheet), findsNothing);
    await tester.tap(find.text('Signaler'));
    await tester.pumpAndSettle();

    expect(find.byType(FeedbackSheet), findsOneWidget);
    expect(find.text('Signaler un problème'), findsOneWidget);
  });

  group('FeedbackOverlay', () {
    testWidgets('paints the FAB above the child by default', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            authProvider.overrideWith((_) => _FakeAuthNotifier()),
            appRouterProvider.overrideWithValue(_stubRouter()),
          ],
          child: MaterialApp(
            theme: AppTheme.light,
            locale: const Locale('fr'),
            localizationsDelegates: const [
              AppLocalizations.delegate,
              GlobalMaterialLocalizations.delegate,
              GlobalWidgetsLocalizations.delegate,
              GlobalCupertinoLocalizations.delegate,
            ],
            supportedLocales: const [Locale('en'), Locale('fr')],
            home: const FeedbackOverlay(
              child: Scaffold(body: SizedBox.expand()),
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.byType(FeedbackFab), findsOneWidget);
    });

    testWidgets('hidden=true removes the FAB', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            authProvider.overrideWith((_) => _FakeAuthNotifier()),
            appRouterProvider.overrideWithValue(_stubRouter()),
          ],
          child: MaterialApp(
            theme: AppTheme.light,
            locale: const Locale('fr'),
            localizationsDelegates: const [
              AppLocalizations.delegate,
              GlobalMaterialLocalizations.delegate,
              GlobalWidgetsLocalizations.delegate,
              GlobalCupertinoLocalizations.delegate,
            ],
            supportedLocales: const [Locale('en'), Locale('fr')],
            home: const FeedbackOverlay(
              hidden: true,
              child: Scaffold(body: SizedBox.expand()),
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.byType(FeedbackFab), findsNothing);
    });
  });
}
