/// The two kinds of report a user can file from the in-app "Signaler"
/// surface. Mirrors the backend `type` enum (`bug` | `security`) — the
/// wire value is the lowercase enum name, sent verbatim in the submit
/// payload.
enum FeedbackType {
  /// Something is broken or behaving unexpectedly.
  bug,

  /// A potential security flaw or vulnerability.
  security;

  /// The exact lowercase token the API expects in `type`.
  String get wireValue => name;
}
