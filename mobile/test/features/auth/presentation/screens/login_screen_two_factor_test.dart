// Tests for the 2FA branch of the login screen state machine.
//
// Mounting strategy mirrors `login_screen_test.dart`: a real AuthNotifier
// with fake deps, then we force the state via the protected setter so the
// initial render exercises the OTP form path.

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:marketplace_mobile/core/network/api_client.dart';
import 'package:marketplace_mobile/core/storage/secure_storage.dart';
import 'package:marketplace_mobile/core/theme/app_theme.dart';
import 'package:marketplace_mobile/features/auth/data/two_factor_api.dart';
import 'package:marketplace_mobile/features/auth/presentation/providers/auth_provider.dart';
import 'package:marketplace_mobile/features/auth/presentation/screens/login_screen.dart';
import 'package:marketplace_mobile/l10n/app_localizations.dart';

class _FakeSecureStorage extends SecureStorageService {
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

class _FakeApiClient extends ApiClient {
  _FakeApiClient() : super(storage: _FakeSecureStorage());

  @override
  Future<Response<T>> get<T>(
    String path, {
    Map<String, dynamic>? queryParameters,
    Options? options,
  }) async {
    throw DioException(
      requestOptions: RequestOptions(path: path),
      type: DioExceptionType.connectionError,
    );
  }

  @override
  Future<Response<T>> post<T>(String path, {dynamic data}) async {
    throw DioException(
      requestOptions: RequestOptions(path: path),
      type: DioExceptionType.connectionError,
    );
  }
}

AuthNotifier _notifierWithState(AuthState state) {
  final notifier = AuthNotifier(
    apiClient: _FakeApiClient(),
    storage: _FakeSecureStorage(),
  );
  // ignore: invalid_use_of_protected_member
  notifier.state = state;
  return notifier;
}

Widget _buildScreen(AuthState state) {
  return ProviderScope(
    overrides: [
      authProvider.overrideWith((_) => _notifierWithState(state)),
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
      home: const LoginScreen(),
    ),
  );
}

void main() {
  group('LoginScreen 2FA branch', () {
    testWidgets('renders OTP form when pendingTwoFactor is set',
        (tester) async {
      await tester.pumpWidget(
        _buildScreen(
          const AuthState(
            status: AuthStatus.unauthenticated,
            pendingTwoFactor: TwoFactorChallenge(
              userId: 'u-1',
              challengeId: 'c-1',
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();

      // The 2FA title replaces the password title.
      expect(find.text("Confirm it's really you."), findsOneWidget);
      // Verify CTA shows.
      expect(
        find.widgetWithText(ElevatedButton, 'Verify code'),
        findsOneWidget,
      );
      // Resend pill shows.
      expect(
        find.widgetWithText(OutlinedButton, 'Resend code'),
        findsOneWidget,
      );
      // Back link shows.
      expect(
        find.widgetWithText(TextButton, 'Back to sign in'),
        findsOneWidget,
      );
      // Password field is gone.
      expect(find.text('Your password'), findsNothing);
    });

    testWidgets('OTP field rejects fewer than 6 digits', (tester) async {
      await tester.pumpWidget(
        _buildScreen(
          const AuthState(
            status: AuthStatus.unauthenticated,
            pendingTwoFactor: TwoFactorChallenge(
              userId: 'u-1',
              challengeId: 'c-1',
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();

      final field = find.byType(TextFormField);
      expect(field, findsOneWidget);
      await tester.enterText(field, '123');

      // Tap the Verify button to trigger validation.
      await tester.tap(find.widgetWithText(ElevatedButton, 'Verify code'));
      await tester.pump();

      expect(find.text('The code must be 6 digits.'), findsOneWidget);
    });

    testWidgets('renders password form when pendingTwoFactor is null',
        (tester) async {
      await tester.pumpWidget(
        _buildScreen(
          const AuthState(status: AuthStatus.unauthenticated),
        ),
      );
      await tester.pumpAndSettle();

      // Password placeholder is back.
      expect(find.text('Your password'), findsOneWidget);
      // 2FA-specific buttons are absent.
      expect(
        find.widgetWithText(ElevatedButton, 'Verify code'),
        findsNothing,
      );
    });

    testWidgets('Back to sign in clears pending 2FA via cancelPendingTwoFactor',
        (tester) async {
      await tester.pumpWidget(
        _buildScreen(
          const AuthState(
            status: AuthStatus.unauthenticated,
            pendingTwoFactor: TwoFactorChallenge(
              userId: 'u-1',
              challengeId: 'c-1',
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(TextButton, 'Back to sign in'));
      await tester.pumpAndSettle();

      // Password field is now visible.
      expect(find.text('Your password'), findsOneWidget);
      // The Verify CTA is gone.
      expect(
        find.widgetWithText(ElevatedButton, 'Verify code'),
        findsNothing,
      );
    });
  });
}
