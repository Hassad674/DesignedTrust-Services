import 'package:json_annotation/json_annotation.dart';

import '../../domain/entities/feedback_attachment.dart';
import '../../domain/entities/submit_feedback_input.dart';

part 'submit_feedback_request.g.dart';

/// Wire-format DTO for the body of `POST /api/v1/feedback`.
///
/// Mirrors `SubmitFeedbackRequest` in
/// `backend/internal/handler/dto/request/feedback.go` (authoritative:
/// `openapi.golden.json`). Every required wire field is always present
/// — empty strings rather than omitted keys — and `context` is null
/// when no context is captured.
@JsonSerializable(explicitToJson: true, createFactory: false)
class SubmitFeedbackRequest {
  const SubmitFeedbackRequest({
    required this.type,
    required this.title,
    required this.description,
    required this.pageUrl,
    required this.context,
    required this.reporterEmail,
    required this.attachmentKeys,
    required this.hp,
  });

  final String type;
  final String title;
  final String description;

  @JsonKey(name: 'page_url')
  final String pageUrl;

  final FeedbackContextRequest? context;

  @JsonKey(name: 'reporter_email')
  final String reporterEmail;

  @JsonKey(name: 'attachment_keys')
  final List<SubmitFeedbackAttachmentRequest> attachmentKeys;

  /// Honeypot — always empty from a real client.
  final String hp;

  factory SubmitFeedbackRequest.fromDomain(SubmitFeedbackInput input) {
    final ctx = input.context;
    return SubmitFeedbackRequest(
      type: input.type.wireValue,
      title: input.title,
      description: input.description,
      pageUrl: input.pageUrl,
      context: ctx == null ? null : FeedbackContextRequest.fromDomain(ctx),
      reporterEmail: input.reporterEmail,
      attachmentKeys: input.attachments
          .map(SubmitFeedbackAttachmentRequest.fromDomain)
          .toList(),
      hp: input.honeypot,
    );
  }

  Map<String, dynamic> toJson() => _$SubmitFeedbackRequestToJson(this);
}

/// A single media reference echoed back in `attachment_keys`. Mirrors
/// the inline object in the golden schema (all four fields required).
@JsonSerializable(createFactory: false)
class SubmitFeedbackAttachmentRequest {
  const SubmitFeedbackAttachmentRequest({
    required this.kind,
    required this.objectKey,
    required this.contentType,
    required this.sizeBytes,
  });

  final String kind;

  @JsonKey(name: 'object_key')
  final String objectKey;

  @JsonKey(name: 'content_type')
  final String contentType;

  @JsonKey(name: 'size_bytes')
  final int sizeBytes;

  factory SubmitFeedbackAttachmentRequest.fromDomain(
    FeedbackAttachment attachment,
  ) =>
      SubmitFeedbackAttachmentRequest(
        kind: attachment.kind.wireValue,
        objectKey: attachment.objectKey,
        contentType: attachment.contentType,
        sizeBytes: attachment.sizeBytes,
      );

  Map<String, dynamic> toJson() =>
      _$SubmitFeedbackAttachmentRequestToJson(this);
}

/// Structured client context. All six fields are required at the wire
/// level when `context` is present (matches the golden oneOf object
/// branch); empty strings are sent for fields a native client cannot
/// supply (viewport / user_agent / app_version).
@JsonSerializable(createFactory: false)
class FeedbackContextRequest {
  const FeedbackContextRequest({
    required this.role,
    required this.locale,
    required this.platform,
    required this.appVersion,
    required this.viewport,
    required this.userAgent,
  });

  final String role;
  final String locale;
  final String platform;

  @JsonKey(name: 'app_version')
  final String appVersion;

  final String viewport;

  @JsonKey(name: 'user_agent')
  final String userAgent;

  factory FeedbackContextRequest.fromDomain(FeedbackContextInput input) =>
      FeedbackContextRequest(
        role: input.role,
        locale: input.locale,
        platform: input.platform,
        appVersion: input.appVersion,
        viewport: input.viewport,
        userAgent: input.userAgent,
      );

  Map<String, dynamic> toJson() => _$FeedbackContextRequestToJson(this);
}
