import 'package:freezed_annotation/freezed_annotation.dart';

import 'feedback_attachment_kind.dart';

part 'presigned_feedback_upload.freezed.dart';

/// The presign envelope returned by `POST
/// /api/v1/feedback/attachments/presign` (auth required): a short-lived
/// PUT [uploadUrl] the client streams the bytes to, plus the
/// server-minted [objectKey] that must be echoed back in the submit
/// payload.
@freezed
class PresignedFeedbackUpload with _$PresignedFeedbackUpload {
  const factory PresignedFeedbackUpload({
    required String uploadUrl,
    required String objectKey,
    required FeedbackAttachmentKind kind,
  }) = _PresignedFeedbackUpload;
}
