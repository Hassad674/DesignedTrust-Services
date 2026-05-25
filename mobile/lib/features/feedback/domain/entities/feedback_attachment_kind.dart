/// Media kind for a feedback attachment. Mirrors the backend
/// `kind` enum (`image` | `video`) used by both the presign request
/// and the submit `attachment_keys[].kind` field.
enum FeedbackAttachmentKind {
  image,
  video;

  /// The exact lowercase token the API expects in `kind`.
  String get wireValue => name;
}
