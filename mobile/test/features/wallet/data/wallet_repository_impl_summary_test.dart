import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:marketplace_mobile/features/wallet/data/exceptions/commission_kyc_required_exception.dart';
import 'package:marketplace_mobile/features/wallet/data/wallet_repository_impl.dart';

import '../../../helpers/fake_api_client.dart';

/// WALLET-UNIFY Run D — covers the new fetchSummary + withdraw
/// methods on [WalletRepositoryImpl] (POST /wallet/summary,
/// POST /wallet/withdraw).
void main() {
  late FakeApiClient fakeApi;
  late WalletRepositoryImpl repo;

  setUp(() {
    fakeApi = FakeApiClient();
    repo = WalletRepositoryImpl(fakeApi);
  });

  group('WalletRepositoryImpl.fetchSummary', () {
    test('GET unwraps the data envelope', () async {
      fakeApi.getHandlers['/api/v1/wallet/summary'] = (qs) async {
        return Response(
          requestOptions: RequestOptions(path: '/api/v1/wallet/summary'),
          statusCode: 200,
          data: {
            'data': {
              'currency': 'EUR',
              'total_cents': 1000,
              'available_cents': 500,
              'breakdown': {
                'missions': {'total_cents': 600, 'available_cents': 300},
                'commissions': {'total_cents': 400, 'available_cents': 200},
              },
              'recent_transactions': [],
            },
          },
        );
      };
      final summary = await repo.fetchSummary();
      expect(summary.currency, 'EUR');
      expect(summary.totalCents, 1000);
      expect(summary.missions.totalCents, 600);
      expect(summary.commissions.totalCents, 400);
    });

    test('forwards cursor as a query parameter', () async {
      Map<String, dynamic>? captured;
      fakeApi.getHandlers['/api/v1/wallet/summary'] = (qs) async {
        captured = qs;
        return Response(
          requestOptions: RequestOptions(path: '/api/v1/wallet/summary'),
          statusCode: 200,
          data: {'data': {}},
        );
      };
      await repo.fetchSummary(cursor: 'CURSOR123');
      expect(captured, isNotNull);
      expect(captured!['cursor'], 'CURSOR123');
    });

    test('handles a bare body without the data envelope', () async {
      fakeApi.getHandlers['/api/v1/wallet/summary'] = (_) async {
        return Response(
          requestOptions: RequestOptions(path: '/api/v1/wallet/summary'),
          statusCode: 200,
          data: {
            'currency': 'USD',
            'total_cents': 200,
          },
        );
      };
      final summary = await repo.fetchSummary();
      expect(summary.currency, 'USD');
      expect(summary.totalCents, 200);
    });
  });

  group('WalletRepositoryImpl.withdraw', () {
    test('returns the WithdrawResult on 200 full success', () async {
      fakeApi.postHandlers['/api/v1/wallet/withdraw'] = (_) async {
        return Response(
          requestOptions: RequestOptions(path: '/api/v1/wallet/withdraw'),
          statusCode: 200,
          data: {
            'data': {
              'drained_cents': 500,
              'missions_cents': 300,
              'commissions_cents': 200,
              'stripe_transfer_ids': ['tr_1'],
              'currency': 'EUR',
              'errors': [],
            },
          },
        );
      };
      final result = await repo.withdraw();
      expect(result.isFullSuccess, isTrue);
      expect(result.drainedCents, 500);
    });

    test('returns the WithdrawResult on 207 partial success', () async {
      fakeApi.postHandlers['/api/v1/wallet/withdraw'] = (_) async {
        return Response(
          requestOptions: RequestOptions(path: '/api/v1/wallet/withdraw'),
          statusCode: 207,
          data: {
            'data': {
              'drained_cents': 300,
              'missions_cents': 300,
              'commissions_cents': 0,
              'currency': 'EUR',
              'errors': [
                {
                  'source': 'commissions',
                  'code': 'failed',
                  'message': 'Stripe rejected',
                },
              ],
            },
          },
        );
      };
      final result = await repo.withdraw();
      expect(result.isPartialSuccess, isTrue);
      expect(result.errors, hasLength(1));
      expect(result.errors.first.source, 'commissions');
    });

    test('throws CommissionKYCRequiredException on 422', () async {
      fakeApi.postHandlers['/api/v1/wallet/withdraw'] = (_) async {
        throw DioException(
          requestOptions: RequestOptions(path: '/api/v1/wallet/withdraw'),
          response: Response(
            requestOptions:
                RequestOptions(path: '/api/v1/wallet/withdraw'),
            statusCode: 422,
            data: {
              'error': 'kyc_required',
              'message': 'Finalize Stripe onboarding',
              'onboarding_url': 'https://stripe.test/onboard',
            },
          ),
        );
      };
      try {
        await repo.withdraw();
        fail('expected CommissionKYCRequiredException');
      } on CommissionKYCRequiredException catch (kyc) {
        expect(kyc.onboardingUrl, 'https://stripe.test/onboard');
      }
    });

    test('forwards amount_cents in the request body when provided',
        () async {
      dynamic body;
      fakeApi.postHandlers['/api/v1/wallet/withdraw'] = (data) async {
        body = data;
        return Response(
          requestOptions: RequestOptions(path: '/api/v1/wallet/withdraw'),
          statusCode: 200,
          data: {'data': {'drained_cents': 0, 'errors': []}},
        );
      };
      await repo.withdraw(amountCents: 250);
      expect(body, isA<Map>());
      expect((body as Map)['amount_cents'], 250);
    });
  });
}
