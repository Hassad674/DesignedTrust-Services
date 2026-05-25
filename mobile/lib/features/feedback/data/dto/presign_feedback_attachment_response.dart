import 'package:json_annotation/json_annotation.dart';

import '../../domain/entities/feedback_attachment_kind.dart';
import '../../domain/entities/presigned_feedback_upload.dart';

part 'presign_feedback_attachment_response.g.dart';

/// Wire-format DTO for the `data` payload of
/// `POST /api/v1/feedback/attachments/presign`.
///
/// Mirrors `PresignFeedbackAttachmentResponse` in
/// `backend/internal/handler/dto/response/feedback.go` — the `url` JSON
/// key carries the short-lived PUT URL.
@JsonSerializable()
class PresignFeedbackAttachmentResponse {
  const PresignFeedbackAttachmentResponse({
    required this.uploadUrl,
    required this.objectKey,
    required this.kind,
  });

  @JsonKey(name: 'url')
  final String uploadUrl;

  @JsonKey(name: 'object_key')
  final String objectKey;

  final String kind;

  factory PresignFeedbackAttachmentResponse.fromJson(
    Map<String, dynamic> json,
  ) =>
      _$PresignFeedbackAttachmentResponseFromJson(json);

  Map<String, dynamic> toJson() =>
      _$PresignFeedbackAttachmentResponseToJson(this);

  PresignedFeedbackUpload toDomain() => PresignedFeedbackUpload(
        uploadUrl: uploadUrl,
        objectKey: objectKey,
        kind: FeedbackAttachmentKind.values.firstWhere(
          (k) => k.wireValue == kind,
          orElse: () => FeedbackAttachmentKind.image,
        ),
      );
}
