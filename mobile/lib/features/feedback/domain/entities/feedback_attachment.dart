import 'package:freezed_annotation/freezed_annotation.dart';

import 'feedback_attachment_kind.dart';

part 'feedback_attachment.freezed.dart';

/// A media reference that has already been uploaded to storage and is
/// ready to be echoed back in the submit payload's `attachment_keys`.
///
/// The [objectKey] is the server-minted key returned by the presign
/// endpoint — the client never invents it. [sizeBytes] and
/// [contentType] are carried verbatim so the backend can re-validate
/// the allowlist + size cap at submit time.
@freezed
class FeedbackAttachment with _$FeedbackAttachment {
  const factory FeedbackAttachment({
    required FeedbackAttachmentKind kind,
    required String objectKey,
    required String contentType,
    required int sizeBytes,
  }) = _FeedbackAttachment;
}
