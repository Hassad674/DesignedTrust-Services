import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:marketplace_mobile/features/referral/presentation/widgets/end_intro_confirmation_dialog.dart';
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

void main() {
  group('EndIntroConfirmationDialog', () {
    testWidgets('renders title, body, cancel and confirm buttons',
        (tester) async {
      await tester.pumpWidget(_wrap(const EndIntroConfirmationDialog(
        providerName: 'Acme',
        clientName: 'Globex',
      ),),);
      await tester.pumpAndSettle();
      expect(find.text('Terminer la mise en relation'), findsOneWidget);
      expect(find.textContaining('Acme'), findsOneWidget);
      expect(find.textContaining('Globex'), findsOneWidget);
      expect(find.byKey(const ValueKey('end-intro-cancel')), findsOneWidget);
      expect(find.byKey(const ValueKey('end-intro-confirm')), findsOneWidget);
    });

    testWidgets('uses fallback labels when names are missing',
        (tester) async {
      await tester.pumpWidget(_wrap(const EndIntroConfirmationDialog()));
      await tester.pumpAndSettle();
      // FR fallback strings come from the ARB.
      expect(find.textContaining('le prestataire'), findsOneWidget);
      expect(find.textContaining('le client'), findsOneWidget);
    });

    testWidgets('confirm + cancel surface via showDialog result',
        (tester) async {
      bool? confirmed;
      bool? cancelled;

      Future<void> openAndDismiss({required String key}) async {
        confirmed = null;
        cancelled = null;
        late BuildContext capturedCtx;
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
            home: Builder(builder: (ctx) {
              capturedCtx = ctx;
              return const Scaffold(body: SizedBox.shrink());
            },),
          ),
        );
        final future = showEndIntroConfirmationDialog(
          context: capturedCtx,
          providerName: 'Acme',
          clientName: 'Globex',
        );
        await tester.pumpAndSettle();
        await tester.tap(find.byKey(ValueKey(key)));
        final result = await future;
        if (key == 'end-intro-confirm') {
          confirmed = result;
        } else {
          cancelled = result;
        }
      }

      await openAndDismiss(key: 'end-intro-confirm');
      expect(confirmed, isTrue);

      await openAndDismiss(key: 'end-intro-cancel');
      expect(cancelled, isFalse);
    });

    testWidgets('confirm button disabled while pending=true',
        (tester) async {
      await tester.pumpWidget(_wrap(const EndIntroConfirmationDialog(
        pending: true,
      ),),);
      // Spinner animates forever — single pump is enough to assert
      // the disabled state without hanging pumpAndSettle.
      await tester.pump();
      final confirmButton = tester.widget<FilledButton>(
        find.byKey(const ValueKey('end-intro-confirm')),
      );
      expect(confirmButton.onPressed, isNull);
      expect(find.byType(CircularProgressIndicator), findsOneWidget);
    });
  });
}
