import 'package:json_annotation/json_annotation.dart';

import '../../domain/entities/feedback_submission.dart';
import '../../domain/entities/feedback_type.dart';

part 'submit_feedback_response.g.dart';

/// Wire-format DTO for the `data` payload of `POST /api/v1/feedback`.
///
/// Mirrors `SubmitFeedbackResponse` in
/// `backend/internal/handler/dto/response/feedback.go`. The `type` is
/// the lowercase enum token; an unrecognised value defaults to
/// [FeedbackType.bug] so a forward-compatible server change never
/// crashes the client.
@JsonSerializable()
class SubmitFeedbackResponse {
  const SubmitFeedbackResponse({
    required this.id,
    required this.type,
    required this.status,
    required this.createdAt,
  });

  final String id;
  final String type;
  final String status;

  @JsonKey(name: 'created_at')
  final String createdAt;

  factory SubmitFeedbackResponse.fromJson(Map<String, dynamic> json) =>
      _$SubmitFeedbackResponseFromJson(json);

  Map<String, dynamic> toJson() => _$SubmitFeedbackResponseToJson(this);

  FeedbackSubmission toDomain() => FeedbackSubmission(
        id: id,
        type: FeedbackType.values.firstWhere(
          (t) => t.wireValue == type,
          orElse: () => FeedbackType.bug,
        ),
        status: status,
        createdAt: DateTime.tryParse(createdAt) ?? DateTime.now(),
      );
}
