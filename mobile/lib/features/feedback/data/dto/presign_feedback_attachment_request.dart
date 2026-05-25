import 'package:json_annotation/json_annotation.dart';

part 'presign_feedback_attachment_request.g.dart';

/// Wire-format DTO for the body of
/// `POST /api/v1/feedback/attachments/presign` (auth required).
///
/// Mirrors `PresignFeedbackAttachmentRequest` in
/// `backend/internal/handler/dto/request/feedback.go`. The filename is
/// accepted only for the human-facing UI — it never influences the
/// server-minted storage key.
@JsonSerializable(createFactory: false)
class PresignFeedbackAttachmentRequest {
  const PresignFeedbackAttachmentRequest({
    required this.kind,
    required this.contentType,
    required this.sizeBytes,
    required this.filename,
  });

  final String kind;

  @JsonKey(name: 'content_type')
  final String contentType;

  @JsonKey(name: 'size_bytes')
  final int sizeBytes;

  final String filename;

  Map<String, dynamic> toJson() =>
      _$PresignFeedbackAttachmentRequestToJson(this);
}
