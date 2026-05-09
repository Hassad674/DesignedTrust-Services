import '../entities/job_entity.dart';
import '../entities/job_application_entity.dart';

class CreateJobData {
  const CreateJobData({
    required this.title,
    required this.description,
    required this.skills,
    required this.applicantType,
    required this.budgetType,
    required this.minBudget,
    required this.maxBudget,
    this.paymentFrequency,
    this.durationWeeks,
    this.isIndefinite = false,
    this.descriptionType = 'text',
    this.videoUrl,
  });

  final String title;
  final String description;
  final List<String> skills;
  final String applicantType;
  final String budgetType;
  final int minBudget;
  final int maxBudget;
  final String? paymentFrequency;
  final int? durationWeeks;
  final bool isIndefinite;
  final String descriptionType;
  final String? videoUrl;
}

abstract class JobRepository {
  Future<JobEntity> createJob(CreateJobData data);
  Future<JobEntity> updateJob(String id, CreateJobData data);
  Future<JobEntity> getJob(String id);
  Future<List<JobEntity>> listMyJobs();
  Future<void> closeJob(String id);
  Future<void> reopenJob(String id);
  Future<void> deleteJob(String id);

  // Job applications
  Future<List<JobEntity>> listOpenJobs({String? cursor});
  /// Apply to a job. [applicantKind] is optional — when omitted the
  /// backend derives the persona from the user's role. Pass an explicit
  /// kind only for the referrer-enabled provider radio.
  Future<JobApplicationEntity> applyToJob(
    String jobId, {
    required String message,
    String? videoUrl,
    ApplicantKind? applicantKind,
  });
  Future<void> withdrawApplication(String applicationId);
  /// List applications for a job. [kindFilter] narrows the rows to a
  /// single applicant_kind; pass null for the unfiltered "Tous" view.
  Future<List<ApplicationWithProfile>> listJobApplications(
    String jobId, {
    String? cursor,
    ApplicantKind? kindFilter,
  });
  Future<List<ApplicationWithJob>> listMyApplications({String? cursor});
  Future<String> contactApplicant(String jobId, String applicantId);
  Future<bool> hasApplied(String jobId);
  Future<void> markApplicationsViewed(String jobId);

  // Credits
  Future<int> getCredits();
}
