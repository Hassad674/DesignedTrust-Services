import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:marketplace_mobile/features/wallet/data/exceptions/commission_kyc_required_exception.dart';
import 'package:marketplace_mobile/features/wallet/domain/entities/wallet_entity.dart';
import 'package:marketplace_mobile/features/wallet/domain/entities/wallet_summary_entity.dart';
import 'package:marketplace_mobile/features/wallet/domain/repositories/wallet_repository.dart';
import 'package:marketplace_mobile/features/wallet/presentation/providers/wallet_provider.dart';
import 'package:marketplace_mobile/features/wallet/presentation/widgets/wallet_unified_header.dart';
import 'package:marketplace_mobile/features/wallet/presentation/widgets/wallet_unified_history.dart';
import 'package:marketplace_mobile/l10n/app_localizations.dart';

/// Integration test that wires the WalletUnifiedHeader to the
/// walletSummaryProvider through a stub repository and asserts the
/// withdraw round-trip flips the screen state from idle → pending →
/// fresh summary on success. Mirrors the e2e flow without spinning
/// a real Stripe round-trip.
class _Recorder implements WalletRepository {
  WalletSummary current = const WalletSummary(
    totalCents: 1000,
    availableCents: 500,
    escrowedCents: 200,
    transmittedCents: 300,
    recentTransactions: [],
  );

  WithdrawResult Function()? withdrawResult;
  CommissionKYCRequiredException? withdrawKycError;
  int withdrawCalls = 0;

  @override
  Future<WalletSummary> fetchSummary({String? cursor}) async => current;

  @override
  Future<WithdrawResult> withdraw({int? amountCents}) async {
    withdrawCalls++;
    if (withdrawKycError != null) throw withdrawKycError!;
    final r = withdrawResult?.call() ?? const WithdrawResult();
    current = WalletSummary(
      totalCents: current.totalCents - r.drainedCents,
      availableCents: current.availableCents - r.drainedCents,
      escrowedCents: current.escrowedCents,
      transmittedCents: current.transmittedCents + r.drainedCents,
    );
    return r;
  }

  @override
  Future<WalletOverview> getWallet() async => throw UnimplementedError();
  @override
  Future<void> requestPayout() async => throw UnimplementedError();
  @override
  Future<void> retryFailedTransfer(String recordId) async =>
      throw UnimplementedError();
  @override
  Future<void> retryCommission(String commissionId) async =>
      throw UnimplementedError();
}

// Simple driver widget that reads the summary provider, renders the
// header, and calls withdraw on tap. Smaller than the full
// WalletScreen but exercises the same plumbing.
class _Driver extends ConsumerStatefulWidget {
  const _Driver();

  @override
  ConsumerState<_Driver> createState() => _DriverState();
}

class _DriverState extends ConsumerState<_Driver> {
  bool pending = false;
  String? error;

  Future<void> _withdraw() async {
    setState(() => pending = true);
    try {
      await ref.read(walletRepositoryProvider).withdraw();
      ref.invalidate(walletSummaryProvider);
    } on CommissionKYCRequiredException catch (e) {
      error = 'kyc:${e.onboardingUrl}';
    } finally {
      if (mounted) setState(() => pending = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final asyncSummary = ref.watch(walletSummaryProvider(null));
    return asyncSummary.when(
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (_, __) => const Text('error'),
      data: (summary) => Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          WalletUnifiedHeader(
            summary: summary,
            payoutPending: pending,
            canWithdraw: true,
            onWithdraw: _withdraw,
          ),
          if (error != null) Text('ERROR=$error'),
          const Expanded(child: WalletUnifiedHistory()),
        ],
      ),
    );
  }
}

void main() {
  testWidgets('full withdraw flow — full success drops available to 0',
      (tester) async {
    tester.view.physicalSize = const Size(900, 1800);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.reset);

    final repo = _Recorder()
      ..withdrawResult = () => const WithdrawResult(
            drainedCents: 500,
            missionsCents: 300,
            commissionsCents: 200,
          );
    await tester.pumpWidget(ProviderScope(
      overrides: [walletRepositoryProvider.overrideWithValue(repo)],
      child: const MaterialApp(
        localizationsDelegates: [
          AppLocalizations.delegate,
          GlobalMaterialLocalizations.delegate,
          GlobalWidgetsLocalizations.delegate,
          GlobalCupertinoLocalizations.delegate,
        ],
        supportedLocales: [Locale('fr'), Locale('en')],
        locale: Locale('fr'),
        home: Scaffold(body: _Driver()),
      ),
    ),);
    await tester.pumpAndSettle();
    // Initial state — Retirer enabled.
    expect(
      find.byKey(const ValueKey('wallet-unified-withdraw')),
      findsOneWidget,
    );
    // Tap to trigger withdraw.
    await tester.tap(
      find.byKey(const ValueKey('wallet-unified-withdraw')),
    );
    await tester.pumpAndSettle();
    expect(repo.withdrawCalls, 1);
    // Summary refreshed — available_cents should now be 0 → Retirer
    // is disabled.
    final button = tester.widget<ElevatedButton>(
      find.byKey(const ValueKey('wallet-unified-withdraw')),
    );
    expect(button.onPressed, isNull,
        reason:
            'After a successful drain, the next render must disable Retirer (available=0)',);
  });

  testWidgets('withdraw surfaces the KYC exception via setState',
      (tester) async {
    tester.view.physicalSize = const Size(900, 1800);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.reset);
    final repo = _Recorder()
      ..withdrawKycError = CommissionKYCRequiredException(
        message: 'Finalize Stripe onboarding',
        onboardingUrl: 'https://stripe.test/onboard',
      );
    await tester.pumpWidget(ProviderScope(
      overrides: [walletRepositoryProvider.overrideWithValue(repo)],
      child: const MaterialApp(
        localizationsDelegates: [
          AppLocalizations.delegate,
          GlobalMaterialLocalizations.delegate,
          GlobalWidgetsLocalizations.delegate,
          GlobalCupertinoLocalizations.delegate,
        ],
        supportedLocales: [Locale('fr'), Locale('en')],
        locale: Locale('fr'),
        home: Scaffold(body: _Driver()),
      ),
    ),);
    await tester.pumpAndSettle();
    await tester.tap(
      find.byKey(const ValueKey('wallet-unified-withdraw')),
    );
    await tester.pumpAndSettle();
    expect(
      find.text('ERROR=kyc:https://stripe.test/onboard'),
      findsOneWidget,
    );
  });
}
