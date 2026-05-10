// Widget tests for the TwoFactorSection (Sécurité toggle).
//
// The widget owns a local `_enabled` flag (see widget docstring for the
// caveat — `/auth/me` does not yet expose the persisted flag). These
// tests cover the rendered states + the dialog flow without hitting the
// network: TwoFactorApi calls are stubbed via a Riverpod override that
// substitutes the apiClientProvider with a fake recording client.

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:marketplace_mobile/core/network/api_client.dart';
import 'package:marketplace_mobile/core/storage/secure_storage.dart';
import 'package:marketplace_mobile/core/theme/app_theme.dart';
import 'package:marketplace_mobile/features/auth/presentation/widgets/two_factor_section.dart';
import 'package:marketplace_mobile/l10n/app_localizations.dart';

class _FakeStorage extends SecureStorageService {
  @override
  Future<void> saveTokens(String accessToken, String refreshToken) async {}
  @override
  Future<String?> getAccessToken() async => null;
  @override
  Future<String?> getRefreshToken() async => null;
  @override
  Future<void> clearTokens() async {}
  @override
  Future<bool> hasTokens() async => false;
  @override
  Future<void> saveUser(Map<String, dynamic> userJson) async {}
  @override
  Future<Map<String, dynamic>?> getUser() async => null;
  @override
  Future<void> clearAll() async {}
}

/// Recording fake — captures path + body, returns a 200 with empty body.
class _RecordingApiClient extends ApiClient {
  _RecordingApiClient() : super(storage: _FakeStorage());

  final List<({String path, dynamic data})> calls = [];

  @override
  Future<Response<T>> post<T>(String path, {dynamic data}) async {
    calls.add((path: path, data: data));
    return Response<T>(
      requestOptions: RequestOptions(path: path),
      statusCode: 200,
    );
  }

  @override
  Future<Response<T>> get<T>(
    String path, {
    Map<String, dynamic>? queryParameters,
    Options? options,
  }) async {
    return Response<T>(
      requestOptions: RequestOptions(path: path),
      statusCode: 200,
    );
  }
}

Widget _buildHost({required _RecordingApiClient api}) {
  return ProviderScope(
    overrides: [
      apiClientProvider.overrideWithValue(api),
    ],
    child: MaterialApp(
      theme: AppTheme.light,
      localizationsDelegates: const [
        AppLocalizations.delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      supportedLocales: const [Locale('en')],
      home: const Scaffold(
        body: Padding(
          padding: EdgeInsets.all(16),
          child: TwoFactorSection(),
        ),
      ),
    ),
  );
}

void main() {
  group('TwoFactorSection initial render', () {
    testWidgets('shows OFF copy and an unchecked switch', (tester) async {
      final api = _RecordingApiClient();
      await tester.pumpWidget(_buildHost(api: api));
      await tester.pumpAndSettle();

      expect(find.text('Email 2FA'), findsOneWidget);
      expect(
        find.text('Inactive. Enable 2FA to harden your account.'),
        findsOneWidget,
      );

      final sw = tester.widget<Switch>(find.byType(Switch));
      expect(sw.value, isFalse);
      // No network call yet.
      expect(api.calls, isEmpty);
    });
  });

  group('TwoFactorSection enable flow', () {
    testWidgets(
      'tapping the switch posts to /me/two-factor/enable then opens code dialog',
      (tester) async {
        final api = _RecordingApiClient();
        await tester.pumpWidget(_buildHost(api: api));
        await tester.pumpAndSettle();

        await tester.tap(find.byType(Switch));
        await tester.pumpAndSettle();

        // First call kicks off the challenge.
        expect(api.calls, hasLength(1));
        expect(api.calls.first.path, '/api/v1/me/two-factor/enable');

        // Modal is open.
        expect(find.text('Two-factor authentication'), findsOneWidget);
        expect(
          find.textContaining(
            'We just emailed a 6-digit code',
          ),
          findsOneWidget,
        );
      },
    );

    testWidgets(
      'submitting an invalid (short) code shows the length error',
      (tester) async {
        final api = _RecordingApiClient();
        await tester.pumpWidget(_buildHost(api: api));
        await tester.pumpAndSettle();

        await tester.tap(find.byType(Switch));
        await tester.pumpAndSettle();

        final codeField = find.byType(TextField).first;
        await tester.enterText(codeField, '123');
        await tester.tap(find.widgetWithText(FilledButton, 'Enable 2FA'));
        await tester.pump();

        expect(find.text('The code must be 6 digits.'), findsOneWidget);
      },
    );

    testWidgets(
      'submitting a valid code POSTs again and flips the switch ON',
      (tester) async {
        final api = _RecordingApiClient();
        await tester.pumpWidget(_buildHost(api: api));
        await tester.pumpAndSettle();

        await tester.tap(find.byType(Switch));
        await tester.pumpAndSettle();

        final codeField = find.byType(TextField).first;
        await tester.enterText(codeField, '654321');
        await tester.tap(find.widgetWithText(FilledButton, 'Enable 2FA'));
        await tester.pumpAndSettle();

        // Two calls: start + confirm.
        expect(api.calls, hasLength(2));
        expect(api.calls[1].path, '/api/v1/me/two-factor/enable');
        expect(api.calls[1].data, {'code': '654321'});

        // Switch is now ON.
        final sw = tester.widget<Switch>(find.byType(Switch));
        expect(sw.value, isTrue);

        // ON description rendered.
        expect(
          find.text('Active. A code will be required at every sign in.'),
          findsOneWidget,
        );
      },
    );

    testWidgets(
      'cancelling the dialog leaves the switch OFF',
      (tester) async {
        final api = _RecordingApiClient();
        await tester.pumpWidget(_buildHost(api: api));
        await tester.pumpAndSettle();

        await tester.tap(find.byType(Switch));
        await tester.pumpAndSettle();

        await tester.tap(find.widgetWithText(TextButton, 'Cancel'));
        await tester.pumpAndSettle();

        final sw = tester.widget<Switch>(find.byType(Switch));
        expect(sw.value, isFalse);
        // Only the start call happened.
        expect(api.calls, hasLength(1));
      },
    );
  });
}
