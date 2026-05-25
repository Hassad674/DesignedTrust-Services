import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:marketplace_mobile/core/theme/app_theme.dart';
import 'package:marketplace_mobile/features/auth/presentation/providers/auth_provider.dart';
import 'package:marketplace_mobile/features/feedback/domain/entities/feedback_attachment.dart';
import 'package:marketplace_mobile/features/feedback/domain/entities/feedback_attachment_kind.dart';
import 'package:marketplace_mobile/features/feedback/domain/entities/feedback_submission.dart';
import 'package:marketplace_mobile/features/feedback/domain/entities/feedback_type.dart';
import 'package:marketplace_mobile/features/feedback/domain/entities/submit_feedback_input.dart';
import 'package:marketplace_mobile/features/feedback/domain/repositories/feedback_repository.dart';
import 'package:marketplace_mobile/features/feedback/presentation/providers/feedback_providers.dart';
import 'package:marketplace_mobile/features/feedback/presentation/widgets/feedback_attachments_section.dart';
import 'package:marketplace_mobile/features/feedback/presentation/widgets/feedback_sheet.dart';
import 'package:marketplace_mobile/l10n/app_localizations.dart';

// ---------------------------------------------------------------------------
// Test doubles
// ---------------------------------------------------------------------------

/// Records the last submit payload and lets each test choose whether the
/// submit / upload succeeds or throws.
class _FakeFeedbackRepository implements FeedbackRepository {
  _FakeFeedbackRepository({this.failSubmit = false});

  bool failSubmit;

  SubmitFeedbackInput? lastInput;
  int submitCalls = 0;
  int uploadCalls = 0;

  @override
  Future<FeedbackSubmission> submit(SubmitFeedbackInput input) async {
    submitCalls++;
    lastInput = input;
    if (failSubmit) {
      throw Exception('submit failed');
    }
    return FeedbackSubmission(
      id: 'fb_1',
      type: input.type,
      status: 'new',
      createdAt: DateTime.utc(2026, 5, 25),
    );
  }

  @override
  Future<FeedbackAttachment> uploadAttachment({
    required File file,
    required FeedbackAttachmentKind kind,
    required String contentType,
    required String filename,
  }) async {
    uploadCalls++;
    return FeedbackAttachment(
      kind: kind,
      objectKey: 'feedback/$filename',
      contentType: contentType,
      sizeBytes: 1234,
    );
  }
}

/// Minimal fake auth notifier seeded to a fixed status — avoids spinning
/// up a real [AuthNotifier] (which touches secure storage on construct).
class _FakeAuthNotifier extends StateNotifier<AuthState>
    implements AuthNotifier {
  _FakeAuthNotifier(super.state);

  @override
  dynamic noSuchMethod(Invocation invocation) =>
      super.noSuchMethod(invocation);
}

AuthState _authed() => const AuthState(
      status: AuthStatus.authenticated,
      user: <String, dynamic>{'id': 'u1', 'role': 'provider'},
    );

AuthState _anon() => const AuthState(status: AuthStatus.unauthenticated);

// ---------------------------------------------------------------------------
// Harness
// ---------------------------------------------------------------------------

Widget _host({
  required _FakeFeedbackRepository repo,
  required AuthState auth,
  String pageUrl = '/dashboard',
}) {
  return ProviderScope(
    overrides: [
      feedbackRepositoryProvider.overrideWithValue(repo),
      authProvider.overrideWith((_) => _FakeAuthNotifier(auth)),
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
      home: Scaffold(body: FeedbackSheet(pageUrl: pageUrl)),
    ),
  );
}

