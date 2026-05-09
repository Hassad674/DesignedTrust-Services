/// Unit tests for the ApplicantKind enum + JobApplicationEntity.fromJson.
///
/// Asserts:
///   - the wire mapping covers all three personas
///   - missing/unknown applicant_kind defaults to freelance
///   - JSON deserialisation pulls the kind through
///   - JSON deserialisation tolerates an absent kind (legacy backend
///     responses pre-138 migration)
import 'package:flutter_test/flutter_test.dart';
import 'package:marketplace_mobile/features/job/domain/entities/job_application_entity.dart';

void main() {
  group('ApplicantKind', () {
    test('wire serialisation', () {
      expect(ApplicantKind.freelance.wire, 'freelance');
      expect(ApplicantKind.agency.wire, 'agency');
      expect(ApplicantKind.referrer.wire, 'referrer');
    });

    test('fromWire deserialisation', () {
      expect(ApplicantKind.fromWire('freelance'), ApplicantKind.freelance);
      expect(ApplicantKind.fromWire('agency'), ApplicantKind.agency);
      expect(ApplicantKind.fromWire('referrer'), ApplicantKind.referrer);
    });

    test('fromWire defaults to freelance for null/unknown', () {
      expect(ApplicantKind.fromWire(null), ApplicantKind.freelance);
      expect(ApplicantKind.fromWire('hacker'), ApplicantKind.freelance);
      expect(ApplicantKind.fromWire(''), ApplicantKind.freelance);
    });
  });

  group('JobApplicationEntity.fromJson', () {
    Map<String, dynamic> baseJson() => <String, dynamic>{
          'id': 'app-1',
          'job_id': 'job-1',
          'applicant_id': 'org-1',
          'message': 'hello',
          'created_at': '2026-04-01T00:00:00Z',
        };

    test('reads the applicant_kind value', () {
      final json = baseJson()..['applicant_kind'] = 'referrer';
      final entity = JobApplicationEntity.fromJson(json);
      expect(entity.applicantKind, ApplicantKind.referrer);
    });

    test('defaults to freelance when applicant_kind is missing', () {
      final json = baseJson();
      final entity = JobApplicationEntity.fromJson(json);
      expect(entity.applicantKind, ApplicantKind.freelance);
    });

    test('defaults to freelance when applicant_kind is null', () {
      final json = baseJson()..['applicant_kind'] = null;
      final entity = JobApplicationEntity.fromJson(json);
      expect(entity.applicantKind, ApplicantKind.freelance);
    });

    test('preserves agency kind', () {
      final json = baseJson()..['applicant_kind'] = 'agency';
      final entity = JobApplicationEntity.fromJson(json);
      expect(entity.applicantKind, ApplicantKind.agency);
    });
  });
}
