import '../entities/profile_completion_report.dart';

/// Domain port for the profile-completion endpoint. Single read
/// method — no mutations; the bar only displays the report computed
/// by the backend.
abstract class ProfileCompletionRepository {
  /// Returns the report for the authenticated user's current org.
  Future<ProfileCompletionReport> getMy();
}