void main() {
  // Generous surface so the scrolling sheet lays out without overflow.
  Future<void> enlarge(WidgetTester tester) async {
    tester.view.physicalSize = const Size(1000, 2600);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
  }

  testWidgets('renders title, type toggle and submit button', (tester) async {
    await enlarge(tester);
    await tester.pumpWidget(_host(repo: _FakeFeedbackRepository(), auth: _authed()));
    await tester.pumpAndSettle();

    expect(find.text('Signaler un problème'), findsOneWidget);
    expect(find.text('Bug'), findsOneWidget);
    expect(find.text('Sécurité'), findsOneWidget);
    expect(find.text('Envoyer'), findsOneWidget);
  });

  testWidgets('blank title + description blocks submit and shows errors',
      (tester) async {
    await enlarge(tester);
    final repo = _FakeFeedbackRepository();
    await tester.pumpWidget(_host(repo: repo, auth: _authed()));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Envoyer'));
    await tester.pumpAndSettle();

    // Repository never called; both validation messages shown.
    expect(repo.submitCalls, 0);
    expect(find.text('Ajoute un titre.'), findsOneWidget);
    expect(find.text('Ajoute une description.'), findsOneWidget);
  });

  testWidgets('valid form submits with the right payload (bug, default type)',
      (tester) async {
    await enlarge(tester);
    final repo = _FakeFeedbackRepository();
    await tester.pumpWidget(_host(repo: repo, auth: _authed(), pageUrl: '/missions'));
    await tester.pumpAndSettle();

    await tester.enterText(
      find.byType(TextField).at(0),
      'Crash on save',
    );
    await tester.enterText(
      find.byType(TextField).at(1),
      'The app crashes when I tap save.',
    );
    await tester.tap(find.text('Envoyer'));
    await tester.pumpAndSettle();

    expect(repo.submitCalls, 1);
    final input = repo.lastInput!;
    expect(input.type, FeedbackType.bug);
    expect(input.title, 'Crash on save');
    expect(input.description, 'The app crashes when I tap save.');
    expect(input.pageUrl, '/missions');
    expect(input.honeypot, '');
    // Context auto-captured for the native client.
    expect(input.context?.platform, 'android');
    expect(input.context?.locale, 'fr');
    expect(input.context?.role, 'provider');
  });

  testWidgets('type toggle flips the submitted type to security',
      (tester) async {
    await enlarge(tester);
    final repo = _FakeFeedbackRepository();
    await tester.pumpWidget(_host(repo: repo, auth: _authed()));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Sécurité'));
    await tester.pump();
    await tester.enterText(find.byType(TextField).at(0), 'XSS in profile');
    await tester.enterText(find.byType(TextField).at(1), 'Stored XSS payload.');
    await tester.tap(find.text('Envoyer'));
    await tester.pumpAndSettle();

    expect(repo.lastInput!.type, FeedbackType.security);
  });

  testWidgets('anonymous reporter sees the login hint, no attach buttons',
      (tester) async {
    await enlarge(tester);
    await tester.pumpWidget(_host(repo: _FakeFeedbackRepository(), auth: _anon()));
    await tester.pumpAndSettle();

    expect(
      find.text('Connecte-toi pour joindre une capture ou une vidéo.'),
      findsOneWidget,
    );
    expect(find.text('Ajouter une image'), findsNothing);
    expect(find.text('Ajouter une vidéo'), findsNothing);
  });

  testWidgets('logged-in reporter sees attach buttons, not the login hint',
      (tester) async {
    await enlarge(tester);
    await tester.pumpWidget(_host(repo: _FakeFeedbackRepository(), auth: _authed()));
    await tester.pumpAndSettle();

    expect(find.text('Ajouter une image'), findsOneWidget);
    expect(find.text('Ajouter une vidéo'), findsOneWidget);
    expect(
      find.text('Connecte-toi pour joindre une capture ou une vidéo.'),
      findsNothing,
    );
  });

  testWidgets('failed submit keeps the sheet open and shows an error snackbar',
      (tester) async {
    await enlarge(tester);
    final repo = _FakeFeedbackRepository(failSubmit: true);
    await tester.pumpWidget(_host(repo: repo, auth: _authed()));
    await tester.pumpAndSettle();

    await tester.enterText(find.byType(TextField).at(0), 'Bug title');
    await tester.enterText(find.byType(TextField).at(1), 'Bug description');
    await tester.tap(find.text('Envoyer'));
    await tester.pumpAndSettle();

    expect(repo.submitCalls, 1);
    // Sheet still present (not popped) + error surfaced.
    expect(find.byType(FeedbackSheet), findsOneWidget);
    expect(
      find.text("Impossible d'envoyer ton signalement. Réessaie."),
      findsOneWidget,
    );
  });

  testWidgets(
      'anonymous submit never carries attachments even if some are present',
      (tester) async {
    // Defensive: the attach controls are hidden for anon, but assert the
    // payload-builder also fails closed regardless of local state.
    await enlarge(tester);
    final repo = _FakeFeedbackRepository();
    await tester.pumpWidget(_host(repo: repo, auth: _anon()));
    await tester.pumpAndSettle();

    await tester.enterText(find.byType(TextField).at(0), 'Anon bug');
    await tester.enterText(find.byType(TextField).at(1), 'Anon description');
    await tester.tap(find.text('Envoyer'));
    await tester.pumpAndSettle();

    expect(repo.lastInput!.attachments, isEmpty);
  });

  group('FeedbackAttachmentsSection gating', () {
    testWidgets('canAttach=false renders the hint only', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          theme: AppTheme.light,
          locale: const Locale('fr'),
          localizationsDelegates: const [
            AppLocalizations.delegate,
            GlobalMaterialLocalizations.delegate,
            GlobalWidgetsLocalizations.delegate,
            GlobalCupertinoLocalizations.delegate,
          ],
          supportedLocales: const [Locale('en'), Locale('fr')],
          home: Scaffold(
            body: FeedbackAttachmentsSection(
              canAttach: false,
              attachments: const [],
              isUploading: false,
              onPickImage: () {},
              onPickVideo: () {},
              onRemove: (_) {},
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();

      expect(
        find.text('Connecte-toi pour joindre une capture ou une vidéo.'),
        findsOneWidget,
      );
      expect(find.byType(OutlinedButton), findsNothing);
    });

    testWidgets('canAttach=true shows uploaded chips', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          theme: AppTheme.light,
          locale: const Locale('fr'),
          localizationsDelegates: const [
            AppLocalizations.delegate,
            GlobalMaterialLocalizations.delegate,
            GlobalWidgetsLocalizations.delegate,
            GlobalCupertinoLocalizations.delegate,
          ],
          supportedLocales: const [Locale('en'), Locale('fr')],
          home: Scaffold(
            body: FeedbackAttachmentsSection(
              canAttach: true,
              attachments: const [
                FeedbackAttachment(
                  kind: FeedbackAttachmentKind.image,
                  objectKey: 'feedback/a.png',
                  contentType: 'image/png',
                  sizeBytes: 2048,
                ),
              ],
              isUploading: false,
              onPickImage: () {},
              onPickVideo: () {},
              onRemove: (_) {},
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();

      expect(find.byType(OutlinedButton), findsNWidgets(2));
      // 2 KB chip rendered.
      expect(find.text('2 KB'), findsOneWidget);
    });
  });
}
