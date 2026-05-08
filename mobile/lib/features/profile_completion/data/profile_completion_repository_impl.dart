import '../../../core/network/api_client.dart';
import '../domain/entities/profile_completion_report.dart';
import '../domain/repositories/profile_completion_repository.dart';

/// Dio-backed implementation of [ProfileCompletionRepository].
///
/// Endpoint:
///   GET /api/v1/me/profile/completion -> getMy
///
/// Tolerates both `{ "data": ... }` envelopes and raw payloads so a
/// future envelope flip on the backend does not break the screen.
class ProfileCompletionRepositoryImpl implements ProfileCompletionRepository {
  ProfileCompletionRepositoryImpl(this._api);

  final ApiClient _api;

  @override
  Future<ProfileCompletionReport> getMy() async {
    final response = await _api.get('/api/v1/me/profile/completion');
    final body = _unwrapMap(response.data);
    if (body == null) return ProfileCompletionReport.empty;
    return ProfileCompletionReport.fromJson(body);
  }

  Map<String, dynamic>? _unwrapMap(Object? raw) {
    if (raw is Map<String, dynamic>) {
      final data = raw['data'];
      if (data is Map<String, dynamic>) return data;
      return raw;
    }
    return null;
  }
}
