import 'feedback_attachment.dart';
import 'feedback_type.dart';

/// The structured client context captured with a submission. Every
/// field is optional at the wire level; on mobile we always send
/// `platform = "android"` plus the active locale and current route as
/// [pageUrl]. [appVersion] is omitted (no package_info dependency) and
/// [viewport] / [userAgent] are not meaningful on a native client.
class FeedbackContextInput {
  const FeedbackContextInput({
    this.role = '',
    this.locale = '',
    this.platform = '',
    this.appVersion = '',
    this.viewport = '',
    this.userAgent = '',
  });

  final String role;
  final String locale;
  final String platform;
  final String appVersion;
  final String viewport;
  final String userAgent;
}

/// Everything needed to file a report via `POST /api/v1/feedback`.
///
/// [attachments] are honoured only for an authenticated reporter; the
/// presentation layer keeps the list empty for anonymous users (the
/// attach control is disabled), and the backend rejects a non-empty
/// list from an anonymous caller regardless.
///
/// [honeypot] is the hidden anti-bot field — always the empty string
/// from a real client; a non-empty value makes the server silently drop
/// the submission.
class SubmitFeedbackInput {
  const SubmitFeedbackInput({
    required this.type,
    required this.title,
    required this.description,
    this.pageUrl = '',
    this.context,
    this.reporterEmail = '',
    this.attachments = const [],
    this.honeypot = '',
  });

  final FeedbackType type;
  final String title;
  final String description;
  final String pageUrl;
  final FeedbackContextInput? context;
  final String reporterEmail;
  final List<FeedbackAttachment> attachments;
  final String honeypot;
}
