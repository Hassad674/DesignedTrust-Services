import 'package:freezed_annotation/freezed_annotation.dart';

import 'feedback_type.dart';

part 'feedback_submission.freezed.dart';

/// The minimal acknowledgement returned after a successful `POST
/// /api/v1/feedback`. Reporters never receive triage internals — just
/// the id, type and status so the UI can show a confirmation.
@freezed
class FeedbackSubmission with _$FeedbackSubmission {
  const factory FeedbackSubmission({
    required String id,
    required FeedbackType type,
    required String status,
    required DateTime createdAt,
  }) = _FeedbackSubmission;
}
