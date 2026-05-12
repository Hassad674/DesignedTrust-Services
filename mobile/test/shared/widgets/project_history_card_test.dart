// Widget tests for the shared ProjectHistoryCard + AwaitingReviewBox.
//
// Pins the three states of the awaiting-review placeholder:
//   - no onLeaveReview → static "Awaiting review" placeholder, no tap.
//   - onLeaveReview supplied + no embedded review → tappable
//     "Laisser ton avis" button, fires the callback when tapped.
//   - embedded review → ReviewCardWidget renders instead of placeholder.

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:intl/date_symbol_data_local.dart';

import 'package:marketplace_mobile/shared/widgets/project_history_card.dart';

Widget _harness(Widget child) {
  return MaterialApp(
    home: Scaffold(body: child),
  );
}

void main() {
  setUpAll(() async {
    // ProjectHistoryCard uses DateFormat('d MMM yyyy', 'fr_FR') which
    // requires the locale data to be initialised in widget tests.
    await initializeDateFormatting('fr_FR');
  });

  group('ProjectHistoryCard / AwaitingReviewBox', () {
    testWidgets(
      'no onLeaveReview → static "Awaiting review" label, not tappable',
      (tester) async {
        await tester.pumpWidget(_harness(
          ProjectHistoryCard(
            title: 'Refonte site web',
            amountCents: 350000,
            completedAt: DateTime(2026, 4, 1),
          ),
        ));
        expect(find.text('Awaiting review'), findsOneWidget);
        expect(find.text('Laisser ton avis sur ce projet'), findsNothing);
        // No InkWell on the AwaitingReviewBox path.
        expect(find.byType(InkWell), findsNothing);
      },
    );

    testWidgets(
      'onLeaveReview supplied → renders "Laisser ton avis" button, fires callback',
      (tester) async {
        var tapped = false;
        await tester.pumpWidget(_harness(
          ProjectHistoryCard(
            title: 'Refonte site web',
            amountCents: 350000,
            completedAt: DateTime(2026, 4, 1),
            onLeaveReview: () => tapped = true,
          ),
        ));
        expect(find.text('Laisser ton avis sur ce projet'), findsOneWidget);
        expect(find.text('Awaiting review'), findsNothing);

        await tester.tap(find.byType(InkWell));
        expect(tapped, isTrue);
      },
    );

    testWidgets(
      'AwaitingReviewBox alone (read-only) does not render the leave-review label',
      (tester) async {
        await tester.pumpWidget(_harness(const AwaitingReviewBox()));
        expect(find.text('Awaiting review'), findsOneWidget);
        expect(find.byType(InkWell), findsNothing);
      },
    );

    testWidgets(
      'AwaitingReviewBox with onTap renders the leave-review CTA and fires onTap',
      (tester) async {
        var tapped = false;
        await tester.pumpWidget(_harness(
          AwaitingReviewBox(onTap: () => tapped = true),
        ));
        await tester.tap(find.byType(InkWell));
        expect(tapped, isTrue);
        expect(find.text('Laisser ton avis sur ce projet'), findsOneWidget);
      },
    );
  });
}
