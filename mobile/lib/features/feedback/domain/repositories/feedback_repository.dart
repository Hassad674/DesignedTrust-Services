import 'dart:io';

import '../entities/feedback_attachment.dart';
import '../entities/feedback_attachment_kind.dart';
import '../entities/feedback_submission.dart';
import '../entities/submit_feedback_input.dart';

/// Persistence contract for the feedback (bug & vulnerability reporting)
/// feature. Two operations back the in-app "Signaler" surface:
///
///  * [submit] — POST the report (anonymous allowed).
///  * [uploadAttachment] — presign + PUT a single media file (auth
///    required) and return the [FeedbackAttachment] ready to attach to
///    a subsequent [submit] call.
///
/// Implementations live in `data/` and depend on `ApiClient`; the
/// presentation layer talks only to this interface.
abstract class FeedbackRepository {
  /// Files a bug or security report. Returns the server acknowledgement
  /// (id / type / status / created_at) on success.
  Future<FeedbackSubmission> submit(SubmitFeedbackInput input);

  /// Presigns then uploads [file] as a feedback attachment of [kind].
  ///
  /// AUTH REQUIRED — the presign endpoint rejects anonymous callers with
  /// 401. The presentation layer must therefore only invoke this for a
  /// logged-in reporter. Returns the uploaded [FeedbackAttachment] (with
  /// the server-minted object key) to include in the submit payload.
  Future<FeedbackAttachment> uploadAttachment({
    required File file,
    required FeedbackAttachmentKind kind,
    required String contentType,
    required String filename,
  });
}
